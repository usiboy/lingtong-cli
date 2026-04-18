// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package scene

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

func TestNewCmdSceneList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"list": []map[string]interface{}{
					{"id": 1, "name": "Order Sync"},
					{"id": 2, "name": "Data Pipeline"},
				},
				"total": 2,
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdSceneList_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Internal server error",
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneList(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdSceneCreate_MissingName(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdSceneCreate(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error about name, got: %v", err)
	}
}

func TestNewCmdSceneCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"id":   100,
				"name": "Test Scene",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--name", "Test Scene", "--description", "Test description"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdSceneCreate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to create scene",
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneCreate(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--name", "Test Scene"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}

func TestNewCmdSceneInfo_MissingSceneId(t *testing.T) {
	f := newTestFactory("http://example.com")
	cmd := newCmdSceneInfo(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --scene-id")
	}
}

func TestNewCmdSceneInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"id":          100,
				"name":        "Test Scene",
				"description": "Test description",
			},
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--scene-id", "100"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCmdSceneInfo_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Scene not found",
		})
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	cmd := newCmdSceneInfo(f)
	cmd.Flags().String("format", "json", "Output format")
	cmd.SetArgs([]string{"--scene-id", "999"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	t.Logf("Result: err=%v", err)
}
