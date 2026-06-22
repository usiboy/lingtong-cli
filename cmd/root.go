// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/lingtong/cli/cmd/api"
	"github.com/lingtong/cli/cmd/app"
	"github.com/lingtong/cli/cmd/auth"
	"github.com/lingtong/cli/cmd/completion"
	"github.com/lingtong/cli/cmd/config"
	"github.com/lingtong/cli/cmd/connector"
	"github.com/lingtong/cli/cmd/doctor"
	"github.com/lingtong/cli/cmd/factory"
	"github.com/lingtong/cli/cmd/model"
	"github.com/lingtong/cli/cmd/scene"
	"github.com/lingtong/cli/cmd/schema"
	"github.com/lingtong/cli/cmd/service"
	"github.com/lingtong/cli/cmd/skills"
	"github.com/lingtong/cli/cmd/table"
	"github.com/lingtong/cli/cmd/update"
	"github.com/lingtong/cli/cmd/workflow"
	ltauth "github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/build"
	"github.com/lingtong/cli/internal/cmdutil"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/notice"
	"github.com/lingtong/cli/internal/openapi"
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
    --envelope            wrap output in {ok, data, error} envelope
    --jq, -q <expr>       filter output with jq expression

AI AGENT SKILLS:
    lingtong-cli pairs with AI agent skills that teach the agent
    Lingtong API patterns, best practices, and workflows.

    Install them into your AI editor (Claude Code, OpenCode, Qoder,
    Cursor, Trae, Codex) — skills are embedded, no npx required:
        lingtong-cli skills install

    Bundled skills:
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

		// Resolve --profile flag
		if profileName, _ := cmd.Flags().GetString("profile"); profileName != "" {
			f.Profile = profileName
		}
		// Apply profile: override host/brand from profile
		if host, brand := f.Config.ResolveProfile(f.Profile); host != "" {
			f.Config.Host = host
			if brand != "" {
				f.Config.Brand = brand
			}
		}
		// Load the profile-scoped token so switching profiles also switches
		// credentials. An explicit LINGTONG_TOKEN env var always wins, and the
		// default-profile token loaded in NewDefault is left untouched.
		if effective := f.EffectiveProfile(); effective != "" && os.Getenv("LINGTONG_TOKEN") == "" {
			if token, err := ltauth.GetTokenForProfile(effective); err == nil && token != "" {
				f.Config.Token = token
			}
		}

		// Only override OmitNull if the flag was explicitly set by the user
		if cmd.Flags().Changed("omit-null") {
			omitNull, _ := cmd.Flags().GetBool("omit-null")
			f.Config.OmitNull = omitNull
		}

		// Handle --envelope flag
		if cmd.Flags().Changed("envelope") {
			env, _ := cmd.Flags().GetBool("envelope")
			f.Envelope = env
		}

		// Handle --jq flag (validate early)
		if jqExpr, _ := cmd.Flags().GetString("jq"); jqExpr != "" {
			if err := output.ValidateJqExpression(jqExpr); err != nil {
				fmt.Fprintln(f.IOStreams.ErrOut, "Error:", err)
				os.Exit(output.ExitValidation)
			}
			f.JqExpr = jqExpr
		}

		// Start a best-effort, non-blocking notice fetch (only when envelope is
		// enabled). The result is injected into output if it is ready by the
		// time we write; it never delays the command (see Factory.NewWriter).
		if f.Envelope {
			ch := make(chan *output.Notice, 1)
			go func() {
				ch <- notice.FetchNotices(context.Background())
			}()
			f.NoticeChan = ch
		}
	}

	// Register subcommands
	rootCmd.AddCommand(config.NewCmdConfig(f))
	rootCmd.AddCommand(auth.NewCmdAuth(f))
	rootCmd.AddCommand(connector.NewCmdConnector(f))
	rootCmd.AddCommand(scene.NewCmdScene(f))
	rootCmd.AddCommand(workflow.NewCmdWorkflow(f))
	rootCmd.AddCommand(table.NewCmdTable(f))
	rootCmd.AddCommand(factory.NewCmdFactory(f))
	rootCmd.AddCommand(model.NewCmdModel(f))
	rootCmd.AddCommand(api.NewCmdApi(f))
	rootCmd.AddCommand(app.NewCmdApp(f))

	// Register P0 enhancement commands
	rootCmd.AddCommand(completion.NewCmdCompletion(f))
	rootCmd.AddCommand(doctor.NewCmdDoctor(f))
	rootCmd.AddCommand(skills.NewCmdSkills(f))

	// Register P2 enhancement commands
	rootCmd.AddCommand(schema.NewCmdSchema(f))
	rootCmd.AddCommand(update.NewCmdUpdate(f))

	// Register P0.5 auto-generated service commands.
	// The spec is loaded from a runtime override if present, otherwise from
	// the copy embedded in the binary, so `service` works out of the box.
	if spec, err := openapi.LoadSpec(); err == nil && spec != nil {
		service.RegisterServiceCommands(rootCmd, f, spec)
	}

	// Deprecated: basicdata has been merged into 'table data'.
	rootCmd.AddCommand(&cobra.Command{
		Use:   "basicdata",
		Short: "Deprecated: use 'table data' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("'basicdata' command has been removed. Use 'table data' instead. Run 'lingtong-cli table data --help' for details")
		},
	})

	// Register shortcuts
	shortcuts.RegisterShortcuts(rootCmd, f)

	// Add global flags
	rootCmd.PersistentFlags().Bool("omit-null", true, "Omit null fields in JSON output (default: true)")
	rootCmd.PersistentFlags().Bool("envelope", false, "Wrap output in standard envelope {ok, data, error}")
	rootCmd.PersistentFlags().StringP("jq", "q", "", "jq expression to filter output (implies JSON output)")
	rootCmd.PersistentFlags().String("profile", "", "Use a named configuration profile (e.g., dev, staging, prod)")

	return rootCmd
}

// handleRootError dispatches a command error to the appropriate handler.
// Returns the process exit code based on the error type.
//
// The raw error is first classified (cobra flag errors and hand-written
// validation checks become typed) so the exit code is meaningful. When
// --envelope is enabled, a machine-readable failure envelope is written to
// stdout; otherwise a human-readable message (plus hint) goes to stderr.
func handleRootError(f *cmdutil.Factory, err error) int {
	classified := lterrors.Classify(err)
	exitCode := output.ExitCodeOf(classified)

	if f.Envelope {
		_ = output.WriteErrorEnvelope(f.IOStreams.Out, "", classified)
		return exitCode
	}

	errOut := f.IOStreams.ErrOut
	fmt.Fprintln(errOut, "Error:", err)
	if p, ok := lterrors.ProblemOf(classified); ok && p.Hint != "" {
		fmt.Fprintf(errOut, "Hint: %s\n", p.Hint)
	}
	return exitCode
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

// The OpenAPI spec used to generate `service` commands is resolved by
// openapi.LoadSpec (env override → ~/.lingtong-cli/openapi.json → embedded),
// shared with the `schema` and `doctor` commands so they all agree on the API.
