// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package model

import (
	"encoding/json"
	"fmt"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// NewCmdModel creates the model command.
func NewCmdModel(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage model metadata",
		Long:  "Query interface models, domain models, and dynamic model schemas.",
	}

	cmd.AddCommand(newCmdModelInterfaceList(f))
	cmd.AddCommand(newCmdModelDomainGet(f))
	cmd.AddCommand(newCmdModelDynamicView(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func newCmdModelInterfaceList(f *cmdutil.Factory) *cobra.Command {
	var connector, modelType, filterModelType, authAccountId string
	cmd := &cobra.Command{
		Use:   "interface list",
		Short: "List interface models for a connector",
		Long: `Query interface model list for a connector.

EXAMPLES:
    lingtong-cli model interface list --connector kmerp --filter-model-type all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connector == "" {
				return fmt.Errorf("--connector is required")
			}
			if filterModelType == "" {
				return fmt.Errorf("--filter-model-type is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := fmt.Sprintf("/gw/ai/model/interface/list?connector=%s&filterModelType=%s", connector, filterModelType)
			if modelType != "" {
				path += "&modelType=" + modelType
			}
			if authAccountId != "" {
				path += "&authAccountId=" + authAccountId
			}

			resp, err := c.Get(path, nil)
			if err != nil {
				return err
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().StringVar(&connector, "connector", "", "Connector name (required)")
	cmd.Flags().StringVar(&modelType, "model-type", "", "Model type (optional)")
	cmd.Flags().StringVar(&filterModelType, "filter-model-type", "", "Filter model type (required)")
	cmd.Flags().StringVar(&authAccountId, "auth-account-id", "", "Auth account ID")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("filter-model-type")
	return cmd
}

func newCmdModelDomainGet(f *cmdutil.Factory) *cobra.Command {
	var connector, business, authAccountId string
	cmd := &cobra.Command{
		Use:   "domain get",
		Short: "Get domain model",
		Long: `Get domain model by connector and business.

EXAMPLES:
    lingtong-cli model domain get --connector kmerp --business order`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connector == "" || business == "" {
				return fmt.Errorf("--connector and --business are required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := fmt.Sprintf("/gw/ai/model/domain/get?connector=%s&business=%s", connector, business)
			if authAccountId != "" {
				path += "&authAccountId=" + authAccountId
			}

			resp, err := c.Get(path, nil)
			if err != nil {
				return err
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().StringVar(&connector, "connector", "", "Connector name (required)")
	cmd.Flags().StringVar(&business, "business", "", "Business name (required)")
	cmd.Flags().StringVar(&authAccountId, "auth-account-id", "", "Auth account ID")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("business")
	return cmd
}

func newCmdModelDynamicView(f *cmdutil.Factory) *cobra.Command {
	var connector, modelName, businessObjectName string
	var authAccountId int
	cmd := &cobra.Command{
		Use:   "dynamic view",
		Short: "Query dynamic data structure view",
		Long: `Query dynamic data structure view for a connector.

EXAMPLES:
    lingtong-cli model dynamic view --connector kmerp --auth-account-id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connector == "" || authAccountId == 0 {
				return fmt.Errorf("--connector and --auth-account-id are required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := fmt.Sprintf("/gw/ai/model/dynamic/view?connector=%s&authAccountId=%d", connector, authAccountId)
			if modelName != "" {
				path += "&modelName=" + modelName
			}
			if businessObjectName != "" {
				path += "&businessObjectName=" + businessObjectName
			}

			resp, err := c.Get(path, nil)
			if err != nil {
				return err
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().StringVar(&connector, "connector", "", "Connector name (required)")
	cmd.Flags().IntVar(&authAccountId, "auth-account-id", 0, "Auth account ID (required)")
	cmd.Flags().StringVar(&modelName, "model-name", "", "Model name (optional)")
	cmd.Flags().StringVar(&businessObjectName, "business-object-name", "", "Business object name (optional)")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("auth-account-id")
	return cmd
}
