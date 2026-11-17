// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	lterrors "github.com/lingtong/cli/internal/errors"
)

// UploadTableFile uses the same temporary bucket and object prefix as
// lingtong-app-view/src/utils/uploadOss.js. Platform tokens never go to OSS.
func (c *Client) UploadTableFile(ctx context.Context, file *os.File) (string, error) {
	uploadClient := &http.Client{
		Timeout:       2 * time.Minute,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return c.uploadTableFile(ctx, file, uploadClient)
}

func (c *Client) uploadTableFile(ctx context.Context, file *os.File, uploadClient *http.Client) (string, error) {
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 50*1024*1024 {
		return "", lterrors.NewValidationError(lterrors.SubtypeInvalidArgument, "file must be a nonempty regular file of at most 50 MB")
	}
	contentTypes := map[string]string{
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xls":  "application/vnd.ms-excel", ".csv": "text/csv",
	}
	contentType, ok := contentTypes[strings.ToLower(filepath.Ext(info.Name()))]
	if !ok {
		return "", lterrors.NewValidationError(lterrors.SubtypeInvalidArgument, "file must have .xlsx, .xls or .csv extension")
	}
	// Content-MD5 detects changes between reading the file and uploading it.
	digest := md5.New()
	if _, err := io.Copy(digest, io.NewSectionReader(file, 0, info.Size())); err != nil {
		return "", fmt.Errorf("read import file: %w", err)
	}
	contentMD5 := base64.StdEncoding.EncodeToString(digest.Sum(nil))
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	const bucket = "erp-tmp-zb"
	objectName := fmt.Sprintf("report_temporary/%d-%s/%s", time.Now().UnixMilli(), hex.EncodeToString(nonce[:]), info.Name())
	objectURL := (&url.URL{Scheme: "https", Host: bucket + ".oss-cn-zhangjiakou.aliyuncs.com", Path: "/" + objectName}).String()
	for attempt := 0; attempt < 3; attempt++ {
		raw, err := c.DoContext(ctx, http.MethodGet, "/storage/oss/getStsToken", map[string]interface{}{"_t": time.Now().UnixMilli()}, nil)
		if err != nil {
			return "", fmt.Errorf("get upload credentials: %w", err)
		}
		credentials, err := parseSTSCredentials(raw, time.Now())
		if err != nil {
			return "", err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL, io.NewSectionReader(file, 0, info.Size()))
		if err != nil {
			return "", err
		}
		req.ContentLength = info.Size()
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Content-MD5", contentMD5)
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
		req.Header.Set("x-oss-security-token", credentials.SecurityToken)
		// OSS signature V1 matches the frontend ali-oss client. See
		// https://www.alibabacloud.com/help/en/oss/include-signatures-in-the-authorization-header
		toSign := "PUT\n" + contentMD5 + "\n" + contentType + "\n" + req.Header.Get("Date") +
			"\nx-oss-security-token:" + credentials.SecurityToken + "\n/" + bucket + "/" + objectName
		mac := hmac.New(sha1.New, []byte(credentials.AccessKeySecret))
		_, _ = mac.Write([]byte(toSign))
		req.Header.Set("Authorization", "OSS "+credentials.AccessKeyID+":"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
		resp, err := uploadClient.Do(req)
		if err != nil {
			return "", classifyNetworkError(err)
		}
		// Do not echo OSS bodies: signature errors can contain credential details.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return objectURL, nil
		}
		if resp.StatusCode == http.StatusForbidden && attempt < 2 {
			continue // Refresh STS and retry the same object, never the import task.
		}
		return "", lterrors.NewAPIError(lterrors.SubtypeServerError, "OSS upload failed (HTTP %d)", resp.StatusCode)
	}
	return "", lterrors.NewAPIError(lterrors.SubtypeServerError, "OSS upload failed")
}

type stsCredentials struct {
	AccessKeyID, AccessKeySecret, SecurityToken string
}

func parseSTSCredentials(raw []byte, now time.Time) (stsCredentials, error) {
	invalid := func(message string) (stsCredentials, error) {
		return stsCredentials{}, lterrors.NewAPIError(lterrors.SubtypeServerError, "upload credentials: %s", message)
	}
	// Hosts return either an object or a JSON string, optionally wrapped in
	// result/data and credentials/Credentials (same variants as the frontend).
	var data interface{}
	if json.Unmarshal(raw, &data) != nil {
		return invalid("invalid JSON response")
	}
	for depth := 0; depth < 8; depth++ {
		if s, ok := data.(string); ok {
			if json.Unmarshal([]byte(s), &data) != nil {
				return invalid("invalid JSON string")
			}
			continue
		}
		m, ok := data.(map[string]interface{})
		if !ok {
			return invalid("missing credentials")
		}
		if success, ok := m["success"].(bool); ok && !success {
			return invalid("server rejected STS request")
		}
		wrapped := false
		for _, key := range []string{"result", "data", "credentials", "Credentials"} {
			if nested, ok := m[key]; ok && nested != nil {
				data, wrapped = nested, true
				break
			}
		}
		if wrapped {
			continue
		}
		get := func(keys ...string) string {
			for _, key := range keys {
				if s, ok := m[key].(string); ok && s != "" {
					return s
				}
			}
			return ""
		}
		credentials := stsCredentials{get("accessKeyId", "AccessKeyId"), get("accessKeySecret", "AccessKeySecret"), get("securityToken", "SecurityToken", "stsToken")}
		if credentials.AccessKeyID == "" || credentials.AccessKeySecret == "" || credentials.SecurityToken == "" {
			return invalid("incomplete credentials")
		}
		if expiration := get("expiration", "Expiration"); expiration != "" {
			t, err := time.Parse(time.RFC3339, expiration)
			if err != nil || !t.After(now.Add(time.Minute)) {
				return invalid("expired or invalid expiration")
			}
		}
		return credentials, nil
	}
	return invalid("unsupported response structure")
}
