// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
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
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
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
