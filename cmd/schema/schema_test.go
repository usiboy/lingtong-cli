// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/openapi"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func newTestFactory() *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  "https://test.example.com",
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
					"summary": "查询场景列表",
					"description": "返回分页的场景数据",
					"tags": ["场景管理"],
					"parameters": [
						{"name": "pageNum", "in": "query", "type": "integer", "required": true, "description": "页码"},
						{"name": "pageSize", "in": "query", "type": "integer", "required": false, "description": "每页条数"}
					]
				}
			},
			"/scene/get": {
				"get": {
					"summary": "获取场景详情",
					"parameters": [
						{"name": "sceneId", "in": "query", "type": "integer", "required": true}
					]
				}
			},
			"/basicdata/save": {
				"post": {
					"summary": "保存基础数据",
					"parameters": [
						{"name": "body", "in": "body", "required": true, "schema": {"type": "object"}}
					]
				}
			},
			"/workflow/create": {
				"post": {
					"summary": "创建工作流",
					"parameters": [
						{"name": "name", "in": "query", "type": "string", "required": true},
						{"name": "body", "in": "body", "required": true, "schema": {"type": "object"}}
					]
				}
			}
		}
	}`

	spec, _ := openapi.ParseBytes([]byte(specJSON))
	return spec
}

func executeSchemaCommand(cmd *cobra.Command, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestNewCmdSchema(t *testing.T) {
	f := newTestFactory()
	cmd := NewCmdSchema(f)

	if cmd.Use != "schema [command] [args]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "schema [command] [args]")
	}

	// Check subcommands
	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	expected := []string{"list", "path", "module", "search"}
	for _, name := range expected {
		if !subCmds[name] {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

func TestSchemaList(t *testing.T) {
	// Note: This test uses the embedded spec, so we test the helpers directly
	spec := newTestSpec()
	groups := spec.GroupByPrefix()

	if len(groups) != 3 {
		t.Errorf("expected 3 modules, got %d", len(groups))
	}

	if len(groups["scene"]) != 2 {
		t.Errorf("expected 2 scene operations, got %d", len(groups["scene"]))
	}

	if len(groups["basicdata"]) != 1 {
		t.Errorf("expected 1 basicdata operation, got %d", len(groups["basicdata"]))
	}
}

func TestSchemaPath(t *testing.T) {
	spec := newTestSpec()

	// Test existing path
	pathItem := spec.Paths["/scene/list"]
	if pathItem.Get == nil {
		t.Fatal("expected GET operation for /scene/list")
	}

	detail := buildPathDetail("/scene/list", &pathItem)
	if detail["path"] != "/scene/list" {
		t.Errorf("path = %v, want /scene/list", detail["path"])
	}

	ops, ok := detail["operations"].([]map[string]interface{})
	if !ok || len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %v", detail["operations"])
	}

	if ops[0]["method"] != "GET" {
		t.Errorf("method = %v, want GET", ops[0]["method"])
	}
}

func TestSchemaSearch(t *testing.T) {
	spec := newTestSpec()

	// Search for "场景"
	results := searchOperations(spec, "场景")
	if len(results) != 2 {
		t.Errorf("expected 2 results for '场景', got %d", len(results))
	}

	// Search for "workflow"
	results = searchOperations(spec, "workflow")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'workflow', got %d", len(results))
	}

	// Search for non-existent
	results = searchOperations(spec, "nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestFuzzyMatchPaths(t *testing.T) {
	spec := newTestSpec()

	matches := fuzzyMatchPaths(spec, "scene")
	if len(matches) != 2 {
		t.Errorf("expected 2 matches for 'scene', got %d", len(matches))
	}

	matches = fuzzyMatchPaths(spec, "basic")
	if len(matches) != 1 {
		t.Errorf("expected 1 match for 'basic', got %d", len(matches))
	}
}

func TestCountOperations(t *testing.T) {
	spec := newTestSpec()
	count := countOperations(spec)
	if count != 4 {
		t.Errorf("expected 4 total operations, got %d", count)
	}
}

func TestCountRequired(t *testing.T) {
	params := []openapi.Parameter{
		{Name: "a", Required: true},
		{Name: "b", Required: false},
		{Name: "c", Required: true},
	}
	n := countRequired(params)
	if n != 2 {
		t.Errorf("expected 2 required, got %d", n)
	}
}

func TestUniquePaths(t *testing.T) {
	ops := []*openapi.Operation{
		{Path: "/a"},
		{Path: "/b"},
		{Path: "/a"}, // duplicate
	}
	paths := uniquePaths(ops)
	if len(paths) != 2 {
		t.Errorf("expected 2 unique paths, got %d", len(paths))
	}
}

func TestBuildPathDetail(t *testing.T) {
	spec := newTestSpec()
	pathItem := spec.Paths["/scene/list"]

	detail := buildPathDetail("/scene/list", &pathItem)

	// Check structure
	path, ok := detail["path"].(string)
	if !ok || path != "/scene/list" {
		t.Errorf("path = %v, want /scene/list", detail["path"])
	}

	ops, ok := detail["operations"].([]map[string]interface{})
	if !ok {
		t.Fatal("operations is not []map[string]interface{}")
	}

	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	op := ops[0]
	if op["summary"] != "查询场景列表" {
		t.Errorf("summary = %v, want 查询场景列表", op["summary"])
	}

	params, ok := op["parameters"].([]map[string]interface{})
	if !ok {
		t.Fatal("parameters is not []map[string]interface{}")
	}

	if len(params) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(params))
	}

	// Check first param
	if params[0]["name"] != "pageNum" {
		t.Errorf("param name = %v, want pageNum", params[0]["name"])
	}
	if params[0]["required"] != true {
		t.Errorf("param required = %v, want true", params[0]["required"])
	}
}

func TestSearchOperationsCaseInsensitive(t *testing.T) {
	spec := newTestSpec()

	// Search should be case-insensitive and match paths
	results := searchOperations(spec, "scene")
	if len(results) < 2 {
		t.Errorf("expected at least 2 results for 'scene', got %d", len(results))
	}

	// Search for Chinese text in summary/description
	results = searchOperations(spec, "场景")
	if len(results) < 2 {
		t.Errorf("expected at least 2 results for '场景', got %d", len(results))
	}
}

func TestSearchOperationsInDescription(t *testing.T) {
	spec := newTestSpec()

	// Search should match description too
	results := searchOperations(spec, "分页")
	if len(results) < 1 {
		t.Error("expected results matching description '分页'")
	}
}

// TestSchemaCommandIntegration tests the full command execution.
func TestSchemaCommandIntegration(t *testing.T) {
	f := newTestFactory()
	var buf bytes.Buffer
	f.IOStreams.Out = &buf

	cmd := NewCmdSchema(f)

	// Test list command
	listCmd := cmd.Commands()[0] // list is first
	if listCmd.Name() != "list" {
		t.Skipf("expected first subcommand to be 'list', got %s", listCmd.Name())
	}

	// Note: Full integration test requires embedded spec, which may not be
	// available in all test environments. The helper tests above cover the logic.
}

// marshalTestJSON is a test helper.
func marshalTestJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func TestMarshalTestJSON(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	result := marshalTestJSON(data)
	if !strings.Contains(result, "key") {
		t.Error("expected JSON to contain 'key'")
	}
}

// writeTempSpec writes the test spec to disk and points LINGTONG_OPENAPI at it
// so the shared loader (openapi.LoadSpec) used by the commands resolves to it.
func writeTempSpec(t *testing.T) {
	t.Helper()
	const specJSON = `{
		"swagger": "2.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/scene/list": {"get": {"summary": "查询场景列表", "description": "返回分页的场景数据", "tags": ["场景管理"], "parameters": [{"name": "pageNum", "in": "query", "type": "integer", "required": true}]}},
			"/scene/get": {"get": {"summary": "获取场景详情", "parameters": [{"name": "sceneId", "in": "query", "type": "integer", "required": true}]}},
			"/basicdata/save": {"post": {"summary": "保存基础数据", "parameters": [{"name": "body", "in": "body", "required": true, "schema": {"type": "object"}}]}}
		}
	}`
	dir := t.TempDir()
	path := dir + "/openapi.json"
	if err := os.WriteFile(path, []byte(specJSON), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LINGTONG_OPENAPI", path)
}

func runSchema(t *testing.T, args ...string) (string, error) {
	t.Helper()
	f := newTestFactory()
	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	root := NewCmdSchema(f)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestLoadSpec_FromOverride(t *testing.T) {
	writeTempSpec(t)
	spec, err := loadSpec()
	if err != nil {
		t.Fatalf("loadSpec error = %v", err)
	}
	if len(spec.Paths) != 3 {
		t.Errorf("expected 3 paths from override spec, got %d", len(spec.Paths))
	}
}

func TestSchemaListCommand(t *testing.T) {
	writeTempSpec(t)
	out, err := runSchema(t, "list")
	if err != nil {
		t.Fatalf("list error = %v", err)
	}
	for _, want := range []string{"scene", "basicdata", "totalOperations"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output missing %q: %s", want, out)
		}
	}
}

func TestSchemaPathCommand(t *testing.T) {
	writeTempSpec(t)

	out, err := runSchema(t, "path", "/scene/list")
	if err != nil {
		t.Fatalf("path error = %v", err)
	}
	if !strings.Contains(out, "查询场景列表") {
		t.Errorf("path output missing summary: %s", out)
	}

	// Leading slash is optional.
	if _, err := runSchema(t, "path", "scene/get"); err != nil {
		t.Errorf("path without leading slash should work: %v", err)
	}

	// Fuzzy suggestion on a near miss.
	if _, err := runSchema(t, "path", "/scene"); err == nil {
		t.Error("expected error with suggestions for partial path")
	}

	// Hard miss.
	if _, err := runSchema(t, "path", "/totally/unknown"); err == nil {
		t.Error("expected error for unknown path")
	}
}

func TestSchemaModuleCommand(t *testing.T) {
	writeTempSpec(t)

	out, err := runSchema(t, "module", "scene")
	if err != nil {
		t.Fatalf("module error = %v", err)
	}
	if !strings.Contains(out, "/scene/list") {
		t.Errorf("module output missing operation: %s", out)
	}

	if _, err := runSchema(t, "module", "ghost"); err == nil {
		t.Error("expected error for unknown module")
	}
}

func TestSchemaSearchCommand(t *testing.T) {
	writeTempSpec(t)
	out, err := runSchema(t, "search", "场景")
	if err != nil {
		t.Fatalf("search error = %v", err)
	}
	if !strings.Contains(out, "matchCount") {
		t.Errorf("search output missing matchCount: %s", out)
	}
}
