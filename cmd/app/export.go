// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdAppExport(f *cmdutil.Factory) *cobra.Command {
	var appID int
	var output string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export application configuration",
		Long: `Export an application configuration to a JSON file.

EXAMPLES:
    lingtong-cli app export --app-id 123 --output app.json
    lingtong-cli app export --app-id 123 --output app.json --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appID == 0 {
				return fmt.Errorf("--app-id is required")
			}
			if output == "" {
				return fmt.Errorf("--output is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)

			// Step 1: Get export sign via proxy
			signResp, err := c.Get("/gw/ai/application/export/sign", nil)
			if err != nil {
				return fmt.Errorf("failed to get export sign: %w", err)
			}

			var signResult map[string]interface{}
			if err := json.Unmarshal(signResp, &signResult); err != nil {
				return fmt.Errorf("invalid sign response: %w", err)
			}

			// Backend wraps responses in Result{result: {...}}
			result, ok := signResult["result"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("missing result in sign response")
			}
			sign, ok := result["sign"].(string)
			if !ok || sign == "" {
				return fmt.Errorf("missing sign in response")
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Would GET: /application/export?sign=%s&appId=%d\n", sign, appID)
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Would save to: %s\n", output)
				return nil
			}

			// Step 2: Export application via proxy
			resp, err := c.Get("/gw/ai/application/export", map[string]interface{}{
				"sign":  sign,
				"appId": appID,
			})
			if err != nil {
				return fmt.Errorf("failed to export application: %w", err)
			}

			// Validate JSON before saving
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return fmt.Errorf("invalid JSON response: %w", err)
			}

			// Format JSON with indentation
			formatted, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}

			if err := os.WriteFile(output, formatted, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", output, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Exported application to %s\n", output)
			return nil
		},
	}

	cmd.Flags().IntVar(&appID, "app-id", 0, "Application ID (required)")
	cmd.Flags().StringVar(&output, "output", "", "Output file path (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be done without executing")
	_ = cmd.MarkFlagRequired("app-id")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}
