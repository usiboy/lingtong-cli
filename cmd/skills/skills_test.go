// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package skills

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
	skillspkg "github.com/lingtong/cli/internal/skills"
)

func fakeSkillsFS() fstest.MapFS {
	return fstest.MapFS{
		"skills/lingtong-shared/SKILL.md": {Data: []byte("---\nname: lingtong-shared\ndescription: \"shared base\"\n---\n")},
		"skills/lingtong-scene/SKILL.md":  {Data: []byte("---\nname: lingtong-scene\ndescription: \"scenes\"\n---\n")},
	}
}

// newTestFactory returns a factory whose stdout is captured.
func newTestFactory() (*cmdutil.Factory, *bytes.Buffer) {
	var out bytes.Buffer
	f := &cmdutil.Factory{
		Config: &config.Config{},
		IOStreams: &output.IOStreams{
			In:     &bytes.Buffer{},
			Out:    &out,
			ErrOut: &bytes.Buffer{},
		},
	}
	return f, &out
}

// isolate installs the fake FS, a temp HOME, and a temp cwd for a test.
func isolate(t *testing.T) (cwd, home string) {
	t.Helper()
	prev := skillspkg.FS()
	skillspkg.SetEmbeddedFS(fakeSkillsFS(), "skills")
	t.Cleanup(func() { skillspkg.SetEmbeddedFS(prev, "") })

	home = t.TempDir()
	cwd = t.TempDir()
	t.Setenv("HOME", home)
	chdir(t, cwd)
	return cwd, home
}

// chdir switches into dir for the duration of the test, restoring the original
// working directory afterward. (Avoids testing.T.Chdir, which requires go1.24.)
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	f, out := newTestFactory()
	cmd := NewCmdSkills(f)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func TestNewCmdSkills_Subcommands(t *testing.T) {
	f, _ := newTestFactory()
	cmd := NewCmdSkills(f)
	if cmd.Use != "skills" {
		t.Errorf("Use = %q", cmd.Use)
	}
	want := map[string]bool{"install": false, "list": false, "status": false, "uninstall": false}
	for _, c := range cmd.Commands() {
		want[c.Name()] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestSkillsList(t *testing.T) {
	isolate(t)
	out, err := run(t, "list")
	if err != nil {
		t.Fatalf("list error = %v", err)
	}
	if !strings.Contains(out, "lingtong-shared") || !strings.Contains(out, "lingtong-scene") {
		t.Errorf("list output missing skills:\n%s", out)
	}
	if !strings.Contains(out, "Embedded skills (2)") {
		t.Errorf("expected count header:\n%s", out)
	}
}

func TestSkillsInstall_ForcedProject(t *testing.T) {
	cwd, _ := isolate(t)
	out, err := run(t, "install", "--editor", "opencode", "--scope", "project")
	if err != nil {
		t.Fatalf("install error = %v", err)
	}
	if !strings.Contains(out, "[OK] opencode") {
		t.Errorf("missing opencode OK line:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".opencode", "skills", "lingtong-shared", "SKILL.md")); err != nil {
		t.Errorf("skill not installed: %v", err)
	}
}

func TestSkillsInstall_DryRun(t *testing.T) {
	cwd, _ := isolate(t)
	out, err := run(t, "install", "--editor", "claude", "--scope", "project", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run error = %v", err)
	}
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected dry-run banner:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); !os.IsNotExist(err) {
		t.Error("dry-run must not create directories")
	}
}

func TestSkillsInstall_InvalidScope(t *testing.T) {
	isolate(t)
	if _, err := run(t, "install", "--scope", "bogus"); err == nil {
		t.Error("invalid scope should error")
	}
}

func TestSkillsInstall_UnknownEditor(t *testing.T) {
	isolate(t)
	if _, err := run(t, "install", "--editor", "nope"); err == nil {
		t.Error("unknown editor should error")
	}
}

func TestSkillsStatusAndUninstall(t *testing.T) {
	cwd, _ := isolate(t)
	if _, err := run(t, "install", "--editor", "opencode", "--scope", "project"); err != nil {
		t.Fatal(err)
	}

	statusOut, err := run(t, "status", "--editor", "opencode", "--scope", "project")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(statusOut, "opencode") || !strings.Contains(statusOut, "skills @") {
		t.Errorf("status missing install info:\n%s", statusOut)
	}

	uOut, err := run(t, "uninstall", "--editor", "opencode", "--scope", "project")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(uOut, "removed") {
		t.Errorf("uninstall output:\n%s", uOut)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".opencode", "skills", "lingtong-shared")); !os.IsNotExist(err) {
		t.Error("skill should be removed after uninstall")
	}
}

func TestSkillsStatus_NothingInstalled(t *testing.T) {
	isolate(t)
	out, err := run(t, "status", "--editor", "qoder", "--scope", "project")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "nothing installed") {
		t.Errorf("expected empty status message:\n%s", out)
	}
}

func TestParseEditors(t *testing.T) {
	tests := []struct {
		in   string
		want int // 0 means nil (all)
	}{
		{"all", 0},
		{"", 0},
		{"claude", 1},
		{"claude, opencode , trae", 3},
	}
	for _, tt := range tests {
		if got := parseEditors(tt.in); len(got) != tt.want {
			t.Errorf("parseEditors(%q) len = %d, want %d", tt.in, len(got), tt.want)
		}
	}
}
