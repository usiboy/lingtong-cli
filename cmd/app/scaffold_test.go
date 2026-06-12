// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
)

func assertScaffoldOutputValidates(t *testing.T, data []byte) {
	t.Helper()

	if err := ValidateRawJSON(data); err != nil {
		t.Fatalf("scaffold output failed raw validation: %v", err)
	}

	var appExport AppExport
	if err := json.Unmarshal(data, &appExport); err != nil {
		t.Fatalf("failed to unmarshal scaffold output: %v", err)
	}

	if errors := validateAppExport(&appExport); len(errors) > 0 {
		t.Fatalf("scaffold output failed app validation: %v", errors)
	}
}

func TestNewCmdAppScaffold_MissingName(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppScaffold(f)
	cmd.SetArgs([]string{"--source", "src", "--target", "tgt"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error about name, got: %v", err)
	}
}

func TestNewCmdAppScaffold_MissingSource(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppScaffold(f)
	cmd.SetArgs([]string{"--name", "my-app", "--target", "tgt"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --source")
	}
	if !strings.Contains(err.Error(), "source") {
		t.Errorf("expected error about source, got: %v", err)
	}
}

func TestNewCmdAppScaffold_MissingTarget(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppScaffold(f)
	cmd.SetArgs([]string{"--name", "my-app", "--source", "src"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --target")
	}
	if !strings.Contains(err.Error(), "target") {
		t.Errorf("expected error about target, got: %v", err)
	}
}

func TestNewCmdAppScaffold_MissingOutputWithoutDryRun(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppScaffold(f)
	cmd.SetArgs([]string{"--name", "my-app", "--source", "src", "--target", "tgt"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --output without --dry-run")
	}
	if !strings.Contains(err.Error(), "output") {
		t.Errorf("expected error about output, got: %v", err)
	}
}

func TestNewCmdAppScaffold_DryRun(t *testing.T) {
	var out strings.Builder
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://example.com",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    &out,
			ErrOut: io.Discard,
		},
	}

	cmd := newCmdAppScaffold(f)
	cmd.SetArgs([]string{"--name", "my-app", "--source", "src-conn", "--target", "tgt-conn", "--dry-run"})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	assertScaffoldOutputValidates(t, []byte(output))

	if result["appName"] != "my-app" {
		t.Errorf("expected appName 'my-app', got: %v", result["appName"])
	}

	connectors, ok := result["appConnectors"].([]interface{})
	if !ok || len(connectors) != 2 {
		t.Fatalf("expected 2 appConnectors, got: %v", result["appConnectors"])
	}
	if connectors[0] != "src-conn" || connectors[1] != "tgt-conn" {
		t.Errorf("expected connectors [src-conn, tgt-conn], got: %v", connectors)
	}
}

func TestNewCmdAppScaffold_Success(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "app.json")

	var out strings.Builder
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://example.com",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    &out,
			ErrOut: io.Discard,
		},
	}

	cmd := newCmdAppScaffold(f)
	cmd.SetArgs([]string{"--name", "my-app", "--source", "src-conn", "--target", "tgt-conn", "--output", outputFile})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify success message
	outputStr := out.String()
	if !strings.Contains(outputStr, "Generated application scaffold") {
		t.Errorf("expected success message, got: %s", outputStr)
	}

	// Verify file was created and contains valid JSON
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output file contains invalid JSON: %v", err)
	}
	assertScaffoldOutputValidates(t, data)

	if result["appName"] != "my-app" {
		t.Errorf("expected appName 'my-app', got: %v", result["appName"])
	}

	// Verify scaffold emits a default validator-compatible scene.
	scenes, ok := result["scenes"].([]interface{})
	if !ok || len(scenes) != 1 {
		t.Fatalf("expected one default scene, got: %v", result["scenes"])
	}
	scene, ok := scenes[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected scene to be an object, got: %v", scenes[0])
	}
	if scene["name"] != "Default Sync" {
		t.Errorf("expected default scene name, got: %v", scene["name"])
	}
	source, ok := scene["connectorSource"].(map[string]interface{})
	if !ok || source["name"] != "src-conn" {
		t.Errorf("expected connectorSource.name src-conn, got: %v", scene["connectorSource"])
	}
	target, ok := scene["connectorTarget"].(map[string]interface{})
	if !ok || target["name"] != "tgt-conn" {
		t.Errorf("expected connectorTarget.name tgt-conn, got: %v", scene["connectorTarget"])
	}
	if basicDatas, ok := result["basicDatas"].([]interface{}); !ok || len(basicDatas) != 0 {
		t.Errorf("expected empty basicDatas array, got: %v", result["basicDatas"])
	}
}

func TestNewCmdAppScaffold_TemplateOutputValidates(t *testing.T) {
	var out strings.Builder
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://example.com",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    &out,
			ErrOut: io.Discard,
		},
	}

	cmd := newCmdAppScaffold(f)
	cmd.SetArgs([]string{"--name", "kuaimai-kingdee", "--template", "kuaimai-kingdee", "--dry-run"})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []byte(out.String())
	assertScaffoldOutputValidates(t, data)

	var appExport AppExport
	if err := json.Unmarshal(data, &appExport); err != nil {
		t.Fatalf("failed to unmarshal template output: %v", err)
	}
	if len(appExport.Scenes) != 7 {
		t.Fatalf("expected kuaimai-kingdee template to contain 7 scenes, got %d", len(appExport.Scenes))
	}
	for i, scene := range appExport.Scenes {
		if sourceConnectorName(scene) == "" || targetConnectorName(scene) == "" {
			t.Fatalf("template scene %d missing connector refs: %+v", i, scene)
		}
	}
}

func TestNewCmdAppScaffold_DryRunOutputStructure(t *testing.T) {
	tests := []struct {
		name           string
		appName        string
		source         string
		target         string
		wantAppName    string
		wantConnectors []string
	}{
		{
			name:           "basic scaffold",
			appName:        "test-app",
			source:         "connector-a",
			target:         "connector-b",
			wantAppName:    "test-app",
			wantConnectors: []string{"connector-a", "connector-b"},
		},
		{
			name:           "different connectors",
			appName:        "another-app",
			source:         "salesforce",
			target:         "hubspot",
			wantAppName:    "another-app",
			wantConnectors: []string{"salesforce", "hubspot"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			f := &cmdutil.Factory{
				Config: &config.Config{
					Host:  "http://example.com",
					Token: "test-token",
				},
				IOStreams: &output.IOStreams{
					In:     io.NopCloser(strings.NewReader("")),
					Out:    &out,
					ErrOut: io.Discard,
				},
			}

			cmd := newCmdAppScaffold(f)
			cmd.SetArgs([]string{
				"--name", tt.appName,
				"--source", tt.source,
				"--target", tt.target,
				"--dry-run",
			})
			cmd.SetOut(&out)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal([]byte(out.String()), &result); err != nil {
				t.Fatalf("output is not valid JSON: %v", err)
			}

			if result["appName"] != tt.wantAppName {
				t.Errorf("expected appName %q, got %q", tt.wantAppName, result["appName"])
			}

			connectors, ok := result["appConnectors"].([]interface{})
			if !ok {
				t.Fatalf("appConnectors is not an array")
			}
			if len(connectors) != len(tt.wantConnectors) {
				t.Fatalf("expected %d connectors, got %d", len(tt.wantConnectors), len(connectors))
			}
			for i, c := range tt.wantConnectors {
				if connectors[i] != c {
					t.Errorf("connector[%d]: expected %q, got %q", i, c, connectors[i])
				}
			}

			// Verify all required fields exist
			requiredFields := []string{"appName", "appConnectors", "basicDatas", "scenes", "workflows", "workflowConnectors"}
			for _, field := range requiredFields {
				if _, exists := result[field]; !exists {
					t.Errorf("missing required field: %s", field)
				}
			}
		})
	}
}
