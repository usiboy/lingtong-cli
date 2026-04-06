// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package shortcuts

import (
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// RegisterShortcuts registers all shortcut commands to the root command.
func RegisterShortcuts(rootCmd *cobra.Command, f *cmdutil.Factory) {
	rootCmd.AddCommand(newShortcutConnectorInfo(f))
	rootCmd.AddCommand(newShortcutSceneList(f))
	rootCmd.AddCommand(newShortcutWorkflowExecute(f))
}

func newShortcutConnectorInfo(f *cmdutil.Factory) *cobra.Command {
	var connector, env string
	cmd := &cobra.Command{
		Use:   "+connector-info",
		Short: "Quick query connector info",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to connector info command
			cmd.Root().SetArgs([]string{"connector", "info", "--connector", connector, "--env", env})
			return cmd.Root().Execute()
		},
	}
	cmd.Flags().StringVar(&connector, "connector", "", "Connector identifier")
	cmd.Flags().StringVar(&env, "env", "test", "Environment")
	_ = cmd.MarkFlagRequired("connector")
	return cmd
}

func newShortcutSceneList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "+scene-list",
		Short: "Quick list scenes",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Root().SetArgs([]string{"scene", "list"})
			return cmd.Root().Execute()
		},
	}
	return cmd
}

func newShortcutWorkflowExecute(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	cmd := &cobra.Command{
		Use:   "+workflow-execute",
		Short: "Quick execute workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Root().SetArgs([]string{"workflow", "execute", "--workflow-id", string(rune(workflowId))})
			return cmd.Root().Execute()
		},
	}
	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}
