// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package config

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	ltconfig "github.com/lingtong/cli/internal/config"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func newTestFactory() *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &ltconfig.Config{
			Host:     "https://default.example.com",
			Brand:    "lingtong",
			OmitNull: true,
			Profiles: map[string]ltconfig.Profile{
				"dev":  {Host: "https://dev.example.com", Brand: "lingtong"},
				"prod": {Host: "https://prod.example.com", Brand: "lingtong"},
			},
		},
		IOStreams: &output.IOStreams{
			In:     &bytes.Buffer{},
			Out:    io.Discard,
			ErrOut: io.Discard,
		},
	}
}

func TestNewCmdProfile(t *testing.T) {
	f := newTestFactory()
	cmd := NewCmdProfile(f)

	if cmd.Use != "profile" {
		t.Errorf("Use = %q, want %q", cmd.Use, "profile")
	}

	// Check subcommands
	subCmds := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	expected := []string{"list", "current", "use", "add", "remove"}
	for _, name := range expected {
		if !subCmds[name] {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

func TestProfileList(t *testing.T) {
	f := newTestFactory()
	cmd := NewCmdProfile(f)

	// Find the list subcommand
	var listCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "list" {
			listCmd = sub
			break
		}
	}
	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	listCmd.SetOut(&buf)
	listCmd.SetErr(&buf)

	if err := listCmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestProfileCurrent(t *testing.T) {
	f := newTestFactory()
	f.Config.CurrentProfile = "dev"

	cmd := NewCmdProfile(f)
	var currentCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "current" {
			currentCmd = sub
			break
		}
	}
	if currentCmd == nil {
		t.Fatal("current subcommand not found")
	}

	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	currentCmd.SetOut(&buf)
	currentCmd.SetErr(&buf)

	if err := currentCmd.Execute(); err != nil {
		t.Fatalf("current command failed: %v", err)
	}
}

func TestProfileUse(t *testing.T) {
	f := newTestFactory()

	// Test the core logic directly: setting CurrentProfile
	f.Config.CurrentProfile = "dev"
	if f.Config.CurrentProfile != "dev" {
		t.Errorf("CurrentProfile = %q, want %q", f.Config.CurrentProfile, "dev")
	}

	// Test ResolveProfile after setting
	host, _ := f.Config.ResolveProfile("")
	if host != "https://dev.example.com" {
		t.Errorf("ResolveProfile after use = %q, want %q", host, "https://dev.example.com")
	}
}

func TestProfileUseNotFound(t *testing.T) {
	f := newTestFactory()

	// Test that resolving a nonexistent profile falls back to top-level
	host, _ := f.Config.ResolveProfile("nonexistent")
	if host != "https://default.example.com" {
		t.Errorf("ResolveProfile(nonexistent) = %q, want %q", host, "https://default.example.com")
	}

	// Test GetProfile returns nil for nonexistent
	p := f.Config.GetProfile("nonexistent")
	if p != nil {
		t.Error("GetProfile(nonexistent) should return nil")
	}
}

func TestConfigProfileMethods(t *testing.T) {
	cfg := &ltconfig.Config{
		Host: "https://default.example.com",
		Profiles: map[string]ltconfig.Profile{
			"dev":  {Host: "https://dev.example.com"},
			"prod": {Host: "https://prod.example.com"},
		},
	}

	// Test GetProfile
	p := cfg.GetProfile("dev")
	if p == nil {
		t.Fatal("GetProfile(dev) returned nil")
	}
	if p.Host != "https://dev.example.com" {
		t.Errorf("dev host = %q, want %q", p.Host, "https://dev.example.com")
	}

	// Test GetProfile not found
	p = cfg.GetProfile("nonexistent")
	if p != nil {
		t.Error("GetProfile(nonexistent) should return nil")
	}

	// Test ListProfiles
	names := cfg.ListProfiles()
	if len(names) != 2 {
		t.Errorf("ListProfiles() returned %d names, want 2", len(names))
	}

	// Test SetProfile
	cfg.SetProfile("staging", ltconfig.Profile{Host: "https://staging.example.com"})
	p = cfg.GetProfile("staging")
	if p == nil {
		t.Fatal("SetProfile failed")
	}

	// Test RemoveProfile
	removed := cfg.RemoveProfile("staging")
	if !removed {
		t.Error("RemoveProfile(staging) returned false")
	}
	p = cfg.GetProfile("staging")
	if p != nil {
		t.Error("profile should be removed")
	}

	// Test RemoveProfile not found
	removed = cfg.RemoveProfile("nonexistent")
	if removed {
		t.Error("RemoveProfile(nonexistent) should return false")
	}

	// Test ResolveProfile
	host, _ := cfg.ResolveProfile("dev")
	if host != "https://dev.example.com" {
		t.Errorf("ResolveProfile(dev) host = %q, want %q", host, "https://dev.example.com")
	}

	// Test ResolveProfile with empty name (falls back to top-level)
	host, _ = cfg.ResolveProfile("")
	if host != "https://default.example.com" {
		t.Errorf("ResolveProfile('') host = %q, want %q", host, "https://default.example.com")
	}

	// Test ResolveProfile with nonexistent name (falls back to top-level)
	host, _ = cfg.ResolveProfile("nonexistent")
	if host != "https://default.example.com" {
		t.Errorf("ResolveProfile(nonexistent) host = %q, want %q", host, "https://default.example.com")
	}
}

// runSub executes a profile subcommand through the profile command group,
// capturing output. Executing through the parent (not the child directly)
// ensures cobra parses the provided args rather than the test binary's os.Args.
func runSub(t *testing.T, f *cmdutil.Factory, name string, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	root := NewCmdProfile(f)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(append([]string{name}, args...))
	err := root.Execute()
	return buf.String(), err
}

func TestProfileCurrent_NoProfile(t *testing.T) {
	f := newTestFactory()
	f.Config.CurrentProfile = "" // top-level config
	out, err := runSub(t, f, "current")
	if err != nil {
		t.Fatalf("current error = %v", err)
	}
	if !strings.Contains(out, "no profile active") {
		t.Errorf("expected top-level message, got: %s", out)
	}
}

func TestProfileCurrent_Missing(t *testing.T) {
	f := newTestFactory()
	f.Config.CurrentProfile = "ghost" // set but not present
	_, err := runSub(t, f, "current")
	if lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error for missing current profile, got %v", err)
	}
}

func TestProfileUse_Command(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f := newTestFactory()
	out, err := runSub(t, f, "use", "dev")
	if err != nil {
		t.Fatalf("use error = %v", err)
	}
	if f.Config.CurrentProfile != "dev" {
		t.Errorf("CurrentProfile = %q, want dev", f.Config.CurrentProfile)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("output should mention the profile, got: %s", out)
	}
}

func TestProfileUse_NotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f := newTestFactory()
	_, err := runSub(t, f, "use", "ghost")
	if lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestProfileAdd_Command(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f := newTestFactory()
	out, err := runSub(t, f, "add", "--name", "staging", "--host", "https://staging.example.com")
	if err != nil {
		t.Fatalf("add error = %v", err)
	}
	if f.Config.GetProfile("staging") == nil {
		t.Error("profile not added")
	}
	if !strings.Contains(out, "staging") {
		t.Errorf("output should mention the profile, got: %s", out)
	}
}

func TestProfileAdd_Validation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Missing --name.
	if _, err := runSub(t, newTestFactory(), "add", "--host", "https://x"); lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("missing name: expected validation, got %v", err)
	}
	// Missing --host.
	if _, err := runSub(t, newTestFactory(), "add", "--name", "x"); lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("missing host: expected validation, got %v", err)
	}
	// Duplicate.
	if _, err := runSub(t, newTestFactory(), "add", "--name", "dev", "--host", "https://x"); lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("duplicate: expected validation, got %v", err)
	}
}

func TestProfileRemove_Command(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f := newTestFactory()
	out, err := runSub(t, f, "remove", "--name", "dev", "--yes")
	if err != nil {
		t.Fatalf("remove error = %v", err)
	}
	if f.Config.GetProfile("dev") != nil {
		t.Error("profile not removed")
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("output should mention the profile, got: %s", out)
	}
}

func TestProfileRemove_NeedsConfirmation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f := newTestFactory()
	_, err := runSub(t, f, "remove", "--name", "dev") // no --yes
	if output.ExitCodeOf(err) != output.ExitConfirmationRequired {
		t.Errorf("expected confirmation-required exit, got %v (exit %d)", err, output.ExitCodeOf(err))
	}
	if f.Config.GetProfile("dev") == nil {
		t.Error("profile should NOT be removed without --yes")
	}
}

func TestProfileRemove_Guards(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Missing --name.
	if _, err := runSub(t, newTestFactory(), "remove", "--yes"); lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("missing name: expected validation, got %v", err)
	}
	// Not found.
	if _, err := runSub(t, newTestFactory(), "remove", "--name", "ghost", "--yes"); lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("not found: expected validation, got %v", err)
	}
	// Cannot remove the active profile.
	f := newTestFactory()
	f.Config.CurrentProfile = "dev"
	if _, err := runSub(t, f, "remove", "--name", "dev", "--yes"); lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("active profile: expected validation, got %v", err)
	}
}

func TestProfileList_Command(t *testing.T) {
	f := newTestFactory()
	out, err := runSub(t, f, "list")
	if err != nil {
		t.Fatalf("list error = %v", err)
	}
	if !strings.Contains(out, "dev") || !strings.Contains(out, "prod") {
		t.Errorf("list should include profiles, got: %s", out)
	}
}
