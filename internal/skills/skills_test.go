// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package skills

import (
	"testing"
	"testing/fstest"
)

// fakeSkills returns an in-memory skills tree with two skills (one with a
// nested references file) plus an empty directory that must be ignored.
func fakeSkills() fstest.MapFS {
	return fstest.MapFS{
		"skills/lingtong-shared/SKILL.md":           {Data: []byte("---\nname: lingtong-shared\ndescription: \"共享基础。当用户首次配置时触发。\"\n---\n\n# shared\n")},
		"skills/lingtong-scene/SKILL.md":            {Data: []byte("---\nname: lingtong-scene\ndescription: 'scene mgmt'\n---\n")},
		"skills/lingtong-scene/references/notes.md": {Data: []byte("notes")},
		// A directory entry with no SKILL.md must not count as a skill.
		"skills/not-a-skill/readme.txt": {Data: []byte("x")},
	}
}

// withFakeFS installs the fake embedded FS for a test and restores it after.
func withFakeFS(t *testing.T, m fstest.MapFS) {
	t.Helper()
	prev := embeddedFS
	SetEmbeddedFS(m, "skills")
	t.Cleanup(func() { embeddedFS = prev })
}

func TestSetEmbeddedFS_Reroots(t *testing.T) {
	withFakeFS(t, fakeSkills())
	if _, err := embeddedFS.Open("lingtong-shared/SKILL.md"); err != nil {
		t.Fatalf("expected re-rooted FS to expose skill at top level: %v", err)
	}
}

func TestSetEmbeddedFS_Nil(t *testing.T) {
	prev := embeddedFS
	t.Cleanup(func() { embeddedFS = prev })
	SetEmbeddedFS(nil, "skills")
	if embeddedFS != nil {
		t.Error("nil FS should clear embeddedFS")
	}
	if list, err := List(); err != nil || list != nil {
		t.Errorf("List with nil FS = (%v, %v), want (nil, nil)", list, err)
	}
}

func TestList(t *testing.T) {
	withFakeFS(t, fakeSkills())
	list, err := List()
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d skills, want 2 (empty/non-skill dirs ignored): %+v", len(list), list)
	}
	// Sorted by name: lingtong-scene before lingtong-shared.
	if list[0].Name != "lingtong-scene" || list[1].Name != "lingtong-shared" {
		t.Errorf("unexpected order/names: %+v", list)
	}
	if list[1].Description != "共享基础。当用户首次配置时触发。" {
		t.Errorf("description = %q", list[1].Description)
	}
}

func TestParseDescription(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want string
	}{
		{"double quoted", "---\ndescription: \"hello world\"\n---\n", "hello world"},
		{"single quoted", "---\ndescription: 'hi'\n---\n", "hi"},
		{"unquoted", "---\nname: x\ndescription: plain\n---\n", "plain"},
		{"no frontmatter", "# title\ndescription: nope\n", ""},
		{"missing description", "---\nname: x\n---\n", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDescription(tt.md); got != tt.want {
				t.Errorf("parseDescription = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEditorByID(t *testing.T) {
	if _, ok := EditorByID("claude"); !ok {
		t.Error("claude should be a known editor")
	}
	if _, ok := EditorByID("nope"); ok {
		t.Error("unknown editor should not resolve")
	}
	if len(Editors()) != 6 {
		t.Errorf("want 6 editors, got %d", len(Editors()))
	}
}

func TestResolveEditors(t *testing.T) {
	all, err := resolveEditors(nil)
	if err != nil || len(all) != 6 {
		t.Fatalf("resolveEditors(nil) = (%d, %v), want (6, nil)", len(all), err)
	}
	if got, _ := resolveEditors([]string{"all"}); len(got) != 6 {
		t.Errorf("'all' should expand to 6, got %d", len(got))
	}
	got, err := resolveEditors([]string{"claude", "claude", "codex"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("duplicates should collapse, got %d", len(got))
	}
	if _, err := resolveEditors([]string{"bogus"}); err == nil {
		t.Error("unknown editor id should error")
	}
}

func TestResolve_Paths(t *testing.T) {
	claude, _ := EditorByID("claude")
	g := claude.Resolve(ScopeGlobal, "/proj", "/home")
	if g.SkillsDir != "/home/.claude/skills" {
		t.Errorf("claude global skills = %q", g.SkillsDir)
	}
	if g.InstructionsPath != "/home/CLAUDE.md" {
		t.Errorf("claude global instructions = %q (CLAUDE.md is at root)", g.InstructionsPath)
	}
	p := claude.Resolve(ScopeProject, "/proj", "/home")
	if p.SkillsDir != "/proj/.claude/skills" || p.InstructionsPath != "/proj/CLAUDE.md" {
		t.Errorf("claude project = %q / %q", p.SkillsDir, p.InstructionsPath)
	}

	codex, _ := EditorByID("codex")
	cp := codex.Resolve(ScopeProject, "/proj", "/home")
	if cp.SkillsDir != "" {
		t.Errorf("codex must have no skills dir, got %q", cp.SkillsDir)
	}
	if cp.InstructionsPath != "/proj/AGENTS.md" {
		t.Errorf("codex project instructions = %q", cp.InstructionsPath)
	}
	cg := codex.Resolve(ScopeGlobal, "/proj", "/home")
	if cg.InstructionsPath != "/home/.codex/AGENTS.md" {
		t.Errorf("codex global instructions = %q", cg.InstructionsPath)
	}

	oc, _ := EditorByID("opencode")
	og := oc.Resolve(ScopeGlobal, "/proj", "/home")
	if og.SkillsDir != "/home/.config/opencode/skills" || og.InstructionsPath != "/home/.config/opencode/AGENTS.md" {
		t.Errorf("opencode global = %q / %q", og.SkillsDir, og.InstructionsPath)
	}
}
