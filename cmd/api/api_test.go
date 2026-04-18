// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// TestParseData tests the parseData function with various inputs
func TestParseData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantType string // "map", "string", or "nil"
	}{
		{
			name:     "empty string",
			input:    "",
			wantNil:  true,
			wantType: "nil",
		},
		{
			name:     "valid JSON object",
			input:    `{"name":"test","value":123}`,
			wantNil:  false,
			wantType: "map",
		},
		{
			name:     "valid JSON with nested object",
			input:    `{"user":{"name":"test"},"active":true}`,
			wantNil:  false,
			wantType: "map",
		},
		{
			name:     "valid JSON with array",
			input:    `{"items":[1,2,3]}`,
			wantNil:  false,
			wantType: "map",
		},
		{
			name:     "invalid JSON returns as string",
			input:    `{invalid json}`,
			wantNil:  false,
			wantType: "string",
		},
		{
			name:     "plain text returns as string",
			input:    "hello world",
			wantNil:  false,
			wantType: "string",
		},
		{
			name:     "JSON array returns as string (not a map)",
			input:    `[1,2,3]`,
			wantNil:  false,
			wantType: "string",
		},
		{
			name:     "whitespace only returns as string",
			input:    "   ",
			wantNil:  false,
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseData(tt.input)

			if tt.wantNil {
				if result != nil {
					t.Errorf("parseData(%q) = %v, want nil", tt.input, result)
				}
				return
			}

			switch tt.wantType {
			case "map":
				if _, ok := result.(map[string]interface{}); !ok {
					t.Errorf("parseData(%q) type = %T, want map[string]interface{}", tt.input, result)
				}
			case "string":
				if _, ok := result.(string); !ok {
					t.Errorf("parseData(%q) type = %T, want string", tt.input, result)
				}
			}
		})
	}
}

// TestParseParams tests the parseParams function with various inputs
func TestParseParams(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantLen int
	}{
		{
			name:    "empty string",
			input:   "",
			wantNil: true,
			wantLen: 0,
		},
		{
			name:    "valid JSON params",
			input:   `{"page":1,"limit":10}`,
			wantNil: false,
			wantLen: 2,
		},
		{
			name:    "valid JSON with string values",
			input:   `{"name":"test","status":"active"}`,
			wantNil: false,
			wantLen: 2,
		},
		{
			name:    "valid JSON with mixed types",
			input:   `{"page":1,"name":"test","active":true}`,
			wantNil: false,
			wantLen: 3,
		},
		{
			name:    "invalid JSON returns nil",
			input:   `{invalid}`,
			wantNil: true,
			wantLen: 0,
		},
		{
			name:    "JSON array returns nil",
			input:   `[1,2,3]`,
			wantNil: true,
			wantLen: 0,
		},
		{
			name:    "plain text returns nil",
			input:   "hello",
			wantNil: true,
			wantLen: 0,
		},
		{
			name:    "empty JSON object",
			input:   `{}`,
			wantNil: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseParams(tt.input)

			if tt.wantNil {
				if result != nil {
					t.Errorf("parseParams(%q) = %v, want nil", tt.input, result)
				}
				return
			}

			if result == nil {
				t.Errorf("parseParams(%q) = nil, want non-nil map", tt.input)
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("parseParams(%q) len = %d, want %d", tt.input, len(result), tt.wantLen)
			}
		})
	}
}

// TestNewCmdApi tests the api command creation
func TestNewCmdApi(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://localhost:8080",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)

	// Verify command properties
	if cmd.Use != "api <method> <path>" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "api <method> <path>")
	}

	if cmd.Short == "" {
		t.Error("cmd.Short should not be empty")
	}

	if cmd.Long == "" {
		t.Error("cmd.Long should not be empty")
	}

	// Verify flags exist
	dataFlag := cmd.Flags().Lookup("data")
	if dataFlag == nil {
		t.Error("missing --data flag")
	}

	paramsFlag := cmd.Flags().Lookup("params")
	if paramsFlag == nil {
		t.Error("missing --params flag")
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("missing --format flag")
	}

	// Verify default format value
	if formatFlag.DefValue != "json" {
		t.Errorf("format default = %q, want %q", formatFlag.DefValue, "json")
	}
}

// TestNewCmdApiArgsValidation tests argument validation
func TestNewCmdApiArgsValidation(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://localhost:8080",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one argument",
			args:    []string{"GET"},
			wantErr: true,
		},
		{
			name:    "two arguments (valid)",
			args:    []string{"GET", "/api/test"},
			wantErr: false,
		},
		{
			name:    "three arguments",
			args:    []string{"GET", "/api/test", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("cmd.Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

// TestApiCommandGET tests GET request execution
func TestApiCommandGET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request (proxy mode), got %s", r.Method)
		}

		// Verify authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want %q", authHeader, "Bearer test-token")
		}

		// Parse and verify proxy body
		var proxyBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&proxyBody); err != nil {
			t.Fatalf("failed to decode proxy body: %v", err)
		}

		if proxyBody["method"] != "GET" {
			t.Errorf("proxy method = %v, want GET", proxyBody["method"])
		}

		if proxyBody["path"] != "/api/connector/info" {
			t.Errorf("proxy path = %v, want /api/connector/info", proxyBody["path"])
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"id":      123,
				"name":    "test-connector",
				"status":  "active",
			},
		})
	}))
	defer server.Close()

	// Create factory with test server URL
	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/connector/info"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	// Verify output contains expected data
	outputStr := out.String()
	if !strings.Contains(outputStr, "test-connector") {
		t.Errorf("output should contain 'test-connector', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "123") {
		t.Errorf("output should contain '123', got: %s", outputStr)
	}
}

// TestApiCommandPOST tests POST request execution
func TestApiCommandPOST(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request is POST
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Parse and verify proxy body
		var proxyBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&proxyBody); err != nil {
			t.Fatalf("failed to decode proxy body: %v", err)
		}

		if proxyBody["method"] != "POST" {
			t.Errorf("proxy method = %v, want POST", proxyBody["method"])
		}

		// Verify request body
		body, ok := proxyBody["body"].(map[string]interface{})
		if !ok {
			t.Fatalf("proxy body is not a map, got %T", proxyBody["body"])
		}

		if body["name"] != "new-scene" {
			t.Errorf("body name = %v, want new-scene", body["name"])
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"id":   456,
				"name": "new-scene",
			},
		})
	}))
	defer server.Close()

	// Create factory with test server URL
	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"POST", "/api/scene/create", "--data", `{"name":"new-scene"}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	// Verify output
	outputStr := out.String()
	if !strings.Contains(outputStr, "new-scene") {
		t.Errorf("output should contain 'new-scene', got: %s", outputStr)
	}
}

// TestApiCommandPUT tests PUT request execution
func TestApiCommandPUT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request (proxy mode), got %s", r.Method)
		}

		var proxyBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&proxyBody); err != nil {
			t.Fatalf("failed to decode proxy body: %v", err)
		}

		if proxyBody["method"] != "PUT" {
			t.Errorf("proxy method = %v, want PUT", proxyBody["method"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "updated",
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"PUT", "/api/scene/123", "--data", `{"name":"updated-scene"}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "updated") {
		t.Errorf("output should contain 'updated', got: %s", outputStr)
	}
}

// TestApiCommandDELETE tests DELETE request execution
func TestApiCommandDELETE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request (proxy mode), got %s", r.Method)
		}

		var proxyBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&proxyBody); err != nil {
			t.Fatalf("failed to decode proxy body: %v", err)
		}

		if proxyBody["method"] != "DELETE" {
			t.Errorf("proxy method = %v, want DELETE", proxyBody["method"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "deleted",
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"DELETE", "/api/scene/123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "deleted") {
		t.Errorf("output should contain 'deleted', got: %s", outputStr)
	}
}

// TestApiCommandUnsupportedMethod tests unsupported HTTP method error
func TestApiCommandUnsupportedMethod(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://localhost:8080",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &errOut,
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"PATCH", "/api/test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported method, got nil")
	}

	expectedErr := "unsupported HTTP method: PATCH"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("error = %v, want to contain %q", err, expectedErr)
	}
}

// TestApiCommandCaseInsensitiveMethod tests that HTTP methods are case-insensitive
func TestApiCommandCaseInsensitiveMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proxyBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&proxyBody)

		// Method should be uppercased
		if proxyBody["method"] != "GET" {
			t.Errorf("proxy method = %v, want GET (uppercased)", proxyBody["method"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	// Test lowercase method
	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"get", "/api/test"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() with lowercase 'get' error = %v", err)
	}
}

// TestApiCommandWithParams tests GET request with query parameters
func TestApiCommandWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proxyBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&proxyBody); err != nil {
			t.Fatalf("failed to decode proxy body: %v", err)
		}

		// Verify params
		params, ok := proxyBody["params"].(map[string]interface{})
		if !ok {
			t.Fatalf("proxy params is not a map, got %T", proxyBody["params"])
		}

		if params["page"] != float64(1) {
			t.Errorf("params page = %v, want 1", params["page"])
		}

		if params["limit"] != float64(10) {
			t.Errorf("params limit = %v, want 10", params["limit"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"items": []interface{}{},
				"total": 0,
			},
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/items", "--params", `{"page":1,"limit":10}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}
}

// TestApiCommandWithPrettyFormat tests output with pretty format
func TestApiCommandWithPrettyFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": "success",
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/test", "--format", "pretty"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "success") {
		t.Errorf("output should contain 'success', got: %s", outputStr)
	}
}

// TestApiCommandWithTableFormat tests output with table format
func TestApiCommandWithTableFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": "success",
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/test", "--format", "table"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}

	// Table format falls back to JSON in current implementation
	outputStr := out.String()
	if outputStr == "" {
		t.Error("expected non-empty output for table format")
	}
}

// TestApiCommandServerError tests handling of server errors
func TestApiCommandServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	var out bytes.Buffer
	var errOut bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &errOut,
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/error"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code 500, got: %v", err)
	}
}

// TestApiCommandConnectionError tests handling of connection errors
func TestApiCommandConnectionError(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://localhost:1", // Invalid port
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &errOut,
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for connection failure, got nil")
	}
}

// TestApiCommandInvalidJSONData tests handling of invalid JSON data
func TestApiCommandInvalidJSONData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	// Invalid JSON should be sent as string
	cmd.SetArgs([]string{"POST", "/api/test", "--data", `{invalid json}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}
}

// TestApiCommandInvalidParamsJSON tests handling of invalid params JSON
func TestApiCommandInvalidParamsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proxyBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&proxyBody)

		// Invalid params should result in nil params
		if _, hasParams := proxyBody["params"]; hasParams {
			t.Error("expected no params for invalid JSON, but got some")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/test", "--params", `{invalid}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}
}

// TestApiCommandInvalidJSONResponse tests handling of non-JSON response
func TestApiCommandInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
	}
}

// TestApiCommandWithAuthHeader tests that authorization header is set correctly
func TestApiCommandWithAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer my-secret-token"
		if authHeader != expectedAuth {
			t.Errorf("Authorization = %q, want %q", authHeader, expectedAuth)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", contentType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"auth": "verified"})
	}))
	defer server.Close()

	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  server.URL,
			Token: "my-secret-token",
		},
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdApi(f)
	cmd.SetArgs([]string{"GET", "/api/verify"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cmd.Execute() error = %v", err)
	}
}

// TestApiCommandTableDriven tests various HTTP methods using table-driven tests
func TestApiCommandTableDriven(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		data         string
		params       string
		expectMethod string // The method that should be in the proxy body
		serverResp   map[string]interface{}
		wantErr      bool
	}{
		{
			name:         "simple GET",
			method:       "GET",
			path:         "/api/users",
			expectMethod: "GET",
			serverResp:   map[string]interface{}{"users": []interface{}{}},
			wantErr:      false,
		},
		{
			name:         "POST with data",
			method:       "POST",
			path:         "/api/users",
			data:         `{"name":"test"}`,
			expectMethod: "POST",
			serverResp:   map[string]interface{}{"id": 1},
			wantErr:      false,
		},
		{
			name:         "PUT with data",
			method:       "PUT",
			path:         "/api/users/1",
			data:         `{"name":"updated"}`,
			expectMethod: "PUT",
			serverResp:   map[string]interface{}{"updated": true},
			wantErr:      false,
		},
		{
			name:         "DELETE",
			method:       "DELETE",
			path:         "/api/users/1",
			expectMethod: "DELETE",
			serverResp:   map[string]interface{}{"deleted": true},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var proxyBody map[string]interface{}
				json.NewDecoder(r.Body).Decode(&proxyBody)

				if proxyBody["method"] != tt.expectMethod {
					t.Errorf("proxy method = %v, want %v", proxyBody["method"], tt.expectMethod)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResp)
			}))
			defer server.Close()

			var out bytes.Buffer
			f := &cmdutil.Factory{
				Config: &config.Config{
					Host:  server.URL,
					Token: "test-token",
				},
				IOStreams: &output.IOStreams{
					In:     os.Stdin,
					Out:    &out,
					ErrOut: &bytes.Buffer{},
				},
			}

			cmd := NewCmdApi(f)
			args := []string{tt.method, tt.path}
			if tt.data != "" {
				args = append(args, "--data", tt.data)
			}
			if tt.params != "" {
				args = append(args, "--params", tt.params)
			}
			cmd.SetArgs(args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("cmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestApiCommandOutputFormat tests different output formats
func TestApiCommandOutputFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "test",
			"id":   123,
		})
	}))
	defer server.Close()

	formats := []string{"json", "pretty", "table"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			var out bytes.Buffer
			f := &cmdutil.Factory{
				Config: &config.Config{
					Host:  server.URL,
					Token: "test-token",
				},
				IOStreams: &output.IOStreams{
					In:     os.Stdin,
					Out:    &out,
					ErrOut: &bytes.Buffer{},
				},
			}

			cmd := NewCmdApi(f)
			cmd.SetArgs([]string{"GET", "/api/test", "--format", format})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("cmd.Execute() with format %s error = %v", format, err)
			}

			outputStr := out.String()
			if !strings.Contains(outputStr, "test") {
				t.Errorf("output should contain 'test' for format %s, got: %s", format, outputStr)
			}
		})
	}
}

// TestClientGetWithTestServer tests the client.Get method directly
func TestClientGetWithTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST (proxy mode), got %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["path"] != "/api/test" {
			t.Errorf("body path = %v, want /api/test", body["path"])
		}
		if body["method"] != "GET" {
			t.Errorf("body method = %v, want GET", body["method"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"result": "ok"})
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token")
	c.DisableProxy() // Test in direct mode for clarity

	// In direct mode, GET request
	resp, err := c.Get("/api/test", nil)
	if err != nil {
		t.Fatalf("client.Get() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["result"] != "ok" {
		t.Errorf("result = %v, want ok", result["result"])
	}
}

// TestClientPostWithTestServer tests the client.Post method directly
func TestClientPostWithTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["name"] != "test" {
			t.Errorf("body name = %v, want test", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token")
	c.DisableProxy()

	resp, err := c.Post("/api/create", map[string]interface{}{"name": "test"})
	if err != nil {
		t.Fatalf("client.Post() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["id"] != float64(1) {
		t.Errorf("result id = %v, want 1", result["id"])
	}
}

// TestClientPutWithTestServer tests the client.Put method directly
func TestClientPutWithTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"updated": true})
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token")
	c.DisableProxy()

	resp, err := c.Put("/api/update/1", map[string]interface{}{"name": "updated"})
	if err != nil {
		t.Fatalf("client.Put() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["updated"] != true {
		t.Errorf("result updated = %v, want true", result["updated"])
	}
}

// TestClientDeleteWithTestServer tests the client.Delete method directly
func TestClientDeleteWithTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"deleted": true})
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token")
	c.DisableProxy()

	resp, err := c.Delete("/api/delete/1")
	if err != nil {
		t.Fatalf("client.Delete() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["deleted"] != true {
		t.Errorf("result deleted = %v, want true", result["deleted"])
	}
}

// TestClientErrorHandling tests client error scenarios
func TestClientErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		serverFunc http.HandlerFunc
		wantErr    bool
		errContain string
	}{
		{
			name: "server returns 404",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"not found"}`))
			},
			wantErr:    true,
			errContain: "404",
		},
		{
			name: "server returns 500",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"server error"}`))
			},
			wantErr:    true,
			errContain: "500",
		},
		{
			name: "server returns 401",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
			},
			wantErr:    true,
			errContain: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverFunc)
			defer server.Close()

			c := client.NewClient(server.URL, "token")
			c.DisableProxy()

			_, err := c.Get("/api/test", nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errContain != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error should contain %q, got: %v", tt.errContain, err)
				}
			}
		})
	}
}

// TestClientProxyMode tests proxy mode request formatting
func TestClientProxyMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All requests should be POST to /gw/ai/proxy
		if r.Method != http.MethodPost {
			t.Errorf("expected POST in proxy mode, got %s", r.Method)
		}

		if !strings.HasSuffix(r.URL.Path, "/gw/ai/proxy") {
			t.Errorf("expected path to end with /gw/ai/proxy, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		// Verify proxy body structure
		if body["method"] != "GET" {
			t.Errorf("proxy body method = %v, want GET", body["method"])
		}
		if body["path"] != "/api/test" {
			t.Errorf("proxy body path = %v, want /api/test", body["path"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "token")
	// Proxy mode is enabled by default

	resp, err := c.Get("/api/test", nil)
	if err != nil {
		t.Fatalf("client.Get() in proxy mode error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

// TestOutputWriter tests the output writer functionality
func TestOutputWriter(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
		data   interface{}
	}{
		{
			name:   "JSON format",
			format: output.FormatJSON,
			data:   map[string]interface{}{"key": "value"},
		},
		{
			name:   "Pretty format",
			format: output.FormatPretty,
			data:   map[string]interface{}{"key": "value"},
		},
		{
			name:   "Table format",
			format: output.FormatTable,
			data:   map[string]interface{}{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ioStreams := &output.IOStreams{
				In:     os.Stdin,
				Out:    &buf,
				ErrOut: &bytes.Buffer{},
			}

			w := output.NewWriter(ioStreams, tt.format)
			err := w.Write(tt.data)
			if err != nil {
				t.Fatalf("writer.Write() error = %v", err)
			}

			if buf.Len() == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

// TestApiCommandEdgeCases tests edge cases for the api command
func TestApiCommandEdgeCases(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		var out bytes.Buffer
		f := &cmdutil.Factory{
			Config: &config.Config{
				Host:  server.URL,
				Token: "test-token",
			},
			IOStreams: &output.IOStreams{
				In:     os.Stdin,
				Out:    &out,
				ErrOut: &bytes.Buffer{},
			},
		}

		cmd := NewCmdApi(f)
		cmd.SetArgs([]string{"GET", "/"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("cmd.Execute() error = %v", err)
		}
	})

	t.Run("path with query string", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"query":"test"}`))
		}))
		defer server.Close()

		var out bytes.Buffer
		f := &cmdutil.Factory{
			Config: &config.Config{
				Host:  server.URL,
				Token: "test-token",
			},
			IOStreams: &output.IOStreams{
				In:     os.Stdin,
				Out:    &out,
				ErrOut: &bytes.Buffer{},
			},
		}

		cmd := NewCmdApi(f)
		cmd.SetArgs([]string{"GET", "/api/search?q=test"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("cmd.Execute() error = %v", err)
		}
	})

	t.Run("path with special characters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"special":"ok"}`))
		}))
		defer server.Close()

		var out bytes.Buffer
		f := &cmdutil.Factory{
			Config: &config.Config{
				Host:  server.URL,
				Token: "test-token",
			},
			IOStreams: &output.IOStreams{
				In:     os.Stdin,
				Out:    &out,
				ErrOut: &bytes.Buffer{},
			},
		}

		cmd := NewCmdApi(f)
		cmd.SetArgs([]string{"GET", "/api/test%20space"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("cmd.Execute() error = %v", err)
		}
	})
}

// TestParseDataEdgeCases tests edge cases for parseData
func TestParseDataEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"null value", "null"},
		{"numeric string", "123"},
		{"boolean string", "true"},
		{"complex nested JSON", `{"a":{"b":{"c":"d"}}}`},
		{"JSON with null value", `{"key":null}`},
		{"JSON with boolean", `{"active":true}`},
		{"JSON with number", `{"count":42}`},
		{"empty object", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseData(tt.input)
			// Just verify it doesn't panic
			if result == nil && tt.input != "" {
				t.Logf("parseData(%q) returned nil", tt.input)
			}
		})
	}
}

// TestParseParamsEdgeCases tests edge cases for parseParams
func TestParseParamsEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
	}{
		{"null value", "null", true},
		{"numeric string", "123", true},
		{"array", `[1,2,3]`, true},
		{"complex nested JSON", `{"a":{"b":1}}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseParams(tt.input)
			isNil := result == nil
			if isNil != tt.wantNil {
				t.Errorf("parseParams(%q) isNil = %v, wantNil %v", tt.input, isNil, tt.wantNil)
			}
		})
	}
}
