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

func TestNewCmdAppDiff_MissingFileA(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppDiff(f)
	cmd.SetArgs([]string{"--file-b", "b.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --file-a")
	}
	if !strings.Contains(err.Error(), "file-a") {
		t.Errorf("expected error about file-a, got: %v", err)
	}
}

func TestNewCmdAppDiff_MissingFileB(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppDiff(f)
	cmd.SetArgs([]string{"--file-a", "a.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --file-b")
	}
	if !strings.Contains(err.Error(), "file-b") {
		t.Errorf("expected error about file-b, got: %v", err)
	}
}

func TestNewCmdAppDiff_FileNotFound(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppDiff(f)
	cmd.SetArgs([]string{"--file-a", "/nonexistent/a.json", "--file-b", "/nonexistent/b.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent files")
	}
	if !strings.Contains(err.Error(), "read file") {
		t.Errorf("expected error about reading file, got: %v", err)
	}
}

func TestNewCmdAppDiff_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	fileA := filepath.Join(tmpDir, "a.json")
	os.WriteFile(fileA, []byte("not valid json"), 0644)

	fileB := filepath.Join(tmpDir, "b.json")
	os.WriteFile(fileB, []byte(`{"appName":"test"}`), 0644)

	f := newTestFactory("http://example.com")
	cmd := newCmdAppDiff(f)
	cmd.SetArgs([]string{"--file-a", fileA, "--file-b", fileB})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected error about invalid JSON, got: %v", err)
	}
}

func TestNewCmdAppDiff_NoDifferences(t *testing.T) {
	tmpDir := t.TempDir()

	appData := map[string]interface{}{
		"appName":            "Test App",
		"appConnectors":      []string{"conn1", "conn2"},
		"basicDatas":         []interface{}{},
		"scenes":             []interface{}{},
		"workflows":          []interface{}{},
		"workflowConnectors": []interface{}{},
	}

	fileA := filepath.Join(tmpDir, "a.json")
	fileB := filepath.Join(tmpDir, "b.json")
	data, _ := json.Marshal(appData)
	os.WriteFile(fileA, data, 0644)
	os.WriteFile(fileB, data, 0644)

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

	cmd := newCmdAppDiff(f)
	cmd.SetArgs([]string{"--file-a", fileA, "--file-b", fileB})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No differences found") {
		t.Errorf("expected no differences message, got: %s", output)
	}
}

func TestNewCmdAppDiff_AddedScenes(t *testing.T) {
	tmpDir := t.TempDir()

	appA := map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"conn1"},
		"basicDatas":    []interface{}{},
		"scenes":        []interface{}{},
	}

	appB := map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"conn1"},
		"basicDatas":    []interface{}{},
		"scenes": []interface{}{
			map[string]interface{}{
				"name":            "New Scene",
				"connectorSource": map[string]interface{}{"name": "conn1"},
				"connectorTarget": map[string]interface{}{"name": "conn1"},
				"fieldMappings":   []interface{}{},
			},
		},
	}

	fileA := filepath.Join(tmpDir, "a.json")
	fileB := filepath.Join(tmpDir, "b.json")
	dataA, _ := json.Marshal(appA)
	dataB, _ := json.Marshal(appB)
	os.WriteFile(fileA, dataA, 0644)
	os.WriteFile(fileB, dataB, 0644)

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

	cmd := newCmdAppDiff(f)
	cmd.SetArgs([]string{"--file-a", fileA, "--file-b", fileB})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Added scenes") {
		t.Errorf("expected added scenes, got: %s", output)
	}
	if !strings.Contains(output, "New Scene") {
		t.Errorf("expected 'New Scene' in output, got: %s", output)
	}
}

func TestNewCmdAppDiff_RemovedConnectors(t *testing.T) {
	tmpDir := t.TempDir()

	appA := map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"conn1", "conn2", "conn3"},
		"basicDatas":    []interface{}{},
		"scenes":        []interface{}{},
	}

	appB := map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"conn1"},
		"basicDatas":    []interface{}{},
		"scenes":        []interface{}{},
	}

	fileA := filepath.Join(tmpDir, "a.json")
	fileB := filepath.Join(tmpDir, "b.json")
	dataA, _ := json.Marshal(appA)
	dataB, _ := json.Marshal(appB)
	os.WriteFile(fileA, dataA, 0644)
	os.WriteFile(fileB, dataB, 0644)

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

	cmd := newCmdAppDiff(f)
	cmd.SetArgs([]string{"--file-a", fileA, "--file-b", fileB})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Removed connectors") {
		t.Errorf("expected removed connectors, got: %s", output)
	}
	if !strings.Contains(output, "conn2") || !strings.Contains(output, "conn3") {
		t.Errorf("expected conn2 and conn3 in output, got: %s", output)
	}
}

func TestDiffApps_Comprehensive(t *testing.T) {
	tests := []struct {
		name                 string
		appA                 AppExport
		appB                 AppExport
		wantAddedScenes      []string
		wantRemovedScenes    []string
		wantAddedConns       []string
		wantRemovedConns     []string
		wantAddedBasicData   int
		wantRemovedBasicData int
	}{
		{
			name: "identical apps",
			appA: AppExport{
				AppName:       "App",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes:        []SceneExport{{Name: "s1"}},
			},
			appB: AppExport{
				AppName:       "App",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes:        []SceneExport{{Name: "s1"}},
			},
			wantAddedScenes:   nil,
			wantRemovedScenes: nil,
			wantAddedConns:    nil,
			wantRemovedConns:  nil,
		},
		{
			name: "added scene and connector",
			appA: AppExport{
				AppConnectors: []string{"c1"},
				Scenes:        []SceneExport{{Name: "s1"}},
			},
			appB: AppExport{
				AppConnectors: []string{"c1", "c2"},
				Scenes:        []SceneExport{{Name: "s1"}, {Name: "s2"}},
			},
			wantAddedScenes: []string{"s2"},
			wantAddedConns:  []string{"c2"},
		},
		{
			name: "removed scene",
			appA: AppExport{
				Scenes: []SceneExport{{Name: "s1"}, {Name: "s2"}},
			},
			appB: AppExport{
				Scenes: []SceneExport{{Name: "s1"}},
			},
			wantRemovedScenes: []string{"s2"},
		},
		{
			name: "basicDatas count change",
			appA: AppExport{
				BasicDatas: []interface{}{},
			},
			appB: AppExport{
				BasicDatas: []interface{}{"data1", "data2"},
			},
			wantAddedBasicData: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := diffApps(&tt.appA, &tt.appB)

			if len(report.AddedScenes) != len(tt.wantAddedScenes) {
				t.Errorf("added scenes: got %d, want %d", len(report.AddedScenes), len(tt.wantAddedScenes))
			}
			if len(report.RemovedScenes) != len(tt.wantRemovedScenes) {
				t.Errorf("removed scenes: got %d, want %d", len(report.RemovedScenes), len(tt.wantRemovedScenes))
			}
			if len(report.AddedConnectors) != len(tt.wantAddedConns) {
				t.Errorf("added connectors: got %d, want %d", len(report.AddedConnectors), len(tt.wantAddedConns))
			}
			if len(report.RemovedConnectors) != len(tt.wantRemovedConns) {
				t.Errorf("removed connectors: got %d, want %d", len(report.RemovedConnectors), len(tt.wantRemovedConns))
			}
			if len(report.AddedBasicDatas) != tt.wantAddedBasicData {
				t.Errorf("added basicDatas: got %d, want %d", len(report.AddedBasicDatas), tt.wantAddedBasicData)
			}
			if len(report.RemovedBasicDatas) != tt.wantRemovedBasicData {
				t.Errorf("removed basicDatas: got %d, want %d", len(report.RemovedBasicDatas), tt.wantRemovedBasicData)
			}
		})
	}
}
