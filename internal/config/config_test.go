// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultOmitNull(t *testing.T) {
	// When no config file exists, OmitNull should default to true
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.OmitNull {
		t.Errorf("OmitNull should default to true, got %v", cfg.OmitNull)
	}
}

func TestLoad_DefaultHost(t *testing.T) {
	// When no config file exists, Host should default to DefaultHost
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Host != DefaultHost {
		t.Errorf("Host should default to %q, got %q", DefaultHost, cfg.Host)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".lingtong-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")

	// Test with omitNull explicitly set to false
	content := "host: https://example.com\nomitNull: false\n"
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Override HOME to use temp directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.OmitNull {
		t.Errorf("OmitNull should be false when explicitly set in config, got %v", cfg.OmitNull)
	}

	if cfg.Host != "https://example.com" {
		t.Errorf("Host = %v, want https://example.com", cfg.Host)
	}
}

func TestLoad_WithConfigFileOmitNullTrue(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".lingtong-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")

	content := "host: https://example.com\nomitNull: true\n"
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.OmitNull {
		t.Errorf("OmitNull should be true, got %v", cfg.OmitNull)
	}
}

func TestProfileMethods(t *testing.T) {
	cfg := &Config{Host: "https://top.example.com", Brand: "lingtong"}

	// Empty config: lookups are safe and return zero values.
	if cfg.GetProfile("x") != nil {
		t.Error("GetProfile on empty config should be nil")
	}
	if cfg.ListProfiles() != nil {
		t.Error("ListProfiles on empty config should be nil")
	}
	if cfg.RemoveProfile("x") {
		t.Error("RemoveProfile on empty config should be false")
	}

	// Set, get, list (sorted), remove.
	cfg.SetProfile("prod", Profile{Host: "https://prod.example.com", Brand: "p"})
	cfg.SetProfile("dev", Profile{Host: "https://dev.example.com"})

	p := cfg.GetProfile("dev")
	if p == nil || p.Host != "https://dev.example.com" {
		t.Fatalf("GetProfile(dev) = %v", p)
	}

	names := cfg.ListProfiles()
	if len(names) != 2 || names[0] != "dev" || names[1] != "prod" {
		t.Errorf("ListProfiles() = %v, want sorted [dev prod]", names)
	}

	if !cfg.RemoveProfile("dev") {
		t.Error("RemoveProfile(dev) should be true")
	}
	if cfg.GetProfile("dev") != nil {
		t.Error("dev should be gone after remove")
	}
}

func TestResolveProfile(t *testing.T) {
	cfg := &Config{
		Host:  "https://top.example.com",
		Brand: "topbrand",
		Profiles: map[string]Profile{
			"dev":      {Host: "https://dev.example.com", Brand: "devbrand"},
			"hostonly": {Host: "https://hostonly.example.com"},
		},
	}

	// Explicit name with both fields.
	if h, b := cfg.ResolveProfile("dev"); h != "https://dev.example.com" || b != "devbrand" {
		t.Errorf("ResolveProfile(dev) = (%q,%q)", h, b)
	}

	// Profile with empty brand falls back to top-level brand.
	if _, b := cfg.ResolveProfile("hostonly"); b != "topbrand" {
		t.Errorf("ResolveProfile(hostonly) brand = %q, want topbrand", b)
	}

	// Empty name with CurrentProfile set.
	cfg.CurrentProfile = "dev"
	if h, _ := cfg.ResolveProfile(""); h != "https://dev.example.com" {
		t.Errorf("ResolveProfile('') with current=dev host = %q", h)
	}

	// Empty name and no CurrentProfile → top-level.
	cfg.CurrentProfile = ""
	if h, b := cfg.ResolveProfile(""); h != "https://top.example.com" || b != "topbrand" {
		t.Errorf("ResolveProfile('') = (%q,%q), want top-level", h, b)
	}

	// Unknown name → top-level fallback.
	if h, _ := cfg.ResolveProfile("ghost"); h != "https://top.example.com" {
		t.Errorf("ResolveProfile(ghost) host = %q, want top-level", h)
	}
}

func TestSaveAndLoad_WithProfiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &Config{
		Host:           "https://top.example.com",
		Brand:          "lingtong",
		OmitNull:       true,
		CurrentProfile: "dev",
		Profiles: map[string]Profile{
			"dev": {Host: "https://dev.example.com", Brand: "lingtong"},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.CurrentProfile != "dev" {
		t.Errorf("CurrentProfile = %q, want dev", loaded.CurrentProfile)
	}
	if p := loaded.GetProfile("dev"); p == nil || p.Host != "https://dev.example.com" {
		t.Errorf("round-tripped profile = %v", p)
	}
}
