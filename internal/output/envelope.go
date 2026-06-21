// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"io"

	lterrors "github.com/lingtong/cli/internal/errors"
)

// Envelope is the standard response wrapper for all CLI output.
// When --envelope is enabled, every successful command output is wrapped
// in this structure so that AI Agents can reliably parse results.
type Envelope struct {
	OK       bool                   `json:"ok"`
	Identity string                 `json:"identity,omitempty"`
	Data     interface{}            `json:"data,omitempty"`
	Error    interface{}            `json:"error,omitempty"`
	Notice   map[string]interface{} `json:"_notice,omitempty"`
}

// ErrorDetail is the structured error payload inside an envelope.
type ErrorDetail struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype,omitempty"`
	Message   string `json:"message"`
	Hint      string `json:"hint,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
}

// ErrorDetailFromError builds an ErrorDetail from any error. Typed errors
// contribute their category, subtype, hint and retryable flag; untyped errors
// fall back to the "internal" type.
func ErrorDetailFromError(err error) ErrorDetail {
	if p, ok := lterrors.ProblemOf(err); ok {
		return ErrorDetail{
			Type:      string(p.Category),
			Subtype:   string(p.Subtype),
			Message:   p.Message,
			Hint:      p.Hint,
			Retryable: p.Retryable,
		}
	}
	return ErrorDetail{
		Type:    string(lterrors.CategoryInternal),
		Message: err.Error(),
	}
}

// WriteErrorEnvelope writes a failure envelope ({ok:false, error:{...}}) as
// indented JSON to w. Used when --envelope is enabled so AI agents receive a
// machine-readable error on the error path.
func WriteErrorEnvelope(w io.Writer, identity string, err error) error {
	env := Envelope{
		OK:       false,
		Identity: identity,
		Error:    ErrorDetailFromError(err),
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}
