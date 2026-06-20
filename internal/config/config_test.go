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
