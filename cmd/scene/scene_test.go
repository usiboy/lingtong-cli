// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package scene

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
	"github.com/spf13/cobra"
)

type sceneProxyRequest struct {
	Path      string
	Method    string
	Body      map[string]interface{}
	HasParams bool
	HasBody   bool
}

func newTestFactory(serverURL string) *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  serverURL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    io.Discard,
			ErrOut: io.Discard,
		},
	}
}

func executeSceneCommand(cmd *cobra.Command, args ...string) error {
	if cmd.Flag("format") == nil {
		cmd.Flags().String("format", "json", "Output format")
	}
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd.Execute()
}

func decodeSceneProxyRequest(t *testing.T, r *http.Request) sceneProxyRequest {
	t.Helper()

	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		t.Fatalf("failed to decode proxy request: %v", err)
	}

	path, _ := raw["path"].(string)
	method, _ := raw["method"].(string)
	body, hasBody := raw["body"].(map[string]interface{})
	_, hasParams := raw["params"]
	return sceneProxyRequest{
		Path:      path,
		Method:    method,
		Body:      body,
		HasParams: hasParams,
		HasBody:   hasBody,
	}
}

func assertSceneProxyRequest(t *testing.T, r *http.Request, wantMethod, wantPath string) sceneProxyRequest {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Errorf("expected incoming request method POST, got %s", r.Method)
	}
	if r.URL.Path != "/gw/ai/proxy" {
		t.Errorf("expected incoming request path /gw/ai/proxy, got %s", r.URL.Path)
	}

	proxyReq := decodeSceneProxyRequest(t, r)
	if proxyReq.Path != wantPath {
		t.Errorf("expected proxy path %s, got %s", wantPath, proxyReq.Path)
	}
	if proxyReq.Method != wantMethod {
		t.Errorf("expected proxy method %s, got %s", wantMethod, proxyReq.Method)
	}
	return proxyReq
}

func writeSceneJSON(t *testing.T, w http.ResponseWriter, payload map[string]interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("failed to write response: %v", err)
	}
}

func TestNewCmdSceneList_Success(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		proxyReq := assertSceneProxyRequest(t, r, "GET", "/scene/list")
		if proxyReq.HasParams {
			t.Error("expected no top-level params key for GET proxy request")
		}
		if proxyReq.HasBody {
			t.Error("expected no body key for scene list GET proxy request")
		}

		query := r.URL.Query()
		if got := query.Get("pageNum"); got != "1" {
			t.Errorf("expected query pageNum=1, got %s", got)
		}
		if got := query.Get("pageSize"); got != "20" {
			t.Errorf("expected query pageSize=20, got %s", got)
		}
		if values, ok := query["page"]; ok {
			t.Errorf("expected no legacy page query parameter, got %v", values)
		}

		w.WriteHeader(http.StatusOK)
		writeSceneJSON(t, w, map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{
					{"id": 1, "name": "Order Sync"},
					{"id": 2, "name": "Data Pipeline"},
				},
				"total": 2,
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneList(f)

	err := executeSceneCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 request, got %d", requestCount)
	}
}

func TestNewCmdSceneList_WithAppID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := assertSceneProxyRequest(t, r, "GET", "/scene/list")
		if proxyReq.HasParams {
			t.Error("expected no top-level params key for GET proxy request")
		}

		query := r.URL.Query()
		if got := query.Get("pageNum"); got != "2" {
			t.Errorf("expected query pageNum=2, got %s", got)
		}
		if got := query.Get("pageSize"); got != "5" {
			t.Errorf("expected query pageSize=5, got %s", got)
		}
		if got := query.Get("appId"); got != "app-42" {
			t.Errorf("expected query appId=app-42, got %s", got)
		}

		w.WriteHeader(http.StatusOK)
		writeSceneJSON(t, w, map[string]interface{}{
			"result": map[string]interface{}{
				"list":  []map[string]interface{}{},
				"total": 0,
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneList(f)

	if err := executeSceneCommand(cmd, "--page", "2", "--page-size", "5", "--app-id", "app-42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdSceneList_InvalidPagination(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "page less than one", args: []string{"--page", "0"}, wantErr: "--page must be at least 1"},
		{name: "page size less than one", args: []string{"--page-size", "0"}, wantErr: "--page-size must be at least 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
			}))
			defer server.Close()

			f := newTestFactory(server.URL)
			cmd := newCmdSceneList(f)

			err := executeSceneCommand(cmd, tt.args...)
			if err == nil {
				t.Fatal("expected pagination validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if requestCount != 0 {
				t.Fatalf("expected no network request, got %d", requestCount)
			}
		})
	}
}

func TestNewCmdSceneList_NoHostConfigured(t *testing.T) {
	f := newTestFactory("")
	cmd := newCmdSceneList(f)

	err := executeSceneCommand(cmd)
	if err == nil {
		t.Fatal("expected missing host configuration error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected no host configured error, got %v", err)
	}
	if !strings.Contains(err.Error(), "lingtong-cli config init") {
		t.Fatalf("expected config init guidance, got %v", err)
	}
	if strings.Contains(err.Error(), "unsupported protocol scheme") {
		t.Fatalf("expected friendly config error, got raw network error: %v", err)
	}
}

func TestNewCmdSceneCreate_NoHostConfigured(t *testing.T) {
	f := newTestFactory("")
	cmd := newCmdSceneCreate(f)

	err := executeSceneCommand(cmd, "--name", "Test Scene")
	if err == nil {
		t.Fatal("expected missing host configuration error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected no host configured error, got %v", err)
	}
}

func TestNewCmdSceneInfo_NoHostConfigured(t *testing.T) {
	f := newTestFactory("")
	cmd := newCmdSceneInfo(f)

	err := executeSceneCommand(cmd, "--scene-id", "100")
	if err == nil {
		t.Fatal("expected missing host configuration error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected no host configured error, got %v", err)
	}
}

func TestNewCmdSceneList_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeSceneJSON(t, w, map[string]interface{}{
			"error": "Internal server error",
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneList(f)

	err := executeSceneCommand(cmd)
	if err == nil {
		t.Fatal("expected server error")
	}
	if !strings.Contains(err.Error(), "API error (500)") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestNewCmdSceneCreate_MissingName(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdSceneCreate(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error about name, got: %v", err)
	}
}

func TestNewCmdSceneCreate_Success(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		proxyReq := assertSceneProxyRequest(t, r, "POST", "/scene/model/save")
		if proxyReq.HasParams {
			t.Error("expected no top-level params key for scene create POST proxy request")
		}
		if !proxyReq.HasBody {
			t.Fatal("expected body key for scene create POST proxy request")
		}
		if got := proxyReq.Body["name"]; got != "Test Scene" {
			t.Errorf("expected body name Test Scene, got %v", got)
		}
		if got := proxyReq.Body["description"]; got != "Test description" {
			t.Errorf("expected body description Test description, got %v", got)
		}

		w.WriteHeader(http.StatusOK)
		writeSceneJSON(t, w, map[string]interface{}{
			"result": map[string]interface{}{
				"id":   100,
				"name": "Test Scene",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneCreate(f)

	err := executeSceneCommand(cmd, "--name", "Test Scene", "--description", "Test description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 request, got %d", requestCount)
	}
}

func TestNewCmdSceneCreate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeSceneJSON(t, w, map[string]interface{}{
			"error": "Failed to create scene",
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneCreate(f)

	err := executeSceneCommand(cmd, "--name", "Test Scene")
	if err == nil {
		t.Fatal("expected server error")
	}
	if !strings.Contains(err.Error(), "API error (500)") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestNewCmdSceneInfo_MissingSceneId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdSceneInfo(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --scene-id")
	}
}

func TestNewCmdSceneInfo_Success(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		proxyReq := assertSceneProxyRequest(t, r, "GET", "/scene/detail/get")
		if proxyReq.HasParams {
			t.Error("expected no top-level params key for GET proxy request")
		}
		query := r.URL.Query()
		if got := query.Get("sceneId"); got != "100" {
			t.Errorf("expected query sceneId=100, got %s", got)
		}

		w.WriteHeader(http.StatusOK)
		writeSceneJSON(t, w, map[string]interface{}{
			"result": map[string]interface{}{
				"id":          100,
				"name":        "Test Scene",
				"description": "Test description",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneInfo(f)

	err := executeSceneCommand(cmd, "--scene-id", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 request, got %d", requestCount)
	}
}

func TestNewCmdSceneInfo_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeSceneJSON(t, w, map[string]interface{}{
			"error": "Scene not found",
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneInfo(f)

	err := executeSceneCommand(cmd, "--scene-id", "999")
	if err == nil {
		t.Fatal("expected server error")
	}
	if !strings.Contains(err.Error(), "API error (500)") {
		t.Fatalf("expected API error, got %v", err)
	}
}
