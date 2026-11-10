package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lingtong/cli/internal/skills"
)

// Exercise the real embedded package, not only an in-memory installer fixture.
func TestDashboardSkillEmbeddedInstall(t *testing.T) {
	for _, editor := range skills.Editors() {
		for _, scope := range []skills.Scope{skills.ScopeProject, skills.ScopeGlobal} {
			t.Run(editor.ID+"/"+string(scope), func(t *testing.T) {
				cwd, home := t.TempDir(), t.TempDir()
				opts := skills.Options{Editors: []string{editor.ID}, Scope: scope, CWD: cwd, Home: home}
				res, err := skills.Install(opts)
				if err != nil || len(res.Targets) != 1 || res.Targets[0].Err != nil {
					t.Fatalf("install: %v, %+v", err, res)
				}
				target := editor.Resolve(scope, cwd, home)
				if res.Targets[0].Editor != editor.ID || res.Targets[0].SkillsDir != target.SkillsDir {
					t.Fatalf("install did not respect selected editor: %+v", res.Targets[0])
				}
				// Selecting one editor must not install skills into another editor or scope.
				for _, other := range skills.Editors() {
					for _, otherScope := range []skills.Scope{skills.ScopeProject, skills.ScopeGlobal} {
						dir := other.Resolve(otherScope, cwd, home).SkillsDir
						if dir == "" || dir == target.SkillsDir {
							continue
						}
						if _, err := os.Stat(dir); !os.IsNotExist(err) {
							t.Fatalf("unexpected install at %s: %v", dir, err)
						}
					}
				}
				instructions, err := os.ReadFile(target.InstructionsPath)
				if err != nil || !bytes.Contains(instructions, []byte("lingtong-dashboard")) {
					t.Fatalf("dashboard missing from selected editor's instructions: %v", err)
				}
				installed := filepath.Join(res.Targets[0].SkillsDir, "lingtong-dashboard")
				count := 0
				err = fs.WalkDir(skills.FS(), "lingtong-dashboard", func(path string, d fs.DirEntry, err error) error {
					if err != nil || d.IsDir() {
						return err
					}
					want, err := fs.ReadFile(skills.FS(), path)
					if err != nil {
						return err
					}
					got, err := os.ReadFile(filepath.Join(res.Targets[0].SkillsDir, filepath.FromSlash(path)))
					if err != nil {
						return err
					}
					if !bytes.Equal(want, got) {
						t.Errorf("installed content differs: %s", path)
					}
					count++
					return nil
				})
				if err != nil || count == 0 {
					t.Fatalf("embedded package files: count=%d err=%v", count, err)
				}
				node, err := exec.LookPath("node")
				if err != nil {
					t.Log("Node unavailable; run make test-dashboard-skill to verify the validator")
				} else {
					command := exec.Command(node, filepath.Join(installed, "scripts", "validate-dashboard.mjs"),
						filepath.Join(installed, "assets", "examples", "operating-overview.json"),
						"--schemas", filepath.Join(installed, "assets", "examples", "metadata.json"),
						"--pivot-configs", filepath.Join(installed, "assets", "examples", "pivot-configs.json"),
						"--require-bindings")
					if out, err := command.CombinedOutput(); err != nil {
						t.Fatalf("installed validator: %v\n%s", err, out)
					}
				}
				res, err = skills.Install(opts)
				if err != nil || res.Targets[0].Err != nil {
					t.Fatalf("reinstall: %v, %+v", err, res)
				}
			})
		}
	}
}
