// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package skills implements the `lingtong-cli skills` command group, which
// installs the embedded Agent Skills into AI coding editors (Claude Code,
// OpenCode, Qoder, Cursor, Trae, Codex) so those agents can discover the CLI,
// its commands, and its skills.
package skills

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lingtong/cli/internal/cmdutil"
	skillspkg "github.com/lingtong/cli/internal/skills"
	"github.com/spf13/cobra"
)

// NewCmdSkills creates the `skills` command group.
func NewCmdSkills(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Install lingtong skills into AI editors (Claude Code, OpenCode, Qoder, Cursor, Trae, Codex)",
		Long: `Deploy the lingtong-cli Agent Skills so AI coding editors can discover the
CLI, its commands, and its skills.

The skills are embedded in the binary, so no source tree or npx is required.
Supported editors: claude, opencode, qoder, cursor, trae, codex.

EXAMPLES:
    # Auto-detect editors in the current project and the user home, install to all
    lingtong-cli skills install

    # Only specific editors
    lingtong-cli skills install --editor claude,opencode

    # Force a scope (global = user home, project = current repo)
    lingtong-cli skills install --scope global
    lingtong-cli skills install --scope project

    # Preview without writing
    lingtong-cli skills install --dry-run

    # Inspect / remove
    lingtong-cli skills list
    lingtong-cli skills status
    lingtong-cli skills uninstall --editor cursor`,
	}

	cmd.AddCommand(newInstallCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newStatusCmd(f))
	cmd.AddCommand(newUninstallCmd(f))
	return cmd
}

// editorFlag, scopeFlag, dryRunFlag are shared across subcommands.
type commonFlags struct {
	editor string
	scope  string
	dryRun bool
}

func addCommonFlags(cmd *cobra.Command, fl *commonFlags, withDryRun bool) {
	cmd.Flags().StringVar(&fl.editor, "editor", "all", "Comma-separated editors (claude,opencode,qoder,cursor,trae,codex) or 'all'")
	cmd.Flags().StringVar(&fl.scope, "scope", "", "Install scope: global | project (default: auto-detect)")
	if withDryRun {
		cmd.Flags().BoolVar(&fl.dryRun, "dry-run", false, "Show what would change without writing")
	}
}

func (fl commonFlags) options() (skillspkg.Options, error) {
	scope, err := parseScope(fl.scope)
	if err != nil {
		return skillspkg.Options{}, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return skillspkg.Options{}, fmt.Errorf("resolving working directory: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return skillspkg.Options{}, fmt.Errorf("resolving home directory: %w", err)
	}
	return skillspkg.Options{
		Editors: parseEditors(fl.editor),
		Scope:   scope,
		CWD:     cwd,
		Home:    home,
		DryRun:  fl.dryRun,
	}, nil
}

func parseEditors(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" || v == "all" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseScope(v string) (skillspkg.Scope, error) {
	switch strings.TrimSpace(v) {
	case "":
		return "", nil
	case "global":
		return skillspkg.ScopeGlobal, nil
	case "project":
		return skillspkg.ScopeProject, nil
	default:
		return "", fmt.Errorf("invalid --scope %q (want: global or project)", v)
	}
}

func newInstallCmd(f *cmdutil.Factory) *cobra.Command {
	var fl commonFlags
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install skills into AI editors",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := fl.options()
			if err != nil {
				return err
			}
			res, err := skillspkg.Install(opts)
			if err != nil {
				return err
			}
			printInstall(f.IOStreams.Out, res, opts.DryRun)
			return firstErr(res)
		},
	}
	addCommonFlags(cmd, &fl, true)
	return cmd
}

func newUninstallCmd(f *cmdutil.Factory) *cobra.Command {
	var fl commonFlags
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed skills and the managed instructions block",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := fl.options()
			if err != nil {
				return err
			}
			res, err := skillspkg.Uninstall(opts)
			if err != nil {
				return err
			}
			printUninstall(f.IOStreams.Out, res, opts.DryRun)
			return firstErr(res)
		},
	}
	addCommonFlags(cmd, &fl, true)
	return cmd
}

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	var fl commonFlags
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show where skills are installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := fl.options()
			if err != nil {
				return err
			}
			res, err := skillspkg.Status(opts)
			if err != nil {
				return err
			}
			printStatus(f.IOStreams.Out, res)
			return nil
		},
	}
	addCommonFlags(cmd, &fl, false)
	return cmd
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the skills embedded in this binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := skillspkg.List()
			if err != nil {
				return err
			}
			out := f.IOStreams.Out
			if len(list) == 0 {
				fmt.Fprintln(out, "No skills are embedded in this build.")
				return nil
			}
			fmt.Fprintf(out, "Embedded skills (%d):\n", len(list))
			for _, s := range list {
				fmt.Fprintf(out, "  - %s\n", s.Name)
				if s.Description != "" {
					fmt.Fprintf(out, "      %s\n", s.Description)
				}
			}
			return nil
		},
	}
}

func printInstall(out io.Writer, res *skillspkg.Result, dryRun bool) {
	if dryRun {
		fmt.Fprintln(out, "Skills install (dry-run — no files written):")
	} else {
		fmt.Fprintln(out, "Skills install:")
	}
	for _, t := range res.Targets {
		if t.Skipped {
			fmt.Fprintf(out, "  [--] %-9s skipped: %s\n", t.Editor, t.SkippedReason)
			continue
		}
		if t.Err != nil {
			fmt.Fprintf(out, "  [XX] %-9s (%s): %v\n", t.Editor, t.Scope, t.Err)
			continue
		}
		fmt.Fprintf(out, "  [OK] %-9s (%s)\n", t.Editor, t.Scope)
		if t.SkillsDir != "" {
			fmt.Fprintf(out, "         %d skills -> %s\n", t.SkillsInstalled, t.SkillsDir)
		}
		if t.InstructionsWritten {
			fmt.Fprintf(out, "         instructions -> %s\n", t.InstructionsPath)
		}
	}
}

func printUninstall(out io.Writer, res *skillspkg.Result, dryRun bool) {
	if dryRun {
		fmt.Fprintln(out, "Skills uninstall (dry-run — no files removed):")
	} else {
		fmt.Fprintln(out, "Skills uninstall:")
	}
	for _, t := range res.Targets {
		if t.Err != nil {
			fmt.Fprintf(out, "  [XX] %-9s (%s): %v\n", t.Editor, t.Scope, t.Err)
			continue
		}
		if t.Removed == 0 && !t.InstructionsWritten {
			continue
		}
		fmt.Fprintf(out, "  [OK] %-9s (%s): removed %d skills", t.Editor, t.Scope, t.Removed)
		if t.InstructionsWritten {
			fmt.Fprintf(out, ", cleared instructions block")
		}
		fmt.Fprintln(out)
	}
}

func printStatus(out io.Writer, res *skillspkg.Result) {
	fmt.Fprintln(out, "Skills status:")
	any := false
	for _, t := range res.Targets {
		if t.SkillsInstalled == 0 && !t.InstructionsWritten {
			continue
		}
		any = true
		marker := "instructions only"
		if t.SkillsDir != "" {
			marker = fmt.Sprintf("%d skills @ %s", t.SkillsInstalled, t.SkillsDir)
		}
		fmt.Fprintf(out, "  [OK] %-9s (%s): %s\n", t.Editor, t.Scope, marker)
	}
	if !any {
		fmt.Fprintln(out, "  (nothing installed — run 'lingtong-cli skills install')")
	}
}

// firstErr surfaces the first per-target error so the command exits non-zero
// when any target failed, after the full report has been printed.
func firstErr(res *skillspkg.Result) error {
	for _, t := range res.Targets {
		if t.Err != nil {
			return t.Err
		}
	}
	return nil
}
