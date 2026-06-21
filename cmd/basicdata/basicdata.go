// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package basicdata provides hand-written commands for the platform's
// basic-data record operations (/basicdata/record/*) — the highest-frequency
// API surface. These commands favour strong validation, a cohesive CRUD shape,
// and confirmation gating on destructive operations.
package basicdata

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/spf13/cobra"
)

// Thin aliases over the shared cmdutil helpers, kept for readability at the
// many call sites in this package.
func requireHostConfigured(host string) error { return cmdutil.RequireHostConfigured(host) }
func requireFlag(ok bool, name string) error  { return cmdutil.RequireFlag(ok, name) }
func confirmDestructive(yes bool, action string) error {
	return cmdutil.ConfirmDestructive(yes, action)
}
func parseDataObject(s string) (map[string]interface{}, error) {
	return cmdutil.ParseDataObject(s)
}

// NewCmdBasicdata creates the basicdata command group.
func NewCmdBasicdata(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "basicdata",
		Short: "Manage basic-data records",
		Long: `Record CRUD for basic-data tables (/basicdata/record/*).

A "basic-data table" is identified by its --basic-data-id; its column layout is
identified by --schema-id. Field values are passed as a JSON object keyed by
field id (e.g. '{"0":"name","1":"value"}').

These commands complement 'lingtong-cli table data' with record counting, single
-record update, and confirmation-gated deletes.`,
	}

	cmd.AddCommand(newCmdRecord(f))
	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func newCmdRecord(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Basic-data record operations",
	}
	cmd.AddCommand(newCmdRecordList(f))
	cmd.AddCommand(newCmdRecordCount(f))
	cmd.AddCommand(newCmdRecordSave(f))
	cmd.AddCommand(newCmdRecordUpdate(f))
	cmd.AddCommand(newCmdRecordDelete(f))
	cmd.AddCommand(newCmdRecordBatchDelete(f))
	return cmd
}

// writeResponse writes a raw API response through the factory's shared writer.
func writeResponse(cmd *cobra.Command, f *cmdutil.Factory, resp []byte) error {
	return f.WriteResponse(cmd, resp, "basicdata."+cmd.Name())
}

// ==================== record list ====================

func newCmdRecordList(f *cmdutil.Factory) *cobra.Command {
	var basicDataID, viewID, page, pageSize int
	var filter, ids, orderBy string
	var orderAsc bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List/query records (paginated)",
		Long: `Query records of a basic-data table with paging, filtering and sorting.

Wraps POST /basicdata/record/listNew.

EXAMPLES:
    lingtong-cli basicdata record list --basic-data-id 123
    lingtong-cli basicdata record list --basic-data-id 123 --filter '{"1":"系统订单"}' --page 1 --page-size 50
    lingtong-cli basicdata record list --basic-data-id 123 --ids 1001,1002
    lingtong-cli basicdata record list --basic-data-id 123 --order-by created --order-asc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag(basicDataID != 0, "basic-data-id"); err != nil {
				return err
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			body := map[string]interface{}{
				"basicDataId": basicDataID,
				"pageNum":     page,
				"pageSize":    pageSize,
			}
			if filter != "" {
				var fm map[string]interface{}
				if err := json.Unmarshal([]byte(filter), &fm); err != nil {
					return lterrors.NewValidationError(lterrors.SubtypeInvalidArgument,
						"--filter must be a JSON object").WithCause(err)
				}
				body["text"] = fm
			}
			if viewID != 0 {
				body["viewId"] = viewID
			}
			if ids != "" {
				parsed, err := parseIDs(ids)
				if err != nil {
					return err
				}
				body["ids"] = parsed
			}
			if orderBy != "" {
				body["column"] = orderBy
				body["asc"] = orderAsc
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/basicdata/record/listNew", body)
			if err != nil {
				return err
			}
			return writeResponse(cmd, f, resp)
		},
	}

	cmd.Flags().IntVar(&basicDataID, "basic-data-id", 0, "Basic-data (table) ID (required)")
	cmd.Flags().StringVar(&filter, "filter", "", "Field-level exact match (JSON: {\"fieldId\":\"value\"})")
	cmd.Flags().IntVar(&viewID, "view-id", 0, "View ID (loads saved filter/sort)")
	cmd.Flags().StringVar(&ids, "ids", "", "Comma-separated record IDs (e.g. 1001,1002)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Sort field (e.g. created, updated)")
	cmd.Flags().BoolVar(&orderAsc, "order-asc", false, "Sort ascending (default: descending)")
	_ = cmd.MarkFlagRequired("basic-data-id")
	return cmd
}

// ==================== record count ====================

func newCmdRecordCount(f *cmdutil.Factory) *cobra.Command {
	var schemaID, version, basicDataID int
	var filter string

	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count records matching a query",
		Long: `Count records of a table without fetching them.

Wraps GET /basicdata/record/count. Requires --schema-id and --version.

EXAMPLES:
    lingtong-cli basicdata record count --schema-id 2546 --version 1
    lingtong-cli basicdata record count --schema-id 2546 --version 1 --filter '{"1":"系统订单"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag(schemaID != 0, "schema-id"); err != nil {
				return err
			}
			if err := requireFlag(version != 0, "version"); err != nil {
				return err
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			params := map[string]interface{}{
				"schemaId": schemaID,
				"version":  version,
			}
			if basicDataID != 0 {
				params["basicDataId"] = basicDataID
			}
			if filter != "" {
				var fm map[string]interface{}
				if err := json.Unmarshal([]byte(filter), &fm); err != nil {
					return lterrors.NewValidationError(lterrors.SubtypeInvalidArgument,
						"--filter must be a JSON object").WithCause(err)
				}
				params["dataFrom"] = fm
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/basicdata/record/count", params)
			if err != nil {
				return err
			}
			return writeResponse(cmd, f, resp)
		},
	}

	cmd.Flags().IntVar(&schemaID, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Data version number (required)")
	cmd.Flags().IntVar(&basicDataID, "basic-data-id", 0, "Basic-data (table) ID")
	cmd.Flags().StringVar(&filter, "filter", "", "Field-level exact match (JSON: {\"fieldId\":\"value\"})")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("version")
	return cmd
}

// ==================== record save ====================

func newCmdRecordSave(f *cmdutil.Factory) *cobra.Command {
	var basicDataID, schemaID, version int
	var data string

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Create a new record",
		Long: `Create a new record in a basic-data table.

Wraps POST /basicdata/record/save. --data is a JSON object keyed by field id.
--version is optional; when omitted the version check is skipped (recommended).

EXAMPLES:
    lingtong-cli basicdata record save --basic-data-id 123 --schema-id 1 --data '{"0":"v1","1":"v2"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag(basicDataID != 0, "basic-data-id"); err != nil {
				return err
			}
			if err := requireFlag(schemaID != 0, "schema-id"); err != nil {
				return err
			}
			if err := requireFlag(data != "", "data"); err != nil {
				return err
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			dataMap, err := parseDataObject(data)
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"basicDataId": basicDataID,
				"schemaId":    schemaID,
				"data":        dataMap,
			}
			if version > 0 {
				body["version"] = version
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/basicdata/record/save", body)
			if err != nil {
				return err
			}
			return writeResponse(cmd, f, resp)
		},
	}

	cmd.Flags().IntVar(&basicDataID, "basic-data-id", 0, "Basic-data (table) ID (required)")
	cmd.Flags().IntVar(&schemaID, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Data version number (optional)")
	cmd.Flags().StringVar(&data, "data", "", "Record field values (JSON object, required)")
	_ = cmd.MarkFlagRequired("basic-data-id")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

// ==================== record update ====================

func newCmdRecordUpdate(f *cmdutil.Factory) *cobra.Command {
	var basicDataID, schemaID, id, version int
	var data string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a single record",
		Long: `Update one existing record by ID.

Wraps POST /basicdata/record/update. --data is a JSON object keyed by field id
containing only the fields to change.

EXAMPLES:
    lingtong-cli basicdata record update --basic-data-id 123 --schema-id 1 --id 33832272 --data '{"1":"newValue"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag(basicDataID != 0, "basic-data-id"); err != nil {
				return err
			}
			if err := requireFlag(schemaID != 0, "schema-id"); err != nil {
				return err
			}
			if err := requireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := requireFlag(data != "", "data"); err != nil {
				return err
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			dataMap, err := parseDataObject(data)
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"basicDataId": basicDataID,
				"schemaId":    schemaID,
				"id":          id,
				"data":        dataMap,
			}
			if version > 0 {
				body["version"] = version
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/basicdata/record/update", body)
			if err != nil {
				return err
			}
			return writeResponse(cmd, f, resp)
		},
	}

	cmd.Flags().IntVar(&basicDataID, "basic-data-id", 0, "Basic-data (table) ID (required)")
	cmd.Flags().IntVar(&schemaID, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&id, "id", 0, "Record ID to update (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Data version number (optional)")
	cmd.Flags().StringVar(&data, "data", "", "Fields to update (JSON object, required)")
	_ = cmd.MarkFlagRequired("basic-data-id")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

// ==================== record delete ====================

func newCmdRecordDelete(f *cmdutil.Factory) *cobra.Command {
	var schemaID, id int
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a single record (requires --yes)",
		Long: `Delete one record by ID.

Wraps POST /basicdata/record/batchDelete with a single ID. This is destructive
and requires --yes to proceed.

EXAMPLES:
    lingtong-cli basicdata record delete --schema-id 2546 --id 33832272 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag(schemaID != 0, "schema-id"); err != nil {
				return err
			}
			if err := requireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := confirmDestructive(yes, "delete record"); err != nil {
				return err
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			body := map[string]interface{}{
				"schemaId": schemaID,
				"ids":      []int{id},
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/basicdata/record/batchDelete", body)
			if err != nil {
				return err
			}
			return writeResponse(cmd, f, resp)
		},
	}

	cmd.Flags().IntVar(&schemaID, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&id, "id", 0, "Record ID to delete (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm the destructive operation")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// ==================== record batch-delete ====================

func newCmdRecordBatchDelete(f *cmdutil.Factory) *cobra.Command {
	var schemaID int
	var ids string
	var yes bool

	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Delete multiple records (requires --yes)",
		Long: `Delete several records by ID in one call.

Wraps POST /basicdata/record/batchDelete. This is destructive and requires --yes.

EXAMPLES:
    lingtong-cli basicdata record batch-delete --schema-id 2546 --ids 1001,1002,1003 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag(schemaID != 0, "schema-id"); err != nil {
				return err
			}
			if err := requireFlag(ids != "", "ids"); err != nil {
				return err
			}
			parsed, err := parseIDs(ids)
			if err != nil {
				return err
			}
			if err := confirmDestructive(yes, fmt.Sprintf("delete %d records", len(parsed))); err != nil {
				return err
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			body := map[string]interface{}{
				"schemaId": schemaID,
				"ids":      parsed,
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/basicdata/record/batchDelete", body)
			if err != nil {
				return err
			}
			return writeResponse(cmd, f, resp)
		},
	}

	cmd.Flags().IntVar(&schemaID, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().StringVar(&ids, "ids", "", "Comma-separated record IDs (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm the destructive operation")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}

// ==================== helpers ====================

// parseIDs converts a comma-separated id list into []int, failing with a
// validation error on any malformed entry.
func parseIDs(s string) ([]int, error) {
	var ids []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, lterrors.NewValidationError(lterrors.SubtypeInvalidArgument,
				"invalid id %q in --ids", part).WithCause(err)
		}
		ids = append(ids, n)
	}
	if len(ids) == 0 {
		return nil, lterrors.NewValidationError(lterrors.SubtypeInvalidArgument,
			"--ids contained no valid IDs")
	}
	return ids, nil
}
