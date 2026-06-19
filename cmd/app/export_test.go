// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
)

func newTestFactory(serverURL string) *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  serverURL,
			Token: "test-token",
		},
	}
}

func signResponse(sign string) map[string]interface{} {
	return map[string]interface{}{
		"result": map[string]interface{}{"sign": sign},
	}
}

func exportResponse() map[string]interface{} {
	return map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"connector1", "connector2"},
		"basicDatas":    []interface{}{},
		"scenes":        []interface{}{},
	}
}

func TestNewCmdAppExport_MissingFlags(t *testing.T) {
	f := newTestFactory("http://localhost")
	cmd := newCmdAppExport(f)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing flags")
	}
}

func TestNewCmdAppExport_NoHostConfigured(t *testing.T) {
	f := newTestFactory("")
	cmd := newCmdAppExport(f)
	cmd.SetArgs([]string{"--app-id", "42", "--output", "app.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	assertAppMissingHostError(t, err)
}

func TestNewCmdAppExport_DryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gw/ai/proxy" {
			t.Errorf("expected proxy path /gw/ai/proxy, got: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got: %s", r.Method)
		}

		var proxyReq map[string]interface{}
		json.NewDecoder(r.Body).Decode(&proxyReq)

		w.WriteHeader(http.StatusOK)
		if proxyReq["path"] == "/gw/ai/application/export/sign" {
			json.NewEncoder(w).Encode(signResponse("test-sign-123"))
		}
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdAppExport(f)
	cmd.SetArgs([]string{"--app-id", "42", "--output", "app.json", "--dry-run"})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "appId=42") {
		t.Errorf("expected app-id in output, got: %s", output)
	}
}

func TestNewCmdAppExport_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gw/ai/proxy" {
			t.Errorf("expected proxy path /gw/ai/proxy, got: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got: %s", r.Method)
		}

		var proxyReq map[string]interface{}
		json.NewDecoder(r.Body).Decode(&proxyReq)

		w.WriteHeader(http.StatusOK)
		if proxyReq["path"] == "/gw/ai/application/export/sign" {
			json.NewEncoder(w).Encode(signResponse("test-sign-123"))
		} else if strings.Contains(proxyReq["path"].(string), "/application/export") {
			json.NewEncoder(w).Encode(exportResponse())
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "app.json")

	f := newTestFactory(server.URL)
	cmd := newCmdAppExport(f)
	cmd.SetArgs([]string{"--app-id", "42", "--output", outputFile})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output file contains invalid JSON: %v", err)
	}

	if result["appName"] != "Test App" {
		t.Errorf("expected appName 'Test App', got: %v", result["appName"])
	}
}

func TestNewCmdAppExport_SignError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "sign failed"})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdAppExport(f)
	cmd.SetArgs([]string{"--app-id", "42", "--output", "app.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for sign failure")
	}
	if !strings.Contains(err.Error(), "sign") {
		t.Errorf("expected error about sign, got: %v", err)
	}
}

func TestNewCmdAppExport_ExportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proxyReq map[string]interface{}
		json.NewDecoder(r.Body).Decode(&proxyReq)

		if proxyReq["path"] == "/gw/ai/application/export/sign" {
			json.NewEncoder(w).Encode(signResponse("test-sign"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "export failed"})
		}
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdAppExport(f)
	cmd.SetArgs([]string{"--app-id", "42", "--output", "app.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for export failure")
	}
	if !strings.Contains(err.Error(), "export") {
		t.Errorf("expected error about export, got: %v", err)
	}
}

func TestNewCmdAppExport_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proxyReq map[string]interface{}
		json.NewDecoder(r.Body).Decode(&proxyReq)

		if proxyReq["path"] == "/gw/ai/application/export/sign" {
			json.NewEncoder(w).Encode(signResponse("test-sign"))
		} else {
			w.Write([]byte("not valid json"))
		}
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdAppExport(f)
	cmd.SetArgs([]string{"--app-id", "42", "--output", "app.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "JSON") {
		t.Errorf("expected error about JSON, got: %v", err)
	}
}
