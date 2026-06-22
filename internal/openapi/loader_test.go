// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

const miniSpec = `{
	"swagger": "2.0",
	"info": {"title": "Mini", "version": "1.0"},
	"paths": {"/ping": {"get": {"summary": "ping"}}}
}`

func TestOverridePath_Env(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(path, []byte(miniSpec), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LINGTONG_OPENAPI", path)
	if got := OverridePath(); got != path {
		t.Errorf("OverridePath() = %q, want %q", got, path)
	}
}

func TestOverridePath_HomeFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LINGTONG_OPENAPI", "")

	dir := filepath.Join(home, ".lingtong-cli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "openapi.json")
	if err := os.WriteFile(path, []byte(miniSpec), 0644); err != nil {
		t.Fatal(err)
	}
	if got := OverridePath(); got != path {
		t.Errorf("OverridePath() = %q, want home fallback %q", got, path)
	}
}

func TestOverridePath_NoneWhenEnvMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("LINGTONG_OPENAPI", "/nonexistent/spec.json")
	if got := OverridePath(); got != "" {
		t.Errorf("OverridePath() = %q, want empty (nonexistent env path ignored)", got)
	}
}

func TestLoadSpec_PrefersOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(path, []byte(miniSpec), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LINGTONG_OPENAPI", path)

	spec, err := LoadSpec()
	if err != nil {
		t.Fatalf("LoadSpec error = %v", err)
	}
	if spec == nil || len(spec.Paths) != 1 {
		t.Fatalf("expected the override spec (1 path), got %v", spec)
	}
	if _, ok := spec.Paths["/ping"]; !ok {
		t.Error("override spec should contain /ping")
	}
}

func TestLoadSpec_FallsBackToEmbedded(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("LINGTONG_OPENAPI", "")

	spec, err := LoadSpec()
	if err != nil {
		t.Fatalf("LoadSpec error = %v", err)
	}
	if HasEmbeddedSpec() {
		if spec == nil {
			t.Error("expected embedded spec when no override present")
		}
	}
}

func TestLoadSpec_MalformedOverrideFallsThrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LINGTONG_OPENAPI", path)

	// A malformed override must not error out; it falls through to embedded.
	spec, err := LoadSpec()
	if err != nil {
		t.Fatalf("LoadSpec should not error on malformed override, got %v", err)
	}
	if HasEmbeddedSpec() && spec == nil {
		t.Error("expected fall-through to embedded spec")
	}
}
