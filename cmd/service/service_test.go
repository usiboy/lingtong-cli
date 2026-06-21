// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package service

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/openapi"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func newTestFactory(serverURL string) *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  serverURL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     &bytes.Buffer{},
			Out:    io.Discard,
			ErrOut: io.Discard,
		},
	}
}

func newTestSpec() *openapi.Spec {
	specJSON := `{
		"swagger": "2.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/scene/list": {
				"get": {
					"summary": "List scenes",
					"parameters": [
						{"name": "pageNum", "in": "query", "type": "integer", "required": true},
						{"name": "pageSize", "in": "query", "type": "integer", "required": false}
					]
				}
			},
			"/scene/get": {
				"get": {
					"summary": "Get scene by ID",
					"parameters": [
						{"name": "sceneId", "in": "query", "type": "integer", "required": true}
					]
				}
			},
			"/basicdata/save": {
				"post": {
					"summary": "Save basic data",
					"parameters": [
						{"name": "body", "in": "body", "required": true, "schema": {"type": "object"}}
					]
				}
			}
		}
	}`

	spec, _ := openapi.ParseBytes([]byte(specJSON))
	return spec
}

func TestRegisterServiceCommands(t *testing.T) {
	f := newTestFactory("http://localhost:8080")
	spec := newTestSpec()

	root := &cobra.Command{Use: "test"}
	RegisterServiceCommands(root, f, spec)

	// Check that service command was added
	var serviceCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "service" {
			serviceCmd = cmd
			break
		}
	}
	if serviceCmd == nil {
		t.Fatal("service command not found")
	}

	// Check that module commands were added
	moduleNames := make(map[string]bool)
	for _, cmd := range serviceCmd.Commands() {
		moduleNames[cmd.Name()] = true
	}

	if !moduleNames["scene"] {
		t.Error("scene module not found")
	}
	if !moduleNames["basicdata"] {
		t.Error("basicdata module not found")
	}
}

func TestBuildModuleCommand(t *testing.T) {
	f := newTestFactory("http://localhost:8080")
	spec := newTestSpec()

	groups := spec.GroupByPrefix()
	sceneOps := groups["scene"]

	cmd := buildModuleCommand("scene", sceneOps, f)

	if cmd.Use != "scene" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "scene")
	}

	// Check that operation commands were added
	cmdNames := make(map[string]bool)
	for _, subCmd := range cmd.Commands() {
		cmdNames[subCmd.Name()] = true
	}

	if !cmdNames["list"] {
		t.Error("list command not found")
	}
	if !cmdNames["get"] {
		t.Error("get command not found")
	}
}

func TestBuildOperationCommand(t *testing.T) {
	f := newTestFactory("http://localhost:8080")

	op := &openapi.Operation{
		Summary:     "List scenes",
		Description: "Returns a paginated list of scenes",
		Path:        "/scene/list",
		Method:      "GET",
		Parameters: []openapi.Parameter{
			{Name: "pageNum", In: "query", Type: "integer", Required: true, Description: "Page number"},
			{Name: "pageSize", In: "query", Type: "integer", Required: false, Description: "Page size"},
		},
	}

	cmd := buildOperationCommand(op, "list", f)

	if cmd.Use != "list" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "list")
	}
	if cmd.Short != "List scenes" {
		t.Errorf("cmd.Short = %q, want %q", cmd.Short, "List scenes")
	}

	// Check flags
	pageNumFlag := cmd.Flags().Lookup("pageNum")
	if pageNumFlag == nil {
		t.Error("pageNum flag not found")
	}

	pageSizeFlag := cmd.Flags().Lookup("pageSize")
	if pageSizeFlag == nil {
		t.Error("pageSize flag not found")
	}
}

func TestExecuteOperationGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if !strings.Contains(r.URL.String(), "pageNum=1") {
			t.Errorf("expected pageNum=1 in URL, got %s", r.URL.String())
		}
		if !strings.Contains(r.URL.String(), "pageSize=10") {
			t.Errorf("expected pageSize=10 in URL, got %s", r.URL.String())
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {"list": [], "total": 0}}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	var buf bytes.Buffer
	f.IOStreams.Out = &buf

	op := &openapi.Operation{
		Summary: "List scenes",
		Path:    "/scene/list",
		Method:  "GET",
		Parameters: []openapi.Parameter{
			{Name: "pageNum", In: "query", Type: "integer", Required: true},
			{Name: "pageSize", In: "query", Type: "integer", Required: false},
		},
	}

	cmd := buildOperationCommand(op, "list", f)
	cmd.SetArgs([]string{"--pageNum", "1", "--pageSize", "10"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check output contains expected data
	output := buf.String()
	if !strings.Contains(output, "result") {
		t.Errorf("output should contain 'result', got: %s", output)
	}
}

func TestExecuteOperationPOST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true, "result": {"id": 123}}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	var buf bytes.Buffer
	f.IOStreams.Out = &buf

	op := &openapi.Operation{
		Summary: "Save data",
		Path:    "/basicdata/save",
		Method:  "POST",
		Parameters: []openapi.Parameter{
			{Name: "body", In: "body", Required: true, Schema: &openapi.Schema{Type: "object"}},
		},
	}

	cmd := buildOperationCommand(op, "save", f)
	cmd.SetArgs([]string{"--body", `{"name": "test"}`})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "success") {
		t.Errorf("output should contain 'success', got: %s", output)
	}
}

func TestPathParamSubstitution(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Proxy mode wraps the real path in the JSON body; capture it.
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		gotPath = buf.String()
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	var out bytes.Buffer
	f.IOStreams.Out = &out

	op := &openapi.Operation{
		Summary: "Get model",
		Path:    "/meta/model/{connector}/get",
		Method:  "GET",
		Parameters: []openapi.Parameter{
			{Name: "connector", In: "path", Type: "string", Required: true},
		},
	}
	cmd := buildOperationCommand(op, "model-connector-get", f)
	cmd.SetArgs([]string{"--connector", "push_state"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(gotPath, "/meta/model/push_state/get") {
		t.Errorf("path param not substituted, proxy body = %s", gotPath)
	}
	if strings.Contains(gotPath, "{connector}") {
		t.Errorf("literal brace remained in path: %s", gotPath)
	}
}

func TestPathParamRequired(t *testing.T) {
	f := newTestFactory("http://localhost:8080")
	op := &openapi.Operation{
		Summary:    "Get model",
		Path:       "/meta/model/{connector}/get",
		Method:     "GET",
		Parameters: []openapi.Parameter{{Name: "connector", In: "path", Type: "string", Required: true}},
	}
	cmd := buildOperationCommand(op, "model-connector-get", f)
	cmd.SetArgs([]string{}) // omit required path param
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when required path param missing")
	}
}

func TestWildcardPathSkipped(t *testing.T) {
	specJSON := `{
		"swagger": "2.0",
		"paths": {
			"/noco/datasource/app/deploy": {"post": {"summary": "deploy"}},
			"/noco/datasource/app/{id}/**": {"get": {"summary": "wild"}}
		}
	}`
	spec, _ := openapi.ParseBytes([]byte(specJSON))
	f := newTestFactory("http://localhost:8080")
	cmd := buildModuleCommand("noco", spec.GroupByPrefix()["noco"], f)

	names := make(map[string]bool)
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
		if strings.ContainsAny(c.Name(), "{}*") {
			t.Errorf("generated command with illegal chars: %q", c.Name())
		}
	}
	if !names["datasource-app-deploy"] {
		t.Error("expected datasource-app-deploy command")
	}
	if len(cmd.Commands()) != 1 {
		t.Errorf("expected wildcard path skipped (1 command), got %d", len(cmd.Commands()))
	}
}

func TestExecuteOperationBodyFromFields(t *testing.T) {
	var bodySeen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		bodySeen = buf.String()
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	var out bytes.Buffer
	f.IOStreams.Out = &out

	op := &openapi.Operation{
		Summary: "Save",
		Path:    "/basicdata/save",
		Method:  "POST",
		Parameters: []openapi.Parameter{
			{Name: "body", In: "body", Required: true, Schema: &openapi.Schema{
				Type: "object",
				Properties: map[string]*openapi.Schema{
					"name":    {Type: "string"},
					"count":   {Type: "integer"},
					"enabled": {Type: "boolean"},
				},
			}},
		},
	}
	cmd := buildOperationCommand(op, "save", f)
	cmd.SetArgs([]string{"--name", "widget", "--count", "5", "--enabled", "true"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, want := range []string{`"name":"widget"`, `"count":5`, `"enabled":true`} {
		if !strings.Contains(strings.ReplaceAll(bodySeen, " ", ""), want) {
			t.Errorf("body missing %s, got %s", want, bodySeen)
		}
	}
}

func TestExecuteOperationInvalidBody(t *testing.T) {
	f := newTestFactory("http://localhost:8080")
	op := &openapi.Operation{
		Summary:    "Save",
		Path:       "/basicdata/save",
		Method:     "POST",
		Parameters: []openapi.Parameter{{Name: "body", In: "body", Schema: &openapi.Schema{Type: "object"}}},
	}
	cmd := buildOperationCommand(op, "save", f)
	cmd.SetArgs([]string{"--body", "{not json"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid --body JSON") {
		t.Errorf("expected invalid --body JSON error, got %v", err)
	}
}

func TestExecuteOperationDELETE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"deleted": true}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	var out bytes.Buffer
	f.IOStreams.Out = &out

	op := &openapi.Operation{Summary: "Delete", Path: "/scene/delete", Method: "DELETE"}
	cmd := buildOperationCommand(op, "delete", f)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("output missing deleted: %s", out.String())
	}
}

func TestExecuteOperationQueryTypes(t *testing.T) {
	var gotURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	var out bytes.Buffer
	f.IOStreams.Out = &out

	op := &openapi.Operation{
		Summary: "Search",
		Path:    "/scene/search",
		Method:  "GET",
		Parameters: []openapi.Parameter{
			{Name: "active", In: "query", Type: "boolean"},
			{Name: "ids", In: "query", Type: "array"},
		},
	}
	cmd := buildOperationCommand(op, "search", f)
	cmd.SetArgs([]string{"--active", "true", "--ids", "a,b"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(gotURL, "active=true") {
		t.Errorf("missing active=true: %s", gotURL)
	}
	if !strings.Contains(gotURL, "ids=a") || !strings.Contains(gotURL, "ids=b") {
		t.Errorf("missing array ids: %s", gotURL)
	}
}

func TestCommandNameGeneration(t *testing.T) {
	tests := []struct {
		path   string
		prefix string
		want   string
	}{
		{"/scene/list", "scene", "list"},
		{"/scene/get", "scene", "get"},
		{"/basicdata/record/listNew", "basicdata", "record-listnew"},
		{"/factory/http/list", "factory", "http-list"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := openapi.CommandName(tt.path, tt.prefix)
			if got != tt.want {
				t.Errorf("CommandName(%q, %q) = %q, want %q", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}
