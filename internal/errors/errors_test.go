// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError(SubtypeMissingField, "field %s is required", "name")
	if err.Category != CategoryValidation {
		t.Errorf("Category = %v, want %v", err.Category, CategoryValidation)
	}
	if err.Subtype != SubtypeMissingField {
		t.Errorf("Subtype = %v, want %v", err.Subtype, SubtypeMissingField)
	}
	if err.Error() != "field name is required" {
		t.Errorf("Error() = %q, want %q", err.Error(), "field name is required")
	}
}

func TestNewAuthenticationError(t *testing.T) {
	err := NewAuthenticationError(SubtypeTokenExpired, "token expired")
	if err.Category != CategoryAuthentication {
		t.Errorf("Category = %v, want %v", err.Category, CategoryAuthentication)
	}
	if err.Subtype != SubtypeTokenExpired {
		t.Errorf("Subtype = %v, want %v", err.Subtype, SubtypeTokenExpired)
	}
}

func TestNewNetworkError(t *testing.T) {
	err := NewNetworkError(SubtypeNetworkTimeout, "connection timed out after %ds", 30)
	if err.Category != CategoryNetwork {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNetwork)
	}
	if err.Subtype != SubtypeNetworkTimeout {
		t.Errorf("Subtype = %v, want %v", err.Subtype, SubtypeNetworkTimeout)
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(SubtypeNotFound, "resource %d not found", 42)
	if err.Category != CategoryAPI {
		t.Errorf("Category = %v, want %v", err.Category, CategoryAPI)
	}
	if err.Subtype != SubtypeNotFound {
		t.Errorf("Subtype = %v, want %v", err.Subtype, SubtypeNotFound)
	}
}

func TestWithHint(t *testing.T) {
	err := NewValidationError(SubtypeMissingField, "missing --name").
		WithHint("Run: lingtong-cli config init --name <value>")
	if err.Hint != "Run: lingtong-cli config init --name <value>" {
		t.Errorf("Hint = %q, want hint text", err.Hint)
	}
}

func TestWithCause(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewAPIError(SubtypeServerError, "request failed").
		WithCause(cause)
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is(err, cause) = false, want true")
	}
}

func TestProblemOf(t *testing.T) {
	// Typed error
	typed := NewValidationError(SubtypeMissingField, "missing field")
	p, ok := ProblemOf(typed)
	if !ok {
		t.Fatal("ProblemOf(typed) = false, want true")
	}
	if p.Category != CategoryValidation {
		t.Errorf("Category = %v, want %v", p.Category, CategoryValidation)
	}

	// Plain error
	plain := fmt.Errorf("plain error")
	p, ok = ProblemOf(plain)
	if ok {
		t.Error("ProblemOf(plain) = true, want false")
	}
	if p != nil {
		t.Errorf("Problem = %v, want nil", p)
	}

	// nil error
	p, ok = ProblemOf(nil)
	if ok {
		t.Error("ProblemOf(nil) = true, want false")
	}

	// Wrapped typed error
	wrapped := fmt.Errorf("wrapped: %w", typed)
	p, ok = ProblemOf(wrapped)
	if !ok {
		t.Fatal("ProblemOf(wrapped typed) = false, want true")
	}
	if p.Category != CategoryValidation {
		t.Errorf("Category = %v, want %v", p.Category, CategoryValidation)
	}
}

func TestCategoryOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Category
	}{
		{"validation", NewValidationError(SubtypeMissingField, "x"), CategoryValidation},
		{"auth", NewAuthenticationError(SubtypeTokenInvalid, "x"), CategoryAuthentication},
		{"network", NewNetworkError(SubtypeNetworkTimeout, "x"), CategoryNetwork},
		{"api", NewAPIError(SubtypeServerError, "x"), CategoryAPI},
		{"plain", fmt.Errorf("plain"), CategoryInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CategoryOf(tt.err)
			if got != tt.want {
				t.Errorf("CategoryOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypedErrorInterface(t *testing.T) {
	// Ensure all types implement TypedError
	var _ TypedError = &ValidationError{}
	var _ TypedError = &AuthenticationError{}
	var _ TypedError = &NetworkError{}
	var _ TypedError = &APIError{}
}

func TestErrorsAsCompatibility(t *testing.T) {
	inner := NewValidationError(SubtypeMissingField, "inner")
	wrapped := fmt.Errorf("outer: %w", inner)

	var ve *ValidationError
	if !errors.As(wrapped, &ve) {
		t.Fatal("errors.As failed to unwrap ValidationError")
	}
	if ve.Subtype != SubtypeMissingField {
		t.Errorf("Subtype = %v, want %v", ve.Subtype, SubtypeMissingField)
	}
}

func TestErrorWithCauseFormatting(t *testing.T) {
	cause := fmt.Errorf("boom")
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"validation", NewValidationError(SubtypeInvalidArgument, "bad").WithCause(cause), "bad: boom"},
		{"auth", NewAuthenticationError(SubtypeTokenInvalid, "denied").WithCause(cause), "denied: boom"},
		{"network", NewNetworkError(SubtypeNetworkTimeout, "slow").WithCause(cause), "slow: boom"},
		{"api", NewAPIError(SubtypeServerError, "fail").WithCause(cause), "fail: boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProblemDetailAndUnwrap(t *testing.T) {
	cause := fmt.Errorf("root")
	auth := NewAuthenticationError(SubtypeTokenExpired, "expired").WithHint("login").WithCause(cause)
	if auth.ProblemDetail().Category != CategoryAuthentication {
		t.Error("auth ProblemDetail category mismatch")
	}
	if auth.Hint != "login" {
		t.Errorf("auth Hint = %q", auth.Hint)
	}
	if !errors.Is(auth, cause) {
		t.Error("auth should unwrap to cause")
	}

	net := NewNetworkError(SubtypeNetworkDNS, "dns").WithHint("check dns")
	if net.ProblemDetail().Subtype != SubtypeNetworkDNS || net.Hint != "check dns" {
		t.Error("network ProblemDetail/hint mismatch")
	}

	api := NewAPIError(SubtypeNotFound, "missing").WithHint("verify id")
	if api.ProblemDetail().Category != CategoryAPI || api.Hint != "verify id" {
		t.Error("api ProblemDetail/hint mismatch")
	}

	// Problem itself implements TypedError.
	var p TypedError = &Problem{Category: CategoryInternal, Message: "x"}
	if p.ProblemDetail().Category != CategoryInternal || p.Error() != "x" {
		t.Error("Problem TypedError behaviour mismatch")
	}
}

func TestConfirmationError(t *testing.T) {
	err := NewConfirmationError("delete %d records", 3).WithHint("re-run with --yes")
	if err.Category != CategoryConfirmation {
		t.Errorf("Category = %v, want %v", err.Category, CategoryConfirmation)
	}
	if err.Error() != "delete 3 records" {
		t.Errorf("Error() = %q", err.Error())
	}
	if err.Hint != "re-run with --yes" {
		t.Errorf("Hint = %q", err.Hint)
	}
	if CategoryOf(err) != CategoryConfirmation {
		t.Errorf("CategoryOf = %v", CategoryOf(err))
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCat  Category
		wantSame bool // expect the same error instance returned
	}{
		{"nil", nil, "", false},
		{"already typed", NewAPIError(SubtypeServerError, "x"), CategoryAPI, true},
		{"english required", fmt.Errorf("--connector is required"), CategoryValidation, false},
		{"cobra required flag", fmt.Errorf(`required flag(s) "scene-id" not set`), CategoryValidation, false},
		{"cobra unknown flag", fmt.Errorf("unknown flag: --foo"), CategoryValidation, false},
		{"cobra accepts args", fmt.Errorf("accepts 1 arg(s), received 0"), CategoryValidation, false},
		{"chinese missing", fmt.Errorf("缺少必填参数: --scene-id"), CategoryValidation, false},
		{"must be", fmt.Errorf("--data must be a JSON object"), CategoryValidation, false},
		{"generic internal", fmt.Errorf("something exploded"), CategoryInternal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.err)
			if tt.err == nil {
				if got != nil {
					t.Fatalf("Classify(nil) = %v, want nil", got)
				}
				return
			}
			if tt.wantSame && got != tt.err {
				t.Errorf("Classify should return the same typed error instance")
			}
			if CategoryOf(got) != tt.wantCat {
				t.Errorf("CategoryOf(Classify()) = %v, want %v", CategoryOf(got), tt.wantCat)
			}
		})
	}
}
