// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// IOStreams wraps the standard IO streams.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// Format represents the output format.
type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
	FormatPretty Format = "pretty"
)

// Writer handles output formatting.
type Writer struct {
	format   Format
	out      io.Writer
	errOut   io.Writer
	omitNull bool
}

// NewWriter creates a new output writer.
func NewWriter(ioStreams *IOStreams, format Format) *Writer {
	return &Writer{
		format: format,
		out:    ioStreams.Out,
		errOut: ioStreams.ErrOut,
	}
}

// NewWriterWithOpts creates a new output writer with options.
func NewWriterWithOpts(ioStreams *IOStreams, format Format, omitNull bool) *Writer {
	return &Writer{
		format:   format,
		out:      ioStreams.Out,
		errOut:   ioStreams.ErrOut,
		omitNull: omitNull,
	}
}

// Write outputs data in the specified format.
func (w *Writer) Write(data interface{}) error {
	// Remove null fields if omitNull is enabled
	if w.omitNull {
		data = removeNullFields(data)
	}

	switch w.format {
	case FormatJSON:
		return w.writeJSON(data)
	case FormatPretty:
		return w.writePretty(data)
	case FormatTable:
		return w.writeTable(data)
	default:
		return w.writeJSON(data)
	}
}

// removeNullFields recursively removes null values from JSON data.
func removeNullFields(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if value != nil {
				result[key] = removeNullFields(value)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			result = append(result, removeNullFields(item))
		}
		return result
	default:
		return data
	}
}

func (w *Writer) writeJSON(data interface{}) error {
	encoder := json.NewEncoder(w.out)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data)
}

func (w *Writer) writePretty(data interface{}) error {
	// TODO: Implement pretty formatting based on data type
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w.out, string(b))
	return nil
}

func (w *Writer) writeTable(data interface{}) error {
	// TODO: Implement table output
	return w.writeJSON(data)
}

// WriteError writes an error to stderr.
func (w *Writer) WriteError(err error) {
	fmt.Fprintln(w.errOut, "Error:", err)
}
