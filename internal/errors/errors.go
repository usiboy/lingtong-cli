// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package errors provides a minimal typed error system for lingtong-cli.
//
// Every typed error carries a Category (top-level classification) and an
// optional Subtype (finer-grained label). All typed errors implement the
// TypedError interface so that callers can extract a Problem detail via
// errors.As.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Category — top-level error classification
// ---------------------------------------------------------------------------

// Category is the top-level error classification.
type Category string

const (
	CategoryValidation     Category = "validation"
	CategoryAuthentication Category = "authentication"
	CategoryNetwork        Category = "network"
	CategoryAPI            Category = "api"
	CategoryInternal       Category = "internal"
	CategoryConfirmation   Category = "confirmation"
)

// ---------------------------------------------------------------------------
// Subtype — second-level classification
// ---------------------------------------------------------------------------

// Subtype provides finer-grained error classification.
type Subtype string

const (
	// Validation subtypes
	SubtypeInvalidArgument Subtype = "invalid_argument"
	SubtypeMissingField    Subtype = "missing_field"

	// Authentication subtypes
	SubtypeTokenMissing Subtype = "token_missing"
	SubtypeTokenExpired Subtype = "token_expired"
	SubtypeTokenInvalid Subtype = "token_invalid"

	// Network subtypes
	SubtypeNetworkTimeout Subtype = "timeout"
	SubtypeNetworkDNS     Subtype = "dns"
	SubtypeNetworkConnRef Subtype = "connection_refused"

	// API subtypes
	SubtypeServerError Subtype = "server_error"
	SubtypeNotFound    Subtype = "not_found"
	SubtypeConflict    Subtype = "conflict"
	SubtypeRateLimit   Subtype = "rate_limit"

	// Generic
	SubtypeUnknown Subtype = "unknown"
)

// ---------------------------------------------------------------------------
// Problem — RFC 7807–inspired detail object
// ---------------------------------------------------------------------------

// Problem carries machine-readable details about an error.
type Problem struct {
	Category  Category `json:"type"`
	Subtype   Subtype  `json:"subtype,omitempty"`
	Message   string   `json:"message"`
	Hint      string   `json:"hint,omitempty"`
	Retryable bool     `json:"retryable,omitempty"`
}

// Error implements the error interface.
func (p *Problem) Error() string { return p.Message }

// ProblemDetail returns itself.
func (p *Problem) ProblemDetail() *Problem { return p }

// WithHint sets a human-readable hint and returns the receiver.
func (p *Problem) WithHint(hint string) *Problem {
	p.Hint = hint
	return p
}

// ---------------------------------------------------------------------------
// TypedError interface
// ---------------------------------------------------------------------------

// TypedError is implemented by all typed errors in this package.
type TypedError interface {
	error
	ProblemDetail() *Problem
}

// ---------------------------------------------------------------------------
// Concrete error types
// ---------------------------------------------------------------------------

// ValidationError represents a parameter validation failure.
type ValidationError struct {
	Problem
	Cause error
}

func (e *ValidationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}
func (e *ValidationError) ProblemDetail() *Problem { return &e.Problem }
func (e *ValidationError) Unwrap() error           { return e.Cause }

// WithHint sets a human-readable hint and returns the receiver.
func (e *ValidationError) WithHint(hint string) *ValidationError {
	e.Hint = hint
	return e
}

// WithCause sets the underlying cause and returns the receiver.
func (e *ValidationError) WithCause(cause error) *ValidationError {
	e.Cause = cause
	return e
}

// AuthenticationError represents an authentication / authorization failure.
type AuthenticationError struct {
	Problem
	Cause error
}

func (e *AuthenticationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}
func (e *AuthenticationError) ProblemDetail() *Problem { return &e.Problem }
func (e *AuthenticationError) Unwrap() error           { return e.Cause }

func (e *AuthenticationError) WithHint(hint string) *AuthenticationError {
	e.Hint = hint
	return e
}

func (e *AuthenticationError) WithCause(cause error) *AuthenticationError {
	e.Cause = cause
	return e
}

// NetworkError represents a network-level failure.
type NetworkError struct {
	Problem
	Cause error
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}
func (e *NetworkError) ProblemDetail() *Problem { return &e.Problem }
func (e *NetworkError) Unwrap() error           { return e.Cause }

func (e *NetworkError) WithHint(hint string) *NetworkError {
	e.Hint = hint
	return e
}

func (e *NetworkError) WithCause(cause error) *NetworkError {
	e.Cause = cause
	return e
}

// APIError represents a remote API error (non-auth, non-network).
type APIError struct {
	Problem
	Cause error
}

func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}
func (e *APIError) ProblemDetail() *Problem { return &e.Problem }
func (e *APIError) Unwrap() error           { return e.Cause }

func (e *APIError) WithHint(hint string) *APIError {
	e.Hint = hint
	return e
}

func (e *APIError) WithCause(cause error) *APIError {
	e.Cause = cause
	return e
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// NewValidationError creates a new ValidationError.
func NewValidationError(subtype Subtype, format string, args ...interface{}) *ValidationError {
	return &ValidationError{
		Problem: Problem{
			Category: CategoryValidation,
			Subtype:  subtype,
			Message:  fmt.Sprintf(format, args...),
		},
	}
}

// NewAuthenticationError creates a new AuthenticationError.
func NewAuthenticationError(subtype Subtype, format string, args ...interface{}) *AuthenticationError {
	return &AuthenticationError{
		Problem: Problem{
			Category: CategoryAuthentication,
			Subtype:  subtype,
			Message:  fmt.Sprintf(format, args...),
		},
	}
}

// NewNetworkError creates a new NetworkError.
func NewNetworkError(subtype Subtype, format string, args ...interface{}) *NetworkError {
	return &NetworkError{
		Problem: Problem{
			Category: CategoryNetwork,
			Subtype:  subtype,
			Message:  fmt.Sprintf(format, args...),
		},
	}
}

// NewAPIError creates a new APIError.
func NewAPIError(subtype Subtype, format string, args ...interface{}) *APIError {
	return &APIError{
		Problem: Problem{
			Category: CategoryAPI,
			Subtype:  subtype,
			Message:  fmt.Sprintf(format, args...),
		},
	}
}

// NewConfirmationError creates an error for a high-risk operation that requires
// explicit confirmation (e.g. --yes). It maps to exit code 10 so AI agents and
// scripts can detect "needs confirmation" distinctly from other failures.
func NewConfirmationError(format string, args ...interface{}) *Problem {
	return &Problem{
		Category: CategoryConfirmation,
		Message:  fmt.Sprintf(format, args...),
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// ProblemOf extracts the Problem from any error that implements TypedError.
// Returns (problem, true) on success, (nil, false) otherwise.
func ProblemOf(err error) (*Problem, bool) {
	if err == nil {
		return nil, false
	}
	var te TypedError
	if errors.As(err, &te) {
		return te.ProblemDetail(), true
	}
	return nil, false
}

// CategoryOf returns the Category of a typed error.
// Non-typed errors return CategoryInternal.
func CategoryOf(err error) Category {
	if p, ok := ProblemOf(err); ok {
		return p.Category
	}
	return CategoryInternal
}

// validationMarkers are substrings that identify a parameter/usage error.
// They cover both cobra's own flag/argument messages and the project's
// hand-written "--x is required" / "invalid ..." checks (English + 中文).
var validationMarkers = []string{
	"is required",
	"required flag",
	"unknown flag",
	"unknown shorthand flag",
	"unknown command",
	"invalid argument",
	"invalid value",
	"accepts ",
	"flag needs an argument",
	"must be",
	"缺少",
	"必填",
	"参数错误",
	"无效的参数",
}

// Classify upgrades an arbitrary error into a typed error so that exit codes
// and envelopes stay meaningful even for call sites that have not yet adopted
// typed errors.
//
//   - nil                  → nil
//   - already TypedError   → returned unchanged
//   - looks like a usage   → wrapped as a ValidationError (exit 2)
//   - anything else        → wrapped as CategoryInternal (exit 5)
//
// Typed errors from the client (network/auth/api) flow through untouched; this
// is primarily a safety net for cobra flag errors and hand-written validation.
func Classify(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := ProblemOf(err); ok {
		return err
	}

	lower := strings.ToLower(err.Error())
	for _, marker := range validationMarkers {
		if strings.Contains(lower, strings.ToLower(marker)) {
			return &ValidationError{
				Problem: Problem{
					Category: CategoryValidation,
					Subtype:  SubtypeInvalidArgument,
					Message:  err.Error(),
				},
			}
		}
	}

	return &Problem{
		Category: CategoryInternal,
		Subtype:  SubtypeUnknown,
		Message:  err.Error(),
	}
}
