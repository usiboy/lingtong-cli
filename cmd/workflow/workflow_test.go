// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package workflow

import (
	"encoding/json"
	"testing"
)

// TestValidateDSL tests the DSL validation function
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
			errorCount:  1, // No start node
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
			errorCount:  1, // No end node
			warnCount:   0,
		},
		{
			name: "multiple start nodes",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start1", "type": "w_start"},
					map[string]interface{}{"id": "start2", "type": "w_start"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{
					map[string]interface{}{"source": "start1", "target": "end"},
					map[string]interface{}{"source": "start2", "target": "end"},
				},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1, // Multiple start nodes
			warnCount:   0,
		},
		{
			name: "invalid node type",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "end", "type": "w_end"},
					map[string]interface{}{"id": "node1", "type": "w_invalid_type"},
				},
				"edges": []interface{}{},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1, // Invalid node type
			warnCount:   0,
		},
		{
			name: "edge references non-existent node",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{
					map[string]interface{}{"source": "start", "target": "nonexistent"},
				},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1, // Non-existent target
			warnCount:   0,
		},
		{
			name: "connector node missing connector field",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "conn", "type": "w_connector", "data": map[string]interface{}{}},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{
					map[string]interface{}{"source": "start", "target": "conn"},
					map[string]interface{}{"source": "conn", "target": "end"},
				},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1, // Missing connector field
			warnCount:   0,
		},
		{
			name: "large workflow warning in strict mode",
			dsl: func() map[string]interface{} {
				nodes := make([]interface{}, 22)
				edges := make([]interface{}, 21)
				nodes[0] = map[string]interface{}{"id": "start", "type": "w_start"}
				for i := 1; i <= 20; i++ {
					nodes[i] = map[string]interface{}{"id": "node", "type": "w_connector", "data": map[string]interface{}{"connector": "test"}}
				}
				nodes[21] = map[string]interface{}{"id": "end", "type": "w_end"}
				for i := 0; i < 21; i++ {
					if i == 0 {
						edges[i] = map[string]interface{}{"source": "start", "target": "node"}
					} else if i == 20 {
						edges[i] = map[string]interface{}{"source": "node", "target": "end"}
					} else {
						edges[i] = map[string]interface{}{"source": "node", "target": "node"}
					}
				}
				return map[string]interface{}{"nodes": nodes, "edges": edges}
			}(),
			strict:      true,
			expectValid: true,
			errorCount:  0,
			warnCount:   1, // Large workflow warning
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

// TestGetWorkflowTemplate tests the template generation function
func TestGetWorkflowTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		expectValid  bool
	}{
		{"simple template", "simple", true},
		{"connector template", "connector", true},
		{"order_sync template", "order_sync", true},
		{"approval template", "approval", true},
		{"data_pipeline template", "data_pipeline", true},
		{"api_wrapper template", "api_wrapper", true},
		{"unknown template", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsl := getWorkflowTemplate(tt.templateName, "Test Workflow", "test")

			if tt.expectValid {
				if dsl == "" {
					t.Error("expected non-empty DSL, got empty string")
					return
				}

				// Verify DSL is valid JSON
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(dsl), &parsed); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				// Verify required fields
				if _, hasNodes := parsed["nodes"]; !hasNodes {
					t.Error("DSL missing 'nodes' field")
				}
				if _, hasEdges := parsed["edges"]; !hasEdges {
					t.Error("DSL missing 'edges' field")
				}
			} else {
				if dsl != "" {
					t.Errorf("expected empty DSL for unknown template, got non-empty string")
				}
			}
		})
	}
}

// TestGetTemplateList tests the template list function
func TestGetTemplateList(t *testing.T) {
	templates := getTemplateList()

	if len(templates) != 6 {
		t.Errorf("expected 6 templates, got %d", len(templates))
	}

	expectedNames := map[string]bool{
		"simple":        false,
		"connector":     false,
		"order_sync":    false,
		"approval":      false,
		"data_pipeline": false,
		"api_wrapper":   false,
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

		// Verify required fields
		if _, hasNodes := tmpl["nodes"]; !hasNodes {
			t.Errorf("template %s missing 'nodes' field", name)
		}
		if _, hasDesc := tmpl["description"]; !hasDesc {
			t.Errorf("template %s missing 'description' field", name)
		}
	}

	// Verify all expected templates are present
	for name, found := range expectedNames {
		if !found {
			t.Errorf("expected template '%s' not found", name)
		}
	}
}

// TestAnalyzeDeps tests the dependency analysis function
func TestAnalyzeDeps(t *testing.T) {
	tests := []struct {
		name           string
		dsl            map[string]interface{}
		expectConns    int
		expectScripts int
	}{
		{
			name: "workflow with connector and script",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{
						"id":   "start",
						"type": "w_start",
					},
					map[string]interface{}{
						"id":   "conn1",
						"type": "w_connector",
						"data": map[string]interface{}{
							"connector":        "kmerp",
							"interfaceModelId": 73,
							"authAccountId":    296,
						},
					},
					map[string]interface{}{
						"id":   "script1",
						"type": "w_script",
						"data": map[string]interface{}{
							"scriptConfig": map[string]interface{}{
								"language": "javascript",
							},
						},
					},
					map[string]interface{}{
						"id":   "end",
						"type": "w_end",
					},
				},
				"edges": []interface{}{},
			},
			expectConns:    1,
			expectScripts: 1,
		},
		{
			name: "workflow without connectors or scripts",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{},
			},
			expectConns:    0,
			expectScripts: 0,
		},
		{
			name: "workflow with modePipe connector",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
					map[string]interface{}{
						"id":   "modepipe1",
						"type": "w_modePipe",
						"data": map[string]interface{}{
							"connector":        "feishu",
							"interfaceModelId": 100,
							"authAccountId":    500,
						},
					},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{},
			},
			expectConns:    1,
			expectScripts: 0,
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

			connCount, _ := result["connectorCount"].(int)
			if connCount != tt.expectConns {
				t.Errorf("expected connectorCount=%d, got %d", tt.expectConns, connCount)
			}

			scriptCount, _ := result["scriptCount"].(int)
			if scriptCount != tt.expectScripts {
				t.Errorf("expected scriptCount=%d, got %d", tt.expectScripts, scriptCount)
			}
		})
	}
}
// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package workflow

import (
	"encoding/json"
	"testing"
)

func TestValidateDSL(t *testing.T) {
	tests := []struct {
		name        string
		dsl         map[string]interface{}
		strict      bool
		expectValid bool
		errorCount  int
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
		},
		{
			name: "missing nodes",
			dsl: map[string]interface{}{
				"edges": []interface{}{},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1,
		},
		{
			name: "no start node",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "node1", "type": "w_connector"},
					map[string]interface{}{"id": "end", "type": "w_end"},
				},
				"edges": []interface{}{},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1,
		},
		{
			name: "no end node",
			dsl: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"id": "start", "type": "w_start"},
				},
				"edges": []interface{}{},
			},
			strict:      false,
			expectValid: false,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateDSL(tt.dsl, tt.strict)
			valid, _ := result["valid"].(bool)
			if valid != tt.expectValid {
				t.Errorf("expected valid=%v, got %v", tt.expectValid, valid)
			}
			errors, _ := result["errors"].([]string)
			if len(errors) != tt.errorCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestGetWorkflowTemplate(t *testing.T) {
	templates := []string{"simple", "connector", "order_sync", "approval", "data_pipeline", "api_wrapper"}
	for _, name := range templates {
		t.Run(name, func(t *testing.T) {
			dsl := getWorkflowTemplate(name, "Test", "test")
			if dsl == "" {
				t.Errorf("template %s returned empty DSL", name)
				return
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(dsl), &parsed); err != nil {
				t.Errorf("invalid JSON: %v", err)
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
}

func TestAnalyzeDeps(t *testing.T) {
	dsl := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{"id": "start", "type": "w_start"},
			map[string]interface{}{
				"id":   "conn1",
				"type": "w_connector",
				"data": map[string]interface{}{
					"connector":        "kmerp",
					"interfaceModelId": 73,
					"authAccountId":    296,
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
