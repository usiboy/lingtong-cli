// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdApp creates the app command with subcommands.
func NewCmdApp(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage Lingtong applications",
		Long:  "Export and validate Lingtong application configurations.",
	}

	cmd.AddCommand(newCmdAppExport(f))
	cmd.AddCommand(newCmdAppImport(f))
	cmd.AddCommand(newCmdAppValidate(f))
	cmd.AddCommand(newCmdAppScaffold(f))
	cmd.AddCommand(newCmdAppDiff(f))
	cmd.AddCommand(newCmdAppScene(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}
