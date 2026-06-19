// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// ImportResult represents the response from the import API.
type ImportResult struct {
	AppName       string   `json:"appName"`
	CreatedScenes []string `json:"createdScenes"`
	CreatedTables []string `json:"createdTables"`
	Message       string   `json:"message"`
}

func newCmdAppImport(f *cmdutil.Factory) *cobra.Command {
	var filePath string
	var dryRun, force bool

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import application configuration from JSON",
		Long: `Import an application configuration from a JSON file into the Lingtong platform.

Performs pre-flight validation before importing unless --force is used.
Use --dry-run to preview what would be imported without making changes.

EXAMPLES:
    lingtong-cli app import --file app.json
    lingtong-cli app import --file app.json --dry-run
    lingtong-cli app import --file app.json --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			// Read and parse the input file
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			var appExport AppExport
			if err := json.Unmarshal(data, &appExport); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			// Pre-flight validation (unless --force)
			// Use ValidateRawJSON for import (checks structure, null values, unknown keys, payload size)
			// Skip circular dependency check since bidirectional sync is common in ERP integrations
			if !force {
				if err := ValidateRawJSON(data); err != nil {
					return fmt.Errorf("pre-flight validation failed: %w", err)
				}
			}

			// Dry-run mode
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Would import application: %s\n", appExport.AppName)
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Connectors: %d\n", len(appExport.AppConnectors))
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Scenes: %d\n", len(appExport.Scenes))
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Workflows: %d\n", len(appExport.Workflows))
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Would POST: /application/import\n")
				return nil
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			// Perform the actual import
			c := client.NewClient(f.Config.Host, f.Config.Token)

			// Parse the raw JSON to send as-is (preserves original structure)
			var rawBody interface{}
			if err := json.Unmarshal(data, &rawBody); err != nil {
				return fmt.Errorf("failed to parse JSON body: %w", err)
			}

			resp, err := c.Post("/application/import", rawBody)
			if err != nil {
				return fmt.Errorf("failed to import application: %w", err)
			}

			// Parse and display the result
			var result ImportResult
			if err := json.Unmarshal(resp, &result); err != nil {
				// If response is not in expected format, print raw response
				fmt.Fprintf(cmd.OutOrStdout(), "Import response: %s\n", string(resp))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully imported application: %s\n", result.AppName)
			if result.Message != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Message: %s\n", result.Message)
			}
			if len(result.CreatedScenes) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Created scenes: %s\n", strings.Join(result.CreatedScenes, ", "))
			}
			if len(result.CreatedTables) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Created tables: %s\n", strings.Join(result.CreatedTables, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Input JSON file path (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be imported without executing")
	cmd.Flags().BoolVar(&force, "force", false, "Skip pre-flight validation checks")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
