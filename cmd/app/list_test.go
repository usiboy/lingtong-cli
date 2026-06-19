// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
)

type appListProxyRequest struct {
	Path      string
	Method    string
	HasParams bool
}

type appListEnvelope struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
}

func appListTestFactory(serverURL string) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &cmdutil.Factory{
		Config: &config.Config{Host: serverURL},
		IOStreams: &output.IOStreams{
			Out:    out,
			ErrOut: errOut,
		},
	}, out, errOut
}

func appListDecodeProxyRequest(t *testing.T, r *http.Request) appListProxyRequest {
	t.Helper()

	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		t.Fatalf("failed to decode proxy request: %v", err)
	}

	path, _ := raw["path"].(string)
	method, _ := raw["method"].(string)
	_, hasParams := raw["params"]
	return appListProxyRequest{
		Path:      path,
		Method:    method,
		HasParams: hasParams,
	}
}

func appListWriteEnvelope(t *testing.T, w http.ResponseWriter, envelope appListEnvelope) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		t.Fatalf("failed to write envelope: %v", err)
	}
}

func TestNewCmdAppList_DefaultSuccess(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/gw/ai/proxy?pageNum=1&pageSize=40" {
			t.Errorf("expected default proxy URI, got %s", got)
		}

		proxyReq := appListDecodeProxyRequest(t, r)
		if proxyReq.Path != "/application/list" {
			t.Errorf("expected path /application/list, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "GET" {
			t.Errorf("expected method GET, got %s", proxyReq.Method)
		}
		if proxyReq.HasParams {
			t.Error("expected no top-level params key for GET proxy request")
		}

		appListWriteEnvelope(t, w, appListEnvelope{
			Code:    10000,
			Msg:     "success",
			Success: true,
			Result:  []map[string]interface{}{{"id": 101, "name": "Alpha"}},
		})
	}))
	defer server.Close()

	f, out, errOut := appListTestFactory(server.URL)
	cmd := newCmdAppList(f)
	cmd.SetErr(errOut)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 request, got %d", requestCount)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("stdout contains invalid JSON: %v", err)
	}
	if got["success"] != true {
		t.Errorf("expected success true in output, got %v", got["success"])
	}
	if !strings.Contains(out.String(), "Alpha") {
		t.Errorf("expected full envelope output to include application name, got %s", out.String())
	}
}

func TestNewCmdAppList_OptionalFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		expected := map[string]string{
			"pageNum":  "2",
			"pageSize": "5",
			"name":     "Sales Sync",
			"appId":    "app-external-42",
			"tenantId": "tenant-7",
		}
		for key, want := range expected {
			if got := query.Get(key); got != want {
				t.Errorf("expected query %s=%s, got %s", key, want, got)
			}
		}

		proxyReq := appListDecodeProxyRequest(t, r)
		if proxyReq.Path != "/application/list" || proxyReq.Method != "GET" {
			t.Errorf("unexpected proxy request: %+v", proxyReq)
		}
		if proxyReq.HasParams {
			t.Error("expected filters in URL query, not proxy body params")
		}

		appListWriteEnvelope(t, w, appListEnvelope{Code: 10000, Msg: "success", Success: true})
	}))
	defer server.Close()

	f, _, errOut := appListTestFactory(server.URL)
	cmd := newCmdAppList(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--page-num", "2", "--page-size", "5", "--name", "Sales Sync", "--app-id", "app-external-42", "--tenant-id", "tenant-7"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdAppList_InvalidPaginationBeforeNetwork(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "page num less than one", args: []string{"--page-num", "0"}, wantErr: "--page-num must be at least 1"},
		{name: "page size less than one", args: []string{"--page-size", "0"}, wantErr: "--page-size must be at least 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
			}))
			defer server.Close()

			f, _, errOut := appListTestFactory(server.URL)
			cmd := newCmdAppList(f)
			cmd.SetOut(errOut)
			cmd.SetErr(errOut)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected pagination validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error %q, got %v", tt.wantErr, err)
			}
			if requestCount != 0 {
				t.Fatalf("expected validation before network, got %d requests", requestCount)
			}
		})
	}
}

func TestNewCmdAppList_NoHostConfigured(t *testing.T) {
	f, _, errOut := appListTestFactory("")
	cmd := NewCmdApp(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"list", "--page-num", "1", "--page-size", "10", "--format", "json"})

	err := cmd.Execute()
	assertAppMissingHostError(t, err)
}

func TestNewCmdAppList_SuccessFalseNoOutputAndErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appListWriteEnvelope(t, w, appListEnvelope{
			Code:    40001,
			Msg:     "application list failed",
			Success: false,
			Result:  map[string]interface{}{"reason": "denied"},
		})
	}))
	defer server.Close()

	f, out, errOut := appListTestFactory(server.URL)
	cmd := newCmdAppList(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected non-zero result when success is false")
	}
	if !strings.Contains(err.Error(), "application list failed") {
		t.Fatalf("expected platform failure message, got %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("expected no stdout output on failure, got %s", out.String())
	}
}

func TestNewCmdAppList_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"success":false,"msg":"server failed"}`)
	}))
	defer server.Close()

	f, _, errOut := appListTestFactory(server.URL)
	cmd := newCmdAppList(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected HTTP error")
	}
	if !strings.Contains(err.Error(), "API error (500)") {
		t.Fatalf("expected API error status, got %v", err)
	}
}

func TestNewCmdAppList_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not valid json")
	}))
	defer server.Close()

	f, _, errOut := appListTestFactory(server.URL)
	cmd := newCmdAppList(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if !strings.Contains(err.Error(), "invalid JSON response") {
		t.Fatalf("expected invalid JSON response error, got %v", err)
	}
}

func TestNewCmdAppList_NilIOStreamsUsesCommandOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appListWriteEnvelope(t, w, appListEnvelope{
			Code:    10000,
			Msg:     "success",
			Success: true,
			Result:  []map[string]interface{}{{"id": 202, "name": "Nil Streams App"}},
		})
	}))
	defer server.Close()

	f := &cmdutil.Factory{Config: &config.Config{Host: server.URL}}
	cmd := newCmdAppList(f)

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Nil Streams App") {
		t.Fatalf("expected command stdout fallback output, got %s", out.String())
	}
}
