// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newProjectHome creates a temp project (cwd) and home for an install test.
func newProjectHome(t *testing.T) (cwd, home string) {
	t.Helper()
	root := t.TempDir()
	cwd = filepath.Join(root, "proj")
	home = filepath.Join(root, "home")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	return cwd, home
}

func targetFor(res *Result, editor string, scope Scope) *TargetResult {
	for i := range res.Targets {
		if res.Targets[i].Editor == editor && res.Targets[i].Scope == scope {
			return &res.Targets[i]
		}
	}
	return nil
}

func TestInstall_ForcedScopeWritesSkillsAndInstructions(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)

	res, err := Install(Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home})
	if err != nil {
		t.Fatalf("Install error = %v", err)
	}
	tr := targetFor(res, "opencode", ScopeProject)
	if tr == nil || tr.Err != nil {
		t.Fatalf("missing/failed opencode target: %+v", tr)
	}
	if tr.SkillsInstalled != 2 {
		t.Errorf("SkillsInstalled = %d, want 2", tr.SkillsInstalled)
	}

	// Skill files (including nested references) must exist.
	want := filepath.Join(cwd, ".opencode", "skills", "lingtong-scene", "references", "notes.md")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("nested reference not copied: %v", err)
	}
	// Managed block present in AGENTS.md.
	agents := filepath.Join(cwd, ".opencode", "AGENTS.md")
	data, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("AGENTS.md not written: %v", err)
	}
	if !strings.Contains(string(data), blockBegin) || !strings.Contains(string(data), "lingtong-cli") {
		t.Error("AGENTS.md missing managed block / CLI mention")
	}
}

func TestInstall_AutoDetectInstallsOnlyPresentEditors(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)
	// Only .opencode exists in the project.
	if err := os.MkdirAll(filepath.Join(cwd, ".opencode"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := Install(Options{CWD: cwd, Home: home}) // all editors, auto scope
	if err != nil {
		t.Fatalf("Install error = %v", err)
	}
	if tr := targetFor(res, "opencode", ScopeProject); tr == nil || tr.SkillsInstalled != 2 {
		t.Errorf("opencode should be auto-installed at project scope: %+v", tr)
	}
	// claude has no dir and wasn't named: it should be skipped, not installed.
	if tr := targetFor(res, "claude", ScopeProject); tr != nil {
		t.Errorf("claude should not be installed (no project dir): %+v", tr)
	}
	cl := targetFor(res, "claude", "")
	if cl == nil || !cl.Skipped {
		t.Errorf("claude should be reported skipped: %+v", cl)
	}
}

func TestInstall_ExplicitUndetectedFallsBackToGlobal(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)

	res, err := Install(Options{Editors: []string{"cursor"}, CWD: cwd, Home: home})
	if err != nil {
		t.Fatalf("Install error = %v", err)
	}
	tr := targetFor(res, "cursor", ScopeGlobal)
	if tr == nil || tr.SkillsInstalled != 2 {
		t.Fatalf("explicit undetected editor should install at global: %+v", tr)
	}
	if _, err := os.Stat(filepath.Join(home, ".cursor", "skills", "lingtong-shared", "SKILL.md")); err != nil {
		t.Errorf("global cursor skill missing: %v", err)
	}
}

func TestInstall_Idempotent(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)
	opts := Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home}
	if _, err := Install(opts); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(opts); err != nil {
		t.Fatal(err)
	}
	agents := filepath.Join(cwd, ".opencode", "AGENTS.md")
	data, _ := os.ReadFile(agents)
	if n := strings.Count(string(data), blockBegin); n != 1 {
		t.Errorf("managed block should appear once after re-install, got %d", n)
	}
}

func TestInstall_PreservesExistingInstructions(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)
	dir := filepath.Join(cwd, ".opencode")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	agents := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agents, []byte("# My rules\nkeep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(agents)
	if !strings.Contains(string(data), "keep me") {
		t.Error("existing user content must be preserved")
	}
	if !strings.Contains(string(data), blockBegin) {
		t.Error("managed block must be appended")
	}
}

func TestInstall_DryRunWritesNothing(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)

	res, err := Install(Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if tr := targetFor(res, "opencode", ScopeProject); tr == nil {
		t.Fatal("expected a target result even in dry-run")
	}
	if _, err := os.Stat(filepath.Join(cwd, ".opencode", "skills")); !os.IsNotExist(err) {
		t.Error("dry-run must not create skill files")
	}
	if _, err := os.Stat(filepath.Join(cwd, ".opencode", "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("dry-run must not write AGENTS.md")
	}
}

func TestInstall_SharedRootAgentsWrittenOnce(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)
	// cursor, trae, codex all target the project-root AGENTS.md.
	res, err := Install(Options{Editors: []string{"cursor", "trae", "codex"}, Scope: ScopeProject, CWD: cwd, Home: home})
	if err != nil {
		t.Fatal(err)
	}
	writes := 0
	for _, tr := range res.Targets {
		if tr.InstructionsWritten {
			writes++
		}
	}
	if writes != 1 {
		t.Errorf("shared root AGENTS.md should be written once, got %d", writes)
	}
	data, _ := os.ReadFile(filepath.Join(cwd, "AGENTS.md"))
	if strings.Count(string(data), blockBegin) != 1 {
		t.Error("root AGENTS.md should contain exactly one managed block")
	}
}

func TestUninstall_RemovesSkillsAndBlock(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)
	opts := Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home}
	if _, err := Install(opts); err != nil {
		t.Fatal(err)
	}

	res, err := Uninstall(opts)
	if err != nil {
		t.Fatal(err)
	}
	tr := targetFor(res, "opencode", ScopeProject)
	if tr == nil || tr.Removed != 2 || !tr.InstructionsWritten {
		t.Fatalf("uninstall result unexpected: %+v", tr)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".opencode", "skills", "lingtong-scene")); !os.IsNotExist(err) {
		t.Error("skill dir should be removed")
	}
	// AGENTS.md had only the managed block, so it should be deleted entirely.
	if _, err := os.Stat(filepath.Join(cwd, ".opencode", "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("AGENTS.md with only managed block should be removed")
	}
}

func TestUninstall_KeepsUserContent(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)
	dir := filepath.Join(cwd, ".opencode")
	_ = os.MkdirAll(dir, 0o755)
	agents := filepath.Join(dir, "AGENTS.md")
	_ = os.WriteFile(agents, []byte("# keep\nmine\n"), 0o644)
	opts := Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home}
	_, _ = Install(opts)

	if _, err := Uninstall(opts); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("AGENTS.md should remain (had user content): %v", err)
	}
	if strings.Contains(string(data), blockBegin) {
		t.Error("managed block should be gone")
	}
	if !strings.Contains(string(data), "mine") {
		t.Error("user content should remain")
	}
}

func TestStatus_ReportsInstalled(t *testing.T) {
	withFakeFS(t, fakeSkills())
	cwd, home := newProjectHome(t)
	opts := Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home}
	_, _ = Install(opts)

	res, err := Status(opts)
	if err != nil {
		t.Fatal(err)
	}
	tr := targetFor(res, "opencode", ScopeProject)
	if tr == nil || tr.SkillsInstalled != 2 || !tr.InstructionsWritten {
		t.Errorf("status should report 2 skills + block: %+v", tr)
	}
}

func TestInstall_NoEmbeddedFSErrors(t *testing.T) {
	prev := embeddedFS
	t.Cleanup(func() { embeddedFS = prev })
	embeddedFS = nil
	cwd, home := newProjectHome(t)

	res, err := Install(Options{Editors: []string{"opencode"}, Scope: ScopeProject, CWD: cwd, Home: home})
	if err != nil {
		t.Fatalf("Install should return per-target error, not top-level: %v", err)
	}
	if tr := targetFor(res, "opencode", ScopeProject); tr == nil || tr.Err == nil {
		t.Errorf("expected per-target error when nothing embedded: %+v", tr)
	}
}

func TestReplaceBlock(t *testing.T) {
	text := "head\n" + blockBegin + "\nold\n" + blockEnd + "\ntail\n"
	out, replaced := replaceBlock(text, blockBegin+"\nnew\n"+blockEnd)
	if !replaced {
		t.Fatal("should report replaced")
	}
	if strings.Contains(out, "old") || !strings.Contains(out, "new") {
		t.Errorf("block not replaced: %q", out)
	}
	if !strings.HasPrefix(out, "head\n") || !strings.HasSuffix(out, "\ntail\n") {
		t.Errorf("surrounding content not preserved: %q", out)
	}
	if _, replaced := replaceBlock("no markers here", "x"); replaced {
		t.Error("should not replace when no markers")
	}
}

func TestFirstSentence(t *testing.T) {
	tests := map[string]string{
		"绫通场景管理。其余内容":      "绫通场景管理",
		"scene mgmt. more": "scene mgmt",
		"single":           "single",
		"":                 "见 SKILL.md",
	}
	for in, want := range tests {
		if got := firstSentence(in); got != want {
			t.Errorf("firstSentence(%q) = %q, want %q", in, got, want)
		}
	}
}
