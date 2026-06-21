// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	lterrors "github.com/lingtong/cli/internal/errors"
)

func TestEnvelopeMode(t *testing.T) {
	var buf bytes.Buffer
	ioStreams := &IOStreams{Out: &buf}
	w := NewWriterWithOptions(ioStreams, FormatJSON,
		WithEnvelope(true, "connector.info"),
		WithOmitNull(false),
	)

	data := map[string]interface{}{"name": "kmerp", "version": "1.0"}
	if err := w.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if !env.OK {
		t.Error("Envelope.OK = false, want true")
	}
	if env.Identity != "connector.info" {
		t.Errorf("Envelope.Identity = %q, want %q", env.Identity, "connector.info")
	}
	if env.Data == nil {
		t.Error("Envelope.Data = nil, want data")
	}
}

func TestEnvelopeModeDisabled(t *testing.T) {
	var buf bytes.Buffer
	ioStreams := &IOStreams{Out: &buf}
	w := NewWriterWithOptions(ioStreams, FormatJSON,
		WithEnvelope(false, ""),
		WithOmitNull(false),
	)

	data := map[string]interface{}{"name": "kmerp"}
	if err := w.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	// Should NOT contain envelope wrapper
	if strings.Contains(output, `"ok"`) {
		t.Errorf("output should not contain envelope when disabled: %s", output)
	}
}

func TestEnvelopeWithOmitNull(t *testing.T) {
	var buf bytes.Buffer
	ioStreams := &IOStreams{Out: &buf}
	w := NewWriterWithOptions(ioStreams, FormatJSON,
		WithEnvelope(true, "test"),
		WithOmitNull(true),
	)

	data := map[string]interface{}{"name": "kmerp", "empty": nil}
	if err := w.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, `"empty"`) {
		t.Errorf("output should not contain null fields: %s", output)
	}
	if !strings.Contains(output, `"ok"`) {
		t.Errorf("output should contain envelope: %s", output)
	}
}

func TestErrorDetail(t *testing.T) {
	detail := ErrorDetail{
		Type:    "validation",
		Subtype: "missing_field",
		Message: "field --name is required",
		Hint:    "Run: lingtong-cli ... --name <value>",
	}
	b, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"type":"validation"`) {
		t.Errorf("missing type field: %s", s)
	}
	if !strings.Contains(s, `"hint":"Run:`) {
		t.Errorf("missing hint field: %s", s)
	}
}

func TestErrorDetailFromError_Typed(t *testing.T) {
	err := lterrors.NewAPIError(lterrors.SubtypeNotFound, "not here").WithHint("check id")
	d := ErrorDetailFromError(err)
	if d.Type != "api" {
		t.Errorf("Type = %q, want api", d.Type)
	}
	if d.Subtype != "not_found" {
		t.Errorf("Subtype = %q, want not_found", d.Subtype)
	}
	if d.Message != "not here" || d.Hint != "check id" {
		t.Errorf("Message/Hint mismatch: %+v", d)
	}
}

func TestErrorDetailFromError_Untyped(t *testing.T) {
	d := ErrorDetailFromError(fmt.Errorf("boom"))
	if d.Type != "internal" {
		t.Errorf("Type = %q, want internal", d.Type)
	}
	if d.Message != "boom" {
		t.Errorf("Message = %q, want boom", d.Message)
	}
}

func TestWriteErrorEnvelope(t *testing.T) {
	var buf bytes.Buffer
	err := lterrors.NewValidationError(lterrors.SubtypeMissingField, "missing --scene-id").
		WithHint("pass --scene-id")
	if writeErr := WriteErrorEnvelope(&buf, "workflow.create", err); writeErr != nil {
		t.Fatalf("WriteErrorEnvelope() error = %v", writeErr)
	}

	var env Envelope
	if uErr := json.Unmarshal(buf.Bytes(), &env); uErr != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", uErr, buf.String())
	}
	if env.OK {
		t.Error("Envelope.OK = true, want false on error path")
	}
	if env.Identity != "workflow.create" {
		t.Errorf("Identity = %q", env.Identity)
	}
	// error field round-trips into a map; verify type + hint.
	s := buf.String()
	if !strings.Contains(s, `"type": "validation"`) {
		t.Errorf("missing validation type: %s", s)
	}
	if !strings.Contains(s, `"hint": "pass --scene-id"`) {
		t.Errorf("missing hint: %s", s)
	}
}
