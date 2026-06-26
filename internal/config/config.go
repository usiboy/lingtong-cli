// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds CLI configuration.
type Config struct {
	Host     string `yaml:"host"`
	Token    string `yaml:"-"`        // Never stored in config file, use keychain
	Brand    string `yaml:"brand"`    // "lingtong" or custom
	OmitNull bool   `yaml:"omitNull"` // Omit null fields in JSON output (default: true)

	// Multi-profile support
	CurrentProfile string             `yaml:"currentProfile,omitempty"`
	Profiles       map[string]Profile `yaml:"profiles,omitempty"`

	// Multi-auth: authenticated identities under the default (top-level) scope
	Auths       map[string]AuthIdentity `yaml:"auths,omitempty"`
	CurrentAuth string                  `yaml:"currentAuth,omitempty"`
}

// Profile represents a named configuration profile (e.g., dev, staging, prod).
type Profile struct {
	Host  string `yaml:"host"`
	Brand string `yaml:"brand,omitempty"`
}

// AuthIdentity represents a single authenticated identity under a config scope.
type AuthIdentity struct {
	Label       string      `yaml:"label"`                // Display alias, e.g. "company-a"
	TokenSuffix string      `yaml:"tokenSuffix"`          // Last 4 chars of token for masked display
	TenantInfo  *TenantInfo `yaml:"tenantInfo,omitempty"` // Cached tenant info (non-sensitive)
	CreatedAt   string      `yaml:"createdAt"`            // ISO-8601 creation timestamp
}

// TenantInfo holds cached tenant information fetched at login time.
type TenantInfo struct {
	TenantId  string `yaml:"tenantId"`
	FetchedAt string `yaml:"fetchedAt"` // ISO-8601 fetch timestamp
}

// NewAuthIdentity creates an AuthIdentity with the current timestamp.
func NewAuthIdentity(label, tokenSuffix, tenantId string) AuthIdentity {
	now := time.Now().UTC().Format(time.RFC3339)
	ai := AuthIdentity{
		Label:       label,
		TokenSuffix: tokenSuffix,
		CreatedAt:   now,
	}
	if tenantId != "" {
		ai.TenantInfo = &TenantInfo{
			TenantId:  tenantId,
			FetchedAt: now,
		}
	}
	return ai
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

// GetProfile returns the profile with the given name, or nil if not found.
func (c *Config) GetProfile(name string) *Profile {
	if c.Profiles == nil {
		return nil
	}
	p, ok := c.Profiles[name]
	if !ok {
		return nil
	}
	return &p
}

// SetProfile sets or updates a profile.
func (c *Config) SetProfile(name string, profile Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}
	c.Profiles[name] = profile
}

// RemoveProfile removes a profile by name. Returns true if it existed.
func (c *Config) RemoveProfile(name string) bool {
	if c.Profiles == nil {
		return false
	}
	_, ok := c.Profiles[name]
	if ok {
		delete(c.Profiles, name)
	}
	return ok
}

// ListProfiles returns all profile names sorted alphabetically.
func (c *Config) ListProfiles() []string {
	if c.Profiles == nil {
		return nil
	}
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	// Sort for deterministic output
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}

// GetAuth returns the auth identity with the given label, or nil if not found.
func (c *Config) GetAuth(label string) *AuthIdentity {
	if c.Auths == nil {
		return nil
	}
	a, ok := c.Auths[label]
	if !ok {
		return nil
	}
	return &a
}

// SetAuth sets or updates an auth identity.
func (c *Config) SetAuth(label string, auth AuthIdentity) {
	if c.Auths == nil {
		c.Auths = make(map[string]AuthIdentity)
	}
	c.Auths[label] = auth
}

// RemoveAuth removes an auth identity by label. Returns true if it existed.
func (c *Config) RemoveAuth(label string) bool {
	if c.Auths == nil {
		return false
	}
	_, ok := c.Auths[label]
	if ok {
		delete(c.Auths, label)
	}
	return ok
}

// ListAuths returns all auth identity labels sorted alphabetically.
func (c *Config) ListAuths() []string {
	if c.Auths == nil {
		return nil
	}
	names := make([]string, 0, len(c.Auths))
	for name := range c.Auths {
		names = append(names, name)
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}

// ResolveAuth returns the auth identity for the given name.
// If name is empty, uses CurrentAuth. Returns nil if not found.
func (c *Config) ResolveAuth(name string) *AuthIdentity {
	if name == "" {
		name = c.CurrentAuth
	}
	if name == "" {
		return nil
	}
	return c.GetAuth(name)
}

// ResolveProfile returns the effective host and brand for the given profile name.
// If profileName is empty, uses CurrentProfile. If CurrentProfile is also empty,
// falls back to the top-level Host/Brand fields for backward compatibility.
func (c *Config) ResolveProfile(profileName string) (host, brand string) {
	// If no profile specified, use top-level config
	if profileName == "" {
		profileName = c.CurrentProfile
	}
	if profileName == "" {
		return c.Host, c.Brand
	}

	// Look up the profile
	p := c.GetProfile(profileName)
	if p == nil {
		// Profile not found, fall back to top-level
		return c.Host, c.Brand
	}

	// Use profile values, falling back to top-level for empty fields
	host = p.Host
	if host == "" {
		host = c.Host
	}
	brand = p.Brand
	if brand == "" {
		brand = c.Brand
	}
	return host, brand
}
