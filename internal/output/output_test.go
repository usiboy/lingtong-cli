// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRemoveNullFields(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
		},
		{
			name:     "simple map with null",
			input:    map[string]interface{}{"a": "value", "b": nil, "c": 123},
			expected: map[string]interface{}{"a": "value", "c": 123},
		},
		{
			name:     "nested map with null",
			input:    map[string]interface{}{"a": map[string]interface{}{"b": nil, "c": "value"}},
			expected: map[string]interface{}{"a": map[string]interface{}{"c": "value"}},
		},
		{
			name:     "array with null",
			input:    []interface{}{"a", nil, "b"},
			expected: []interface{}{"a", nil, "b"},
		},
		{
			name:     "array of maps with null",
			input:    []interface{}{map[string]interface{}{"a": "value", "b": nil}},
			expected: []interface{}{map[string]interface{}{"a": "value"}},
		},
		{
			name:     "complex nested structure",
			input:    map[string]interface{}{"a": []interface{}{map[string]interface{}{"b": nil, "c": "value"}, nil}, "d": nil},
			expected: map[string]interface{}{"a": []interface{}{map[string]interface{}{"c": "value"}, nil}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeNullFields(tt.input)

			// Compare by marshaling to JSON
			expectedJSON, _ := json.Marshal(tt.expected)
			resultJSON, _ := json.Marshal(result)

			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("removeNullFields() = %v, want %v", string(resultJSON), string(expectedJSON))
			}
		})
	}
}

func TestWriterOmitNull(t *testing.T) {
	tests := []struct {
		name        string
		omitNull    bool
		input       interface{}
		contains    string
		notContains string
	}{
		{
			name:        "with omitNull true",
			omitNull:    true,
			input:       map[string]interface{}{"a": "value", "b": nil, "c": 123},
			contains:    "\"a\": \"value\"",
			notContains: "\"b\": null",
		},
		{
			name:        "with omitNull false",
			omitNull:    false,
			input:       map[string]interface{}{"a": "value", "b": nil, "c": 123},
			contains:    "\"b\": null",
			notContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ioStreams := &IOStreams{Out: &buf}
			w := NewWriterWithOpts(ioStreams, FormatJSON, tt.omitNull)

			err := w.Write(tt.input)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			output := buf.String()
			if tt.contains != "" && !bytes.Contains([]byte(output), []byte(tt.contains)) {
				t.Errorf("Write() output should contain %v, got:\n%s", tt.contains, output)
			}
			if tt.notContains != "" && bytes.Contains([]byte(output), []byte(tt.notContains)) {
				t.Errorf("Write() output should not contain %v, got:\n%s", tt.notContains, output)
			}
		})
	}
}

// TestWriteJqOnRawData verifies --jq without --envelope filters the raw data.
func TestWriteJqOnRawData(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriterWithOptions(&IOStreams{Out: &buf}, FormatJSON, WithJq(".name"))
	data := map[string]interface{}{"name": "kmerp", "version": "1.0"}
	if err := w.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if got := bytes.TrimSpace(buf.Bytes()); string(got) != "kmerp" {
		t.Errorf("jq output = %q, want kmerp", got)
	}
}

// TestWriteJqOnEnvelope verifies that with --jq and --envelope together the
// expression runs against the full envelope (so '.data.name' works).
func TestWriteJqOnEnvelope(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriterWithOptions(&IOStreams{Out: &buf}, FormatJSON,
		WithEnvelope(true, "connector.info"),
		WithJq(".data.name"),
	)
	data := map[string]interface{}{"name": "kmerp"}
	if err := w.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if got := bytes.TrimSpace(buf.Bytes()); string(got) != "kmerp" {
		t.Errorf("jq-on-envelope output = %q, want kmerp", got)
	}

	// And '.ok' should be reachable through the envelope.
	buf.Reset()
	w = NewWriterWithOptions(&IOStreams{Out: &buf}, FormatJSON,
		WithEnvelope(true, "x"),
		WithJq(".ok"),
	)
	if err := w.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if got := bytes.TrimSpace(buf.Bytes()); string(got) != "true" {
		t.Errorf("jq .ok = %q, want true", got)
	}
}
