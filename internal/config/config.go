// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds CLI configuration.
type Config struct {
	Host     string `yaml:"host"`
	Token    string `yaml:"-"` // Never stored in config file, use keychain
	Brand    string `yaml:"brand"` // "lingtong" or custom
	OmitNull bool   `yaml:"omitNull"` // Omit null fields in JSON output (default: true)
}

// DefaultHost is the default API host when none is configured.
const DefaultHost = "https://app1.ltpass.com"

// DefaultConfigDir returns the default config directory path.
func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".lingtong-cli"), nil
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load loads configuration from file.
func Load() (*Config, error) {
	// Default: omit null fields in JSON output, use default host
	cfg := &Config{
		Host:     DefaultHost,
		OmitNull: true,
	}

	path, err := DefaultConfigPath()
	if err != nil {
		return cfg, nil // Return default config if path fails
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil // Config file doesn't exist yet, use defaults
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves configuration to file.
func (c *Config) Save() error {
	dir, err := DefaultConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := DefaultConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
