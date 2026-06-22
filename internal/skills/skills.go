// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package skills installs the lingtong-cli Agent Skills into the directories
// that AI coding editors (Claude Code, OpenCode, Qoder, Cursor, Trae, Codex)
// read, so those agents can discover the CLI, its commands, and its skills.
//
// The skills themselves are compiled into the binary via go:embed (the embed
// directive lives in package main, which owns the repo-root skills/ tree) and
// injected here through SetEmbeddedFS. This keeps the package decoupled from the
// embed root and lets tests substitute an in-memory filesystem.
package skills

import (
	"io/fs"
	"sort"
	"strings"
)

// embeddedFS holds the skills compiled into the binary, rooted so that each
// top-level entry is a single skill directory. It is nil until SetEmbeddedFS is
// called (in non-release contexts such as gen-docs it may stay nil).
var embeddedFS fs.FS

// SetEmbeddedFS installs the embedded skills filesystem. root is the directory
// inside f that holds one subdirectory per skill (e.g. "skills"); the stored FS
// is re-rooted there so callers see skill directories at the top level. It is
// called once from package main's init.
func SetEmbeddedFS(f fs.FS, root string) {
	if f == nil {
		embeddedFS = nil
		return
	}
	if root == "" || root == "." {
		embeddedFS = f
		return
	}
	if sub, err := fs.Sub(f, root); err == nil {
		embeddedFS = sub
		return
	}
	embeddedFS = f
}

// FS returns the embedded skills filesystem (top-level entries are skill
// directories), or nil if no skills were compiled in.
func FS() fs.FS { return embeddedFS }

// Skill describes one embedded skill.
type Skill struct {
	Name        string // directory name, e.g. "lingtong-shared"
	Description string // one-line description from SKILL.md frontmatter
}

// List returns the embedded skills sorted by name. A skill is any top-level
// directory that contains a SKILL.md file. Returns an empty slice (no error)
// when nothing is embedded.
func List() ([]Skill, error) {
	if embeddedFS == nil {
		return nil, nil
	}
	entries, err := fs.ReadDir(embeddedFS, ".")
	if err != nil {
		return nil, err
	}

	skills := make([]Skill, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := fs.ReadFile(embeddedFS, e.Name()+"/SKILL.md")
		if err != nil {
			continue // directory without a SKILL.md is not a skill
		}
		skills = append(skills, Skill{
			Name:        e.Name(),
			Description: parseDescription(string(data)),
		})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

// parseDescription extracts the `description:` value from a SKILL.md YAML
// frontmatter block. It avoids a YAML dependency by scanning the leading
// `---`-delimited section line by line. Returns "" when absent.
func parseDescription(md string) string {
	lines := strings.Split(md, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break // end of frontmatter
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != "description" {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return ""
}
