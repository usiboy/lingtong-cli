// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package cmdutil

import (
	"encoding/json"
	"strings"

	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// MissingHostConfigMessage is shown when no API host is configured.
const MissingHostConfigMessage = "no host configured. Run `lingtong-cli config init --host <url>` first"

// RequireHostConfigured returns a validation error (exit 2) when host is empty.
func RequireHostConfigured(host string) error {
	if strings.TrimSpace(host) == "" {
		return lterrors.NewValidationError(lterrors.SubtypeMissingField, MissingHostConfigMessage)
	}
	return nil
}

// RequireFlag returns a validation error (exit 2) when a required flag is unset.
func RequireFlag(ok bool, name string) error {
	if ok {
		return nil
	}
	return lterrors.NewValidationError(lterrors.SubtypeMissingField, "--%s is required", name)
}

// ConfirmDestructive gates a destructive operation behind --yes. Without it the
// command fails with exit code 10 (confirmation required), distinct from other
// errors so agents/scripts can detect the gate.
func ConfirmDestructive(yes bool, action string) error {
	if yes {
		return nil
	}
	return lterrors.NewConfirmationError("%s is destructive and was not confirmed", action).
		WithHint("re-run with --yes to proceed")
}

// ParseDataObject parses a --data JSON object, returning a validation error on
// malformed input.
func ParseDataObject(s string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, lterrors.NewValidationError(lterrors.SubtypeInvalidArgument,
			"--data must be a JSON object").WithCause(err)
	}
	return m, nil
}

// WriteResponse parses a raw API response and writes it through the factory's
// writer, sharing envelope / jq / omit-null with every other command. identity
// labels the command in envelope mode (e.g. "factory.list").
func (f *Factory) WriteResponse(cmd *cobra.Command, resp []byte, identity string) error {
	format := output.Format(cmd.Flag("format").Value.String())
	w := f.NewWriter(format, identity)
	var data interface{}
	if err := json.Unmarshal(resp, &data); err != nil {
		return err
	}
	return w.Write(data)
}
