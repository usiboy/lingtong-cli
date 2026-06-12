// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"

	"github.com/lingtong/cli/cmd/api"
	"github.com/lingtong/cli/cmd/app"
	"github.com/lingtong/cli/cmd/auth"
	"github.com/lingtong/cli/cmd/config"
	"github.com/lingtong/cli/cmd/connector"
	"github.com/lingtong/cli/cmd/model"
	"github.com/lingtong/cli/cmd/scene"
	"github.com/lingtong/cli/cmd/table"
	"github.com/lingtong/cli/cmd/workflow"
	"github.com/lingtong/cli/internal/build"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/lingtong/cli/shortcuts"
	"github.com/spf13/cobra"
)

const rootLong = `lingtong-cli — Lingtong iPaaS CLI tool for AI Agents.

USAGE:
    lingtong-cli <command> [subcommand] [options]
    lingtong-cli api <method> <path> [--params <json>] [--data <json>]

EXAMPLES:
    # Configure lingtong host
    lingtong-cli config init

    # Login and get token
    lingtong-cli auth login

    # Query connector info
    lingtong-cli connector info --connector kmerp

    # List scenes
    lingtong-cli scene list

    # Execute workflow
    lingtong-cli workflow execute --workflow-id 123

FLAGS:
    --format <fmt>        output format: json (default) | table | pretty
    --host <url>          override configured host
    --dry-run             print request without executing

AI AGENT SKILLS:
    lingtong-cli pairs with AI agent skills that teach the agent
    Lingtong API patterns, best practices, and workflows.

    Skills are located in the skills/ directory:
    - lingtong-shared: Auth, config, security rules
    - lingtong-connector: Connector management
    - lingtong-scene: Scene management
    - lingtong-workflow: Workflow orchestration
    - lingtong-table: Table operations

COMMUNITY:
    Docs:       docs/knowledge/lingtong-product-knowledge.md

More help: lingtong-cli <command> --help`

// Execute runs the root command and returns the process exit code.
func Execute() int {
	f := cmdutil.NewDefault()
	rootCmd := NewRootCommand(f)

	if err := rootCmd.Execute(); err != nil {
		return handleRootError(f, err)
	}
	return 0
}

// NewRootCommand creates and returns the root cobra command with all subcommands registered.
// This is exported for use in documentation generation and testing.
func NewRootCommand(f *cmdutil.Factory) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "lingtong-cli",
		Short:   "Lingtong iPaaS CLI — AI Agent integration tool",
		Long:    rootLong,
		Version: build.Version,
	}

	cmdutil.InstallHelpFunc(rootCmd)
	rootCmd.SilenceErrors = true
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	}

	// Register subcommands
	rootCmd.AddCommand(config.NewCmdConfig(f))
	rootCmd.AddCommand(auth.NewCmdAuth(f))
	rootCmd.AddCommand(connector.NewCmdConnector(f))
	rootCmd.AddCommand(scene.NewCmdScene(f))
	rootCmd.AddCommand(workflow.NewCmdWorkflow(f))
	rootCmd.AddCommand(table.NewCmdTable(f))
	rootCmd.AddCommand(model.NewCmdModel(f))
	rootCmd.AddCommand(api.NewCmdApi(f))
	rootCmd.AddCommand(app.NewCmdApp(f))

	// Register shortcuts
	shortcuts.RegisterShortcuts(rootCmd, f)

	return rootCmd
}

// handleRootError dispatches a command error to the appropriate handler.
func handleRootError(f *cmdutil.Factory, err error) int {
	errOut := f.IOStreams.ErrOut
	fmt.Fprintln(errOut, "Error:", err)
	return 1
}

// formatFlag returns the output format from the command flags.
func formatFlag(cmd *cobra.Command) output.Format {
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		return output.FormatJSON
	case "table":
		return output.FormatTable
	case "pretty":
		return output.FormatPretty
	default:
		return output.FormatJSON
	}
}
