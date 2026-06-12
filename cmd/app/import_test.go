// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
)

// validAppExport returns a valid AppExport for testing.
func validAppExport() AppExport {
	return AppExport{
		AppName:       "test-app",
		AppConnectors: []string{"connector-a", "connector-b"},
		BasicDatas:    []interface{}{},
		Scenes: []SceneExport{
			{
				Name:            "test-scene",
				ConnectorSource: ConnectorRef{Name: "connector-a"},
				ConnectorTarget: ConnectorRef{Name: "connector-b"},
				FieldMappings:   []interface{}{},
			},
		},
		Workflows:          []interface{}{},
		WorkflowConnectors: []interface{}{},
	}
}

// writeTempJSON writes data to a temporary JSON file and returns the path.
func writeTempJSON(t *testing.T, data interface{}) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.json")
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

// writeRawTempJSON writes raw JSON string to a temporary file.
func writeRawTempJSON(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestNewCmdAppImport(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		server      func() *httptest.Server
		fileContent string
		wantErr     bool
		errContains string
		outContains string
	}{
		{
			name:        "missing file flag",
			args:        []string{"import"},
			wantErr:     true,
			errContains: `"file" not set`,
		},
		{
			name:        "file not found",
			args:        []string{"import", "--file", "/nonexistent/path.json"},
			wantErr:     true,
			errContains: "failed to read file",
		},
		{
			name:        "invalid JSON",
			args:        []string{"import", "--file"},
			fileContent: `{invalid json`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name: "validation failure - missing appName",
			args: []string{"import", "--file"},
			fileContent: `{
				"appConnectors": ["conn"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "conn"}, "connectorTarget": {"name": "conn"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			wantErr:     true,
			errContains: "pre-flight validation failed",
		},
		{
			name: "validation failure - missing connectors",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "test",
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": ""}, "connectorTarget": {"name": ""}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			wantErr:     true,
			errContains: "required field 'appConnectors' is missing",
		},
		{
			name: "validation failure - missing scenes",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "test",
				"appConnectors": ["conn"],
				"basicDatas": [],
				"workflows": [],
				"workflowConnectors": []
			}`,
			wantErr:     true,
			errContains: "required field 'scenes' is missing",
		},
		{
			name: "force skips validation",
			args: []string{"import", "--file", "--force"},
			fileContent: `{
				"appName": "test",
				"appConnectors": ["conn"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "conn"}, "connectorTarget": {"name": "conn"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/gw/ai/proxy" {
						http.NotFound(w, r)
						return
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ImportResult{
						AppName: "test",
						Message: "Imported successfully",
					})
				}))
			},
			wantErr:     false,
			outContains: "Successfully imported application: test",
		},
		{
			name: "dry-run mode",
			args: []string{"import", "--file", "--dry-run"},
			fileContent: `{
				"appName": "my-app",
				"appConnectors": ["conn-a", "conn-b"],
				"basicDatas": [],
				"scenes": [{"name": "scene-1", "connectorSource": {"name": "conn-a"}, "connectorTarget": {"name": "conn-b"}, "fieldMappings": []}],
				"workflows": [{"id": 1}],
				"workflowConnectors": []
			}`,
			wantErr:     false,
			outContains: "[dry-run] Would import application: my-app",
		},
		{
			name: "dry-run shows connector count",
			args: []string{"import", "--file", "--dry-run"},
			fileContent: `{
				"appName": "app",
				"appConnectors": ["c1", "c2", "c3"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c2"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			wantErr:     false,
			outContains: "[dry-run] Connectors: 3",
		},
		{
			name: "dry-run shows scene count",
			args: []string{"import", "--file", "--dry-run"},
			fileContent: `{
				"appName": "app",
				"appConnectors": ["c1"],
				"basicDatas": [],
				"scenes": [{"name": "s1", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c1"}, "fieldMappings": []}, {"name": "s2", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c1"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			wantErr:     false,
			outContains: "[dry-run] Scenes: 2",
		},
		{
			name: "dry-run shows workflow count",
			args: []string{"import", "--file", "--dry-run"},
			fileContent: `{
				"appName": "app",
				"appConnectors": ["c1"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c1"}, "fieldMappings": []}],
				"workflows": [{"id": 1}, {"id": 2}, {"id": 3}],
				"workflowConnectors": []
			}`,
			wantErr:     false,
			outContains: "[dry-run] Workflows: 3",
		},
		{
			name: "successful import with full result",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "prod-app",
				"appConnectors": ["salesforce", "hubspot"],
				"basicDatas": [],
				"scenes": [{"name": "sync", "connectorSource": {"name": "salesforce"}, "connectorTarget": {"name": "hubspot"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/gw/ai/proxy" {
						http.NotFound(w, r)
						return
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ImportResult{
						AppName:       "prod-app",
						CreatedScenes: []string{"sync"},
						CreatedTables: []string{"contacts", "deals"},
						Message:       "Import completed",
					})
				}))
			},
			wantErr:     false,
			outContains: "Successfully imported application: prod-app",
		},
		{
			name: "import shows created scenes",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "app",
				"appConnectors": ["c1"],
				"basicDatas": [],
				"scenes": [{"name": "scene-a", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c1"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ImportResult{
						AppName:       "app",
						CreatedScenes: []string{"scene-a", "scene-b"},
					})
				}))
			},
			wantErr:     false,
			outContains: "Created scenes: scene-a, scene-b",
		},
		{
			name: "import shows created tables",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "app",
				"appConnectors": ["c1"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c1"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ImportResult{
						AppName:       "app",
						CreatedTables: []string{"table-1", "table-2", "table-3"},
					})
				}))
			},
			wantErr:     false,
			outContains: "Created tables: table-1, table-2, table-3",
		},
		{
			name: "import with non-JSON response",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "app",
				"appConnectors": ["c1"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c1"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status": "ok"}`))
				}))
			},
			wantErr:     false,
			outContains: "Successfully imported application:",
		},
		{
			name: "API error returns error",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "app",
				"appConnectors": ["c1"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "c1"}, "connectorTarget": {"name": "c1"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "internal server error"}`))
				}))
			},
			wantErr:     true,
			errContains: "failed to import application",
		},
		{
			name: "unknown connector reference passes import validation",
			args: []string{"import", "--file"},
			fileContent: `{
				"appName": "test",
				"appConnectors": ["conn-a"],
				"basicDatas": [],
				"scenes": [{"name": "s", "connectorSource": {"name": "unknown-conn"}, "connectorTarget": {"name": "conn-a"}, "fieldMappings": []}],
				"workflows": [],
				"workflowConnectors": []
			}`,
			server: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/gw/ai/proxy" {
						http.NotFound(w, r)
						return
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ImportResult{
						AppName: "test",
						Message: "imported",
					})
				}))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup server if needed
			var serverURL string
			if tt.server != nil {
				srv := tt.server()
				defer srv.Close()
				serverURL = srv.URL
			}

			// Setup file if needed
			if tt.fileContent != "" {
				filePath := writeRawTempJSON(t, tt.fileContent)
				// Insert file path after --file flag
				newArgs := []string{}
				for i, arg := range tt.args {
					newArgs = append(newArgs, arg)
					if arg == "--file" && i+1 < len(tt.args) && tt.args[i+1] != "--force" && tt.args[i+1] != "--dry-run" {
						// Next arg is not another flag, so --file has no value yet
						newArgs = append(newArgs, filePath)
					} else if arg == "--file" && (i+1 >= len(tt.args) || tt.args[i+1] == "--force" || tt.args[i+1] == "--dry-run") {
						// --file has no value, insert it
						newArgs = append(newArgs, filePath)
					}
				}
				tt.args = newArgs
			}

			// Create factory and command
			f := &cmdutil.Factory{
				Config: &config.Config{
					Host:  serverURL,
					Token: "test-token",
				},
			}

			cmd := newCmdAppImport(f)
			cmd.SetArgs(tt.args)

			// Capture output
			var out strings.Builder
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			// Execute
			err := cmd.Execute()

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}

			// Check output expectations
			if tt.outContains != "" {
				output := out.String()
				if !strings.Contains(output, tt.outContains) {
					t.Errorf("output %q does not contain %q", output, tt.outContains)
				}
			}
		})
	}
}

func TestNewCmdAppImport_ValidFile(t *testing.T) {
	// Test with a valid file using the validAppExport helper
	appExport := validAppExport()
	filePath := writeTempJSON(t, appExport)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gw/ai/proxy" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ImportResult{
			AppName: "test-app",
			Message: "OK",
		})
	}))
	defer srv.Close()

	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  srv.URL,
			Token: "test-token",
		},
	}

	cmd := newCmdAppImport(f)
	cmd.SetArgs([]string{"import", "--file", filePath})

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Successfully imported application: test-app") {
		t.Errorf("output %q does not contain success message", output)
	}
}

func TestImportResult_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected ImportResult
	}{
		{
			name: "full result",
			json: `{"appName":"app","createdScenes":["s1"],"createdTables":["t1"],"message":"done"}`,
			expected: ImportResult{
				AppName:       "app",
				CreatedScenes: []string{"s1"},
				CreatedTables: []string{"t1"},
				Message:       "done",
			},
		},
		{
			name:     "empty result",
			json:     `{}`,
			expected: ImportResult{},
		},
		{
			name: "only app name",
			json: `{"appName":"my-app"}`,
			expected: ImportResult{
				AppName: "my-app",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ImportResult
			if err := json.Unmarshal([]byte(tt.json), &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if result.AppName != tt.expected.AppName {
				t.Errorf("AppName: got %q, want %q", result.AppName, tt.expected.AppName)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("Message: got %q, want %q", result.Message, tt.expected.Message)
			}
		})
	}
}
