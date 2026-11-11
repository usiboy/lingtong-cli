// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestConnectorInvokeProxy(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		want     map[string]interface{}
		response string
	}{
		{"parameters", []string{"--connector", "kmerp", "--method", "erp.trade.list.query", "--auth-account", "广州力人服饰", "--env", "test", "--body", `{"pageSize":100,"filter":{"name":"广州"}}`}, map[string]interface{}{"connector": "kmerp", "method": "erp.trade.list.query", "authName": "广州力人服饰", "env": "test", "requestBody": map[string]interface{}{"pageSize": float64(100), "filter": map[string]interface{}{"name": "广州"}}}, `{"success":true,"result":{"receiptStatus":"SUCCESS","responseBody":{"body":"keep","list":[{"id":1}],"total":1}}}`},
		{"built-in defaults", []string{"--connector", "ltTable", "--method", "example"}, map[string]interface{}{"connector": "ltTable", "method": "example", "authName": "", "env": "prod", "requestBody": map[string]interface{}{}}, `{"success":true,"result":{"receiptStatus":"SUCCESS","responseBody":[{"id":1}]}}`},
		{"business failure", []string{"--connector", "kmerp", "--method", "example", "--auth-account", "test"}, map[string]interface{}{"connector": "kmerp", "method": "example", "authName": "test", "env": "prod", "requestBody": map[string]interface{}{}}, `{"success":true,"result":{"receiptStatus":"FAIL","responseBody":{"success":false,"message":"failed"}}}`},
		{"unknown result", []string{"--connector", "kmerp", "--method", "example"}, map[string]interface{}{"connector": "kmerp", "method": "example", "authName": "", "env": "prod", "requestBody": map[string]interface{}{}}, `{"success":false,"msg":"执行结果可能未知，请核实后再操作","result":null}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "POST" || r.URL.Path != "/gw/ai/proxy" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer apk-test123" {
					t.Error("missing bearer token")
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing JSON content type")
				}
				var got map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Error(err)
				}
				want := map[string]interface{}{"path": "/gw/connector/invoke", "method": "POST", "body": tc.want}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("request = %#v, want %#v", got, want)
				}
				w.Write([]byte(tc.response))
			}))
			defer server.Close()
			f := newTestFactory(server.URL)
			var out bytes.Buffer
			f.IOStreams.Out = &out
			cmd := NewCmdConnector(f)
			cmd.SetArgs(append([]string{"invoke"}, tc.args...))
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Errorf("calls = %d, want 1", calls)
			}
			var got, want interface{}
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.response), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response changed: %s", out.String())
			}
		})
	}
}

func TestConnectorInvokeInvalidInput(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.Write([]byte(`{}`)) }))
	defer server.Close()
	for _, args := range [][]string{
		{"--method", "example"}, {"--connector", "kmerp"},
		{"--connector", "kmerp", "--method", "example", "--body", "null"},
		{"--connector", "kmerp", "--method", "example", "--body", "[]"},
		{"--connector", "kmerp", "--method", "example", "--body", "invalid"},
		{"--connector", "kmerp", "--method", "example", "--body", `"text"`},
	} {
		cmd := NewCmdConnector(newTestFactory(server.URL))
		cmd.SetArgs(append([]string{"invoke"}, args...))
		if err := cmd.Execute(); err == nil {
			t.Errorf("expected validation error for %v", args)
		}
	}
	if calls != 0 {
		t.Errorf("invalid inputs sent %d requests", calls)
	}
}

func TestConnectorInvokeErrors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		body, want string
	}{
		{"unauthorized", 401, `{"msg":"denied"}`, "API error (401)"},
		{"server error", 500, `{"msg":"执行结果可能未知"}`, "执行结果可能未知"},
		{"invalid JSON", 200, "invalid", "invalid character"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()
			cmd := NewCmdConnector(newTestFactory(server.URL))
			cmd.SetArgs([]string{"invoke", "--connector", "ltTable", "--method", "example"})
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want %s", err, tc.want)
			}
			if calls != 1 {
				t.Errorf("calls = %d, want 1 without retries", calls)
			}
		})
	}
}

func TestConnectorInvokeForceIgnoresCache(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, err := cacheFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	// A directory in place of the cache makes any attempt to read/write it fail.
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"success":true}`)) }))
	defer server.Close()
	f := newTestFactory(server.URL)
	var out, errOut bytes.Buffer
	f.IOStreams.Out = &out
	f.IOStreams.ErrOut = &errOut
	cmd := NewCmdConnector(f)
	cmd.SetArgs([]string{"invoke", "--connector", "ltTable", "--method", "example", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "deprecated") {
		t.Error("expected deprecation warning on stderr")
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Fatal("invoke modified cache")
	}
}

func TestConnectorLegacyCacheCommands(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	original := &ConnectorCache{Universal: map[string]*ConnectorCacheEntry{"test": {WorkflowId: 123, AppTag: "old", Env: "test"}}, Legacy: map[string]*ConnectorCacheEntry{}}
	if err := saveCache(original); err != nil {
		t.Fatal(err)
	}
	f := newTestFactory("http://unused.invalid")
	var out bytes.Buffer
	f.IOStreams.Out = &out
	cmd := NewCmdConnector(f)
	cmd.SetArgs([]string{"cache", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"workflowId": 123`) {
		t.Fatalf("missing legacy entry: %s", out.String())
	}
	cmd = NewCmdConnector(f)
	cmd.SetArgs([]string{"cache", "clear", "--key", "test"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	cache, err := loadCache()
	if err != nil {
		t.Fatal(err)
	}
	if len(cache.Universal) != 0 {
		t.Fatal("cache not cleared")
	}
}

// timeoutTransport simulates the transport deadline without waiting 120 seconds.
type timeoutTransport struct {
	t     *testing.T
	calls int
}

func (tr *timeoutTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr.calls++
	deadline, ok := r.Context().Deadline()
	if remaining := time.Until(deadline); !ok || remaining < 119*time.Second || remaining > 120*time.Second {
		tr.t.Errorf("expected 120-second invocation deadline, got %v (set=%v)", remaining, ok)
	}
	return nil, context.DeadlineExceeded
}

func TestConnectorInvokeTimeout(t *testing.T) {
	previous := http.DefaultTransport
	transport := &timeoutTransport{t: t}
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
	cmd := NewCmdConnector(newTestFactory("https://unused.invalid"))
	cmd.SetArgs([]string{"invoke", "--connector", "ltTable", "--method", "example"})
	err := cmd.Execute()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected timeout cause, got %v", err)
	}
	if transport.calls != 1 {
		t.Errorf("timeout retried: %d calls", transport.calls)
	}
}
