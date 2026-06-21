// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
)

func newTestFactory(envelope bool) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	f := &cmdutil.Factory{
		Config:    &config.Config{},
		IOStreams: &output.IOStreams{Out: out, ErrOut: errOut},
		Envelope:  envelope,
	}
	return f, out, errOut
}

func TestHandleRootError_ExitCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"validation (hand-written)", fmt.Errorf("--scene-id is required"), output.ExitValidation},
		{"auth", lterrors.NewAuthenticationError(lterrors.SubtypeTokenInvalid, "bad token"), output.ExitAuth},
		{"network", lterrors.NewNetworkError(lterrors.SubtypeNetworkTimeout, "slow"), output.ExitNetwork},
		{"api", lterrors.NewAPIError(lterrors.SubtypeServerError, "boom"), output.ExitAPI},
		{"internal/unknown", fmt.Errorf("something exploded"), output.ExitInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, _ := newTestFactory(false)
			if got := handleRootError(f, tt.err); got != tt.want {
				t.Errorf("exit code = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHandleRootError_HumanOutput(t *testing.T) {
	f, out, errOut := newTestFactory(false)
	err := lterrors.NewAuthenticationError(lterrors.SubtypeTokenExpired, "token expired").
		WithHint("run auth login")
	handleRootError(f, err)

	if out.Len() != 0 {
		t.Errorf("stdout should be empty in non-envelope mode, got: %s", out.String())
	}
	s := errOut.String()
	if !strings.Contains(s, "token expired") {
		t.Errorf("stderr missing message: %s", s)
	}
	if !strings.Contains(s, "Hint: run auth login") {
		t.Errorf("stderr missing hint: %s", s)
	}
}

func TestHandleRootError_EnvelopeOutput(t *testing.T) {
	f, out, errOut := newTestFactory(true)
	err := lterrors.NewValidationError(lterrors.SubtypeMissingField, "missing --name").
		WithHint("pass --name")
	code := handleRootError(f, err)

	if code != output.ExitValidation {
		t.Errorf("exit code = %d, want %d", code, output.ExitValidation)
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr should be empty in envelope mode, got: %s", errOut.String())
	}

	var env output.Envelope
	if uErr := json.Unmarshal(out.Bytes(), &env); uErr != nil {
		t.Fatalf("stdout not valid JSON: %v\n%s", uErr, out.String())
	}
	if env.OK {
		t.Error("envelope OK = true, want false")
	}
	s := out.String()
	if !strings.Contains(s, `"type": "validation"`) {
		t.Errorf("missing validation type: %s", s)
	}
}
