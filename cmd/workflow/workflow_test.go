// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package workflow

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

func TestValidateDSL(t *testing.T) {
	tests := []struct {
		name        string
		dsl         map[string]interface{}
		strict      bool
		expectValid bool
		errorCount  int
		warnCount   int
	}{
		{
			name: "valid simple workflow",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{
					map[string]interface{}{"source": "start", "target": "end"},
				},
			},
			strict:      false,
			expectValid: true,
			errorCount:  0,
			warnCount:   0,
		},
		{
			name: "missing nodes field",
			dsl: map[string]interface{}{
				"edges": []interface{}{},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1,
			warnCount:   0,
		},
		{
			name: "missing edges field",
			dsl: map[string]interface{}{
				"nodes": []interface{}{},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1,
			warnCount:   0,
		},
		{
			name: "no start node",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "node1", "type": "w_connector"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{
					map[string]interface{}{"source": "node1", "target": "end"},
				},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1,
			warnCount:   0,
		},
		{
			name: "no end node",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "node1", "type": "w_connector"},
				},
				"edges": []interface{}{
					map[string]interface{}{"source": "start", "target": "node1"},
				},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1,
			warnCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateDSL(tt.dsl, tt.strict)
			valid, ok := result["valid"].(bool)
			if !ok {
				t.Fatal("result missing 'valid' field")
			}
			if valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, valid)
			}
			errors, _ := result["errors"].([]string)
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
			warnings, _ := result["warnings"].([]string)
			if len(warnings) != tt.warnCount {
				t.Errorf("expected %d warnings, got %d: %v", tt.warnCount, len(warnings), warnings)
			}
		})
	}
}

func TestGetWorkflowTemplate(t *testing.T) {
	templates := []string{"simple", "connector", "order_sync", "approval", "data_pipeline", "api_wrapper"}
	for _, tmpl := range templates {
		t.Run(tmpl+" template", func(t *testing.T) {
			dsl := getWorkflowTemplate(tmpl, "Test Workflow", "test")
			if dsl == "" {
				t.Error("expected non-empty DSL")
				return
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(dsl), &parsed); err != nil {
				t.Errorf("invalid JSON: %v", err)
				return
			}
			if _, hasNodes := parsed["nodes"]; !hasNodes {
				t.Error("DSL missing 'nodes' field")
			}
			if _, hasEdges := parsed["edges"]; !hasEdges {
				t.Error("DSL missing 'edges' field")
			}
		})
	}
	t.Run("unknown template", func(t *testing.T) {
		dsl := getWorkflowTemplate("nonexistent", "Test", "test")
		if dsl != "" {
			t.Error("expected empty DSL for unknown template")
		}
	})
}

func TestGetTemplateList(t *testing.T) {
	templates := getTemplateList()
	if len(templates) != 6 {
		t.Errorf("expected 6 templates, got %d", len(templates))
	}
	expectedNames := map[string]bool{
		"simple": false, "connector": false, "order_sync": false,
		"approval": false, "data_pipeline": false, "api_wrapper": false,
	}
	for _, tmpl := range templates {
		name, ok := tmpl["name"].(string)
		if !ok {
			t.Error("template missing 'name' field")
			continue
		}
		if _, exists := expectedNames[name]; !exists {
			t.Errorf("unexpected template name: %s", name)
		}
		expectedNames[name] = true
		if _, hasDesc := tmpl["description"]; !hasDesc {
			t.Errorf("template %s missing 'description' field", name)
		}
	}
	for name, found := range expectedNames {
		if !found {
			t.Errorf("expected template '%s' not found", name)
		}
	}
}

func TestAnalyzeDeps(t *testing.T) {
	dsl := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{"id": "start", "type": "w_start"},
			map[string]interface{}{
				"id": "conn1", "type": "w_connector",
				"data": map[string]interface{}{
					"connector": "kmerp", "interfaceModelId": 73, "authAccountId": 296,
				},
			},
			map[string]interface{}{"id": "end", "type": "w_end"},
		},
		"edges": []interface{}{},
	}
	result := analyzeDeps(dsl)
	connCount, _ := result["connectorCount"].(int)
	if connCount != 1 {
		t.Errorf("expected 1 connector, got %d", connCount)
	}
}

func TestNewCmdWorkflowList_MissingAppId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowList(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --app-id")
	}
	if !strings.Contains(err.Error(), "app-id") {
		t.Errorf("expected error about app-id, got: %v", err)
	}
}

func TestNewCmdWorkflowList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{
					{"id": 1, "name": "Workflow 1"},
					{"id": 2, "name": "Workflow 2"},
				},
				"total": 2,
			},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--app-id", "165"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowInfo_MissingWorkflowId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowInfo(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --workflow-id")
	}
}

func TestNewCmdWorkflowInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"id": 100, "name": "Test Workflow"},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowPublish_MissingWorkflowId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowPublish(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --workflow-id")
	}
}

func TestNewCmdWorkflowPublish_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"success": true, "version": "v1.0.0"},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowPublish(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--version", "v1.0.0"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowVersions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{
					{"version": "v1.0.0", "memo": "Initial", "publishTime": "2026-04-07", "status": "published"},
				},
			},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowVersions(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowExecute_MissingWorkflowId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowExecute(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --workflow-id")
	}
}

func TestNewCmdWorkflowExecute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"receiptId": "test-receipt-123"},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowExecute(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowLogs_MissingReceiptId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowLogs(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --receipt-id")
	}
}

func TestNewCmdWorkflowLogs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"logs": []map[string]interface{}{{"level": "info", "message": "Started"}}},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowLogs(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--receipt-id", "abc123"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowAPITest_MissingAppTag(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowAPITest(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --app-tag")
	}
}

func TestNewCmdWorkflowAPITest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"success": true},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowAPITest(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--app-tag", "test-app", "--params", `{"key":"value"}`})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowValidate_MissingInput(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowValidate(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing input")
	}
	if !strings.Contains(err.Error(), "either --workflow-id or --dsl-file is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewCmdWorkflowDependencyList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"content": `{"nodes": [{"id": "start", "type": "w_start"}, {"id": "conn1", "type": "w_connector", "data": {"connector": "kmerp"}}, {"id": "end", "type": "w_end"}], "edges": []}`,
			},
		})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowDependencyList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowList_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Internal server error"})
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--app-id", "165"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdWorkflowInfo_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdWorkflowPublish_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Publish failed"}`))
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowPublish(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdWorkflowExecute_CreateFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid workflow ID"}`))
	}))
	defer server.Close()
	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowExecute(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "999"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for failed execution create")
	}
	if !strings.Contains(err.Error(), "failed to create execution task") {
		t.Errorf("expected execution task error, got: %v", err)
	}
}

func TestValidateDSL_ExtendedEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		dsl         map[string]interface{}
		strict      bool
		expectValid bool
		errorCount  int
	}{
		{
			name: "empty DSL", dsl: map[string]interface{}{}, strict: false, expectValid: false, errorCount: 2,
		},
		{
			name: "multiple start nodes",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start1", "type": "w_start"},
					map[string]interface{}{"id": "start2", "type": "w_start"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{map[string]interface{}{"source": "start1", "target": "end"}},
			},
			strict: false, expectValid: false, errorCount: 1,
		},
		{
			name: "invalid node type",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "invalid", "type": "w_invalid_type"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{map[string]interface{}{"source": "start", "target": "end"}},
			},
			strict: false, expectValid: false, errorCount: 1,
		},
		{
			name: "edge references non-existent source",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{map[string]interface{}{"source": "nonexistent", "target": "end"}},
			},
			strict: false, expectValid: false, errorCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateDSL(tt.dsl, tt.strict)
			valid, ok := result["valid"].(bool)
			if !ok {
				t.Fatal("result missing 'valid' field")
			}
			if valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, valid)
			}
			errors, _ := result["errors"].([]string)
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestAnalyzeDeps_ExtendedCases(t *testing.T) {
	tests := []struct {
		name           string
		dsl            map[string]interface{}
		expectConns    int
		expectScripts int
	}{
		{
			name: "multiple connectors and scripts",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "conn1", "type": "w_connector", "data": map[string]interface{}{"connector": "erp", "interfaceModelId": 100, "authAccountId": 200}},
					map[string]interface{}{"id": "script1", "type": "w_script", "data": map[string]interface{}{"scriptConfig": map[string]interface{}{"language": "javascript"}}},
					map[string]interface{}{"id": "conn2", "type": "w_modePipe", "data": map[string]interface{}{"connector": "crm", "interfaceModelId": 300, "authAccountId": 400}},
				},
				"edges": []interface{}{},
			},
			expectConns: 2, expectScripts: 1,
		},
		{
			name: "connector without data",
			dsl: map[string]interface{}{
				"nodes": []interface{}{map[string]interface{}{"id": "conn", "type": "w_connector"}},
				"edges": []interface{}{},
			},
			expectConns: 0, expectScripts: 0,
		},
		{
			name: "missing nodes field",
			dsl: map[string]interface{}{"edges": []interface{}{}},
			expectConns: 0, expectScripts: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeDeps(tt.dsl)
			conns, _ := result["connectors"].([]map[string]interface{})
			if len(conns) != tt.expectConns {
				t.Errorf("expected %d connectors, got %d", tt.expectConns, len(conns))
			}
			scripts, _ := result["scripts"].([]map[string]interface{})
			if len(scripts) != tt.expectScripts {
				t.Errorf("expected %d scripts, got %d", tt.expectScripts, len(scripts))
			}
		})
	}
}

func TestGetWorkflowTemplate_AllTemplates(t *testing.T) {
	templates := []string{"simple", "connector", "order_sync", "approval", "data_pipeline", "api_wrapper"}
	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			dsl := getWorkflowTemplate(tmpl, "Test Workflow", "test")
			if dsl == "" {
				t.Errorf("template '%s' returned empty DSL", tmpl)
				return
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(dsl), &parsed); err != nil {
				t.Errorf("template '%s' has invalid JSON: %v", tmpl, err)
				return
			}
			if _, ok := parsed["nodes"]; !ok {
				t.Errorf("template '%s' missing nodes", tmpl)
			}
			if _, ok := parsed["edges"]; !ok {
				t.Errorf("template '%s' missing edges", tmpl)
			}
		})
	}
}

func TestGetWorkflowTemplate_Unknown(t *testing.T) {
	dsl := getWorkflowTemplate("nonexistent", "Test", "test")
	if dsl != "" {
		t.Error("expected empty DSL for unknown template")
	}
}

func TestGetTemplateList_Count(t *testing.T) {
	templates := getTemplateList()
	if len(templates) != 6 {
		t.Errorf("expected 6 templates, got %d", len(templates))
	}
}

func TestGetTemplateList_RequiredFields(t *testing.T) {
	templates := getTemplateList()
	for i, tmpl := range templates {
		if _, ok := tmpl["name"]; !ok {
			t.Errorf("template %d missing 'name'", i)
		}
		if _, ok := tmpl["description"]; !ok {
			t.Errorf("template %d missing 'description'", i)
		}
		if _, ok := tmpl["nodes"]; !ok {
			t.Errorf("template %d missing 'nodes'", i)
		}
	}
}

// ============================================================================
// CRUD Tests
// ============================================================================

func TestNewCmdWorkflowCreate_MissingName(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowCreate(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected name error, got: %v", err)
	}
}

func TestNewCmdWorkflowCreate_WithTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"id": 100, "name": "Test Workflow"},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--name", "Test Workflow", "--template", "simple"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowCreate_UnknownTemplate(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowCreate(f)
	cmd.SetArgs([]string{"--name", "Test", "--template", "unknown"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("expected unknown template error, got: %v", err)
	}
}

func TestNewCmdWorkflowCreate_NoDSL(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowCreate(f)
	cmd.SetArgs([]string{"--name", "Test"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing DSL source")
	}
	if !strings.Contains(err.Error(), "dsl-file") && !strings.Contains(err.Error(), "dsl-string") && !strings.Contains(err.Error(), "template") {
		t.Errorf("expected DSL source error, got: %v", err)
	}
}

func TestNewCmdWorkflowCreate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to create workflow"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--name", "Test", "--template", "simple"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdWorkflowUpdate_MissingWorkflowId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowUpdate(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --workflow-id")
	}
}

func TestNewCmdWorkflowUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"success": true},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowUpdate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--name", "Updated Name"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowUpdate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Update failed"}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowUpdate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--name", "Updated"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdWorkflowDelete_MissingWorkflowId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowDelete(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --workflow-id")
	}
}

func TestNewCmdWorkflowAPIEnable_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"success": true},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowAPIEnable(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowAPIDisable_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"success": true},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowAPIDisable(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================================
// Template Command Tests
// ============================================================================

func TestNewCmdWorkflowTemplateList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"result": []map[string]interface{}{}})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowTemplateList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowTemplateShow_MissingName(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowTemplateShow(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name")
	}
}

func TestNewCmdWorkflowTemplateShow_UnknownTemplate(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowTemplateShow(f)
	cmd.SetArgs([]string{"--name", "unknown"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestNewCmdWorkflowTemplateShow_Success(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowTemplateShow(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--name", "simple"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowTemplateUse_MissingParams(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowTemplateUse(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing params")
	}
}

// ============================================================================
// Version Rollback Tests
// ============================================================================

func TestNewCmdWorkflowVersionRollback_MissingParams(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowVersionRollback(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing params")
	}
}
func TestNewCmdWorkflowVersionRollback_VersionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty list so v2.0.0 won't be found
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowVersionRollback(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--version", "v2.0.0", "--dry-run"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for version not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestNewCmdWorkflowVersionRollback_DryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"snapshotList": []map[string]interface{}{
					{"id": 1, "version": "v1.0.0"},
				},
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowVersionRollback(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--version", "v1.0.0", "--dry-run"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================================
// Test Run Tests
// ============================================================================

func TestNewCmdWorkflowTestRun_MissingWorkflowId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowTestRun(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --workflow-id")
	}
}

func TestNewCmdWorkflowTestRun_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"receiptId": "test-123",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowTestRun(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--verbose"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount < 1 {
		t.Errorf("expected at least 1 API call, got %d", callCount)
	}
}

// ============================================================================
// Documentation Tests
// ============================================================================

func TestGenerateDoc_BasicWorkflow(t *testing.T) {
	workflow := map[string]interface{}{
		"name": "Test Workflow",
		"id":   float64(100),
		"env":  "test",
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "start",
				"type": "w_start",
				"data": map[string]interface{}{"title": "Start Node"},
			},
			map[string]interface{}{
				"id":   "conn",
				"type": "w_connector",
				"data": map[string]interface{}{
					"title":     "Connector Node",
					"connector": "kmerp",
				},
			},
			map[string]interface{}{
				"id":   "end",
				"type": "w_end",
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"source": "start", "target": "conn"},
			map[string]interface{}{"source": "conn", "target": "end"},
		},
	}

	doc := generateDoc(workflow, false, false)
	if doc == "" {
		t.Error("expected non-empty documentation")
	}
	if !strings.Contains(doc, "Test Workflow") {
		t.Error("documentation missing workflow name")
	}
	if !strings.Contains(doc, "Start Node") {
		t.Error("documentation missing node title")
	}
	if !strings.Contains(doc, "kmerp") {
		t.Error("documentation missing connector name")
	}
}

func TestGenerateDoc_WithAPI(t *testing.T) {
	workflow := map[string]interface{}{
		"name":    "API Workflow",
		"id":      float64(200),
		"openApi": float64(1),
		"nodes":   []interface{}{},
		"edges":   []interface{}{},
	}

	doc := generateDoc(workflow, true, false)
	if !strings.Contains(doc, "API access is enabled") {
		t.Error("documentation should show API is enabled")
	}
}

func TestGenerateDoc_APIDisabled(t *testing.T) {
	workflow := map[string]interface{}{
		"name":    "No API Workflow",
		"id":      float64(300),
		"openApi": float64(0),
		"nodes":   []interface{}{},
		"edges":   []interface{}{},
	}

	doc := generateDoc(workflow, true, false)
	if !strings.Contains(doc, "API access is not enabled") {
		t.Error("documentation should show API is disabled")
	}
}

func TestNewCmdWorkflowDocGenerate_MissingInput(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowDocGenerate(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing input")
	}
	if !strings.Contains(err.Error(), "workflow-id") && !strings.Contains(err.Error(), "dsl-file") {
		t.Errorf("expected input error, got: %v", err)
	}
}

// ============================================================================
// Extended Workflow Execute Tests
// ============================================================================

func TestNewCmdWorkflowExecute_WithParams(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"receiptId": "test-456",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowExecute(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--params", `{"key":"value"}`})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}


func TestNewCmdWorkflowExecute_InvalidParams(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdWorkflowExecute(f)
	cmd.SetArgs([]string{"--workflow-id", "100", "--wait", "--params", "invalid-json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid params")
	}
	// Error could be "invalid params JSON" or "failed to create execution task" depending on execution path
	if !strings.Contains(err.Error(), "invalid params JSON") && !strings.Contains(err.Error(), "failed to create") {
		t.Errorf("expected params or execution error, got: %v", err)
	}
}

// ============================================================================
// Extended Validate Tests
// ============================================================================

func TestNewCmdWorkflowValidate_WithWorkflowId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"content": `{"nodes": [{"id": "start", "type": "w_start"}, {"id": "end", "type": "w_end"}], "edges": []}`,
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowValidate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdWorkflowValidate_StrictMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"dsl": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{"id": "start", "type": "w_start"},
						map[string]interface{}{"id": "end", "type": "w_end"},
					},
					"edges": []interface{}{},
				},
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowValidate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100", "--strict"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================================
// Versions Table Output Tests
// ============================================================================

func TestNewCmdWorkflowVersions_TableFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{
					{"version": "v1.0.0", "memo": "Initial", "publishTime": "2026-04-07", "status": "published"},
				},
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdWorkflowVersions(f)
	cmd.Flags().String("format", "table", "Output format")
	cmd.SetArgs([]string{"--workflow-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
