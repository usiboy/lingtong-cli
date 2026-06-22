// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package connector

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
)

func newTestFactory(serverURL string) *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  serverURL,
			Token: "apk-test123",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    io.Discard,
			ErrOut: io.Discard,
		},
	}
}

func TestNewCmdConnector(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := NewCmdConnector(f)

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	if cmd.Use != "connector" {
		t.Errorf("expected Use 'connector', got '%s'", cmd.Use)
	}

	expectedSubs := []string{"info", "category", "list", "account", "check-auth"}
	for _, sub := range expectedSubs {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == sub {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand '%s' not found", sub)
		}
	}

	formatFlag := cmd.PersistentFlags().Lookup("format")
	if formatFlag == nil {
		t.Error("expected --format flag")
	}
}

func TestNewCmdConnectorInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.URL.Path != "/gw/ai/proxy" {
			t.Errorf("expected /gw/ai/proxy, got %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer apk-test123" {
			t.Errorf("expected Bearer token, got %s", auth)
		}

		var proxyReq map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&proxyReq); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		proxiedPath, ok := proxyReq["path"].(string)
		if !ok {
			t.Errorf("missing or invalid 'path' field in request")
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}

		if !strings.Contains(proxiedPath, "/gw/ai/connector/info") {
			t.Errorf("expected proxied path to contain /gw/ai/connector/info, got %s", proxiedPath)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"data":{"name":"kmerp"}}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--env", "test"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorInfoMissingConnector(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --connector flag")
	}
}

func TestNewCmdConnectorCategoryList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"result":[]}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorCategoryList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"result":[]}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorListWithAppId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"result":[]}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--app-id", "165"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorAccountList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"result":[]}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorAccountList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--env", "test"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorAccountVerify(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		if callCount == 1 {
			w.Write([]byte(`{"success":true,"result":{"id":123}}`))
		} else {
			w.Write([]byte(`{"success":true,"result":{"verified":true}}`))
		}
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorAccountVerify(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--account-id", "123"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorAccountVerifyMissingConnector(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorAccountVerify(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--account-id", "123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --connector")
	}
}

func TestNewCmdConnectorAccountVerifyMissingAccountId(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorAccountVerify(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --account-id")
	}
}

func TestNewCmdConnectorAccountCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"result":{"accountRelationId":"abc123"}}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorAccountCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{
		"--connector", "kmerp",
		"--name", "Test Account",
		"--env", "test",
		"--data", `{"appKey":"xxx"}`,
	})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorAccountCreateMissingConnector(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorAccountCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--name", "test", "--data", "{}"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --connector")
	}
}

func TestNewCmdConnectorAccountCreateMissingName(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorAccountCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--data", "{}"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name")
	}
}

func TestNewCmdConnectorAccountCreateMissingData(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorAccountCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--name", "test"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --data")
	}
}

func TestNewCmdConnectorAccountCreateInvalidJSON(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorAccountCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{
		"--connector", "kmerp",
		"--name", "Test",
		"--data", "invalid-json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNewCmdConnectorCheckAuthWithConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"result":[]}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorCheckAuth(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdConnectorCheckAuthMissingParams(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := newCmdConnectorCheckAuth(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing required params")
	}
}

func TestConnectorAccountSubcommands(t *testing.T) {
	f := newTestFactory("https://test.example.com")
	cmd := NewCmdConnector(f)

	var accountSubCmd interface{}
	for _, c := range cmd.Commands() {
		if c.Name() == "account" {
			accountSubCmd = c
			break
		}
	}

	if accountSubCmd == nil {
		t.Fatal("account subcommand not found")
	}
}

// ============================================================================
// P1: 错误处理测试
// ============================================================================

func TestNewCmdConnectorInfo_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 error, got: %v", err)
	}
}

func TestNewCmdConnectorInfo_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unauthorized")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 error, got: %v", err)
	}
}

func TestNewCmdConnectorInfo_NetworkError(t *testing.T) {
	f := newTestFactory("http://invalid-host-that-does-not-exist.local:12345")
	cmd := newCmdConnectorInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for network failure")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("expected network error, got: %v", err)
	}
}

func TestNewCmdConnectorList_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json response`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid JSON error, got: %v", err)
	}
}

func TestNewCmdConnectorAccountVerify_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorAccountVerify(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--account-id", "123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestNewCmdConnectorAccountCreate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorAccountCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{
		"--connector", "kmerp",
		"--name", "Test Account",
		"--data", `{"appKey":"xxx"}`,
	})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestNewCmdConnectorCheckAuth_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorCheckAuth(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	_ = cmd.Execute() // Just verify no panic
}

func TestNewCmdConnectorInfo_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(``))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for empty response")
	}
}

func TestNewCmdConnectorCategoryList_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdConnectorCategoryList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unauthorized")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 error, got: %v", err)
	}
}

// ============================================================================
// P1: 错误处理测试
// ============================================================================
