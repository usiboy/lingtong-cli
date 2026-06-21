// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"syscall"
	"time"

	lterrors "github.com/lingtong/cli/internal/errors"
)

// Client is the HTTP client for Lingtong API.
type Client struct {
	host       string
	token      string
	httpClient *http.Client
	// Proxy mode: all requests go through /gw/ai/proxy
	useProxy bool
}

// NewClient creates a new API client.
func NewClient(host, token string) *Client {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	return &Client{
		host:       strings.TrimRight(host, "/"),
		token:      token,
		httpClient: httpClient,
		useProxy:   true, // Default to proxy mode for security
	}
}

// NewClientWithTimeout creates a new API client with a custom timeout.
func NewClientWithTimeout(host, token string, timeout time.Duration) *Client {
	c := NewClient(host, token)
	c.httpClient.Timeout = timeout
	return c
}

// NewClientWithProxy creates a new API client with proxy mode enabled.
func NewClientWithProxy(host, token string) *Client {
	return NewClient(host, token)
}

// DisableProxy disables proxy mode (for testing only).
func (c *Client) DisableProxy() {
	c.useProxy = false
}

// Do performs an HTTP request and returns the response.
func (c *Client) Do(method, path string, params map[string]interface{}, body interface{}) ([]byte, error) {
	var url string
	var reqBody io.Reader

	if c.useProxy {
		// Proxy mode: POST to /gw/ai/proxy with path and body in JSON
		url = joinURL(c.host, "/gw/ai/proxy")
		proxyBody := map[string]interface{}{
			"path":   path,
			"method": method,
		}
		if params != nil && len(params) > 0 {
			if method == http.MethodGet {
				url = appendQueryParams(url, params)
			} else {
				proxyBody["params"] = params
			}
		}
		if body != nil {
			proxyBody["body"] = body
		}
		b, err := json.Marshal(proxyBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal proxy request: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
		method = http.MethodPost
	} else {
		// Direct mode (legacy)
		url = joinURL(c.host, path)
		if params != nil && len(params) > 0 {
			url = appendQueryParams(url, params)
		}
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(b)
		}
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Authentication via Bearer token
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, classifyNetworkError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, classifyHTTPError(resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// classifyNetworkError converts a transport-level error into a typed
// NetworkError so the CLI can return the correct exit code (4) and a
// machine-readable error type. The message preserves the "request failed"
// prefix for backward compatibility with existing callers and tests.
func classifyNetworkError(err error) error {
	subtype := lterrors.SubtypeUnknown

	var dnsErr *net.DNSError
	switch {
	case stderrors.As(err, &dnsErr):
		subtype = lterrors.SubtypeNetworkDNS
	case stderrors.Is(err, syscall.ECONNREFUSED):
		subtype = lterrors.SubtypeNetworkConnRef
	default:
		var netErr net.Error
		if stderrors.As(err, &netErr) && netErr.Timeout() {
			subtype = lterrors.SubtypeNetworkTimeout
		}
	}

	ne := lterrors.NewNetworkError(subtype, "request failed: %v", err).
		WithCause(err).
		WithHint("Check your network connection and the configured --host")
	if subtype == lterrors.SubtypeNetworkTimeout {
		ne.Retryable = true
	}
	return ne
}

// classifyHTTPError converts a non-2xx HTTP response into a typed error.
// Authentication failures (401/403) map to AuthenticationError (exit 3);
// everything else maps to APIError (exit 1). The message preserves the
// "API error (<status>)" prefix expected by existing callers and tests.
func classifyHTTPError(status int, body string) error {
	msg := fmt.Sprintf("API error (%d): %s", status, body)

	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return lterrors.NewAuthenticationError(lterrors.SubtypeTokenInvalid, "%s", msg).
			WithHint("Run 'lingtong-cli auth login' to refresh your token")
	case status == http.StatusNotFound:
		return lterrors.NewAPIError(lterrors.SubtypeNotFound, "%s", msg)
	case status == http.StatusConflict:
		return lterrors.NewAPIError(lterrors.SubtypeConflict, "%s", msg)
	case status == http.StatusTooManyRequests:
		e := lterrors.NewAPIError(lterrors.SubtypeRateLimit, "%s", msg)
		e.Retryable = true
		return e
	case status >= 500:
		e := lterrors.NewAPIError(lterrors.SubtypeServerError, "%s", msg)
		e.Retryable = true
		return e
	default:
		return lterrors.NewAPIError(lterrors.SubtypeUnknown, "%s", msg)
	}
}

func joinURL(host, path string) string {
	if host == "" {
		return path
	}
	return strings.TrimRight(host, "/") + "/" + strings.TrimLeft(path, "/")
}

func appendQueryParams(rawURL string, params map[string]interface{}) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	query := parsed.Query()
	for key, value := range params {
		switch typed := value.(type) {
		case nil:
			continue
		case []string:
			for _, item := range typed {
				query.Add(key, item)
			}
		case []interface{}:
			for _, item := range typed {
				query.Add(key, fmt.Sprint(item))
			}
		default:
			query.Set(key, fmt.Sprint(typed))
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// Get performs a GET request.
func (c *Client) Get(path string, params map[string]interface{}) ([]byte, error) {
	return c.Do(http.MethodGet, path, params, nil)
}

// Post performs a POST request.
func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	return c.Do(http.MethodPost, path, nil, body)
}

// Put performs a PUT request.
func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	return c.Do(http.MethodPut, path, nil, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.Do(http.MethodDelete, path, nil, nil)
}
