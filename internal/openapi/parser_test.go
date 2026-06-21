// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

// findSpecPath locates the OpenAPI spec file for testing.
// It searches common locations relative to the test file.
func findSpecPath(t *testing.T) string {
	t.Helper()

	// Try relative to the module root
	candidates := []string{
		"../../../lingtong-skill/lingtong-api/lingtong-app-web_OpenAPI.json",
		"../../../../lingtong-skill/lingtong-api/lingtong-app-web_OpenAPI.json",
	}

	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	t.Skip("OpenAPI spec file not found, skipping integration test")
	return ""
}

func TestParseBytes(t *testing.T) {
	specJSON := `{
		"swagger": "2.0",
		"info": {"title": "Test API", "version": "1.0"},
		"paths": {
			"/scene/list": {
				"get": {
					"summary": "List scenes",
					"tags": ["Scene"],
					"parameters": [
						{"name": "pageNum", "in": "query", "type": "integer", "required": true},
						{"name": "pageSize", "in": "query", "type": "integer", "required": false}
					]
				}
			},
			"/scene/get": {
				"get": {
					"summary": "Get scene",
					"tags": ["Scene"],
					"parameters": [
						{"name": "sceneId", "in": "query", "type": "integer", "required": true}
					]
				}
			},
			"/basicdata/save": {
				"post": {
					"summary": "Save basic data",
					"tags": ["BasicData"],
					"parameters": [
						{"name": "body", "in": "body", "required": true, "schema": {"type": "object"}}
					]
				}
			}
		},
		"tags": [
			{"name": "Scene", "description": "Scene management"},
			{"name": "BasicData", "description": "Basic data operations"}
		]
	}`

	spec, err := ParseBytes([]byte(specJSON))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	// Check basic structure
	if spec.Info.Title != "Test API" {
		t.Errorf("Info.Title = %q, want %q", spec.Info.Title, "Test API")
	}

	// Check paths
	if len(spec.Paths) != 3 {
		t.Errorf("len(Paths) = %d, want 3", len(spec.Paths))
	}

	// Check operation was populated with path/method
	sceneList := spec.Paths["/scene/list"].Get
	if sceneList == nil {
		t.Fatal("scene list operation is nil")
	}
	if sceneList.Path != "/scene/list" {
		t.Errorf("sceneList.Path = %q, want %q", sceneList.Path, "/scene/list")
	}
	if sceneList.Method != "GET" {
		t.Errorf("sceneList.Method = %q, want %q", sceneList.Method, "GET")
	}
}

func TestGroupByPrefix(t *testing.T) {
	specJSON := `{
		"swagger": "2.0",
		"info": {"title": "Test", "version": "1.0"},
		"paths": {
			"/scene/list": {"get": {"summary": "List scenes"}},
			"/scene/get": {"get": {"summary": "Get scene"}},
			"/scene/create": {"post": {"summary": "Create scene"}},
			"/basicdata/save": {"post": {"summary": "Save data"}},
			"/basicdata/list": {"get": {"summary": "List data"}},
			"/factory/http/list": {"get": {"summary": "List HTTP"}}
		}
	}`

	spec, err := ParseBytes([]byte(specJSON))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	groups := spec.GroupByPrefix()

	// Check scene group
	if len(groups["scene"]) != 3 {
		t.Errorf("len(groups[scene]) = %d, want 3", len(groups["scene"]))
	}

	// Check basicdata group
	if len(groups["basicdata"]) != 2 {
		t.Errorf("len(groups[basicdata]) = %d, want 2", len(groups["basicdata"]))
	}

	// Check factory group
	if len(groups["factory"]) != 1 {
		t.Errorf("len(groups[factory]) = %d, want 1", len(groups["factory"]))
	}
}

func TestCommandName(t *testing.T) {
	tests := []struct {
		path   string
		prefix string
		want   string
	}{
		{"/scene/list", "scene", "list"},
		{"/scene/get", "scene", "get"},
		{"/basicdata/record/listNew", "basicdata", "record-listnew"},
		{"/factory/http/list", "factory", "http-list"},
		{"/scene", "scene", "root"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := CommandName(tt.path, tt.prefix)
			if got != tt.want {
				t.Errorf("CommandName(%q, %q) = %q, want %q", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestOperationHelpers(t *testing.T) {
	op := &Operation{
		Parameters: []Parameter{
			{Name: "pageNum", In: "query", Required: true, Type: "integer"},
			{Name: "pageSize", In: "query", Required: false, Type: "integer"},
			{Name: "body", In: "body", Required: true},
		},
	}

	// Test QueryParams
	qp := op.QueryParams()
	if len(qp) != 2 {
		t.Errorf("len(QueryParams) = %d, want 2", len(qp))
	}

	// Test BodyParams
	bp := op.BodyParams()
	if len(bp) != 1 {
		t.Errorf("len(BodyParams) = %d, want 1", len(bp))
	}

	// Test HasRequiredParams
	if !op.HasRequiredParams() {
		t.Error("HasRequiredParams() = false, want true")
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/scene/list", "scene"},
		{"/basicdata/save", "basicdata"},
		{"/factory/http/list", "factory"},
		{"/", "root"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractPrefix(tt.path)
			if got != tt.want {
				t.Errorf("extractPrefix(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestParseRealSpec tests parsing the actual OpenAPI spec file.
// This test is skipped if the spec file is not found.
func TestParseRealSpec(t *testing.T) {
	specPath := findSpecPath(t)

	spec, err := Parse(specPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify we got a reasonable number of paths
	if len(spec.Paths) < 100 {
		t.Errorf("len(Paths) = %d, expected at least 100", len(spec.Paths))
	}

	// Verify grouping works
	groups := spec.GroupByPrefix()
	if len(groups) < 10 {
		t.Errorf("len(groups) = %d, expected at least 10", len(groups))
	}

	// Verify specific expected groups exist
	expectedGroups := []string{"scene", "basicdata", "workflow", "factory"}
	for _, g := range expectedGroups {
		if _, ok := groups[g]; !ok {
			t.Errorf("expected group %q not found", g)
		}
	}
}

func TestCommandNameStripsBracesAndWildcards(t *testing.T) {
	tests := []struct {
		path, prefix, want string
	}{
		{"/meta/model/{connector}/get", "meta", "model-connector-get"},
		{"/noco/datasource/app/{datasourceId}/**", "noco", "datasource-app-datasourceid"},
		{"/test/flowSource/{id}", "test", "flowsource-id"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := CommandName(tt.path, tt.prefix); got != tt.want {
				t.Errorf("CommandName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestPathParams(t *testing.T) {
	op := &Operation{Parameters: []Parameter{
		{Name: "connector", In: "path", Required: true},
		{Name: "page", In: "query"},
		{Name: "body", In: "body"},
	}}
	pp := op.PathParams()
	if len(pp) != 1 || pp[0].Name != "connector" {
		t.Errorf("PathParams() = %+v, want single 'connector'", pp)
	}
}

func TestEmbeddedSpec(t *testing.T) {
	if !HasEmbeddedSpec() {
		t.Fatal("expected an embedded spec to be compiled in")
	}
	spec, err := EmbeddedSpec()
	if err != nil {
		t.Fatalf("EmbeddedSpec() error = %v", err)
	}
	if len(spec.Paths) < 100 {
		t.Errorf("embedded spec has too few paths: %d", len(spec.Paths))
	}
	// Every operation should have Path/Method populated.
	groups := spec.GroupByPrefix()
	if len(groups) < 10 {
		t.Errorf("expected many prefix groups, got %d", len(groups))
	}
}
