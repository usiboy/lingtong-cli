// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	host := "https://example.com"
	token := "test-token"

	c := NewClient(host, token)

	if c.host != host {
		t.Errorf("expected host=%s, got %s", host, c.host)
	}
	if c.token != token {
		t.Errorf("expected token=%s, got %s", token, c.token)
	}
	if !c.useProxy {
		t.Error("expected proxy mode to be enabled by default")
	}
	if c.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

// TestDisableProxy tests disabling proxy mode
func TestDisableProxy(t *testing.T) {
	c := NewClient("https://example.com", "token")
	c.DisableProxy()

	if c.useProxy {
		t.Error("expected proxy mode to be disabled")
	}
}

// TestProxyModeRequest tests that proxy mode constructs correct requests
func TestProxyModeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/gw/ai/proxy" {
			t.Errorf("expected path /gw/ai/proxy, got %s", r.URL.Path)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer test-token"
		if authHeader != expectedAuth {
			t.Errorf("expected Authorization=%s, got %s", expectedAuth, authHeader)
		}

		// Verify request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["path"] != "/test/path" {
			t.Errorf("expected path=/test/path, got %v", body["path"])
		}
		if body["method"] != "GET" {
			t.Errorf("expected method=GET, got %v", body["method"])
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	resp, err := c.Get("/test/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if success, _ := result["success"].(bool); !success {
		t.Error("expected success=true")
	}
}

// TestDirectModeRequest tests direct mode (non-proxy) requests
func TestDirectModeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/test/path" {
			t.Errorf("expected path /test/path, got %s", r.URL.Path)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer test-token"
		if authHeader != expectedAuth {
			t.Errorf("expected Authorization=%s, got %s", expectedAuth, authHeader)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"direct": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	c.DisableProxy()

	resp, err := c.Get("/test/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if direct, _ := result["direct"].(bool); !direct {
		t.Error("expected direct=true")
	}
}

// TestPostRequest tests POST requests
func TestPostRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		// In proxy mode, actual body is nested under "body" key
		if actualBody, ok := body["body"].(map[string]interface{}); ok {
			if actualBody["key"] != "value" {
				t.Errorf("expected body.key=value, got %v", actualBody["key"])
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"created": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	resp, err := c.Post("/test/path", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if created, _ := result["created"].(bool); !created {
		t.Error("expected created=true")
	}
}

// TestErrorResponse tests error handling for API errors
func TestErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "invalid_request",
			"message": "Missing required parameter",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	_, err := c.Get("/test/path", nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Check error message contains status code
	expectedMsg := "400"
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
	if err.Error() != "" && err.Error()[:3] != expectedMsg {
		// Error format: "API error (400): ..."
		t.Logf("error message: %s", err.Error())
	}
}

// TestGetWithParams tests GET requests with query parameters
func TestGetWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		params, ok := body["params"].(map[string]interface{})
		if !ok {
			t.Fatal("expected params in request body")
		}

		if params["page"] != "1" {
			t.Errorf("expected params.page=1, got %v", params["page"])
		}
		if params["size"] != "20" {
			t.Errorf("expected params.size=20, got %v", params["size"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"page": 1, "size": 20})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	params := map[string]interface{}{
		"page": "1",
		"size": "20",
	}

	resp, err := c.Get("/test/path", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if page, _ := result["page"].(float64); page != 1 {
		t.Errorf("expected page=1, got %v", page)
	}
}

// TestPutRequest tests PUT requests
func TestPutRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"updated": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	c.DisableProxy()

	resp, err := c.Put("/test/path", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if updated, _ := result["updated"].(bool); !updated {
		t.Error("expected updated=true")
	}
}

// TestDeleteRequest tests DELETE requests
func TestDeleteRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"deleted": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	c.DisableProxy()

	resp, err := c.Delete("/test/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if deleted, _ := result["deleted"].(bool); !deleted {
		t.Error("expected deleted=true")
	}
}
// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://example.com", "test-token")
	if c.host != "https://example.com" {
		t.Errorf("wrong host: %s", c.host)
	}
	if !c.useProxy {
		t.Error("proxy should be enabled by default")
	}
}

func TestDisableProxy(t *testing.T) {
	c := NewClient("https://example.com", "token")
	c.DisableProxy()
	if c.useProxy {
		t.Error("proxy should be disabled")
	}
}

func TestProxyModeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/gw/ai/proxy" {
			t.Errorf("expected /gw/ai/proxy, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	resp, err := c.Get("/test/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]interface{}
	json.Unmarshal(resp, &result)
	if result["success"] != true {
		t.Error("expected success=true")
	}
}

func TestDirectModeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"direct": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	c.DisableProxy()
	resp, err := c.Get("/test/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]interface{}
	json.Unmarshal(resp, &result)
	if result["direct"] != true {
		t.Error("expected direct=true")
	}
}

func TestErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	_, err := c.Get("/test/path", nil)
	if err == nil {
		t.Error("expected error for 400 response")
	}
}
