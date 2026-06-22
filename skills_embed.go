// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package main

import (
	"embed"

	"github.com/lingtong/cli/internal/skills"
)

// embeddedSkills bundles the Agent Skills (skills/<name>/SKILL.md and their
// references) into the binary so `lingtong-cli skills install` works with no
// source tree or npx. The directive lives in package main because go:embed
// paths cannot escape their package directory, and the canonical skills/ tree
// sits at the repo root. Files beginning with "." (e.g. .DS_Store) are excluded
// by default.
//
//go:embed skills
var embeddedSkills embed.FS

func init() {
	// Re-root at "skills" so internal/skills sees each skill at the top level.
	skills.SetEmbeddedFS(embeddedSkills, "skills")
}
