// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package openapi

import (
	"os"
	"path/filepath"
)

// LoadSpec returns the OpenAPI spec shared by every command that needs it
// (`service` generation, `schema` browsing, `doctor` checks), so they all see
// the same API surface.
//
// Resolution order (first hit wins):
//  1. LINGTONG_OPENAPI environment variable (explicit override / freshest)
//  2. ~/.lingtong-cli/openapi.json (user-managed override)
//  3. the spec embedded in the binary at build time (default, zero-config)
//
// A malformed override falls through to the next source rather than disabling
// spec-backed commands entirely.
func LoadSpec() (*Spec, error) {
	if path := OverridePath(); path != "" {
		if spec, err := Parse(path); err == nil {
			return spec, nil
		}
	}
	if HasEmbeddedSpec() {
		return EmbeddedSpec()
	}
	return nil, nil
}

// OverridePath returns a user-supplied spec path, or "" if none is present.
func OverridePath() string {
	if path := os.Getenv("LINGTONG_OPENAPI"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".lingtong-cli", "openapi.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
