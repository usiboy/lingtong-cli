// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// ScaffoldApp represents a minimal application template for scaffolding.
type ScaffoldApp struct {
	AppName            string          `json:"appName"`
	AppConnectors      []string        `json:"appConnectors"`
	BasicDatas         []interface{}   `json:"basicDatas"`
	Scenes             []ScaffoldScene `json:"scenes"`
	Workflows          []interface{}   `json:"workflows"`
	WorkflowConnectors []interface{}   `json:"workflowConnectors"`
}

// ScaffoldScene is the validator-compatible scene shape emitted by scaffold.
type ScaffoldScene struct {
	Name            string        `json:"name"`
	Description     string        `json:"description,omitempty"`
	ConnectorSource ConnectorRef  `json:"connectorSource"`
	ConnectorTarget ConnectorRef  `json:"connectorTarget"`
	Trigger         string        `json:"trigger,omitempty"`
	FieldMappings   []interface{} `json:"fieldMappings"`
}

func newScaffoldScene(name, description, source, target, trigger string) ScaffoldScene {
	return ScaffoldScene{
		Name:            name,
		Description:     description,
		ConnectorSource: ConnectorRef{Name: source},
		ConnectorTarget: ConnectorRef{Name: target},
		Trigger:         trigger,
		FieldMappings:   []interface{}{},
	}
}

// kuaimaiKingdeeTemplate returns a pre-configured scaffold for the
// Kuaimai ERP ↔ Kingdee Cloud Star integration case.
func kuaimaiKingdeeTemplate(appName string) ScaffoldApp {
	basicDatas := []interface{}{
		map[string]interface{}{
			"name":         "快麦店铺",
			"dataSource":   "connector",
			"connector":    "kmerp",
			"autoSync":     true,
			"syncInterval": "1h",
		},
		map[string]interface{}{
			"name":         "快麦仓库",
			"dataSource":   "connector",
			"connector":    "kmerp",
			"autoSync":     true,
			"syncInterval": "1h",
		},
		map[string]interface{}{
			"name":       "金蝶物料",
			"dataSource": "connector",
			"connector":  "kingDeeCloudStar",
			"autoSync":   true,
		},
		map[string]interface{}{
			"name":       "金蝶仓位",
			"dataSource": "connector",
			"connector":  "kingDeeCloudStar",
			"autoSync":   true,
		},
		map[string]interface{}{
			"name":       "店铺映射表",
			"dataSource": "manual",
			"autoSync":   false,
		},
	}

	scenes := []ScaffoldScene{
		newScaffoldScene("商品同步", "定时查询金蝶商品列表同步到快麦", "kingDeeCloudStar", "kmerp", "scheduled"),
		newScaffoldScene("销售订单同步", "快麦货位进出记录→金蝶销售出库单", "kmerp", "kingDeeCloudStar", "scheduled"),
		newScaffoldScene("销退入库单同步", "快麦销退上架→金蝶销退入库单", "kmerp", "kingDeeCloudStar", "realtime"),
		newScaffoldScene("库存同步", "定时查询金蝶库存修改快麦实际库存", "kingDeeCloudStar", "kmerp", "scheduled"),
		newScaffoldScene("质量管理发货通知单", "定时抓取金蝶已审核发货通知单→快麦系统手工单", "kingDeeCloudStar", "kmerp", "scheduled"),
		newScaffoldScene("线下销售出库单下推", "金蝶发货通知单下推→销售出库单", "kingDeeCloudStar", "kingDeeCloudStar", "manual"),
		newScaffoldScene("线下销售出库单操作", "下推返回数据替换货位映射后保存审核", "kmerp", "kingDeeCloudStar", "manual"),
	}

	return ScaffoldApp{
		AppName:            appName,
		AppConnectors:      []string{"kmerp", "kingDeeCloudStar"},
		BasicDatas:         basicDatas,
		Scenes:             scenes,
		Workflows:          []interface{}{},
		WorkflowConnectors: []interface{}{},
	}
}

func newCmdAppScaffold(f *cmdutil.Factory) *cobra.Command {
	var name, source, target, output, templateFlag string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Generate a minimal application configuration template",
		Long: `Generate a minimal valid application JSON with a default scene and empty basicDatas.

Use --template to generate a pre-configured application from a known integration pattern.

EXAMPLES:
    lingtong-cli app scaffold --name "my-app" --source "connector-a" --target "connector-b" --output app.json
    lingtong-cli app scaffold --name "my-app" --source "connector-a" --target "connector-b" --dry-run
    lingtong-cli app scaffold --name "kuaimai-kingdee" --template kuaimai-kingdee --output app.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Template mode: --template overrides --source/--target
			if templateFlag != "" {
				if name == "" {
					return fmt.Errorf("--name is required")
				}

				var scaffold ScaffoldApp
				switch templateFlag {
				case "kuaimai-kingdee":
					scaffold = kuaimaiKingdeeTemplate(name)
				default:
					return fmt.Errorf("unknown template %q (supported: kuaimai-kingdee)", templateFlag)
				}

				formatted, err := json.MarshalIndent(scaffold, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}

				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", formatted)
					return nil
				}

				if output == "" {
					return fmt.Errorf("--output is required (or use --dry-run)")
				}

				if err := os.WriteFile(output, formatted, 0644); err != nil {
					return fmt.Errorf("failed to write file %s: %w", output, err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Generated application scaffold to %s\n", output)
				return nil
			}

			// Standard mode (existing behavior)
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if source == "" {
				return fmt.Errorf("--source is required")
			}
			if target == "" {
				return fmt.Errorf("--target is required")
			}

			scaffold := ScaffoldApp{
				AppName:            name,
				AppConnectors:      []string{source, target},
				BasicDatas:         []interface{}{},
				Scenes:             []ScaffoldScene{newScaffoldScene("Default Sync", "", source, target, "manual")},
				Workflows:          []interface{}{},
				WorkflowConnectors: []interface{}{},
			}

			formatted, err := json.MarshalIndent(scaffold, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", formatted)
				return nil
			}

			if output == "" {
				return fmt.Errorf("--output is required (or use --dry-run)")
			}

			if err := os.WriteFile(output, formatted, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", output, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated application scaffold to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Application name (required)")
	cmd.Flags().StringVar(&source, "source", "", "Source connector name (required)")
	cmd.Flags().StringVar(&target, "target", "", "Target connector name (required)")
	cmd.Flags().StringVar(&output, "output", "", "Output file path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print generated JSON to stdout instead of file")
	cmd.Flags().StringVar(&templateFlag, "template", "", "Pre-configured template: kuaimai-kingdee")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
