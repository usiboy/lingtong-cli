// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package skills

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Scope selects where skills are installed: the user's home (shared across all
// projects) or the current project working tree.
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
)

// instructionsFile is the agent instructions file name an editor reads.
const (
	fileAGENTS = "AGENTS.md"
	fileCLAUDE = "CLAUDE.md"
)

// Editor describes one AI coding tool and where it expects skills and agent
// instructions, parameterised by scope. layout is resolved lazily so paths can
// be computed against an injected cwd/home (which keeps tests hermetic).
type Editor struct {
	ID   string // stable identifier used by the --editor flag, e.g. "claude"
	Name string // human-readable name

	// projectBase is the editor's config dir relative to the project root
	// (e.g. ".claude"). "" means the project root itself (Codex).
	projectBase string
	// globalBase returns the editor's absolute config dir under home
	// (e.g. <home>/.claude).
	globalBase func(home string) string
	// skillsSubdir is appended to the base to hold skills (e.g. "skills").
	// "" means the editor has no skills directory (Codex).
	skillsSubdir string
	// instructionsName is the agent instructions file the editor reads
	// (AGENTS.md or CLAUDE.md). "" means none.
	instructionsName string
	// instructionsAtRoot places the instructions file at the project/home root
	// rather than inside the editor's config dir (Claude's CLAUDE.md, and the
	// shared root AGENTS.md that Cursor/Trae/Codex read in project scope).
	instructionsAtRoot bool
}

// Target is a fully resolved install destination for one editor at one scope.
type Target struct {
	Editor *Editor
	Scope  Scope
	// SkillsDir is the absolute directory that holds one subdirectory per skill,
	// or "" when the editor has no skills directory.
	SkillsDir string
	// InstructionsPath is the absolute AGENTS.md/CLAUDE.md path, or "".
	InstructionsPath string
	// base is the editor's resolved config dir, used for presence detection.
	base string
}

// registry is the ordered set of supported editors.
var registry = []*Editor{
	{
		ID: "claude", Name: "Claude Code",
		projectBase:        ".claude",
		globalBase:         func(home string) string { return filepath.Join(home, ".claude") },
		skillsSubdir:       "skills",
		instructionsName:   fileCLAUDE,
		instructionsAtRoot: true,
	},
	{
		ID: "opencode", Name: "OpenCode",
		projectBase:      ".opencode",
		globalBase:       func(home string) string { return filepath.Join(home, ".config", "opencode") },
		skillsSubdir:     "skills",
		instructionsName: fileAGENTS,
	},
	{
		ID: "qoder", Name: "Qoder",
		projectBase:      ".qoder",
		globalBase:       func(home string) string { return filepath.Join(home, ".qoder") },
		skillsSubdir:     "skills",
		instructionsName: fileAGENTS,
	},
	{
		ID: "cursor", Name: "Cursor",
		projectBase:        ".cursor",
		globalBase:         func(home string) string { return filepath.Join(home, ".cursor") },
		skillsSubdir:       "skills",
		instructionsName:   fileAGENTS,
		instructionsAtRoot: true,
	},
	{
		ID: "trae", Name: "Trae",
		projectBase:        ".trae",
		globalBase:         func(home string) string { return filepath.Join(home, ".trae") },
		skillsSubdir:       "skills",
		instructionsName:   fileAGENTS,
		instructionsAtRoot: true,
	},
	{
		ID: "codex", Name: "Codex",
		projectBase:      "", // project root
		globalBase:       func(home string) string { return filepath.Join(home, ".codex") },
		skillsSubdir:     "", // no skills directory; AGENTS.md only
		instructionsName: fileAGENTS,
	},
}

// Editors returns the supported editors in registry order.
func Editors() []*Editor {
	out := make([]*Editor, len(registry))
	copy(out, registry)
	return out
}

// EditorByID looks up an editor by its --editor identifier.
func EditorByID(id string) (*Editor, bool) {
	for _, e := range registry {
		if e.ID == id {
			return e, true
		}
	}
	return nil, false
}

// resolveEditors maps the requested ids ("all" or specific ids) to editors,
// erroring on any unknown id.
func resolveEditors(ids []string) ([]*Editor, error) {
	if len(ids) == 0 {
		return Editors(), nil
	}
	for _, id := range ids {
		if id == "all" {
			return Editors(), nil
		}
	}
	seen := map[string]bool{}
	var out []*Editor
	for _, id := range ids {
		e, ok := EditorByID(id)
		if !ok {
			return nil, fmt.Errorf("unknown editor %q (supported: %s)", id, strings.Join(editorIDs(), ", "))
		}
		if seen[e.ID] {
			continue
		}
		seen[e.ID] = true
		out = append(out, e)
	}
	return out, nil
}

func editorIDs() []string {
	ids := make([]string, 0, len(registry))
	for _, e := range registry {
		ids = append(ids, e.ID)
	}
	sort.Strings(ids)
	return ids
}

// Resolve computes the install destination for this editor at the given scope.
// cwd is the project root (project scope); home is the user's home (global).
func (e *Editor) Resolve(scope Scope, cwd, home string) Target {
	var base, root string
	switch scope {
	case ScopeGlobal:
		base = e.globalBase(home)
		root = home
	default: // ScopeProject
		if e.projectBase == "" {
			base = cwd
		} else {
			base = filepath.Join(cwd, e.projectBase)
		}
		root = cwd
	}

	t := Target{Editor: e, Scope: scope, base: base}

	if e.skillsSubdir != "" {
		t.SkillsDir = filepath.Join(base, e.skillsSubdir)
	}
	if e.instructionsName != "" {
		if e.instructionsAtRoot {
			t.InstructionsPath = filepath.Join(root, e.instructionsName)
		} else {
			t.InstructionsPath = filepath.Join(base, e.instructionsName)
		}
	}
	return t
}
