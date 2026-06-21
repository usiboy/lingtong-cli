// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package output

import (
	lterrors "github.com/lingtong/cli/internal/errors"
)

// Exit codes for the CLI process.
//
// Fine-grained error types (permission, not_found, rate_limit, etc.)
// are communicated via the JSON error envelope's "type" field,
// not via exit codes.
const (
	ExitOK                   = 0  // Success
	ExitAPI                  = 1  // API / general error
	ExitValidation           = 2  // Parameter validation failure
	ExitAuth                 = 3  // Authentication failure (invalid/expired token)
	ExitNetwork              = 4  // Network error (timeout, DNS, etc.)
	ExitInternal             = 5  // Internal error (should not happen)
	ExitContentSafety        = 6  // Content safety violation
	ExitConfirmationRequired = 10 // High-risk operation needs --yes confirmation
)

// ExitCodeOf returns the shell exit code for any error.
//   - Typed errors → routed by Category
//   - nil → ExitOK
//   - Untyped errors → ExitInternal
func ExitCodeOf(err error) int {
	if err == nil {
		return ExitOK
	}
	return ExitCodeForCategory(lterrors.CategoryOf(err))
}

// ExitCodeForCategory maps a Category to the corresponding shell exit code.
func ExitCodeForCategory(cat lterrors.Category) int {
	switch cat {
	case lterrors.CategoryValidation:
		return ExitValidation
	case lterrors.CategoryAuthentication:
		return ExitAuth
	case lterrors.CategoryNetwork:
		return ExitNetwork
	case lterrors.CategoryAPI:
		return ExitAPI
	case lterrors.CategoryConfirmation:
		return ExitConfirmationRequired
	case lterrors.CategoryInternal:
		return ExitInternal
	}
	return ExitInternal
}
