// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package scene

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

const missingHostConfigMessage = "no host configured. Run `lingtong-cli config init --host <url>` first"

// NewCmdScene creates the scene command.
func NewCmdScene(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scene",
		Short: "Manage integration scenes",
		Long:  "Create, query, and manage integration scenes (scenarios).",
	}

	cmd.AddCommand(newCmdSceneList(f))
	cmd.AddCommand(newCmdSceneCreate(f))
	cmd.AddCommand(newCmdSceneInfo(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func newCmdSceneList(f *cmdutil.Factory) *cobra.Command {
	var page, pageSize int
	var appID string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List integration scenes",
		Long: `List all integration scenes.

EXAMPLES:
    lingtong-cli scene list
    lingtong-cli scene list --page 1 --page-size 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if page < 1 {
				return fmt.Errorf("--page must be at least 1")
			}
			if pageSize < 1 {
				return fmt.Errorf("--page-size must be at least 1")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			params := map[string]interface{}{
				"pageNum":  page,
				"pageSize": pageSize,
			}
			if appID != "" {
				params["appId"] = appID
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/scene/list", params)
			if err != nil {
				return err
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&appID, "app-id", "", "Application ID filter")
	return cmd
}

func newCmdSceneCreate(f *cmdutil.Factory) *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new integration scene",
		Long: `Create a new integration scene.

EXAMPLES:
    lingtong-cli scene create --name "Order Sync" --description "Sync orders from ERP to CRM"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			body := map[string]interface{}{
				"name":        name,
				"description": description,
			}

			resp, err := c.Post("/scene/model/save", body)
			if err != nil {
				return err
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Scene name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Scene description")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newCmdSceneInfo(f *cmdutil.Factory) *cobra.Command {
	var sceneId int
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Query scene details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sceneId == 0 {
				return fmt.Errorf("--scene-id is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/scene/detail/get", map[string]interface{}{"sceneId": sceneId})
			if err != nil {
				return err
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().IntVar(&sceneId, "scene-id", 0, "Scene ID (required)")
	_ = cmd.MarkFlagRequired("scene-id")
	return cmd
}

func requireHostConfigured(host string) error {
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf(missingHostConfigMessage)
	}
	return nil
}
