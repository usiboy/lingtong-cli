// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type uploadTransport func(*http.Request) (*http.Response, error)

func (fn uploadTransport) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func TestUploadTableFileRefreshesSTSAndPreservesBytes(t *testing.T) {
	content := []byte("商品,金额\n苹果,12.5\n")
	path := filepath.Join(t.TempDir(), "订单 & #+%.CSV")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	stsCalls, uploads := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer platform-token" {
			t.Error("missing platform auth")
		}
		if r.URL.Path != "/gw/ai/proxy" || r.URL.Query().Get("_t") == "" {
			t.Error("invalid STS route")
		}
		var p map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Error(err)
		}
		if p["path"] != "/storage/oss/getStsToken" || p["method"] != "GET" {
			t.Errorf("invalid STS proxy request: %v", p)
		}
		stsCalls++
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"Credentials": map[string]string{
			"AccessKeyId": "temporary-id", "AccessKeySecret": "temporary-secret", "SecurityToken": fmt.Sprintf("sts-%d", stsCalls),
		}})
	}))
	defer server.Close()
	var firstObject string
	uploadClient := &http.Client{Transport: uploadTransport(func(r *http.Request) (*http.Response, error) {
		uploads++
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != string(content) || r.ContentLength != int64(len(content)) {
			t.Errorf("changed upload bytes: %q", data)
		}
		if r.Method != "PUT" || r.URL.Host != "erp-tmp-zb.oss-cn-zhangjiakou.aliyuncs.com" {
			t.Errorf("wrong upload destination: %s", r.URL)
		}
		if !strings.HasSuffix(r.URL.Path, "/订单 & #+%.CSV") || r.URL.RawQuery != "" || r.URL.Fragment != "" {
			t.Errorf("filename encoding: %s", r.URL)
		}
		if !strings.HasPrefix(r.URL.Path, "/report_temporary/") {
			t.Error("incorrect object prefix")
		}
		if uploads == 1 {
			firstObject = r.URL.String()
		} else if firstObject != r.URL.String() {
			t.Error("retry changed object key")
		}
		if r.Header.Get("x-oss-security-token") != fmt.Sprintf("sts-%d", uploads) {
			t.Error("STS was not refreshed")
		}
		if strings.Contains(r.Header.Get("Authorization"), "platform-token") {
			t.Error("platform token leaked to OSS")
		}
		digest := md5.Sum(content)
		if r.Header.Get("Content-MD5") != base64.StdEncoding.EncodeToString(digest[:]) {
			t.Error("invalid checksum")
		}
		if r.Header.Get("Content-Type") != "text/csv" {
			t.Error("invalid MIME type")
		}
		stringToSign := strings.Join([]string{"PUT", r.Header.Get("Content-MD5"), "text/csv", r.Header.Get("Date"), "x-oss-security-token:" + r.Header.Get("x-oss-security-token"), "/erp-tmp-zb" + r.URL.Path}, "\n")
		mac := hmac.New(sha1.New, []byte("temporary-secret"))
		_, _ = mac.Write([]byte(stringToSign))
		if r.Header.Get("Authorization") != "OSS temporary-id:"+base64.StdEncoding.EncodeToString(mac.Sum(nil)) {
			t.Error("invalid OSS signature")
		}
		status := 200
		if uploads == 1 {
			status = 403
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	got, err := NewClient(server.URL, "platform-token").uploadTableFile(context.Background(), file, uploadClient)
	if err != nil {
		t.Fatal(err)
	}
	if got != firstObject || stsCalls != 2 || uploads != 2 {
		t.Fatalf("result=%s sts=%d uploads=%d", got, stsCalls, uploads)
	}
}

func TestParseSTSCredentialsVariantsAndRejections(t *testing.T) {
	now := time.Now()
	valid := `{"accessKeyId":"id","accessKeySecret":"secret","securityToken":"token"}`
	quoted, _ := json.Marshal(valid)
	for _, raw := range []string{valid, string(quoted), `{"result":{"credentials":` + valid + `}}`, `{"data":{"Credentials":` + valid + `}}`, `{"AccessKeyId":"id","AccessKeySecret":"secret","SecurityToken":"token"}`} {
		credentials, err := parseSTSCredentials([]byte(raw), now)
		if err != nil || credentials.AccessKeyID != "id" {
			t.Fatalf("parse variant: %v", err)
		}
	}
	for _, raw := range []string{`null`, `{}`, `"not json"`, `{"success":false,"result":` + valid + `}`, `{"accessKeyId":"secret"}`, strings.TrimSuffix(valid, "}") + `,"expiration":"yesterday"}`, strings.TrimSuffix(valid, "}") + `,"Expiration":"2000-01-01T00:00:00Z"}`} {
		_, err := parseSTSCredentials([]byte(raw), now)
		if err == nil {
			t.Errorf("accepted invalid credentials: %s", raw)
		} else if strings.Contains(err.Error(), "secret") {
			t.Error("error leaked credentials")
		}
	}
}

func TestDoQueryContextPreservesPostParamsAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fileUrl") != "https://example.com/订单 +&.xlsx?x=1&y=2" {
			t.Errorf("query corrupted: %s", r.URL.RawQuery)
		}
		var p map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&p)
		if p["params"] != nil || p["body"] != nil || p["method"] != "POST" || p["path"] != "/basicdata/record/import" {
			t.Errorf("wrong proxy payload: %v", p)
		}
		w.Write([]byte(`{"success":true,"result":"uuid"}`))
	}))
	defer server.Close()
	c := NewClient(server.URL, "token")
	_, err := c.DoQueryContext(context.Background(), "POST", "/basicdata/record/import", map[string]interface{}{"fileUrl": "https://example.com/订单 +&.xlsx?x=1&y=2"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.DoQueryContext(ctx, "POST", "/basicdata/record/import", nil)
	if err == nil {
		t.Fatal("cancelled context sent a request")
	}
}

func TestUploadFailureDoesNotExposeResponse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.xlsx")
	_ = os.WriteFile(path, []byte("test"), 0600)
	f, _ := os.Open(path)
	defer f.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"AccessKeyId":"id","AccessKeySecret":"secret","SecurityToken":"token"}`))
	}))
	defer server.Close()
	uploadClient := &http.Client{Transport: uploadTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("sensitive signature token secret"))}, nil
	})}
	_, err := NewClient(server.URL, "token").uploadTableFile(context.Background(), f, uploadClient)
	if err == nil || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "signature") {
		t.Fatalf("unexpected error: %v", err)
	}
}
