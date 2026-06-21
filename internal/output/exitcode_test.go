// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	"fmt"
	"testing"

	lterrors "github.com/lingtong/cli/internal/errors"
)

func TestExitCodeOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, ExitOK},
		{"validation", lterrors.NewValidationError(lterrors.SubtypeMissingField, "x"), ExitValidation},
		{"auth", lterrors.NewAuthenticationError(lterrors.SubtypeTokenInvalid, "x"), ExitAuth},
		{"network", lterrors.NewNetworkError(lterrors.SubtypeNetworkTimeout, "x"), ExitNetwork},
		{"api", lterrors.NewAPIError(lterrors.SubtypeServerError, "x"), ExitAPI},
		{"plain", fmt.Errorf("plain error"), ExitInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCodeOf(tt.err)
			if got != tt.want {
				t.Errorf("ExitCodeOf() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExitCodeOfWrapped(t *testing.T) {
	inner := lterrors.NewValidationError(lterrors.SubtypeMissingField, "inner")
	wrapped := fmt.Errorf("outer: %w", inner)
	if got := ExitCodeOf(wrapped); got != ExitValidation {
		t.Errorf("ExitCodeOf(wrapped validation) = %d, want %d", got, ExitValidation)
	}
}

func TestExitCodeForCategory(t *testing.T) {
	tests := []struct {
		cat  lterrors.Category
		want int
	}{
		{lterrors.CategoryValidation, ExitValidation},
		{lterrors.CategoryAuthentication, ExitAuth},
		{lterrors.CategoryNetwork, ExitNetwork},
		{lterrors.CategoryAPI, ExitAPI},
		{lterrors.CategoryConfirmation, ExitConfirmationRequired},
		{lterrors.CategoryInternal, ExitInternal},
		{lterrors.Category("unknown"), ExitInternal},
	}
	for _, tt := range tests {
		t.Run(string(tt.cat), func(t *testing.T) {
			got := ExitCodeForCategory(tt.cat)
			if got != tt.want {
				t.Errorf("ExitCodeForCategory(%v) = %d, want %d", tt.cat, got, tt.want)
			}
		})
	}
}
