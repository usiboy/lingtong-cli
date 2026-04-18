// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package model

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
)

func newTestFactory(serverURL string) *cmdutil.Factory {
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

func TestNewCmdModelInterfaceList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{
					{"id": 1, "name": "GetOrder"},
					{"id": 2, "name": "CreateOrder"},
				},
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdModelInterfaceList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--filter-model-type", "domain"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdModelInterfaceList_MissingConnector(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdModelInterfaceList(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --connector")
	}
}

func TestNewCmdModelInterfaceList_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Internal error"})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdModelInterfaceList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--filter-model-type", "domain"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdModelDomainGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"domain": "Order Management",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdModelDomainGet(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--business", "order"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdModelDomainGet_MissingParams(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdModelDomainGet(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing params")
	}
}

func TestNewCmdModelDynamicView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"model": "Order",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdModelDynamicView(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--connector", "kmerp", "--auth-account-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdModelDynamicView_MissingParams(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdModelDynamicView(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing params")
	}
}
