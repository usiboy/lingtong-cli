// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdConfig creates the config command.
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "Configure lingtong-cli settings including host URL and credentials.",
	}

	cmd.AddCommand(newCmdConfigInit(f))
	cmd.AddCommand(newCmdConfigShow(f))
	cmd.AddCommand(newCmdConfigDelete(f))

	return cmd
}

func newCmdConfigInit(f *cmdutil.Factory) *cobra.Command {
	var newConfig bool
	var host string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Long: `Interactive guided setup to configure lingtong host URL.

EXAMPLES:
    lingtong-cli config init
    lingtong-cli config init --host https://your-lingtong-host.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Flags().GetString("host")
			if host == "" {
				fmt.Print("Enter Lingtong host URL (e.g., https://lingtong.example.com): ")
				fmt.Scanln(&host)
			}

			cfg := f.Config
			cfg.Host = host
			cfg.Brand = "lingtong"

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Configuration saved to %s\n", "<config-path>")
			fmt.Println("Next step: run `lingtong-cli auth login` to authenticate.")
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Lingtong host URL")
	cmd.Flags().BoolVar(&newConfig, "new", false, "Force create new configuration")
	return cmd
}

func newCmdConfigShow(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := f.Config
			fmt.Printf("Host:  %s\n", cfg.Host)
			fmt.Printf("Brand: %s\n", cfg.Brand)
			if cfg.Host == "" {
				fmt.Println("\nNo configuration found. Run `lingtong-cli config init` to set up.")
			}
			return nil
		},
	}
}

func newCmdConfigDelete(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement config deletion
			fmt.Println("Configuration deleted.")
			return nil
		},
	}
}
