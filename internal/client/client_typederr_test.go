// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	lterrors "github.com/lingtong/cli/internal/errors"
)

// TestHTTPErrorsAreTyped verifies that non-2xx responses map to the correct
// typed error category so the CLI returns the right exit code.
func TestHTTPErrorsAreTyped(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantCat lterrors.Category
		wantSub lterrors.Subtype
	}{
		{"unauthorized", http.StatusUnauthorized, lterrors.CategoryAuthentication, lterrors.SubtypeTokenInvalid},
		{"forbidden", http.StatusForbidden, lterrors.CategoryAuthentication, lterrors.SubtypeTokenInvalid},
		{"not found", http.StatusNotFound, lterrors.CategoryAPI, lterrors.SubtypeNotFound},
		{"conflict", http.StatusConflict, lterrors.CategoryAPI, lterrors.SubtypeConflict},
		{"rate limit", http.StatusTooManyRequests, lterrors.CategoryAPI, lterrors.SubtypeRateLimit},
		{"server error", http.StatusInternalServerError, lterrors.CategoryAPI, lterrors.SubtypeServerError},
		{"bad request", http.StatusBadRequest, lterrors.CategoryAPI, lterrors.SubtypeUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{"error":"x"}`))
			}))
			defer srv.Close()

			c := NewClient(srv.URL, "tok")
			_, err := c.Do("GET", "/x", nil, nil)
			if err == nil {
				t.Fatal("expected error")
			}
			p, ok := lterrors.ProblemOf(err)
			if !ok {
				t.Fatalf("error is not typed: %v", err)
			}
			if p.Category != tt.wantCat {
				t.Errorf("Category = %v, want %v", p.Category, tt.wantCat)
			}
			if p.Subtype != tt.wantSub {
				t.Errorf("Subtype = %v, want %v", p.Subtype, tt.wantSub)
			}
		})
	}
}

// TestNetworkErrorIsTyped verifies a transport failure becomes a NetworkError.
func TestNetworkErrorIsTyped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // force connection refused

	c := NewClient(url, "")
	_, err := c.Do("GET", "/x", nil, nil)
	if err == nil {
		t.Fatal("expected network error")
	}
	if lterrors.CategoryOf(err) != lterrors.CategoryNetwork {
		t.Errorf("Category = %v, want network", lterrors.CategoryOf(err))
	}
}
