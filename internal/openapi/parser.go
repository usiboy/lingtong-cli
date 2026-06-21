// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package openapi provides a minimal Swagger 2.0 parser for generating
// CLI commands from OpenAPI specifications.
package openapi

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
)

// Spec represents a parsed Swagger 2.0 specification.
type Spec struct {
	Info  Info                `json:"info"`
	Paths map[string]PathItem `json:"paths"`
	Tags  []Tag               `json:"tags"`
}

// Info contains metadata about the API.
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// Tag represents an API tag/group.
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PathItem represents a single API path with its operations.
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operations returns all non-nil operations in this path item.
func (p *PathItem) Operations() []*Operation {
	var ops []*Operation
	if p.Get != nil {
		ops = append(ops, p.Get)
	}
	if p.Post != nil {
		ops = append(ops, p.Post)
	}
	if p.Put != nil {
		ops = append(ops, p.Put)
	}
	if p.Delete != nil {
		ops = append(ops, p.Delete)
	}
	return ops
}

// Operation represents a single API operation.
type Operation struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	Parameters  []Parameter `json:"parameters"`
	OperationID string      `json:"operationId"`
	// Populated during parsing
	Path   string `json:"-"`
	Method string `json:"-"`
}

// Parameter represents an API parameter.
type Parameter struct {
	Name        string        `json:"name"`
	In          string        `json:"in"` // query, path, body, header
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Type        string        `json:"type"`   // for query/path params
	Schema      *Schema       `json:"schema"` // for body params
	Format      string        `json:"format"`
	Enum        []interface{} `json:"enum,omitempty"`
}

// Schema represents a JSON schema (simplified).
type Schema struct {
	Type       string             `json:"type"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Items      *Schema            `json:"items,omitempty"`
	Ref        string             `json:"$ref,omitempty"`
	Example    interface{}        `json:"example,omitempty"`
}

// Parse reads and parses a Swagger 2.0 specification from a file.
func Parse(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(data)
}

// ParseBytes parses a Swagger 2.0 specification from JSON bytes.
func ParseBytes(data []byte) (*Spec, error) {
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	// Populate Path and Method for each operation
	for path, item := range spec.Paths {
		if item.Get != nil {
			item.Get.Path = path
			item.Get.Method = "GET"
		}
		if item.Post != nil {
			item.Post.Path = path
			item.Post.Method = "POST"
		}
		if item.Put != nil {
			item.Put.Path = path
			item.Put.Method = "PUT"
		}
		if item.Delete != nil {
			item.Delete.Path = path
			item.Delete.Method = "DELETE"
		}
	}

	return &spec, nil
}

// GroupByPrefix groups operations by their first-level path prefix.
// For example, /scene/list and /scene/get are grouped under "scene".
func (s *Spec) GroupByPrefix() map[string][]*Operation {
	result := make(map[string][]*Operation)

	for path, item := range s.Paths {
		prefix := extractPrefix(path)
		for _, op := range item.Operations() {
			result[prefix] = append(result[prefix], op)
		}
	}

	// Sort operations within each group by path for deterministic output
	for prefix := range result {
		sort.Slice(result[prefix], func(i, j int) bool {
			return result[prefix][i].Path < result[prefix][j].Path
		})
	}

	return result
}

// extractPrefix extracts the first-level path prefix.
// /scene/list -> scene
// /basicdata/record/listNew -> basicdata
// /factory/http/list -> factory
func extractPrefix(path string) string {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return "root"
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "root"
	}
	return parts[0]
}

// CommandName generates a CLI-friendly command name from a path.
// /scene/list -> list
// /basicdata/record/listNew -> record-listnew
// /factory/http/list -> http-list
func CommandName(path, prefix string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, prefix+"/")
	path = strings.TrimPrefix(path, prefix) // handle case where path is just /prefix

	if path == "" {
		return "root"
	}

	// Convert to lowercase and replace slashes with hyphens
	path = strings.ToLower(path)
	path = strings.ReplaceAll(path, "/", "-")

	// Drop path-parameter braces and wildcards so command names stay clean:
	// /meta/model/{connector}/get -> model-connector-get
	replacer := strings.NewReplacer("{", "", "}", "", "*", "")
	path = replacer.Replace(path)

	// Collapse any double hyphens left by removed wildcard segments.
	for strings.Contains(path, "--") {
		path = strings.ReplaceAll(path, "--", "-")
	}
	path = strings.Trim(path, "-")
	if path == "" {
		return "root"
	}

	return path
}

// PathParams returns only the path parameters from an operation.
func (op *Operation) PathParams() []Parameter {
	var params []Parameter
	for _, p := range op.Parameters {
		if p.In == "path" {
			params = append(params, p)
		}
	}
	return params
}

// QueryParams returns only the query parameters from an operation.
func (op *Operation) QueryParams() []Parameter {
	var params []Parameter
	for _, p := range op.Parameters {
		if p.In == "query" {
			params = append(params, p)
		}
	}
	return params
}

// BodyParams returns only the body parameters from an operation.
func (op *Operation) BodyParams() []Parameter {
	var params []Parameter
	for _, p := range op.Parameters {
		if p.In == "body" {
			params = append(params, p)
		}
	}
	return params
}

// HasRequiredParams returns true if the operation has any required parameters.
func (op *Operation) HasRequiredParams() bool {
	for _, p := range op.Parameters {
		if p.Required {
			return true
		}
	}
	return false
}
