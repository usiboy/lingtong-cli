// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/itchyny/gojq"
)

// JqFilter applies a jq expression to data and writes the results to w.
// Scalar values are printed raw (no quotes for strings), matching jq -r behavior.
// Complex values (maps, arrays) are printed as indented JSON.
//
// Runtime errors (e.g. iterating over null) produce a friendly hint instead of
// a raw gojq error, so users can quickly discover the correct data path.
func JqFilter(w io.Writer, data interface{}, expr string) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid jq expression: %w", err)
	}
	code, err := gojq.Compile(query)
	if err != nil {
		return fmt.Errorf("invalid jq expression: %w", err)
	}

	// Normalize data through toGeneric so typed structs become map[string]any.
	normalized := toGeneric(data)
	// Convert json.Number values to gojq-compatible types.
	normalized = convertNumbers(normalized)

	iter := code.Run(normalized)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if runErr, isErr := v.(error); isErr {
			return enhanceJqError(runErr, expr, normalized)
		}
		if err := writeJqValue(w, v); err != nil {
			return err
		}
	}
	return nil
}

// enhanceJqError converts raw gojq runtime errors into user-friendly messages.
func enhanceJqError(err error, expr string, data interface{}) error {
	msg := err.Error()

	// Detect common "cannot iterate over" errors and provide guidance.
	if strings.Contains(msg, "cannot iterate over") {
		hint := buildDataHint(data)
		return fmt.Errorf("jq error: %w\nHint: %s\nTry '.' first to see the full data structure, or use '.path?[]' to skip nulls", err, hint)
	}

	// Detect type mismatch errors (e.g. "expected an object but got: array")
	if strings.Contains(msg, "expected an") && strings.Contains(msg, "but got") {
		hint := buildDataHint(data)
		return fmt.Errorf("jq error: %w\nHint: %s\nThe data structure may differ from what the jq expression expects", err, hint)
	}

	return fmt.Errorf("jq error: %w", err)
}

// buildDataHint returns a hint string showing the top-level keys of the data.
func buildDataHint(data interface{}) string {
	if m, ok := data.(map[string]interface{}); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		if len(keys) > 0 {
			return fmt.Sprintf("top-level keys are: %s", strings.Join(keys, ", "))
		}
	}
	return "the target field may be null or have a different path"
}

// ValidateJqExpression checks whether a jq expression is syntactically valid.
func ValidateJqExpression(expr string) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid jq expression: %w", err)
	}
	_, err = gojq.Compile(query)
	if err != nil {
		return fmt.Errorf("invalid jq expression: %w", err)
	}
	return nil
}

// toGeneric normalizes any value to a plain map/slice representation via JSON round-trip.
func toGeneric(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var result interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		return v
	}
	return result
}

// convertNumbers recursively converts json.Number values to int or float64
// so that gojq can process them correctly.
func convertNumbers(v interface{}) interface{} {
	switch val := v.(type) {
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return int(i)
		}
		if f, err := val.Float64(); err == nil {
			return f
		}
		return val.String()
	case map[string]interface{}:
		for k, elem := range val {
			val[k] = convertNumbers(elem)
		}
		return val
	case []interface{}:
		for i, elem := range val {
			val[i] = convertNumbers(elem)
		}
		return val
	default:
		return v
	}
}

// writeJqValue writes a single jq result value to w.
// Scalars are printed raw; complex values as indented JSON.
func writeJqValue(w io.Writer, v interface{}) error {
	switch val := v.(type) {
	case nil:
		fmt.Fprintln(w, "null")
	case bool:
		fmt.Fprintln(w, val)
	case int:
		fmt.Fprintln(w, val)
	case float64:
		fmt.Fprintf(w, "%g\n", val)
	case *big.Int:
		fmt.Fprintln(w, val.String())
	case string:
		// Raw output for strings (no quotes), matching jq -r.
		fmt.Fprintln(w, val)
	default:
		// Complex value (map, array): indented JSON.
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			return fmt.Errorf("failed to marshal jq result: %w", err)
		}
	}
	return nil
}
