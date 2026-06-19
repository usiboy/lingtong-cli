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

func newCmdAppList(f *cmdutil.Factory) *cobra.Command {
	var pageNum int
	var pageSize int
	var name string
	var appID string
	var tenantID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List applications",
		Long: `List Lingtong applications without modifying platform state.

EXAMPLES:
    lingtong-cli app list
    lingtong-cli app list --page-num 2 --page-size 20 --name "Sales Sync"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageNum < 1 {
				return fmt.Errorf("--page-num must be at least 1")
			}
			if pageSize < 1 {
				return fmt.Errorf("--page-size must be at least 1")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			params := map[string]interface{}{
				"pageNum":  pageNum,
				"pageSize": pageSize,
			}
			if name != "" {
				params["name"] = name
			}
			if appID != "" {
				params["appId"] = appID
			}
			if tenantID != "" {
				params["tenantId"] = tenantID
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/application/list", params)
			if err != nil {
				return err
			}

			var envelope struct {
				Success *bool   `json:"success"`
				Msg     string  `json:"msg"`
				Data    interface{} `json:"-"`
			}
			var raw map[string]interface{}
			if err := json.Unmarshal(resp, &raw); err != nil {
				return fmt.Errorf("invalid JSON response: %w", err)
			}
			if err := json.Unmarshal(resp, &envelope); err != nil {
				return fmt.Errorf("invalid JSON response: %w", err)
			}
			envelope.Data = raw

			if envelope.Success != nil && !*envelope.Success {
				message := envelope.Msg
				if message == "" {
					message = "success is false"
				}
				return fmt.Errorf("application list failed: %s", message)
			}

			format := output.Format("json")
			if formatFlag := cmd.Flag("format"); formatFlag != nil {
				format = output.Format(formatFlag.Value.String())
			}
			w := output.NewWriter(appListIOStreams(f, cmd), format)
			if err := w.Write(envelope.Data); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&pageNum, "page-num", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 40, "Page size")
	cmd.Flags().StringVar(&name, "name", "", "Application name filter")
	cmd.Flags().StringVar(&appID, "app-id", "", "Application ID filter")
	cmd.Flags().StringVar(&tenantID, "tenant-id", "", "Tenant ID filter")

	return cmd
}

func appListIOStreams(f *cmdutil.Factory, cmd *cobra.Command) *output.IOStreams {
	if f.IOStreams != nil {
		return f.IOStreams
	}
	return &output.IOStreams{
		Out:    cmd.OutOrStdout(),
		ErrOut: cmd.ErrOrStderr(),
	}
}
