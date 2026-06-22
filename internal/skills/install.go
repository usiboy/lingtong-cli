// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package skills

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Managed-block markers delimit the lingtong-cli section inside an editor's
// AGENTS.md/CLAUDE.md so installs are idempotent and never clobber user content.
const (
	blockBegin = "<!-- BEGIN lingtong-cli (managed by `lingtong-cli skills install`) -->"
	blockEnd   = "<!-- END lingtong-cli (managed by `lingtong-cli skills install`) -->"
)

// Options controls install/uninstall/status across editors.
type Options struct {
	Editors []string // editor ids, or {"all"} / empty for all supported editors
	Scope   Scope    // "" means auto-detect; otherwise forces ScopeGlobal/ScopeProject
	CWD     string   // project root for project scope
	Home    string   // user home for global scope
	DryRun  bool     // report intended actions without writing
}

// TargetResult records what happened (or would happen) for one editor+scope.
type TargetResult struct {
	Editor              string
	Scope               Scope
	SkillsDir           string
	SkillsInstalled     int
	InstructionsPath    string
	InstructionsWritten bool
	Removed             int  // skills removed (uninstall)
	Skipped             bool // editor not detected and not requested
	SkippedReason       string
	Err                 error
}

// Result aggregates per-target outcomes.
type Result struct {
	Targets []TargetResult
}

// Install copies the embedded skills and writes the agent-instructions managed
// block into every selected editor's directories. With Options.Scope unset it
// installs wherever an editor is detected; an explicitly named but undetected
// editor defaults to global scope.
func Install(opts Options) (*Result, error) {
	editors, err := resolveEditors(opts.Editors)
	if err != nil {
		return nil, err
	}
	explicit := isExplicit(opts.Editors)
	instr := buildInstructions()
	written := map[string]bool{} // dedupe shared instruction files (e.g. root AGENTS.md)

	res := &Result{}
	for _, e := range editors {
		scopes := scopesToInstall(e, opts, explicit)
		if len(scopes) == 0 {
			res.Targets = append(res.Targets, TargetResult{
				Editor: e.ID, Skipped: true,
				SkippedReason: "not detected (pass --editor or --scope to install anyway)",
			})
			continue
		}
		for _, sc := range scopes {
			res.Targets = append(res.Targets, installTarget(e.Resolve(sc, opts.CWD, opts.Home), instr, opts, written))
		}
	}
	return res, nil
}

func installTarget(t Target, instr string, opts Options, written map[string]bool) TargetResult {
	tr := TargetResult{Editor: t.Editor.ID, Scope: t.Scope, SkillsDir: t.SkillsDir, InstructionsPath: t.InstructionsPath}

	if t.SkillsDir != "" {
		n, err := copySkills(t.SkillsDir, opts.DryRun)
		if err != nil {
			tr.Err = err
			return tr
		}
		tr.SkillsInstalled = n
	}

	if t.InstructionsPath != "" && !written[t.InstructionsPath] {
		if err := upsertManagedBlock(t.InstructionsPath, instr, opts.DryRun); err != nil {
			tr.Err = err
			return tr
		}
		written[t.InstructionsPath] = true
		tr.InstructionsWritten = true
	}
	return tr
}

// Uninstall removes the lingtong skill directories and the managed instructions
// block from the selected editors/scopes.
func Uninstall(opts Options) (*Result, error) {
	editors, err := resolveEditors(opts.Editors)
	if err != nil {
		return nil, err
	}
	names, err := skillNames()
	if err != nil {
		return nil, err
	}
	removed := map[string]bool{}

	res := &Result{}
	for _, e := range editors {
		for _, sc := range scopesToScan(opts) {
			t := e.Resolve(sc, opts.CWD, opts.Home)
			tr := TargetResult{Editor: e.ID, Scope: sc, SkillsDir: t.SkillsDir, InstructionsPath: t.InstructionsPath}
			if t.SkillsDir != "" {
				for _, name := range names {
					dir := filepath.Join(t.SkillsDir, name)
					if dirExists(dir) {
						if !opts.DryRun {
							if err := os.RemoveAll(dir); err != nil {
								tr.Err = err
								break
							}
						}
						tr.Removed++
					}
				}
			}
			if t.InstructionsPath != "" && !removed[t.InstructionsPath] && fileExists(t.InstructionsPath) {
				if !opts.DryRun {
					if err := removeManagedBlock(t.InstructionsPath); err != nil {
						tr.Err = err
					}
				}
				removed[t.InstructionsPath] = true
				tr.InstructionsWritten = true
			}
			res.Targets = append(res.Targets, tr)
		}
	}
	return res, nil
}

// Status reports, for each selected editor/scope, how many lingtong skills are
// present and whether the managed instructions block exists.
func Status(opts Options) (*Result, error) {
	editors, err := resolveEditors(opts.Editors)
	if err != nil {
		return nil, err
	}
	names, err := skillNames()
	if err != nil {
		return nil, err
	}

	res := &Result{}
	for _, e := range editors {
		for _, sc := range scopesToScan(opts) {
			t := e.Resolve(sc, opts.CWD, opts.Home)
			tr := TargetResult{Editor: e.ID, Scope: sc, SkillsDir: t.SkillsDir, InstructionsPath: t.InstructionsPath}
			if t.SkillsDir != "" {
				for _, name := range names {
					if dirExists(filepath.Join(t.SkillsDir, name)) {
						tr.SkillsInstalled++
					}
				}
			}
			tr.InstructionsWritten = t.InstructionsPath != "" && hasManagedBlock(t.InstructionsPath)
			res.Targets = append(res.Targets, tr)
		}
	}
	return res, nil
}

// scopesToInstall picks install scopes for an editor: the forced scope, or the
// detected scopes; an explicitly requested but undetected editor falls back to
// global so `--editor cursor` works even on a clean machine.
func scopesToInstall(e *Editor, opts Options, explicit bool) []Scope {
	if opts.Scope != "" {
		return []Scope{opts.Scope}
	}
	var out []Scope
	for _, sc := range []Scope{ScopeProject, ScopeGlobal} {
		if detected(e.Resolve(sc, opts.CWD, opts.Home)) {
			out = append(out, sc)
		}
	}
	if len(out) == 0 && explicit {
		return []Scope{ScopeGlobal}
	}
	return out
}

func scopesToScan(opts Options) []Scope {
	if opts.Scope != "" {
		return []Scope{opts.Scope}
	}
	return []Scope{ScopeProject, ScopeGlobal}
}

// detected reports whether the editor appears to be set up at this target.
func detected(t Target) bool {
	if t.Editor.skillsSubdir != "" {
		return dirExists(t.base)
	}
	// Codex has no skills dir: detect by its config dir (global) or the
	// project-root AGENTS.md it reads (project).
	if t.Scope == ScopeGlobal {
		return dirExists(t.base)
	}
	return fileExists(t.InstructionsPath)
}

func isExplicit(ids []string) bool {
	for _, id := range ids {
		if id == "all" {
			return false
		}
	}
	return len(ids) > 0
}

// copySkills writes every embedded skill into destDir/<name>, replacing any
// existing copy for a clean reinstall. Returns the number of skills written.
func copySkills(destDir string, dryRun bool) (int, error) {
	if embeddedFS == nil {
		return 0, fmt.Errorf("no skills are embedded in this build")
	}
	names, err := skillNames()
	if err != nil {
		return 0, err
	}
	for _, name := range names {
		if dryRun {
			continue
		}
		dest := filepath.Join(destDir, name)
		if err := os.RemoveAll(dest); err != nil {
			return 0, fmt.Errorf("clearing %s: %w", dest, err)
		}
		if err := copyTree(name, destDir); err != nil {
			return 0, err
		}
	}
	return len(names), nil
}

// copyTree copies the embedded subtree rooted at srcRoot into destBase,
// preserving the srcRoot/... layout.
func copyTree(srcRoot, destBase string) error {
	return fs.WalkDir(embeddedFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(destBase, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(embeddedFS, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// skillNames returns the embedded skill directory names.
func skillNames() ([]string, error) {
	list, err := List()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(list))
	for _, s := range list {
		names = append(names, s.Name)
	}
	return names, nil
}

// buildInstructions renders the managed-block body advertising the CLI and its
// skills to the agent. It is generated from the embedded skill list so it never
// drifts from what is actually installed.
func buildInstructions() string {
	var b strings.Builder
	b.WriteString("## 绫通 lingtong-cli\n\n")
	b.WriteString("`lingtong-cli` 已安装：绫通 (Lingtong) iPaaS 平台官方命令行，供你（AI Agent）在终端操作连接器、场景、工作流、表格、模型等业务域。\n\n")
	b.WriteString("- 验证安装：`lingtong-cli --version`；自检：`lingtong-cli doctor`\n")
	b.WriteString("- 首次使用：`lingtong-cli config init` → `lingtong-cli auth login`\n")
	b.WriteString("- 浏览 API：`lingtong-cli schema list`；机器可读输出：附加 `--envelope`，按 `.ok`/`.error.type` 判定\n\n")

	if list, _ := List(); len(list) > 0 {
		b.WriteString("可用 Skills（详见各 `skills/<name>/SKILL.md`）：\n")
		for _, s := range list {
			b.WriteString(fmt.Sprintf("- `%s` — %s\n", s.Name, firstSentence(s.Description)))
		}
		b.WriteString("\n")
	}

	b.WriteString("安全：禁止输出 Token 明文；写入/删除前先确认意图；危险操作用 `--yes` 或 `--dry-run` 预览。\n")
	return b.String()
}

// firstSentence trims a description to its leading clause for a compact bullet.
func firstSentence(desc string) string {
	if desc == "" {
		return "见 SKILL.md"
	}
	for _, sep := range []string{"。", ". ", "；", ";"} {
		if i := strings.Index(desc, sep); i > 0 {
			return strings.TrimSpace(desc[:i])
		}
	}
	return desc
}

// upsertManagedBlock writes content between the managed markers in path,
// replacing an existing block or appending a new one, preserving other content.
func upsertManagedBlock(path, content string, dryRun bool) error {
	block := blockBegin + "\n" + content + blockEnd
	if dryRun {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return os.WriteFile(path, []byte(block+"\n"), 0o644)
	}

	updated, replaced := replaceBlock(string(existing), block)
	if !replaced {
		sep := "\n"
		if !strings.HasSuffix(updated, "\n") {
			sep = "\n\n"
		} else if !strings.HasSuffix(updated, "\n\n") {
			sep = "\n"
		}
		updated = updated + sep + block + "\n"
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}

// replaceBlock swaps the existing managed region (markers inclusive) for block.
// Returns the new text and whether a region was found.
func replaceBlock(text, block string) (string, bool) {
	start := strings.Index(text, blockBegin)
	if start < 0 {
		return text, false
	}
	end := strings.Index(text[start:], blockEnd)
	if end < 0 {
		return text, false
	}
	end = start + end + len(blockEnd)
	return text[:start] + block + text[end:], true
}

// removeManagedBlock strips the managed region from path; if nothing else
// remains, the file is deleted.
func removeManagedBlock(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	text := string(data)
	start := strings.Index(text, blockBegin)
	if start < 0 {
		return nil
	}
	end := strings.Index(text[start:], blockEnd)
	if end < 0 {
		return nil
	}
	end = start + end + len(blockEnd)
	cleaned := strings.TrimRight(text[:start], "\n")
	rest := strings.TrimLeft(text[end:], "\n")
	if rest != "" {
		cleaned = cleaned + "\n\n" + rest
	}
	if strings.TrimSpace(cleaned) == "" {
		return os.Remove(path)
	}
	if !strings.HasSuffix(cleaned, "\n") {
		cleaned += "\n"
	}
	return os.WriteFile(path, []byte(cleaned), 0o644)
}

func hasManagedBlock(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), blockBegin)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
