// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		token   string
		checkFn func(*testing.T, *Client)
	}{
		{
			name:  "valid client with proxy enabled by default",
			host:  "https://example.com",
			token: "apk-test123",
			checkFn: func(t *testing.T, c *Client) {
				if c.host != "https://example.com" {
					t.Errorf("expected host https://example.com, got %s", c.host)
				}
				if c.token != "apk-test123" {
					t.Errorf("expected token apk-test123, got %s", c.token)
				}
				if !c.useProxy {
					t.Error("expected proxy to be enabled by default")
				}
				if c.httpClient == nil {
					t.Error("expected non-nil httpClient")
				}
			},
		},
		{
			name:  "client trims trailing host slash",
			host:  "https://example.com/",
			token: "apk-test123",
			checkFn: func(t *testing.T, c *Client) {
				if c.host != "https://example.com" {
					t.Errorf("expected trimmed host https://example.com, got %s", c.host)
				}
			},
		},
		{
			name:  "client with empty host",
			host:  "",
			token: "apk-test123",
			checkFn: func(t *testing.T, c *Client) {
				if c.host != "" {
					t.Errorf("expected empty host, got %s", c.host)
				}
			},
		},
		{
			name:  "client with empty token",
			host:  "https://example.com",
			token: "",
			checkFn: func(t *testing.T, c *Client) {
				if c.token != "" {
					t.Errorf("expected empty token, got %s", c.token)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.host, tt.token)
			if client == nil {
				t.Fatal("expected non-nil client")
			}
			tt.checkFn(t, client)
		})
	}
}

// TestNewClientWithProxy tests the proxy client factory
func TestNewClientWithProxy(t *testing.T) {
	client := NewClientWithProxy("https://example.com", "apk-test123")

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if !client.useProxy {
		t.Error("expected proxy to be enabled")
	}
	if client.host != "https://example.com" {
		t.Errorf("expected host https://example.com, got %s", client.host)
	}
}

// TestDisableProxy tests proxy mode toggle
func TestDisableProxy(t *testing.T) {
	client := NewClient("https://example.com", "apk-test123")

	if !client.useProxy {
		t.Fatal("expected proxy to be enabled by default")
	}

	client.DisableProxy()

	if client.useProxy {
		t.Error("expected proxy to be disabled after DisableProxy()")
	}
}

// TestDoProxyMode tests Do method in proxy mode
func TestDoProxyMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/gw/ai/proxy" {
			t.Errorf("expected /gw/ai/proxy path, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer apk-test123" {
			t.Errorf("expected Bearer token, got %s", r.Header.Get("Authorization"))
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode proxy request: %v", err)
		}
		if payload["path"] != "/test/api" {
			t.Errorf("expected proxy path /test/api, got %v", payload["path"])
		}
		if payload["method"] != "GET" {
			t.Errorf("expected proxy method GET, got %v", payload["method"])
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	resp, err := client.Do("GET", "/test/api", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"status":"ok"}` {
		t.Errorf("expected response {\"status\":\"ok\"}, got %s", string(resp))
	}
}

// TestDoDirectModeWithParams tests query param handling outside proxy mode.
func TestDoDirectModeWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page query param 2, got %s", r.URL.Query().Get("page"))
		}
		if values := r.URL.Query()["tag"]; len(values) != 2 || values[0] != "a" || values[1] != "b" {
			t.Errorf("expected tag query params [a b], got %v", values)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", "apk-test123")
	client.DisableProxy()

	params := map[string]interface{}{
		"page": 2,
		"tag":  []string{"a", "b"},
	}
	resp, err := client.Do("GET", "/test/api", params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"status":"ok"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestDoDirectMode tests Do method in direct mode
func TestDoDirectMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/test/api" {
			t.Errorf("expected /test/api path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	client.DisableProxy()

	resp, err := client.Do("GET", "/test/api", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"status":"ok"}` {
		t.Errorf("expected response {\"status\":\"ok\"}, got %s", string(resp))
	}
}

// TestDoWithBody tests Do method with request body
func TestDoWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"created":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	client.DisableProxy()

	body := map[string]interface{}{"name": "test", "value": 123}
	resp, err := client.Do("POST", "/test/api", nil, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"created":true}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestDoWithParams tests Do method with query params in proxy mode
func TestDoWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	params := map[string]interface{}{"page": 1, "size": 20}
	resp, err := client.Do("GET", "/test/api", params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"data":[]}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestDoServerError tests Do method with server error
func TestDoServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	_, err := client.Do("GET", "/test/api", nil, nil)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain '500', got: %v", err)
	}
}

// TestDoBadRequest tests Do method with bad request
func TestDoBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	_, err := client.Do("GET", "/test/api", nil, nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain '400', got: %v", err)
	}
}

// TestDoUnauthorized tests Do method with unauthorized error
func TestDoUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "invalid-token")
	_, err := client.Do("GET", "/test/api", nil, nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to contain '401', got: %v", err)
	}
}

// TestDoNotFound tests Do method with not found error
func TestDoNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	_, err := client.Do("GET", "/nonexistent", nil, nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected error to contain '404', got: %v", err)
	}
}

// TestDoNetworkError tests Do method with network error
func TestDoNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	serverURL := server.URL
	server.Close()

	client := NewClient(serverURL, "")
	_, err := client.Do("GET", "/test/api", nil, nil)
	if err == nil {
		t.Fatal("expected network error")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("expected error to contain 'request failed', got: %v", err)
	}
}

// TestGet tests the Get convenience method
func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"test"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	params := map[string]interface{}{"id": 1}
	resp, err := client.Get("/test/api", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"id":1,"name":"test"}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestGetWithoutParams tests Get method without params
func TestGetWithoutParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	resp, err := client.Get("/test/api", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"items":[]}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestPost tests the Post convenience method
func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":123,"created":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	body := map[string]interface{}{"name": "new item"}
	resp, err := client.Post("/test/api", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"id":123,"created":true}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestPut tests the Put convenience method
func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":123,"updated":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	body := map[string]interface{}{"id": 123, "name": "updated item"}
	resp, err := client.Put("/test/api", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"id":123,"updated":true}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestDelete tests the Delete convenience method
func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	resp, err := client.Delete("/test/api/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"deleted":true}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestClientWithoutToken tests client without authentication token
func TestClientWithoutToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			t.Errorf("expected no Authorization header, got %s", authHeader)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"public":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	resp, err := client.Get("/public/api", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"public":true}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestDoWithEmptyBody tests Do method with nil body in proxy mode
func TestDoWithEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	resp, err := client.Do("DELETE", "/test/api/123", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"success":true}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// TestDoWithComplexParams tests Do method with complex nested params
func TestDoWithComplexParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "apk-test123")
	params := map[string]interface{}{
		"filter": map[string]interface{}{
			"name":  "test",
			"value": 123,
		},
		"sort": []string{"name", "created"},
	}
	resp, err := client.Do("GET", "/test/api", params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"results":[]}` {
		t.Errorf("unexpected response: %s", string(resp))
	}
}
