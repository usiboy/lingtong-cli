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

// Helper: create a temporary app JSON file for testing.
func createTestAppFile(t *testing.T, dir string, name string, scenes []interface{}) string {
	t.Helper()
	appFile := AppFile{
		AppName:            name,
		AppConnectors:      []string{"kmerp", "kingDeeCloudStar"},
		BasicDatas:         []interface{}{},
		Scenes:             scenes,
		Workflows:          []interface{}{},
		WorkflowConnectors: []interface{}{},
	}
	data, err := json.MarshalIndent(appFile, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test app file: %v", err)
	}
	path := filepath.Join(dir, "app.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test app file: %v", err)
	}
	return path
}

func TestNewCmdAppSceneList_MissingFile(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppSceneList(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --file")
	}
	if !strings.Contains(err.Error(), "file") {
		t.Errorf("expected error about file, got: %v", err)
	}
}

func TestNewCmdAppSceneList_Success(t *testing.T) {
	tmpDir := t.TempDir()
	scenes := []interface{}{
		map[string]interface{}{
			"name":        "商品同步",
			"description": "定时查询金蝶商品列表同步到快麦",
			"source":      "kingDeeCloudStar",
			"target":      "kmerp",
			"trigger":     "scheduled",
		},
		map[string]interface{}{
			"name":   "销售订单同步",
			"source": "kmerp",
			"target": "kingDeeCloudStar",
		},
	}
	appPath := createTestAppFile(t, tmpDir, "test-app", scenes)

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

	cmd := newCmdAppSceneList(f)
	cmd.SetArgs([]string{"--file", appPath})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output is valid JSON
	var result []AppScene
	if err := json.Unmarshal([]byte(out.String()), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}

	if len(result) != 2 {
		t.Errorf("expected 2 scenes, got %d", len(result))
	}

	if result[0].Name != "商品同步" {
		t.Errorf("expected first scene name '商品同步', got %q", result[0].Name)
	}
	if result[0].Source != "kingDeeCloudStar" {
		t.Errorf("expected first scene source 'kingDeeCloudStar', got %q", result[0].Source)
	}
}

func TestNewCmdAppSceneList_EmptyScenes(t *testing.T) {
	tmpDir := t.TempDir()
	appPath := createTestAppFile(t, tmpDir, "empty-app", []interface{}{})

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

	cmd := newCmdAppSceneList(f)
	cmd.SetArgs([]string{"--file", appPath})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []AppScene
	if err := json.Unmarshal([]byte(out.String()), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 scenes, got %d", len(result))
	}
}

func TestNewCmdAppSceneList_InvalidFile(t *testing.T) {
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

	cmd := newCmdAppSceneList(f)
	cmd.SetArgs([]string{"--file", "/nonexistent/app.json"})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("expected read error, got: %v", err)
	}
}

func TestNewCmdAppSceneAdd_MissingFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErrPart string
	}{
		{
			name:        "missing file",
			args:        []string{"--name", "test", "--source", "src", "--target", "tgt"},
			wantErrPart: "file",
		},
		{
			name:        "missing name",
			args:        []string{"--file", "app.json", "--source", "src", "--target", "tgt"},
			wantErrPart: "name",
		},
		{
			name:        "missing source",
			args:        []string{"--file", "app.json", "--name", "test", "--target", "tgt"},
			wantErrPart: "source",
		},
		{
			name:        "missing target",
			args:        []string{"--file", "app.json", "--name", "test", "--source", "src"},
			wantErrPart: "target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFactory("http://example.com")
			cmd := newCmdAppSceneAdd(f)
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()
			if err == nil {
				t.Errorf("expected error containing %q", tt.wantErrPart)
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Errorf("expected error about %q, got: %v", tt.wantErrPart, err)
			}
		})
	}
}

func TestNewCmdAppSceneAdd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	initialScenes := []interface{}{
		map[string]interface{}{
			"name":   "商品同步",
			"source": "kingDeeCloudStar",
			"target": "kmerp",
		},
	}
	appPath := createTestAppFile(t, tmpDir, "test-app", initialScenes)

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

	cmd := newCmdAppSceneAdd(f)
	cmd.SetArgs([]string{
		"--file", appPath,
		"--name", "库存同步",
		"--source", "kingDeeCloudStar",
		"--target", "kmerp",
	})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify success message
	if !strings.Contains(out.String(), "Added scene") {
		t.Errorf("expected success message, got: %s", out.String())
	}

	// Verify file was updated
	data, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result AppFile
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output file contains invalid JSON: %v", err)
	}

	if len(result.Scenes) != 2 {
		t.Errorf("expected 2 scenes after add, got %d", len(result.Scenes))
	}
}

func TestNewCmdAppSceneAdd_WithOutput(t *testing.T) {
	tmpDir := t.TempDir()
	appPath := createTestAppFile(t, tmpDir, "test-app", []interface{}{})
	outputPath := filepath.Join(tmpDir, "updated.json")

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

	cmd := newCmdAppSceneAdd(f)
	cmd.SetArgs([]string{
		"--file", appPath,
		"--name", "New Scene",
		"--source", "src",
		"--target", "tgt",
		"--output", outputPath,
	})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original file unchanged
	origData, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("failed to read original file: %v", err)
	}
	var orig AppFile
	if err := json.Unmarshal(origData, &orig); err != nil {
		t.Fatalf("original file invalid JSON: %v", err)
	}
	if len(orig.Scenes) != 0 {
		t.Errorf("original file should have 0 scenes, got %d", len(orig.Scenes))
	}

	// Verify output file has the new scene
	newData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	var newFile AppFile
	if err := json.Unmarshal(newData, &newFile); err != nil {
		t.Fatalf("output file invalid JSON: %v", err)
	}
	if len(newFile.Scenes) != 1 {
		t.Errorf("output file should have 1 scene, got %d", len(newFile.Scenes))
	}
}

func TestNewCmdAppSceneRemove_MissingFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErrPart string
	}{
		{
			name:        "missing file",
			args:        []string{"--name", "test"},
			wantErrPart: "file",
		},
		{
			name:        "missing name",
			args:        []string{"--file", "app.json"},
			wantErrPart: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFactory("http://example.com")
			cmd := newCmdAppSceneRemove(f)
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()
			if err == nil {
				t.Errorf("expected error containing %q", tt.wantErrPart)
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Errorf("expected error about %q, got: %v", tt.wantErrPart, err)
			}
		})
	}
}

func TestNewCmdAppSceneRemove_Success(t *testing.T) {
	tmpDir := t.TempDir()
	scenes := []interface{}{
		map[string]interface{}{
			"name":   "商品同步",
			"source": "kingDeeCloudStar",
			"target": "kmerp",
		},
		map[string]interface{}{
			"name":   "销售订单同步",
			"source": "kmerp",
			"target": "kingDeeCloudStar",
		},
	}
	appPath := createTestAppFile(t, tmpDir, "test-app", scenes)

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

	cmd := newCmdAppSceneRemove(f)
	cmd.SetArgs([]string{
		"--file", appPath,
		"--name", "商品同步",
	})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was updated
	data, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result AppFile
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output file contains invalid JSON: %v", err)
	}

	if len(result.Scenes) != 1 {
		t.Errorf("expected 1 scene after remove, got %d", len(result.Scenes))
	}

	// Verify remaining scene is the correct one
	remainingScenes := extractScenes(result.Scenes)
	if len(remainingScenes) != 1 || remainingScenes[0].Name != "销售订单同步" {
		t.Errorf("expected remaining scene '销售订单同步', got: %v", remainingScenes)
	}
}

func TestNewCmdAppSceneRemove_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	appPath := createTestAppFile(t, tmpDir, "test-app", []interface{}{})

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

	cmd := newCmdAppSceneRemove(f)
	cmd.SetArgs([]string{
		"--file", appPath,
		"--name", "NonExistent",
	})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent scene")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestNewCmdAppSceneRemove_WithOutput(t *testing.T) {
	tmpDir := t.TempDir()
	scenes := []interface{}{
		map[string]interface{}{
			"name":   "商品同步",
			"source": "kingDeeCloudStar",
			"target": "kmerp",
		},
	}
	appPath := createTestAppFile(t, tmpDir, "test-app", scenes)
	outputPath := filepath.Join(tmpDir, "updated.json")

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

	cmd := newCmdAppSceneRemove(f)
	cmd.SetArgs([]string{
		"--file", appPath,
		"--name", "商品同步",
		"--output", outputPath,
	})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original file unchanged
	origData, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("failed to read original file: %v", err)
	}
	var orig AppFile
	if err := json.Unmarshal(origData, &orig); err != nil {
		t.Fatalf("original file invalid JSON: %v", err)
	}
	if len(orig.Scenes) != 1 {
		t.Errorf("original file should have 1 scene, got %d", len(orig.Scenes))
	}

	// Verify output file has 0 scenes
	newData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	var newFile AppFile
	if err := json.Unmarshal(newData, &newFile); err != nil {
		t.Fatalf("output file invalid JSON: %v", err)
	}
	if len(newFile.Scenes) != 0 {
		t.Errorf("output file should have 0 scenes, got %d", len(newFile.Scenes))
	}
}

// Unit tests for helper functions.

func TestExtractScenes_FromMap(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{
			"name":        "Test Scene",
			"description": "A test scene",
			"source":      "src",
			"target":      "tgt",
			"trigger":     "scheduled",
		},
	}

	scenes := extractScenes(raw)

	if len(scenes) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(scenes))
	}

	s := scenes[0]
	if s.Name != "Test Scene" {
		t.Errorf("expected name 'Test Scene', got %q", s.Name)
	}
	if s.Description != "A test scene" {
		t.Errorf("expected description 'A test scene', got %q", s.Description)
	}
	if s.Source != "src" {
		t.Errorf("expected source 'src', got %q", s.Source)
	}
	if s.Target != "tgt" {
		t.Errorf("expected target 'tgt', got %q", s.Target)
	}
	if s.Trigger != "scheduled" {
		t.Errorf("expected trigger 'scheduled', got %q", s.Trigger)
	}
}

func TestExtractScenes_FromAppScene(t *testing.T) {
	raw := []interface{}{
		AppScene{
			Name:        "Typed Scene",
			Description: "Typed description",
			Source:      "src2",
			Target:      "tgt2",
			Trigger:     "manual",
		},
	}

	scenes := extractScenes(raw)

	if len(scenes) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(scenes))
	}

	s := scenes[0]
	if s.Name != "Typed Scene" {
		t.Errorf("expected name 'Typed Scene', got %q", s.Name)
	}
	if s.Trigger != "manual" {
		t.Errorf("expected trigger 'manual', got %q", s.Trigger)
	}
}

func TestExtractScenes_Empty(t *testing.T) {
	scenes := extractScenes([]interface{}{})
	if len(scenes) != 0 {
		t.Errorf("expected 0 scenes, got %d", len(scenes))
	}
}

func TestExtractScenes_MixedTypes(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{
			"name":   "Map Scene",
			"source": "src",
			"target": "tgt",
		},
		"invalid string item",
		42, // invalid numeric item
		AppScene{
			Name:   "Typed Scene",
			Source: "src2",
			Target: "tgt2",
		},
	}

	scenes := extractScenes(raw)

	if len(scenes) != 2 {
		t.Errorf("expected 2 scenes (skipping invalid types), got %d", len(scenes))
	}
	if scenes[0].Name != "Map Scene" {
		t.Errorf("expected first scene 'Map Scene', got %q", scenes[0].Name)
	}
	if scenes[1].Name != "Typed Scene" {
		t.Errorf("expected second scene 'Typed Scene', got %q", scenes[1].Name)
	}
}

func TestRemoveSceneByName(t *testing.T) {
	tests := []struct {
		name          string
		scenes        []interface{}
		removeName    string
		wantLen       int
		wantRemoved   bool
		wantRemaining []string
	}{
		{
			name: "remove from map scenes",
			scenes: []interface{}{
				map[string]interface{}{"name": "Scene A", "source": "src"},
				map[string]interface{}{"name": "Scene B", "source": "src"},
				map[string]interface{}{"name": "Scene C", "source": "src"},
			},
			removeName:    "Scene B",
			wantLen:       2,
			wantRemoved:   true,
			wantRemaining: []string{"Scene A", "Scene C"},
		},
		{
			name: "remove from typed scenes",
			scenes: []interface{}{
				AppScene{Name: "X", Source: "src"},
				AppScene{Name: "Y", Source: "src"},
			},
			removeName:    "X",
			wantLen:       1,
			wantRemoved:   true,
			wantRemaining: []string{"Y"},
		},
		{
			name: "case insensitive",
			scenes: []interface{}{
				map[string]interface{}{"name": "Scene A", "source": "src"},
			},
			removeName:    "scene a",
			wantLen:       0,
			wantRemoved:   true,
			wantRemaining: []string{},
		},
		{
			name: "not found",
			scenes: []interface{}{
				map[string]interface{}{"name": "Scene A", "source": "src"},
			},
			removeName:    "NonExistent",
			wantLen:       1,
			wantRemoved:   false,
			wantRemaining: []string{"Scene A"},
		},
		{
			name:          "empty scenes",
			scenes:        []interface{}{},
			removeName:    "Any",
			wantLen:       0,
			wantRemoved:   false,
			wantRemaining: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, removed := removeSceneByName(tt.scenes, tt.removeName)

			if removed != tt.wantRemoved {
				t.Errorf("removed: got %v, want %v", removed, tt.wantRemoved)
			}
			if len(result) != tt.wantLen {
				t.Errorf("result length: got %d, want %d", len(result), tt.wantLen)
			}

			remaining := extractScenes(result)
			if len(remaining) != len(tt.wantRemaining) {
				t.Errorf("remaining count: got %d, want %d", len(remaining), len(tt.wantRemaining))
			}
			for i, want := range tt.wantRemaining {
				if i < len(remaining) && remaining[i].Name != want {
					t.Errorf("remaining[%d]: got %q, want %q", i, remaining[i].Name, want)
				}
			}
		})
	}
}

func TestLoadAppFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := loadAppFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid application JSON") {
		t.Errorf("expected invalid JSON error, got: %v", err)
	}
}

func TestLoadAppFile_Nonexistent(t *testing.T) {
	_, err := loadAppFile("/nonexistent/path/app.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Errorf("expected read error, got: %v", err)
	}
}

func TestSaveAppFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.json")

	appFile := &AppFile{
		AppName:       "Test App",
		AppConnectors: []string{"a", "b"},
		Scenes:        []interface{}{},
	}

	err := saveAppFile(path, appFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file exists and contains valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	var result AppFile
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("saved file contains invalid JSON: %v", err)
	}

	if result.AppName != "Test App" {
		t.Errorf("expected appName 'Test App', got %q", result.AppName)
	}
}
