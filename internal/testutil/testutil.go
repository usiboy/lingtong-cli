// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package testutil provides common testing utilities for lingtong-cli tests.
package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
)

// NewTestFactory creates a Factory for testing with a mock server URL.
// IOStreams.In is set to a no-op reader to prevent panics on input reads.
// IOStreams.Out and IOStreams.ErrOut are set to io.Discard.
func NewTestFactory(serverURL string) *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  serverURL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    io.Discard,
			ErrOut: io.Discard,
		},
	}
}

// NewTestFactoryWithOutput creates a Factory for testing with output capture.
// Returns the Factory and a strings.Builder to capture output.
func NewTestFactoryWithOutput(serverURL string) (*cmdutil.Factory, *strings.Builder) {
	out := &strings.Builder{}
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  serverURL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    out,
			ErrOut: io.Discard,
		},
	}, out
}

// NewMockServer creates an httptest.Server with the given handler.
// Returns the server and its URL. Caller must call server.Close() when done.
func NewMockServer(handler http.HandlerFunc) (*httptest.Server, string) {
	server := httptest.NewServer(handler)
	return server, server.URL
}

// NewMockServerWithAuth creates an httptest.Server that validates Bearer token authentication.
// The handler is only called if the Authorization header matches the expected token.
func NewMockServerWithAuth(expectedToken string, handler http.HandlerFunc) (*httptest.Server, string) {
	return NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + expectedToken
		if auth != expectedAuth {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		handler(w, r)
	})
}

// MockHandler is a helper to create common mock response patterns.
type MockHandler struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

// ServeHTTP implements http.Handler for MockHandler.
func (m MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for key, value := range m.Headers {
		w.Header().Set(key, value)
	}
	w.WriteHeader(m.StatusCode)
	w.Write([]byte(m.Body))
}

// SuccessHandler returns a MockHandler for successful responses (200 OK).
func SuccessHandler(body string) MockHandler {
	return MockHandler{
		StatusCode: http.StatusOK,
		Body:       body,
	}
}

// ServerErrorHandler returns a MockHandler for server errors (500 Internal Server Error).
func ServerErrorHandler(errorMsg string) MockHandler {
	return MockHandler{
		StatusCode: http.StatusInternalServerError,
		Body:       `{"error":"` + errorMsg + `"}`,
	}
}

// UnauthorizedHandler returns a MockHandler for unauthorized errors (401 Unauthorized).
func UnauthorizedHandler() MockHandler {
	return MockHandler{
		StatusCode: http.StatusUnauthorized,
		Body:       `{"error":"unauthorized"}`,
	}
}

// NotFoundHandler returns a MockHandler for not found errors (404 Not Found).
func NotFoundHandler() MockHandler {
	return MockHandler{
		StatusCode: http.StatusNotFound,
		Body:       `{"error":"not found"}`,
	}
}
