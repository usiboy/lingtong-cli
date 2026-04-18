// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
	"strings"
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

func TestNewCmdTableList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{
					{"id": 1, "name": "Orders"},
					{"id": 2, "name": "Customers"},
				},
				"total": 2,
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdTableList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdTableList_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Internal error"})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdTableList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdTableDataQuery_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 1, "name": "Item 1"},
				},
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdTableDataQuery(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--table-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataQuery_MissingTableId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdTableDataQuery(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --table-id")
	}
}

func TestNewCmdTableDataCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"id": 500},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdTableDataCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--table-id", "100", "--data", `{"name":"test"}`})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataCreate_MissingTableId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdTableDataCreate(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --table-id")
	}
}
