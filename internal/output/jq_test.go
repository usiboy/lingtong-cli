// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestJqFilterBasicPath(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"name":    "kmerp",
		"version": "1.0",
	}
	if err := JqFilter(&buf, data, ".name"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "kmerp" {
		t.Errorf("JqFilter(.name) = %q, want %q", got, "kmerp")
	}
}

func TestJqFilterNestedPath(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"result": map[string]interface{}{
			"id":   42,
			"name": "test",
		},
	}
	if err := JqFilter(&buf, data, ".result.id"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "42" {
		t.Errorf("JqFilter(.result.id) = %q, want %q", got, "42")
	}
}

func TestJqFilterArray(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "a"},
			map[string]interface{}{"name": "b"},
		},
	}
	if err := JqFilter(&buf, data, ".items[].name"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "a" || lines[1] != "b" {
		t.Errorf("lines = %v, want [a, b]", lines)
	}
}

func TestJqFilterArrayIndex(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"items": []interface{}{"first", "second", "third"},
	}
	if err := JqFilter(&buf, data, ".items[0]"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "first" {
		t.Errorf("JqFilter(.items[0]) = %q, want %q", got, "first")
	}
}

func TestJqFilterComplexValue(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"config": map[string]interface{}{
			"host": "localhost",
			"port": 8080,
		},
	}
	if err := JqFilter(&buf, data, ".config"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"host"`) {
		t.Errorf("expected JSON object output, got: %s", output)
	}
	if !strings.Contains(output, `"localhost"`) {
		t.Errorf("expected host value in output, got: %s", output)
	}
}

func TestJqFilterNull(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{"name": "test"}
	if err := JqFilter(&buf, data, ".missing"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "null" {
		t.Errorf("JqFilter(.missing) = %q, want %q", got, "null")
	}
}

func TestJqFilterBoolean(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{"active": true}
	if err := JqFilter(&buf, data, ".active"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "true" {
		t.Errorf("JqFilter(.active) = %q, want %q", got, "true")
	}
}

func TestJqFilterPipe(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "a", "value": 1},
			map[string]interface{}{"name": "b", "value": 2},
		},
	}
	if err := JqFilter(&buf, data, ".items[] | .name"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 || lines[0] != "a" || lines[1] != "b" {
		t.Errorf("pipe result = %v, want [a, b]", lines)
	}
}

func TestJqFilterInvalidExpression(t *testing.T) {
	var buf bytes.Buffer
	err := JqFilter(&buf, nil, ".[invalid")
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
	if !strings.Contains(err.Error(), "invalid jq expression") {
		t.Errorf("error = %q, want 'invalid jq expression'", err.Error())
	}
}

func TestJqFilterRuntimeError(t *testing.T) {
	var buf bytes.Buffer
	// Trying to index a string should cause a runtime error
	data := map[string]interface{}{"name": "test"}
	err := JqFilter(&buf, data, ".name | length | .foo")
	if err == nil {
		t.Error("expected runtime error, got nil")
	}
}

func TestJqFilterIterateNull(t *testing.T) {
	var buf bytes.Buffer
	// .result is null, iterating over it should give a friendly error
	data := map[string]interface{}{"success": true, "message": "ok"}
	err := JqFilter(&buf, data, ".result[].name")
	if err == nil {
		t.Error("expected error when iterating over null, got nil")
	}
	// Error should contain helpful hint
	if !strings.Contains(err.Error(), "Hint:") {
		t.Errorf("error should contain Hint, got: %v", err)
	}
	if !strings.Contains(err.Error(), "top-level keys") {
		t.Errorf("error should list top-level keys, got: %v", err)
	}
}

func TestJqFilterIterateObject(t *testing.T) {
	var buf bytes.Buffer
	// .result is an object (not array), iterating gives values then .name fails on non-objects
	data := map[string]interface{}{
		"result": map[string]interface{}{
			"list":  []interface{}{map[string]interface{}{"name": "a"}},
			"total": 1,
		},
	}
	err := JqFilter(&buf, data, ".result[].name")
	if err == nil {
		t.Error("expected error when iterating over object, got nil")
	}
	// Error should mention the type mismatch
	if !strings.Contains(err.Error(), "jq error") {
		t.Errorf("error should contain 'jq error', got: %v", err)
	}
}

func TestJqFilterCorrectPath(t *testing.T) {
	var buf bytes.Buffer
	// Correct path for scene list structure: .result.list[].name
	data := map[string]interface{}{
		"result": map[string]interface{}{
			"list": []interface{}{
				map[string]interface{}{"name": "Order Sync"},
				map[string]interface{}{"name": "Data Pipeline"},
			},
			"total": 2,
		},
	}
	if err := JqFilter(&buf, data, ".result.list[].name"); err != nil {
		t.Fatalf("JqFilter() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 || lines[0] != "Order Sync" || lines[1] != "Data Pipeline" {
		t.Errorf("result = %v, want [Order Sync, Data Pipeline]", lines)
	}
}

func TestValidateJqExpression(t *testing.T) {
	tests := []struct {
		expr    string
		wantErr bool
	}{
		{".name", false},
		{".items[].name", false},
		{". | select(.active)", false},
		{".[invalid", true},
		{"", true}, // empty is invalid for gojq (missing query)
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			err := ValidateJqExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJqExpression(%q) error = %v, wantErr = %v", tt.expr, err, tt.wantErr)
			}
		})
	}
}
