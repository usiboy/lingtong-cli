// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// MaxPayloadSize defines the maximum allowed size for app export JSON (10MB).
const MaxPayloadSize = 10 * 1024 * 1024

// validNamePattern matches names that contain only safe characters.
// Allows letters, numbers, spaces, hyphens, underscores, and dots.
var validNamePattern = regexp.MustCompile(`^[\p{L}\p{N} \-_.]+$`)

// AppExport represents the structure of an exported application.
type AppExport struct {
	AppName            string        `json:"appName"`
	AppConnectors      []string      `json:"appConnectors"`
	BasicDatas         []interface{} `json:"basicDatas"`
	Scenes             []SceneExport `json:"scenes"`
	Workflows          []interface{} `json:"workflows"`
	WorkflowConnectors []interface{} `json:"workflowConnectors"`
}

// SceneExport represents a scene in the exported application.
type SceneExport struct {
	Name            string        `json:"name"`
	Source          string        `json:"source,omitempty"`
	Target          string        `json:"target,omitempty"`
	ConnectorSource ConnectorRef  `json:"connectorSource"`
	ConnectorTarget ConnectorRef  `json:"connectorTarget"`
	DependsOn       []string      `json:"dependsOn,omitempty"`
	FieldMappings   []interface{} `json:"fieldMappings"`
}

// ConnectorRef represents a connector reference in a scene.
type ConnectorRef struct {
	Name string `json:"name"`
}

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func newCmdAppValidate(f *cmdutil.Factory) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate application export file",
		Long: `Validate the structure and references of an application export JSON file.

EXAMPLES:
    lingtong-cli app validate --file app.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			if err := ValidateRawJSON(data); err != nil {
				return err
			}

			var appExport AppExport
			if err := json.Unmarshal(data, &appExport); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			errors := validateAppExport(&appExport)

			if len(errors) > 0 {
				var msgs []string
				for _, e := range errors {
					msgs = append(msgs, e.Error())
				}
				return fmt.Errorf("validation failed:\n%s", strings.Join(msgs, "\n"))
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Validation passed: %s\n", appExport.AppName)
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Input JSON file path (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// validateAppExport validates the exported application structure.
func validateAppExport(app *AppExport) []ValidationError {
	var errors []ValidationError

	// Validate appName
	if app.AppName == "" {
		errors = append(errors, ValidationError{
			Field:   "appName",
			Message: "appName is required and must be a non-empty string",
		})
	}

	// Validate appConnectors
	if app.AppConnectors == nil || len(app.AppConnectors) == 0 {
		errors = append(errors, ValidationError{
			Field:   "appConnectors",
			Message: "appConnectors is required and must be a non-empty array",
		})
	}

	// Validate basicDatas
	if app.BasicDatas == nil {
		errors = append(errors, ValidationError{
			Field:   "basicDatas",
			Message: "basicDatas is required",
		})
	}

	// Validate scenes
	if app.Scenes == nil || len(app.Scenes) == 0 {
		errors = append(errors, ValidationError{
			Field:   "scenes",
			Message: "scenes is required and must be a non-empty array",
		})
	}

	// Build connector set for reference validation
	connectorSet := make(map[string]bool)
	for _, c := range app.AppConnectors {
		connectorSet[c] = true
	}

	// Validate scene connector references
	for i, scene := range app.Scenes {
		if scene.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("scenes[%d].name", i),
				Message: "scene name is required",
			})
		}

		if sourceConnectorName(scene) == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("scenes[%d].connectorSource.name", i),
				Message: "connectorSource.name is required",
			})
		}

		if targetConnectorName(scene) == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("scenes[%d].connectorTarget.name", i),
				Message: "connectorTarget.name is required",
			})
		}
	}

	// Edge case 1: Circular scene dependency refs
	errors = append(errors, validateCircularConnectorRefs(app.Scenes)...)

	// Edge case 2: Missing connector preflight
	errors = append(errors, validateConnectorPreflight(app.Scenes, connectorSet)...)

	// Edge case 4: Special characters in names
	errors = append(errors, validateSpecialCharacters(app)...)

	// Edge case 6: Duplicate scene names
	errors = append(errors, validateDuplicateSceneNames(app.Scenes)...)

	// Edge case 3: Field mapping断裂
	errors = append(errors, validateFieldMappings(app.Scenes)...)

	return errors
}

// validateCircularConnectorRefs detects circular scene dependency references.
func validateCircularConnectorRefs(scenes []SceneExport) []ValidationError {
	var errors []ValidationError
	// Build a graph: scene name -> depended-on scene names.
	edges := make(map[string][]string)
	for _, scene := range scenes {
		if scene.Name == "" {
			continue
		}
		for _, dependency := range scene.DependsOn {
			dependency = strings.TrimSpace(dependency)
			if dependency != "" {
				edges[scene.Name] = append(edges[scene.Name], dependency)
			}
		}
	}
	// DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var detectCycle func(node string) bool
	detectCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		for _, neighbor := range edges[node] {
			if !visited[neighbor] {
				if detectCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}
		recStack[node] = false
		return false
	}
	for node := range edges {
		if !visited[node] {
			if detectCycle(node) {
				errors = append(errors, ValidationError{
					Field:   "scenes",
					Message: "circular scene dependency detected among scenes",
				})
				return errors
			}
		}
	}
	return errors
}

// validateConnectorPreflight ensures all connectors referenced in scenes exist in appConnectors.
func validateConnectorPreflight(scenes []SceneExport, connectorSet map[string]bool) []ValidationError {
	var errors []ValidationError
	for i, scene := range scenes {
		source := sourceConnectorName(scene)
		target := targetConnectorName(scene)
		if source != "" && !connectorSet[source] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("scenes[%d].connectorSource.name", i),
				Message: fmt.Sprintf("connector '%s' not declared in appConnectors (preflight check)", source),
			})
		}
		if target != "" && !connectorSet[target] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("scenes[%d].connectorTarget.name", i),
				Message: fmt.Sprintf("connector '%s' not declared in appConnectors (preflight check)", target),
			})
		}
	}
	return errors
}

// validateFieldMappings checks that field mappings reference valid source/target fields.
func validateFieldMappings(scenes []SceneExport) []ValidationError {
	var errors []ValidationError
	for i, scene := range scenes {
		for j, fm := range scene.FieldMappings {
			if fmMap, ok := fm.(map[string]interface{}); ok {
				sourceField, hasSource := fmMap["sourceField"]
				targetField, hasTarget := fmMap["targetField"]
				expression, hasExpression := fmMap["expression"]
				if hasSource && (sourceField == nil || sourceField == "") {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("scenes[%d].fieldMappings[%d].sourceField", i, j),
						Message: "sourceField is empty or null",
					})
				}
				if hasTarget && (targetField == nil || targetField == "") {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("scenes[%d].fieldMappings[%d].targetField", i, j),
						Message: "targetField is empty or null",
					})
				}
				if hasExpression {
					expressionValue, ok := expression.(string)
					if !ok || strings.TrimSpace(expressionValue) == "" {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("scenes[%d].fieldMappings[%d].expression", i, j),
							Message: "expression must be a non-empty string",
						})
					} else if err := validateExpressionSyntax(expressionValue); err != nil {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("scenes[%d].fieldMappings[%d].expression", i, j),
							Message: fmt.Sprintf("invalid expression syntax: %s", err),
						})
					}
				}
			}
		}
	}
	return errors
}

func sourceConnectorName(scene SceneExport) string {
	if scene.ConnectorSource.Name != "" {
		return scene.ConnectorSource.Name
	}
	return scene.Source
}

func targetConnectorName(scene SceneExport) string {
	if scene.ConnectorTarget.Name != "" {
		return scene.ConnectorTarget.Name
	}
	return scene.Target
}

func validateExpressionSyntax(expression string) error {
	closingToOpening := map[rune]rune{
		')': '(',
		']': '[',
		'}': '{',
	}
	var stack []rune
	for _, r := range expression {
		if unicode.IsControl(r) {
			return fmt.Errorf("control character is not allowed")
		}
		switch r {
		case '(', '[', '{':
			stack = append(stack, r)
		case ')', ']', '}':
			if len(stack) == 0 || stack[len(stack)-1] != closingToOpening[r] {
				return fmt.Errorf("unmatched delimiter %q", r)
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		return fmt.Errorf("unclosed delimiter %q", stack[len(stack)-1])
	}
	return nil
}

// validateSpecialCharacters checks for dangerous special characters in names.
func validateSpecialCharacters(app *AppExport) []ValidationError {
	var errors []ValidationError
	if !validNamePattern.MatchString(app.AppName) {
		errors = append(errors, ValidationError{
			Field:   "appName",
			Message: "appName contains invalid characters (quotes, slashes, or other special chars)",
		})
	}
	for i, scene := range app.Scenes {
		if scene.Name != "" && !validNamePattern.MatchString(scene.Name) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("scenes[%d].name", i),
				Message: "scene name contains invalid characters",
			})
		}
	}
	return errors
}

// validateDuplicateSceneNames detects duplicate scene names.
func validateDuplicateSceneNames(scenes []SceneExport) []ValidationError {
	var errors []ValidationError
	seen := make(map[string]int)
	for i, scene := range scenes {
		if scene.Name != "" {
			if firstIdx, exists := seen[scene.Name]; exists {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("scenes[%d].name", i),
					Message: fmt.Sprintf("duplicate scene name '%s' (first seen at scenes[%d])", scene.Name, firstIdx),
				})
			} else {
				seen[scene.Name] = i
			}
		}
	}
	return errors
}

// hasSpecialChars checks if a string contains quotes, slashes, or control characters.
func hasSpecialChars(s string) bool {
	for _, r := range s {
		if r == '"' || r == '\'' || r == '\\' || r == '/' || unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// ValidatePayloadSize checks if the JSON payload exceeds the maximum allowed size.
func ValidatePayloadSize(data []byte) error {
	if len(data) > MaxPayloadSize {
		return fmt.Errorf("payload size %d bytes exceeds maximum allowed size of %d bytes (10MB)", len(data), MaxPayloadSize)
	}
	return nil
}

// ValidateRawJSON performs pre-unmarshal validation on raw JSON data.
func ValidateRawJSON(data []byte) error {
	// Edge case 10: Large payload
	if err := ValidatePayloadSize(data); err != nil {
		return err
	}

	// Edge case 7: Invalid JSON structure - check for unknown top-level keys
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	allowedKeys := map[string]bool{
		"appName":            true,
		"appConnectors":      true,
		"basicDatas":         true,
		"scenes":             true,
		"workflows":          true,
		"workflowConnectors": true,
	}

	var unknownKeys []string
	for key := range rawMap {
		if !allowedKeys[key] {
			unknownKeys = append(unknownKeys, key)
		}
	}

	if len(unknownKeys) > 0 {
		return fmt.Errorf("unknown top-level keys: %s", strings.Join(unknownKeys, ", "))
	}

	if connectorErrors := validateConnectorType(rawMap); len(connectorErrors) > 0 {
		return connectorErrors[0]
	}

	// Edge case 8: Null values for required fields
	requiredFields := []string{"appName", "appConnectors", "basicDatas", "scenes"}
	for _, field := range requiredFields {
		val, exists := rawMap[field]
		if !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
		if val == nil {
			return fmt.Errorf("required field '%s' is set to null instead of a valid value", field)
		}
	}

	return nil
}

// Edge case 9 is handled by JSON unmarshal type checking.
// When appConnectors is an object instead of array, json.Unmarshal will fail
// with a type mismatch error. We add a specific check here for clarity.
func validateConnectorType(rawMap map[string]interface{}) []ValidationError {
	var errors []ValidationError
	if connectors, exists := rawMap["appConnectors"]; exists {
		if connectors != nil {
			if _, ok := connectors.([]interface{}); !ok {
				errors = append(errors, ValidationError{
					Field:   "appConnectors",
					Message: "appConnectors must be an array of strings, not an object",
				})
			}
		}
	}
	return errors
}
