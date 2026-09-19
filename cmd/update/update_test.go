// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package update

import (
	"bytes"
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

func newTestFactory() (*cmdutil.Factory, *bytes.Buffer) {
	var buf bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{Host: "https://test.example.com", OmitNull: true},
		IOStreams: &output.IOStreams{
			In:     &bytes.Buffer{},
			Out:    &buf,
			ErrOut: io.Discard,
		},
	}
	return f, &buf
}

func TestBuildUpdateResult(t *testing.T) {
	// Update needed.
	result := buildUpdateResult("1.0.0", "1.1.0", true)
	if result["currentVersion"] != "1.0.0" {
		t.Errorf("currentVersion = %v, want 1.0.0", result["currentVersion"])
	}
	if result["latestVersion"] != "1.1.0" {
		t.Errorf("latestVersion = %v, want 1.1.0", result["latestVersion"])
	}
	if result["upToDate"] != false {
		t.Errorf("upToDate = %v, want false", result["upToDate"])
	}
	if result["installInstructions"] == nil {
		t.Error("installInstructions should not be nil when update needed")
	}
	if result["releaseURL"] == "" {
		t.Error("releaseURL should not be empty when update needed")
	}

	// Up to date.
	result = buildUpdateResult("1.1.0", "1.1.0", false)
	if result["upToDate"] != true {
		t.Errorf("upToDate = %v, want true", result["upToDate"])
	}
	if result["installInstructions"] != nil {
		t.Error("installInstructions should be nil when up to date")
	}
}

func TestGetInstallInstructions(t *testing.T) {
	instructions := getInstallInstructions("v1.3.0")
	methods := map[string]bool{}
	npmCommand := ""
	for _, inst := range instructions {
		methods[inst["method"]] = true
		if inst["method"] == "npm" {
			npmCommand = inst["command"]
		}
		if inst["command"] == "" {
			t.Errorf("command for %q should not be empty", inst["method"])
		}
	}
	for _, m := range []string{"npm", "go install", "binary"} {
		if !methods[m] {
			t.Errorf("missing install method %q", m)
		}
	}
	if npmCommand != "npm install -g @lingtong-cli/cli" {
		t.Errorf("npm command = %q, want scoped package", npmCommand)
	}
}

func TestResolveCheckURL(t *testing.T) {
	const publicReleaseURL = "https://api.github.com/repos/usiboy/lingtong-cli/releases/latest"
	if CheckURL != publicReleaseURL {
		t.Fatalf("CheckURL = %q, want %q", CheckURL, publicReleaseURL)
	}
	if got := resolveCheckURL(); got != CheckURL {
		t.Errorf("resolveCheckURL() = %q, want default %q", got, CheckURL)
	}
	t.Setenv("LINGTONG_UPDATE_URL", "https://override.example.com/x")
	if got := resolveCheckURL(); got != "https://override.example.com/x" {
		t.Errorf("resolveCheckURL() with env = %q, want override", got)
	}
}

func TestFetchLatestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v2.5.0"})
	}))
	defer srv.Close()

	got, err := fetchLatestVersion(srv.URL)
	if err != nil {
		t.Fatalf("fetchLatestVersion error = %v", err)
	}
	if got != "v2.5.0" {
		t.Errorf("got %q, want v2.5.0", got)
	}
}

func TestFetchLatestVersion_Errors(t *testing.T) {
	// Non-200 status.
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()
	if _, err := fetchLatestVersion(bad.URL); err == nil {
		t.Error("expected error on 500 status")
	}

	// Empty tag.
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":""}`))
	}))
	defer empty.Close()
	if _, err := fetchLatestVersion(empty.URL); err == nil {
		t.Error("expected error on empty tag")
	}

	// Unreachable host.
	if _, err := fetchLatestVersion("http://127.0.0.1:0"); err == nil {
		t.Error("expected error on unreachable host")
	}
}

func TestNewCmdUpdate_TargetVersion(t *testing.T) {
	f, buf := newTestFactory()
	cmd := NewCmdUpdate(f)
	cmd.SetArgs([]string{"--version", "v9.9.9"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute error = %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if out["latestVersion"] != "v9.9.9" {
		t.Errorf("latestVersion = %v, want v9.9.9", out["latestVersion"])
	}
}

func TestNewCmdUpdate_LiveCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v100.0.0"}`))
	}))
	defer srv.Close()
	t.Setenv("LINGTONG_UPDATE_URL", srv.URL)

	f, buf := newTestFactory()
	cmd := NewCmdUpdate(f)
	cmd.SetArgs([]string{"--check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute error = %v", err)
	}
	if !strings.Contains(buf.String(), "v100.0.0") {
		t.Errorf("expected latest version in output, got: %s", buf.String())
	}
}

func TestNewCmdUpdate_CheckError(t *testing.T) {
	t.Setenv("LINGTONG_UPDATE_URL", "http://127.0.0.1:0")
	f, buf := newTestFactory()
	cmd := NewCmdUpdate(f)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute should degrade gracefully, got error = %v", err)
	}
	if !strings.Contains(buf.String(), "checkError") {
		t.Errorf("expected checkError in graceful-degradation output, got: %s", buf.String())
	}
}

func TestNewCmdUpdate_Structure(t *testing.T) {
	f, _ := newTestFactory()
	cmd := NewCmdUpdate(f)
	if cmd.Use != "update" {
		t.Errorf("Use = %q, want update", cmd.Use)
	}
	for _, flag := range []string{"check", "version", "format"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("missing flag --%s", flag)
		}
	}
}
