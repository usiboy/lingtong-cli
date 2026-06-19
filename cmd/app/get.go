// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"fmt"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func newCmdAppGet(f *cmdutil.Factory) *cobra.Command {
	var applicationID string
	var idAlias string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get application details",
		Long: `Get Lingtong application details without modifying platform state.

EXAMPLES:
    lingtong-cli app get --application-id 123
    lingtong-cli app get --id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedID, err := appGetSelectedID(cmd, applicationID, idAlias)
			if err != nil {
				return err
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/application/get", map[string]interface{}{"applicationId": selectedID})
			if err != nil {
				return err
			}

			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return fmt.Errorf("invalid JSON response: %w", err)
			}

			format := output.Format("json")
			if formatFlag := cmd.Flag("format"); formatFlag != nil {
				format = output.Format(formatFlag.Value.String())
			}
			w := output.NewWriter(appGetIOStreams(f, cmd), format)
			if err := w.Write(data); err != nil {
				return err
			}

			var envelope struct {
				Success *bool  `json:"success"`
				Msg     string `json:"msg"`
			}
			if err := json.Unmarshal(resp, &envelope); err != nil {
				return fmt.Errorf("invalid JSON response: %w", err)
			}
			if envelope.Success != nil && !*envelope.Success {
				message := envelope.Msg
				if message == "" {
					message = "success is false"
				}
				return fmt.Errorf("application get failed: %s", message)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&applicationID, "application-id", "", "Application ID (required)")
	cmd.Flags().StringVar(&idAlias, "id", "", "Alias for --application-id")

	return cmd
}

func appGetSelectedID(cmd *cobra.Command, applicationID string, idAlias string) (string, error) {
	applicationIDChanged := cmd.Flags().Changed("application-id")
	idAliasChanged := cmd.Flags().Changed("id")

	if applicationIDChanged && idAliasChanged && applicationID != idAlias {
		return "", fmt.Errorf("--application-id and --id must match when both are set")
	}

	selectedID := applicationID
	if selectedID == "" {
		selectedID = idAlias
	}
	if selectedID == "" {
		return "", fmt.Errorf("--application-id is required")
	}

	return selectedID, nil
}

func appGetIOStreams(f *cmdutil.Factory, cmd *cobra.Command) *output.IOStreams {
	if f.IOStreams != nil {
		return f.IOStreams
	}
	return &output.IOStreams{
		Out:    cmd.OutOrStdout(),
		ErrOut: cmd.ErrOrStderr(),
	}
}
