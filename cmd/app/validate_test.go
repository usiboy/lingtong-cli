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

func TestNewCmdAppValidate_MissingFile(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppValidate(f)
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

func TestNewCmdAppValidate_FileNotFound(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdAppValidate(f)
	cmd.SetArgs([]string{"--file", "/nonexistent/file.json"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "read file") {
		t.Errorf("expected error about reading file, got: %v", err)
	}
}

func TestNewCmdAppValidate_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(testFile, []byte("not valid json"), 0644)

	f := newTestFactory("http://example.com")
	cmd := newCmdAppValidate(f)
	cmd.SetArgs([]string{"--file", testFile})
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

func TestNewCmdAppValidate_UnknownTopLevelKey(t *testing.T) {
	input := `{"appName":"Test App","appConnectors":["c1"],"basicDatas":[],"scenes":[],"unexpected":true}`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "unknown-key.json")
	os.WriteFile(testFile, []byte(input), 0644)

	f := newTestFactory("http://example.com")
	cmd := newCmdAppValidate(f)
	cmd.SetArgs([]string{"--file", testFile})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown top-level key")
	}
	if !strings.Contains(err.Error(), "unknown top-level keys") {
		t.Errorf("expected unknown key error, got: %v", err)
	}
}

func TestNewCmdAppValidate_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		wantErr  bool
		errMatch string
	}{
		{
			name:     "empty object",
			input:    map[string]interface{}{},
			wantErr:  true,
			errMatch: "appName",
		},
		{
			name: "missing appName",
			input: map[string]interface{}{
				"appConnectors": []string{"c1"},
				"basicDatas":    []interface{}{},
				"scenes":        []interface{}{},
			},
			wantErr:  true,
			errMatch: "appName",
		},
		{
			name: "missing appConnectors",
			input: map[string]interface{}{
				"appName":    "Test App",
				"basicDatas": []interface{}{},
				"scenes":     []interface{}{},
			},
			wantErr:  true,
			errMatch: "appConnectors",
		},
		{
			name: "empty appConnectors",
			input: map[string]interface{}{
				"appName":       "Test App",
				"appConnectors": []string{},
				"basicDatas":    []interface{}{},
				"scenes":        []interface{}{},
			},
			wantErr:  true,
			errMatch: "appConnectors",
		},
		{
			name: "missing basicDatas",
			input: map[string]interface{}{
				"appName":       "Test App",
				"appConnectors": []string{"c1"},
				"scenes":        []interface{}{},
			},
			wantErr:  true,
			errMatch: "basicDatas",
		},
		{
			name: "missing scenes",
			input: map[string]interface{}{
				"appName":       "Test App",
				"appConnectors": []string{"c1"},
				"basicDatas":    []interface{}{},
			},
			wantErr:  true,
			errMatch: "scenes",
		},
		{
			name: "empty scenes",
			input: map[string]interface{}{
				"appName":       "Test App",
				"appConnectors": []string{"c1"},
				"basicDatas":    []interface{}{},
				"scenes":        []interface{}{},
			},
			wantErr:  true,
			errMatch: "scenes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.json")

			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal test input: %v", err)
			}
			os.WriteFile(testFile, data, 0644)

			f := &cmdutil.Factory{
				Config: &config.Config{
					Host:  "http://example.com",
					Token: "test-token",
				},
				IOStreams: &output.IOStreams{
					In:     io.NopCloser(strings.NewReader("")),
					Out:    io.Discard,
					ErrOut: io.Discard,
				},
			}

			cmd := newCmdAppValidate(f)
			cmd.SetArgs([]string{"--file", testFile})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err = cmd.Execute()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMatch) {
				t.Errorf("expected error containing %q, got: %v", tt.errMatch, err)
			}
		})
	}
}

func TestNewCmdAppValidate_InvalidConnectorReference(t *testing.T) {
	input := map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"connector1", "connector2"},
		"basicDatas":    []interface{}{},
		"scenes": []interface{}{
			map[string]interface{}{
				"name": "Test Scene",
				"connectorSource": map[string]interface{}{
					"name": "nonexistent-connector",
				},
				"connectorTarget": map[string]interface{}{
					"name": "connector1",
				},
				"fieldMappings": []interface{}{},
			},
		},
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	data, _ := json.Marshal(input)
	os.WriteFile(testFile, data, 0644)

	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://example.com",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    io.Discard,
			ErrOut: io.Discard,
		},
	}

	cmd := newCmdAppValidate(f)
	cmd.SetArgs([]string{"--file", testFile})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid connector reference")
	}
	if !strings.Contains(err.Error(), "nonexistent-connector") {
		t.Errorf("expected error about nonexistent connector, got: %v", err)
	}
}

func TestNewCmdAppValidate_TargetConnectorInvalid(t *testing.T) {
	input := map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"connector1"},
		"basicDatas":    []interface{}{},
		"scenes": []interface{}{
			map[string]interface{}{
				"name": "Test Scene",
				"connectorSource": map[string]interface{}{
					"name": "connector1",
				},
				"connectorTarget": map[string]interface{}{
					"name": "invalid-target",
				},
				"fieldMappings": []interface{}{},
			},
		},
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	data, _ := json.Marshal(input)
	os.WriteFile(testFile, data, 0644)

	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "http://example.com",
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    io.Discard,
			ErrOut: io.Discard,
		},
	}

	cmd := newCmdAppValidate(f)
	cmd.SetArgs([]string{"--file", testFile})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid target connector reference")
	}
	if !strings.Contains(err.Error(), "invalid-target") {
		t.Errorf("expected error about invalid target connector, got: %v", err)
	}
}

func TestNewCmdAppValidate_Success(t *testing.T) {
	input := map[string]interface{}{
		"appName":       "Test App",
		"appConnectors": []string{"connector1", "connector2"},
		"basicDatas":    []interface{}{},
		"scenes": []interface{}{
			map[string]interface{}{
				"name": "Test Scene",
				"connectorSource": map[string]interface{}{
					"name": "connector1",
				},
				"connectorTarget": map[string]interface{}{
					"name": "connector2",
				},
				"fieldMappings": []interface{}{},
			},
		},
		"workflows":          []interface{}{},
		"workflowConnectors": []interface{}{},
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	data, _ := json.Marshal(input)
	os.WriteFile(testFile, data, 0644)

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

	cmd := newCmdAppValidate(f)
	cmd.SetArgs([]string{"--file", testFile})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Validation passed") {
		t.Errorf("expected validation passed message, got: %s", output)
	}
	if !strings.Contains(output, "Test App") {
		t.Errorf("expected app name in output, got: %s", output)
	}
}

func TestValidateAppExport_EmptyAppName(t *testing.T) {
	app := &AppExport{
		AppName:       "",
		AppConnectors: []string{"c1"},
		BasicDatas:    []interface{}{},
		Scenes:        []SceneExport{{Name: "scene1"}},
	}

	errors := validateAppExport(app)
	if len(errors) == 0 {
		t.Fatal("expected validation errors")
	}
	if errors[0].Field != "appName" {
		t.Errorf("expected first error on appName, got: %s", errors[0].Field)
	}
}

func TestValidateAppExport_NilBasicDatas(t *testing.T) {
	app := &AppExport{
		AppName:       "Test App",
		AppConnectors: []string{"c1"},
		BasicDatas:    nil,
		Scenes:        []SceneExport{{Name: "scene1"}},
	}

	errors := validateAppExport(app)
	if len(errors) == 0 {
		t.Fatal("expected validation errors")
	}
	found := false
	for _, e := range errors {
		if e.Field == "basicDatas" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error on basicDatas, got: %v", errors)
	}
}

func TestValidateAppExport_MultipleErrors(t *testing.T) {
	app := &AppExport{
		AppName:       "",
		AppConnectors: nil,
		BasicDatas:    nil,
		Scenes:        nil,
	}

	errors := validateAppExport(app)
	if len(errors) < 3 {
		t.Errorf("expected at least 3 errors, got: %d", len(errors))
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "appName",
		Message: "required",
	}

	expected := "appName: required"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidateAppExport_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		app      *AppExport
		wantErr  bool
		errMatch string
	}{
		{
			name: "circular scene refs",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"A", "B"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{Name: "scene1", ConnectorSource: ConnectorRef{Name: "A"}, ConnectorTarget: ConnectorRef{Name: "B"}, DependsOn: []string{"scene2"}},
					{Name: "scene2", ConnectorSource: ConnectorRef{Name: "B"}, ConnectorTarget: ConnectorRef{Name: "A"}, DependsOn: []string{"scene1"}},
				},
			},
			wantErr:  true,
			errMatch: "circular scene dependency",
		},
		{
			name: "missing connector preflight - source",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"valid-connector"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{Name: "scene1", ConnectorSource: ConnectorRef{Name: "missing-connector"}, ConnectorTarget: ConnectorRef{Name: "valid-connector"}},
				},
			},
			wantErr:  true,
			errMatch: "not declared in appConnectors",
		},
		{
			name: "missing connector preflight - target",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"valid-connector"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{Name: "scene1", ConnectorSource: ConnectorRef{Name: "valid-connector"}, ConnectorTarget: ConnectorRef{Name: "missing-target"}},
				},
			},
			wantErr:  true,
			errMatch: "not declared in appConnectors",
		},
		{
			name: "missing scene connector refs",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"valid-connector"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{Name: "scene1"},
				},
			},
			wantErr:  true,
			errMatch: "connectorSource.name is required",
		},
		{
			name: "field mapping with empty sourceField",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{
						Name:            "scene1",
						ConnectorSource: ConnectorRef{Name: "c1"},
						ConnectorTarget: ConnectorRef{Name: "c1"},
						FieldMappings:   []interface{}{map[string]interface{}{"sourceField": "", "targetField": "valid"}},
					},
				},
			},
			wantErr:  true,
			errMatch: "sourceField is empty or null",
		},
		{
			name: "field mapping with empty targetField",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{
						Name:            "scene1",
						ConnectorSource: ConnectorRef{Name: "c1"},
						ConnectorTarget: ConnectorRef{Name: "c1"},
						FieldMappings:   []interface{}{map[string]interface{}{"sourceField": "valid", "targetField": ""}},
					},
				},
			},
			wantErr:  true,
			errMatch: "targetField is empty or null",
		},
		{
			name: "field mapping with bad expression",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{
						Name:            "scene1",
						ConnectorSource: ConnectorRef{Name: "c1"},
						ConnectorTarget: ConnectorRef{Name: "c1"},
						FieldMappings:   []interface{}{map[string]interface{}{"sourceField": "valid", "targetField": "valid", "expression": "SUM({amount)"}},
					},
				},
			},
			wantErr:  true,
			errMatch: "invalid expression syntax",
		},
		{
			name: "special characters in app name",
			app: &AppExport{
				AppName:       "Test/App\"Name",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes:        []SceneExport{{Name: "scene1"}},
			},
			wantErr:  true,
			errMatch: "invalid characters",
		},
		{
			name: "special characters in scene name",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes:        []SceneExport{{Name: "scene/with'quote"}},
			},
			wantErr:  true,
			errMatch: "scene name contains invalid characters",
		},
		{
			name: "duplicate scene names",
			app: &AppExport{
				AppName:       "Test App",
				AppConnectors: []string{"c1"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{Name: "duplicate-scene"},
					{Name: "duplicate-scene"},
				},
			},
			wantErr:  true,
			errMatch: "duplicate scene name",
		},
		{
			name: "valid app with no edge case errors",
			app: &AppExport{
				AppName:       "Valid App Name",
				AppConnectors: []string{"connector-a", "connector-b"},
				BasicDatas:    []interface{}{},
				Scenes: []SceneExport{
					{
						Name:            "unique-scene-1",
						ConnectorSource: ConnectorRef{Name: "connector-a"},
						ConnectorTarget: ConnectorRef{Name: "connector-b"},
						FieldMappings:   []interface{}{},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateAppExport(tt.app)

			if tt.wantErr && len(errors) == 0 {
				t.Error("expected validation errors, got none")
			}

			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("expected no errors, got: %v", errors)
			}

			if tt.wantErr && tt.errMatch != "" {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.errMatch) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got: %v", tt.errMatch, errors)
				}
			}
		})
	}
}

func TestValidateRawJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		errMatch string
	}{
		{
			name:     "unknown top-level keys",
			input:    `{"appName":"Test","appConnectors":["c1"],"basicDatas":[],"scenes":[],"unknownKey":"value"}`,
			wantErr:  true,
			errMatch: "unknown top-level keys",
		},
		{
			name:     "null appName",
			input:    `{"appName":null,"appConnectors":["c1"],"basicDatas":[],"scenes":[]}`,
			wantErr:  true,
			errMatch: "set to null",
		},
		{
			name:     "null appConnectors",
			input:    `{"appName":"Test","appConnectors":null,"basicDatas":[],"scenes":[]}`,
			wantErr:  true,
			errMatch: "set to null",
		},
		{
			name:     "null basicDatas",
			input:    `{"appName":"Test","appConnectors":["c1"],"basicDatas":null,"scenes":[]}`,
			wantErr:  true,
			errMatch: "set to null",
		},
		{
			name:     "null scenes",
			input:    `{"appName":"Test","appConnectors":["c1"],"basicDatas":[],"scenes":null}`,
			wantErr:  true,
			errMatch: "set to null",
		},
		{
			name:    "valid JSON passes",
			input:   `{"appName":"Test","appConnectors":["c1"],"basicDatas":[],"scenes":[],"workflows":[],"workflowConnectors":[]}`,
			wantErr: false,
		},
		{
			name:     "invalid JSON syntax",
			input:    `{invalid json}`,
			wantErr:  true,
			errMatch: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRawJSON([]byte(tt.input))

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.wantErr && tt.errMatch != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("expected error containing %q, got: %v", tt.errMatch, err)
				}
			}
		})
	}
}

func TestValidatePayloadSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{
			name:    "payload under limit",
			size:    1024,
			wantErr: false,
		},
		{
			name:    "payload at exact limit",
			size:    MaxPayloadSize,
			wantErr: false,
		},
		{
			name:    "payload over limit",
			size:    MaxPayloadSize + 1,
			wantErr: true,
		},
		{
			name:    "payload significantly over limit",
			size:    MaxPayloadSize * 2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			err := ValidatePayloadSize(data)

			if tt.wantErr && err == nil {
				t.Error("expected error for large payload, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "exceeds maximum allowed size") {
					t.Errorf("expected size limit error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateConnectorType(t *testing.T) {
	tests := []struct {
		name     string
		rawMap   map[string]interface{}
		wantErr  bool
		errMatch string
	}{
		{
			name: "appConnectors as array - valid",
			rawMap: map[string]interface{}{
				"appConnectors": []interface{}{"c1", "c2"},
			},
			wantErr: false,
		},
		{
			name: "appConnectors as object - invalid",
			rawMap: map[string]interface{}{
				"appConnectors": map[string]interface{}{"key": "value"},
			},
			wantErr:  true,
			errMatch: "must be an array",
		},
		{
			name: "appConnectors missing - no error",
			rawMap: map[string]interface{}{
				"appName": "Test",
			},
			wantErr: false,
		},
		{
			name: "appConnectors null - no error",
			rawMap: map[string]interface{}{
				"appConnectors": nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateConnectorType(tt.rawMap)

			if tt.wantErr && len(errors) == 0 {
				t.Error("expected validation error, got none")
			}

			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("expected no errors, got: %v", errors)
			}

			if tt.wantErr && tt.errMatch != "" && len(errors) > 0 {
				if !strings.Contains(errors[0].Message, tt.errMatch) {
					t.Errorf("expected error containing %q, got: %v", tt.errMatch, errors[0].Message)
				}
			}
		})
	}
}

func TestValidateCircularConnectorRefs_NoCycle(t *testing.T) {
	scenes := []SceneExport{
		{Name: "s1", ConnectorSource: ConnectorRef{Name: "A"}, ConnectorTarget: ConnectorRef{Name: "B"}},
		{Name: "s2", ConnectorSource: ConnectorRef{Name: "C"}, ConnectorTarget: ConnectorRef{Name: "D"}},
	}

	errors := validateCircularConnectorRefs(scenes)
	if len(errors) > 0 {
		t.Errorf("expected no cycle errors, got: %v", errors)
	}
}

func TestValidateCircularConnectorRefs_EmptyScenes(t *testing.T) {
	errors := validateCircularConnectorRefs([]SceneExport{})
	if len(errors) > 0 {
		t.Errorf("expected no errors for empty scenes, got: %v", errors)
	}
}

func TestValidateCircularConnectorRefs_SelfConnectorSceneNoCycle(t *testing.T) {
	scenes := []SceneExport{
		{Name: "self", ConnectorSource: ConnectorRef{Name: "A"}, ConnectorTarget: ConnectorRef{Name: "A"}},
	}

	errors := validateCircularConnectorRefs(scenes)
	if len(errors) > 0 {
		t.Errorf("expected no cycle errors for a same-connector scene, got: %v", errors)
	}
}

func TestValidateCircularConnectorRefs_DependsOnCycle(t *testing.T) {
	scenes := []SceneExport{
		{Name: "s1", DependsOn: []string{"s2"}},
		{Name: "s2", DependsOn: []string{"s1"}},
	}

	errors := validateCircularConnectorRefs(scenes)
	if len(errors) == 0 {
		t.Fatal("expected circular dependency error for scene dependency refs")
	}
	if !strings.Contains(errors[0].Message, "circular scene dependency") {
		t.Errorf("expected circular dependency error, got: %v", errors)
	}
}

func TestValidateCircularConnectorRefs_LegacySourceTargetNoCycle(t *testing.T) {
	scenes := []SceneExport{
		{Name: "s1", Source: "A", Target: "B"},
		{Name: "s2", Source: "B", Target: "A"},
	}

	errors := validateCircularConnectorRefs(scenes)
	if len(errors) > 0 {
		t.Errorf("expected no dependency cycle errors for legacy source/target flow scenes, got: %v", errors)
	}
}

func TestValidateDuplicateSceneNames_NoDuplicates(t *testing.T) {
	scenes := []SceneExport{
		{Name: "scene-a"},
		{Name: "scene-b"},
		{Name: "scene-c"},
	}

	errors := validateDuplicateSceneNames(scenes)
	if len(errors) > 0 {
		t.Errorf("expected no duplicate errors, got: %v", errors)
	}
}

func TestValidateSpecialCharacters_ValidNames(t *testing.T) {
	app := &AppExport{
		AppName:       "Valid App Name 123",
		AppConnectors: []string{"c1"},
		BasicDatas:    []interface{}{},
		Scenes:        []SceneExport{{Name: "valid-scene_name.test"}},
	}

	errors := validateSpecialCharacters(app)
	if len(errors) > 0 {
		t.Errorf("expected no special char errors for valid names, got: %v", errors)
	}
}

func TestValidateFieldMappings_EmptyMappings(t *testing.T) {
	scenes := []SceneExport{
		{
			Name:            "scene1",
			ConnectorSource: ConnectorRef{Name: "c1"},
			ConnectorTarget: ConnectorRef{Name: "c1"},
			FieldMappings:   []interface{}{},
		},
	}

	errors := validateFieldMappings(scenes)
	if len(errors) > 0 {
		t.Errorf("expected no errors for empty field mappings, got: %v", errors)
	}
}
