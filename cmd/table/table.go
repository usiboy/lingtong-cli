// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"encoding/json"
	"fmt"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// NewCmdTable creates the table command.
func NewCmdTable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "table",
		Short: "Manage tables",
		Long:  "Table CRUD operations and data management.",
	}

	cmd.AddCommand(newCmdTableList(f))
	cmd.AddCommand(newCmdTableDataQuery(f))
	cmd.AddCommand(newCmdTableDataCreate(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func newCmdTableList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/gw/ai/table/list", nil)
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
	return cmd
}

func newCmdTableDataQuery(f *cmdutil.Factory) *cobra.Command {
	var tableId int
	var filter, sort string
	cmd := &cobra.Command{
		Use:   "data query",
		Short: "Query table data",
		Long: `Query table data with optional filters and sorting.

EXAMPLES:
    lingtong-cli table data query --table-id 123
    lingtong-cli table data query --table-id 123 --filter '{"status":"active"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tableId == 0 {
				return fmt.Errorf("--table-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := fmt.Sprintf("/gw/ai/table/data/query?tableId=%d", tableId)
			if filter != "" {
				path += "&filter=" + filter
			}
			if sort != "" {
				path += "&sort=" + sort
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

	cmd.Flags().IntVar(&tableId, "table-id", 0, "Table ID (required)")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter conditions (JSON)")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort configuration (JSON)")
	_ = cmd.MarkFlagRequired("table-id")
	return cmd
}

func newCmdTableDataCreate(f *cmdutil.Factory) *cobra.Command {
	var tableId int
	var data string
	cmd := &cobra.Command{
		Use:   "data create",
		Short: "Create table record",
		Long: `Create a new record in a table.

EXAMPLES:
    lingtong-cli table data create --table-id 123 --data '{"name":"test","value":100}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tableId == 0 {
				return fmt.Errorf("--table-id is required")
			}
			if data == "" {
				return fmt.Errorf("--data is required")
			}

			var dataMap map[string]interface{}
			if err := json.Unmarshal([]byte(data), &dataMap); err != nil {
				return fmt.Errorf("invalid --data JSON: %w", err)
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			body := map[string]interface{}{
				"tableId": tableId,
				"data":    dataMap,
			}

			resp, err := c.Post("/gw/ai/table/data/create", body)
			if err != nil {
				return err
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var result interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return err
			}
			return w.Write(result)
		},
	}

	cmd.Flags().IntVar(&tableId, "table-id", 0, "Table ID (required)")
	cmd.Flags().StringVar(&data, "data", "", "Record data (JSON, required)")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}
