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

type appGetProxyRequest struct {
	Path      string
	Method    string
	HasParams bool
}

type appGetEnvelope struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
}

func appGetTestFactory(serverURL string) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer) {
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

func appGetDecodeProxyRequest(t *testing.T, r *http.Request) appGetProxyRequest {
	t.Helper()

	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		t.Fatalf("failed to decode proxy request: %v", err)
	}

	path, _ := raw["path"].(string)
	method, _ := raw["method"].(string)
	_, hasParams := raw["params"]
	return appGetProxyRequest{
		Path:      path,
		Method:    method,
		HasParams: hasParams,
	}
}

func appGetWriteEnvelope(t *testing.T, w http.ResponseWriter, envelope appGetEnvelope) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		t.Fatalf("failed to write envelope: %v", err)
	}
}

func appGetAssertProxyRequest(t *testing.T, r *http.Request, wantID string) {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Errorf("expected POST method, got %s", r.Method)
	}
	if r.URL.Path != "/gw/ai/proxy" {
		t.Errorf("expected proxy path /gw/ai/proxy, got %s", r.URL.Path)
	}
	query := r.URL.Query()
	if got := query.Get("applicationId"); got != wantID {
		t.Errorf("expected query applicationId=%s, got %s", wantID, got)
	}
	if values, ok := query["id"]; ok {
		t.Errorf("expected no backend id query parameter, got %v", values)
	}

	proxyReq := appGetDecodeProxyRequest(t, r)
	if proxyReq.Path != "/application/get" {
		t.Errorf("expected path /application/get, got %s", proxyReq.Path)
	}
	if proxyReq.Method != "GET" {
		t.Errorf("expected method GET, got %s", proxyReq.Method)
	}
	if proxyReq.HasParams {
		t.Error("expected no top-level params key for GET proxy request")
	}
}

func TestNewCmdAppGet_CanonicalFlagSuccess(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		appGetAssertProxyRequest(t, r, "123")
		appGetWriteEnvelope(t, w, appGetEnvelope{
			Code:    10000,
			Msg:     "success",
			Success: true,
			Result:  map[string]interface{}{"id": 123, "name": "Canonical App"},
		})
	}))
	defer server.Close()

	f, out, errOut := appGetTestFactory(server.URL)
	cmd := newCmdAppGet(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--application-id", "123"})

	if err := cmd.Execute(); err != nil {
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
	if got["code"] != float64(10000) {
		t.Errorf("expected code 10000 in output, got %v", got["code"])
	}
	if !strings.Contains(out.String(), "Canonical App") {
		t.Errorf("expected full envelope output to include application name, got %s", out.String())
	}
}

func TestNewCmdAppGet_AliasFlagSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appGetAssertProxyRequest(t, r, "123")
		appGetWriteEnvelope(t, w, appGetEnvelope{Code: 10000, Msg: "success", Success: true})
	}))
	defer server.Close()

	f, _, errOut := appGetTestFactory(server.URL)
	cmd := newCmdAppGet(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--id", "123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdAppGet_IDValidationBeforeNetwork(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing id", args: []string{}, wantErr: "--application-id is required"},
		{name: "conflicting ids", args: []string{"--application-id", "123", "--id", "456"}, wantErr: "--application-id and --id must match"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
			}))
			defer server.Close()

			f, _, errOut := appGetTestFactory(server.URL)
			cmd := newCmdAppGet(f)
			cmd.SetOut(errOut)
			cmd.SetErr(errOut)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected ID validation error")
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

func TestNewCmdAppGet_NoHostConfigured(t *testing.T) {
	f, _, errOut := appGetTestFactory("")
	cmd := newCmdAppGet(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--application-id", "123"})

	err := cmd.Execute()
	assertAppMissingHostError(t, err)
}

func TestNewCmdAppGet_BothEqualIDsAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appGetAssertProxyRequest(t, r, "123")
		appGetWriteEnvelope(t, w, appGetEnvelope{Code: 10000, Msg: "success", Success: true})
	}))
	defer server.Close()

	f, _, errOut := appGetTestFactory(server.URL)
	cmd := newCmdAppGet(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--application-id", "123", "--id", "123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdAppGet_SuccessFalseWritesEnvelopeAndErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appGetWriteEnvelope(t, w, appGetEnvelope{
			Code:    40001,
			Msg:     "application get failed",
			Success: false,
			Result:  map[string]interface{}{"reason": "denied"},
		})
	}))
	defer server.Close()

	f, out, errOut := appGetTestFactory(server.URL)
	cmd := newCmdAppGet(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--application-id", "123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected non-zero result when success is false")
	}
	if !strings.Contains(err.Error(), "application get failed") {
		t.Fatalf("expected platform failure message, got %v", err)
	}

	var got appGetEnvelope
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("stdout contains invalid JSON: %v", err)
	}
	if got.Success {
		t.Fatal("expected success false in output envelope")
	}
	if got.Code != 40001 || got.Msg != "application get failed" {
		t.Fatalf("expected full failure envelope, got %+v", got)
	}
}

func TestNewCmdAppGet_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"success":false,"msg":"server failed"}`)
	}))
	defer server.Close()

	f, _, errOut := appGetTestFactory(server.URL)
	cmd := newCmdAppGet(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--application-id", "123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected HTTP error")
	}
	if !strings.Contains(err.Error(), "API error (500)") {
		t.Fatalf("expected API error status, got %v", err)
	}
}

func TestNewCmdAppGet_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not valid json")
	}))
	defer server.Close()

	f, _, errOut := appGetTestFactory(server.URL)
	cmd := newCmdAppGet(f)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"--application-id", "123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if !strings.Contains(err.Error(), "invalid JSON response") {
		t.Fatalf("expected invalid JSON response error, got %v", err)
	}
}

func TestNewCmdAppGet_NilIOStreamsUsesCommandOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appGetWriteEnvelope(t, w, appGetEnvelope{
			Code:    10000,
			Msg:     "success",
			Success: true,
			Result:  map[string]interface{}{"id": 123, "name": "Nil Streams App"},
		})
	}))
	defer server.Close()

	f := &cmdutil.Factory{Config: &config.Config{Host: server.URL}}
	cmd := newCmdAppGet(f)
	cmd.SetArgs([]string{"--application-id", "123"})

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
