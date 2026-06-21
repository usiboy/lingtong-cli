// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package scene

import (
	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// This file extends the scene command group with lifecycle operations:
// update, delete, copy, open/close, publish, version, trigger, field-mapping.
// Destructive operations are confirmation-gated (--yes).

// ==================== scene update ====================

func newCmdSceneUpdate(f *cmdutil.Factory) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a scene",
		Long: `Update a scene from a JSON definition. Wraps POST /scene/model/update.

EXAMPLES:
    lingtong-cli scene update --data '{"id":123,"name":"renamed"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(data != "", "data"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			body, err := cmdutil.ParseDataObject(data)
			if err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/scene/model/update", body)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.update")
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "Scene definition (JSON object, required)")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

// ==================== scene delete ====================

func newCmdSceneDelete(f *cmdutil.Factory) *cobra.Command {
	var sceneID int
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a scene (requires --yes)",
		Long: `Delete a scene by ID. Wraps POST /scene/delete.

This is destructive and requires --yes.

EXAMPLES:
    lingtong-cli scene delete --scene-id 123 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(sceneID != 0, "scene-id"); err != nil {
				return err
			}
			if err := cmdutil.ConfirmDestructive(yes, "delete scene"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/scene/delete", sceneID)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.delete")
		},
	}
	cmd.Flags().IntVar(&sceneID, "scene-id", 0, "Scene ID (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm the destructive operation")
	_ = cmd.MarkFlagRequired("scene-id")
	return cmd
}

// ==================== scene copy ====================

func newCmdSceneCopy(f *cmdutil.Factory) *cobra.Command {
	var sceneID int
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy a scene",
		Long: `Duplicate a scene by ID. Wraps POST /scene/copy.

EXAMPLES:
    lingtong-cli scene copy --scene-id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(sceneID != 0, "scene-id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/scene/copy", sceneID)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.copy")
		},
	}
	cmd.Flags().IntVar(&sceneID, "scene-id", 0, "Scene ID (required)")
	_ = cmd.MarkFlagRequired("scene-id")
	return cmd
}

// ==================== scene open (enable/disable) ====================

func newCmdSceneOpen(f *cmdutil.Factory) *cobra.Command {
	var sceneID int
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Toggle a scene on/off",
		Long: `Enable or disable a scene (toggles current state). Wraps POST /scene/open.

EXAMPLES:
    lingtong-cli scene open --scene-id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(sceneID != 0, "scene-id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/scene/open", sceneID)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.open")
		},
	}
	cmd.Flags().IntVar(&sceneID, "scene-id", 0, "Scene ID (required)")
	_ = cmd.MarkFlagRequired("scene-id")
	return cmd
}

// ==================== scene publish ====================

func newCmdScenePublish(f *cmdutil.Factory) *cobra.Command {
	var sceneID int
	var data string
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a scene version",
		Long: `Publish a scene version (release). Wraps POST /scene/publish/release.

Provide --scene-id for a simple release, or --data with a full snapshot payload.

EXAMPLES:
    lingtong-cli scene publish --scene-id 123
    lingtong-cli scene publish --data '{"sceneId":123,"remark":"v2"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			var body interface{}
			if data != "" {
				m, err := cmdutil.ParseDataObject(data)
				if err != nil {
					return err
				}
				body = m
			} else {
				if err := cmdutil.RequireFlag(sceneID != 0, "scene-id"); err != nil {
					return err
				}
				body = map[string]interface{}{"sceneId": sceneID}
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/scene/publish/release", body)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.publish")
		},
	}
	cmd.Flags().IntVar(&sceneID, "scene-id", 0, "Scene ID (required unless --data given)")
	cmd.Flags().StringVar(&data, "data", "", "Full snapshot payload (JSON object)")
	return cmd
}

// ==================== scene version list ====================

func newCmdSceneVersion(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Scene version operations",
	}
	cmd.AddCommand(newCmdSceneVersionList(f))
	return cmd
}

func newCmdSceneVersionList(f *cmdutil.Factory) *cobra.Command {
	var sceneID int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a scene's versions",
		Long: `List version history for a scene. Wraps GET /scene/version/list.

EXAMPLES:
    lingtong-cli scene version list --scene-id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(sceneID != 0, "scene-id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/scene/version/list", map[string]interface{}{"sceneId": sceneID})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.version.list")
		},
	}
	cmd.Flags().IntVar(&sceneID, "scene-id", 0, "Scene ID (required)")
	_ = cmd.MarkFlagRequired("scene-id")
	return cmd
}

// ==================== scene trigger ====================

func newCmdSceneTrigger(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Scene trigger condition operations",
	}
	cmd.AddCommand(newCmdSceneTriggerGet(f))
	cmd.AddCommand(newCmdSceneTriggerSave(f))
	return cmd
}

func newCmdSceneTriggerGet(f *cmdutil.Factory) *cobra.Command {
	var sceneID int
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a scene's trigger condition",
		Long: `Show the trigger condition for a scene. Wraps GET /scene/trigger/condition/get.

EXAMPLES:
    lingtong-cli scene trigger get --scene-id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(sceneID != 0, "scene-id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/scene/trigger/condition/get", map[string]interface{}{"sceneId": sceneID})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.trigger.get")
		},
	}
	cmd.Flags().IntVar(&sceneID, "scene-id", 0, "Scene ID (required)")
	_ = cmd.MarkFlagRequired("scene-id")
	return cmd
}

func newCmdSceneTriggerSave(f *cmdutil.Factory) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save a scene's trigger condition",
		Long: `Save a trigger condition from JSON. Wraps POST /scene/trigger/condition/save.

EXAMPLES:
    lingtong-cli scene trigger save --data '{"sceneId":123,"conditions":[]}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(data != "", "data"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			body, err := cmdutil.ParseDataObject(data)
			if err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/scene/trigger/condition/save", body)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.trigger.save")
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "Trigger condition (JSON object, required)")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

// ==================== scene field-mapping ====================

func newCmdSceneFieldMapping(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "field-mapping",
		Short: "Scene field-mapping operations",
	}
	cmd.AddCommand(newCmdSceneFieldMappingList(f))
	cmd.AddCommand(newCmdSceneFieldMappingExecute(f))
	return cmd
}

func newCmdSceneFieldMappingList(f *cmdutil.Factory) *cobra.Command {
	var sceneID int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List field mapping values",
		Long: `List a scene's field mapping values. Wraps GET /scene/field/value/list.

EXAMPLES:
    lingtong-cli scene field-mapping list --scene-id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(sceneID != 0, "scene-id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/scene/field/value/list", map[string]interface{}{"sceneId": sceneID})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.field-mapping.list")
		},
	}
	cmd.Flags().IntVar(&sceneID, "scene-id", 0, "Scene ID (required)")
	_ = cmd.MarkFlagRequired("scene-id")
	return cmd
}

func newCmdSceneFieldMappingExecute(f *cmdutil.Factory) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Debug-execute field mapping",
		Long: `Run a field-mapping debug execution. Wraps POST /scene/fieldMapping/execute.

EXAMPLES:
    lingtong-cli scene field-mapping execute --data '{"sceneId":123,"record":{}}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(data != "", "data"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			body, err := cmdutil.ParseDataObject(data)
			if err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/scene/fieldMapping/execute", body)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "scene.field-mapping.execute")
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "Field-mapping debug payload (JSON object, required)")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}
