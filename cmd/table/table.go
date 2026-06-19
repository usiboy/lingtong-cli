// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

const missingHostConfigMessage = "no host configured. Run `lingtong-cli config init --host <url>` first"

func requireHostConfigured(host string) error {
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf(missingHostConfigMessage)
	}
	return nil
}

// NewCmdTable creates the table command.
func NewCmdTable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "table",
		Short: "Manage tables",
		Long:  "Table CRUD operations and data management.",
	}

	cmd.AddCommand(newCmdTableList(f))
	cmd.AddCommand(newCmdTableCreate(f))
	cmd.AddCommand(newCmdTableData(f))
	cmd.AddCommand(newCmdTableSchema(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func newCmdTableList(f *cmdutil.Factory) *cobra.Command {
	var appID int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tables",
		Long: `List tables with optional application filter.

EXAMPLES:
    lingtong-cli table list
    lingtong-cli table list --app-id 123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			params := map[string]interface{}{}
			if appID != 0 {
				params["appId"] = appID
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/basicdata/list", params)
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

	cmd.Flags().IntVar(&appID, "app-id", 0, "Application ID filter")
	return cmd
}

// ==================== table create ====================

func newCmdTableCreate(f *cmdutil.Factory) *cobra.Command {
	var appID int
	var name string
	var source int
	var tableType int
	var openHighMode int
	var openConnector int
	var columnsSchema string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new table",
		Long: `Create a new table with optional schema definition.

This command performs two steps:
1. Create the basic data (table) via POST /basicdata/save
2. If --columns-schema is provided, create the schema via POST /basicdata/schema/save

The table creation requires app-id, name, source, and type.
The schema creation is optional and requires --columns-schema JSON.

EXAMPLES:
    # Create a manual table
    lingtong-cli table create --app-id 165 --name "测试表格" --source 1 --type 1

    # Create with high performance mode
    lingtong-cli table create --app-id 165 --name "测试表格" --source 1 --type 1 --open-high-mode 1

    # Create with schema
    lingtong-cli table create --app-id 165 --name "测试表格" --source 1 --type 1 --columns-schema '[{"title":"名称","key":"name","type":"text"}]'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appID == 0 {
				return fmt.Errorf("--app-id is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)

			// Step 1: Create basic data (table)
			basicDataBody := map[string]interface{}{
				"appId":         appID,
				"name":          name,
				"source":        source,
				"type":          tableType,
				"openHighMode":  openHighMode,
				"openConnector": openConnector,
			}

			resp, err := c.Post("/basicdata/save", basicDataBody)
			if err != nil {
				return fmt.Errorf("failed to create table: %w", err)
			}

			var basicDataResult map[string]interface{}
			if err := json.Unmarshal(resp, &basicDataResult); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if creation was successful
			if success, ok := basicDataResult["success"].(bool); ok && !success {
				msg := ""
				if m, ok := basicDataResult["msg"].(string); ok {
					msg = m
				}
				return fmt.Errorf("table creation failed: %s", msg)
			}

			// Extract the table ID from response
			var tableID float64
			if result, ok := basicDataResult["result"].(map[string]interface{}); ok {
				if id, ok := result["id"].(float64); ok {
					tableID = id
				}
			}

			if tableID == 0 {
				return fmt.Errorf("failed to get table ID from response")
			}

			// Step 2: Create schema if columns-schema is provided
			if columnsSchema != "" {
				var columns []interface{}
				if err := json.Unmarshal([]byte(columnsSchema), &columns); err != nil {
					return fmt.Errorf("invalid --columns-schema JSON: %w", err)
				}

				// Apply defaults for each column to ensure visibility and proper display
				for i, col := range columns {
					if colMap, ok := col.(map[string]interface{}); ok {
						// Auto-assign ID if not provided
						if _, hasID := colMap["id"]; !hasID {
							colMap["id"] = i + 1
						}
						// First column is primary key by default
						if i == 0 {
							if _, hasPK := colMap["primaryKey"]; !hasPK {
								colMap["primaryKey"] = true
							}
						} else {
							if _, hasPK := colMap["primaryKey"]; !hasPK {
								colMap["primaryKey"] = false
							}
						}
						// Default to visible (critical: without this, columns are hidden)
						if _, hasShow := colMap["show"]; !hasShow {
							colMap["show"] = true
						}
						// Default width
						if _, hasWidth := colMap["width"]; !hasWidth {
							colMap["width"] = 200
						}
						// Default alignment
						if _, hasAlign := colMap["align"]; !hasAlign {
							colMap["align"] = "left"
						}
						// Default fixed position
						if _, hasFixed := colMap["fixed"]; !hasFixed {
							colMap["fixed"] = "left"
						}
						// Default description
						if _, hasDesc := colMap["description"]; !hasDesc {
							colMap["description"] = ""
						}
						// Default number format
						if _, hasNumFmt := colMap["numberFormat"]; !hasNumFmt {
							colMap["numberFormat"] = "1000"
						}
						// Default date format
						if _, hasDateFmt := colMap["dateFormat"]; !hasDateFmt {
							colMap["dateFormat"] = "YYYY-MM-DD"
						}
						// Default function fields
						if _, hasFunc := colMap["function"]; !hasFunc {
							colMap["function"] = ""
						}
						if _, hasFuncField := colMap["functionFieldType"]; !hasFuncField {
							colMap["functionFieldType"] = "normal"
						}
						// Default keyDisabled
						if _, hasKeyDis := colMap["keyDisabled"]; !hasKeyDis {
							colMap["keyDisabled"] = false
						}
						// Default typeDisabled
						if _, hasTypeDis := colMap["typeDisabled"]; !hasTypeDis {
							colMap["typeDisabled"] = true
						}
						// Default dataDisabled
						if _, hasDataDis := colMap["dataDisabled"]; !hasDataDis {
							colMap["dataDisabled"] = false
						}
					}
				}

				schemaBody := map[string]interface{}{
					"basicDataId":   int(tableID),
					"columnsSchema": columns,
				}

				schemaResp, err := c.Post("/basicdata/schema/save", schemaBody)
				if err != nil {
					return fmt.Errorf("failed to create schema: %w", err)
				}

				var schemaResult map[string]interface{}
				if err := json.Unmarshal(schemaResp, &schemaResult); err != nil {
					return fmt.Errorf("failed to parse schema response: %w", err)
				}

				if success, ok := schemaResult["success"].(bool); ok && !success {
					msg := ""
					if m, ok := schemaResult["msg"].(string); ok {
						msg = m
					}
					return fmt.Errorf("schema creation failed: %s", msg)
				}

				// Merge results
				basicDataResult["schemaResult"] = schemaResult["result"]
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			return w.Write(basicDataResult)
		},
	}

	cmd.Flags().IntVar(&appID, "app-id", 0, "Application ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Table name (required)")
	cmd.Flags().IntVar(&source, "source", 1, "Source: 1-manual, 2-connector (default 1)")
	cmd.Flags().IntVar(&tableType, "type", 1, "Type: 1-regular, 2-structure-mapping (default 1)")
	cmd.Flags().IntVar(&openHighMode, "open-high-mode", 0, "High performance mode: 0-off, 1-on (default 0)")
	cmd.Flags().IntVar(&openConnector, "open-connector", 0, "Connector sync: 0-off, 1-on (default 0)")
	cmd.Flags().StringVar(&columnsSchema, "columns-schema", "", "Schema columns JSON array (optional)")
	_ = cmd.MarkFlagRequired("app-id")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newCmdTableData(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Manage table data",
		Long:  "Table data CRUD operations.",
	}

	cmd.AddCommand(newCmdTableDataQuery(f))
	cmd.AddCommand(newCmdTableDataCreate(f))
	cmd.AddCommand(newCmdTableDataBatchUpdate(f))
	cmd.AddCommand(newCmdTableDataDelete(f))
	cmd.AddCommand(newCmdTableDataBatchDelete(f))
	return cmd
}

func newCmdTableDataQuery(f *cmdutil.Factory) *cobra.Command {
	var basicDataId int
	var text, filter, idsStr, primaryKeyStr, orderBy string
	var viewId int
	var startUpdateTime, endUpdateTime int64
	var page, pageSize int
	var orderAsc bool

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query table data",
		Long: `Query table data with full search, filter, pagination and sorting support.

Uses POST /basicdata/record/listNew endpoint with full backend parameter support.

FILTER MODES:
  --text     Full-text search (MySQL: LIKE match on data field)
  --filter   Field-level exact match (JSON: {"fieldKey":"value"})
             Maps to backend 'text' parameter for precise field filtering.
             In PG mode: data->>'fieldKey' = 'value'
             In MySQL mode: JSON_EXTRACT(data, '$."fieldKey"') = 'value'
  --view-id  Load saved filter/sort config from a table view

EXAMPLES:
    # Basic query (returns first page, 20 records)
    lingtong-cli table data query --basic-data-id 123

    # Field-level exact match (filter by specific field values)
    lingtong-cli table data query --basic-data-id 123 --filter '{"1":"系统订单"}'
    lingtong-cli table data query --basic-data-id 123 --filter '{"1":"系统订单","17":"-2"}'

    # Full-text search (LIKE match in MySQL mode)
    lingtong-cli table data query --basic-data-id 123 --text '系统订单'

    # Load saved view filter (viewId loads pre-configured filters)
    lingtong-cli table data query --basic-data-id 123 --view-id 456

    # Filter by record IDs
    lingtong-cli table data query --basic-data-id 123 --ids 1001,1002,1003

    # Filter by update time range (millisecond timestamps)
    lingtong-cli table data query --basic-data-id 123 --start-update-time 1700000000000 --end-update-time 1700100000000

    # Filter by primary key value
    lingtong-cli table data query --basic-data-id 123 --primary-key '["pk_value"]'

    # Paginate and sort
    lingtong-cli table data query --basic-data-id 123 --page 2 --page-size 50 --order-by created --order-asc

    # Combine filter with pagination and sorting
    lingtong-cli table data query --basic-data-id 123 --filter '{"1":"系统订单"}' --page 1 --page-size 50 --order-by created --order-asc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if basicDataId == 0 {
				return fmt.Errorf("--basic-data-id is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			body := map[string]interface{}{
				"basicDataId": basicDataId,
				"pageNum":     page,
				"pageSize":    pageSize,
			}

			// --filter takes precedence over --text for field-level exact match
			if filter != "" {
				body["text"] = filter
			} else if text != "" {
				body["text"] = text
			}
			if viewId != 0 {
				body["viewId"] = viewId
			}
			if idsStr != "" {
				var ids []int64
				for _, s := range strings.Split(idsStr, ",") {
					s = strings.TrimSpace(s)
					if s == "" {
						continue
					}
					id, err := strconv.ParseInt(s, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid id %q: %w", s, err)
					}
					ids = append(ids, id)
				}
				if len(ids) > 0 {
					body["ids"] = ids
				}
			}
			if startUpdateTime > 0 {
				body["startUpdateTime"] = startUpdateTime
			}
			if endUpdateTime > 0 {
				body["endUpdateTime"] = endUpdateTime
			}
			if primaryKeyStr != "" {
				var pk []interface{}
				if err := json.Unmarshal([]byte(primaryKeyStr), &pk); err != nil {
					return fmt.Errorf("--primary-key must be a JSON array: %w", err)
				}
				body["primaryKey"] = pk
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

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().IntVar(&basicDataId, "basic-data-id", 0, "Basic data ID (table ID, required)")
	cmd.Flags().StringVar(&text, "text", "", "Full-text search (MySQL: LIKE match)")
	cmd.Flags().StringVar(&filter, "filter", "", "Field-level exact match (JSON: {\"fieldKey\":\"value\"})")
	cmd.Flags().IntVar(&viewId, "view-id", 0, "View ID (loads saved filter/sort config)")
	cmd.Flags().StringVar(&idsStr, "ids", "", "Comma-separated record IDs (e.g. 1001,1002)")
	cmd.Flags().Int64Var(&startUpdateTime, "start-update-time", 0, "Update time start (ms timestamp)")
	cmd.Flags().Int64Var(&endUpdateTime, "end-update-time", 0, "Update time end (ms timestamp)")
	cmd.Flags().StringVar(&primaryKeyStr, "primary-key", "", "Primary key filter (JSON array)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number (default 1)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size (default 20)")
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Sort field (e.g. created, updated)")
	cmd.Flags().BoolVar(&orderAsc, "order-asc", false, "Sort ascending (default: descending)")
	_ = cmd.MarkFlagRequired("basic-data-id")
	return cmd
}

func newCmdTableDataCreate(f *cmdutil.Factory) *cobra.Command {
	var basicDataId int
	var schemaId int
	var version int
	var data string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create table record",
		Long: `Create a new record in a table.

The --data flag accepts a JSON map of field ID (as string key) to value.
The --schema-id can be found from the 'table data query' response (schemaId field).
If --version is not specified, the version check is skipped (recommended).

EXAMPLES:
    lingtong-cli table data create --basic-data-id 123 --schema-id 1 --data '{"0":"val1","1":"val2"}'
    lingtong-cli table data create --basic-data-id 123 --schema-id 1 --data '{"0":"val1"}' --version 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if basicDataId == 0 {
				return fmt.Errorf("--basic-data-id is required")
			}
			if schemaId == 0 {
				return fmt.Errorf("--schema-id is required")
			}
			if data == "" {
				return fmt.Errorf("--data is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			var dataMap map[string]interface{}
			if err := json.Unmarshal([]byte(data), &dataMap); err != nil {
				return fmt.Errorf("invalid --data JSON: %w", err)
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			body := map[string]interface{}{
				"basicDataId": basicDataId,
				"schemaId":    schemaId,
				"data":        dataMap,
			}
			// Only include version when explicitly set (non-zero)
			if version > 0 {
				body["version"] = version
			}

			resp, err := c.Post("/basicdata/record/save", body)
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

	cmd.Flags().IntVar(&basicDataId, "basic-data-id", 0, "Basic data ID (table ID, required)")
	cmd.Flags().IntVar(&schemaId, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Schema version (optional, skip check if not set)")
	cmd.Flags().StringVar(&data, "data", "", "Record data as JSON map, e.g. '{\"0\":\"value0\",\"1\":\"value1\"}' (required)")
	_ = cmd.MarkFlagRequired("basic-data-id")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

// ==================== table data batch-update ====================

func newCmdTableDataBatchUpdate(f *cmdutil.Factory) *cobra.Command {
	var basicDataId int
	var schemaId int
	var version int
	var records string
	cmd := &cobra.Command{
		Use:   "batch-update",
		Short: "Batch update table records",
		Long: `Batch update multiple records in a table.

The --records flag accepts a JSON array of update objects. Each object must contain:
- id: record ID (required)
- data: field data as JSON map (required)

Optional fields per record:
- basicDataId: table ID (auto-filled from --basic-data-id if not set)
- schemaId: schema ID (auto-filled from --schema-id if not set)
- version: schema version (auto-filled from --version if not set)

EXAMPLES:
    # Batch update two records
    lingtong-cli table data batch-update --basic-data-id 1633 --schema-id 2546 \
      --records '[{"id":33832272,"data":{"1":"value1","2":10}},{"id":33832273,"data":{"1":"value2","2":20}}]'

    # Batch update with version
    lingtong-cli table data batch-update --basic-data-id 1633 --schema-id 2546 --version 1 \
      --records '[{"id":33832272,"data":{"1":"new"}}]'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if basicDataId == 0 {
				return fmt.Errorf("--basic-data-id is required")
			}
			if schemaId == 0 {
				return fmt.Errorf("--schema-id is required")
			}
			if records == "" {
				return fmt.Errorf("--records is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			var recordList []map[string]interface{}
			if err := json.Unmarshal([]byte(records), &recordList); err != nil {
				return fmt.Errorf("invalid --records JSON: %w", err)
			}

			if len(recordList) == 0 {
				return fmt.Errorf("--records must contain at least one record")
			}

			// Validate and fill defaults for each record
			for i, rec := range recordList {
				// Validate id
				if _, hasID := rec["id"]; !hasID {
					return fmt.Errorf("record %d: missing required field 'id'", i)
				}
				// Validate data
				if _, hasData := rec["data"]; !hasData {
					return fmt.Errorf("record %d: missing required field 'data'", i)
				}
				// Auto-fill basicDataId
				if _, hasBDI := rec["basicDataId"]; !hasBDI {
					rec["basicDataId"] = basicDataId
				}
				// Auto-fill schemaId
				if _, hasSI := rec["schemaId"]; !hasSI {
					rec["schemaId"] = schemaId
				}
				// Auto-fill version if provided
				if version > 0 {
					if _, hasVer := rec["version"]; !hasVer {
						rec["version"] = version
					}
				}
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/basicdata/record/batchUpdate", recordList)
			if err != nil {
				return fmt.Errorf("failed to batch update: %w", err)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var result interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			return w.Write(result)
		},
	}

	cmd.Flags().IntVar(&basicDataId, "basic-data-id", 0, "Table ID (required)")
	cmd.Flags().IntVar(&schemaId, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Schema version (optional, applied to all records)")
	cmd.Flags().StringVar(&records, "records", "", "JSON array of update records (required)")
	_ = cmd.MarkFlagRequired("basic-data-id")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("records")
	return cmd
}

// ==================== table data delete ====================

func newCmdTableDataDelete(f *cmdutil.Factory) *cobra.Command {
	var id int64
	var schemaId int
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a table record",
		Long: `Delete a single record from a table.

Internally wraps the batchDelete API with a single ID.

EXAMPLES:
    # Delete a record by ID
    lingtong-cli table data delete --schema-id 2546 --id 33832272`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if schemaId == 0 {
				return fmt.Errorf("--schema-id is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			body := map[string]interface{}{
				"schemaId": schemaId,
				"ids":      []int64{id},
			}

			resp, err := c.Post("/basicdata/record/batchDelete", body)
			if err != nil {
				return fmt.Errorf("failed to delete record: %w", err)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var result interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			return w.Write(result)
		},
	}

	cmd.Flags().Int64Var(&id, "id", 0, "Record ID to delete (required)")
	cmd.Flags().IntVar(&schemaId, "schema-id", 0, "Schema ID (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("schema-id")
	return cmd
}

// ==================== table data batch-delete ====================

func newCmdTableDataBatchDelete(f *cmdutil.Factory) *cobra.Command {
	var schemaId int
	var ids string
	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Batch delete table records",
		Long: `Batch delete multiple records from a table.

EXAMPLES:
    # Batch delete records by IDs
    lingtong-cli table data batch-delete --schema-id 2546 --ids "33832272,33832273,33832274"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if schemaId == 0 {
				return fmt.Errorf("--schema-id is required")
			}
			if ids == "" {
				return fmt.Errorf("--ids is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			// Parse comma-separated IDs
			idStrs := strings.Split(ids, ",")
			var idList []int64
			for _, s := range idStrs {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				var id int64
				if _, err := fmt.Sscanf(s, "%d", &id); err != nil {
					return fmt.Errorf("invalid ID format: %s", s)
				}
				idList = append(idList, id)
			}

			if len(idList) == 0 {
				return fmt.Errorf("no valid IDs provided")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			body := map[string]interface{}{
				"schemaId": schemaId,
				"ids":      idList,
			}

			resp, err := c.Post("/basicdata/record/batchDelete", body)
			if err != nil {
				return fmt.Errorf("failed to batch delete: %w", err)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := output.NewWriter(f.IOStreams, format)
			var result interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			return w.Write(result)
		},
	}

	cmd.Flags().IntVar(&schemaId, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().StringVar(&ids, "ids", "", "Comma-separated record IDs to delete (required)")
	_ = cmd.MarkFlagRequired("schema-id")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}

// ==================== table schema ====================

func newCmdTableSchema(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage table schema",
		Long:  "Table schema query and update operations.",
	}

	cmd.AddCommand(newCmdTableSchemaQuery(f))
	cmd.AddCommand(newCmdTableSchemaUpdate(f))
	return cmd
}

func newCmdTableSchemaQuery(f *cmdutil.Factory) *cobra.Command {
	var basicDataId int
	var businessType int
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query table schema",
		Long: `Query table schema to understand the field structure.

The schema contains column definitions (id, title, type, key, etc.) that
describe how table data is structured. This is essential for AI to understand
the meaning of field IDs in table data records.

EXAMPLES:
    lingtong-cli table schema query --basic-data-id 123
    lingtong-cli table schema query --basic-data-id 123 --business-type 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if basicDataId == 0 {
				return fmt.Errorf("--basic-data-id is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			// Default to BASIC_DATA (2) if not specified
			if businessType == 0 {
				businessType = 2
			}

			params := map[string]interface{}{
				"businessId":   basicDataId,
				"businessType": businessType,
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/basicdata/schema/get", params)
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

	cmd.Flags().IntVar(&basicDataId, "basic-data-id", 0, "Basic data ID (table ID, required)")
	cmd.Flags().IntVar(&businessType, "business-type", 2, "Business type: 1-scene, 2-basic-data (default 2)")
	_ = cmd.MarkFlagRequired("basic-data-id")
	return cmd
}

func newCmdTableSchemaUpdate(f *cmdutil.Factory) *cobra.Command {
	var schemaId int
	var basicDataId int
	var schema string
	var columnsSchema string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update table schema",
		Long: `Update table schema configuration.

Two modes of operation:
1. Full update: Use --schema with complete BasicDataFormSchemaDto JSON
2. Simple update: Use --columns-schema with just the columns array

If --basic-data-id is provided and version is not in the JSON, the current version
will be auto-fetched from the server to avoid version conflicts.

EXAMPLES:
    # Simple update - just update columns (auto-fetch version)
    lingtong-cli table schema update --schema-id 2189 --basic-data-id 1553 --columns-schema '[{"id":1,"title":"Name","type":"text","key":"name","primaryKey":true,"show":true}]'

    # Full update with complete schema JSON
    lingtong-cli table schema update --schema-id 2189 --schema '{"columnsSchema":[...],"version":1}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if schemaId == 0 {
				return fmt.Errorf("--schema-id is required")
			}
			if schema == "" && columnsSchema == "" {
				return fmt.Errorf("either --schema or --columns-schema is required")
			}
			if err := requireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			var schemaMap map[string]interface{}

			if schema != "" {
				// Full schema mode
				if err := json.Unmarshal([]byte(schema), &schemaMap); err != nil {
					return fmt.Errorf("invalid --schema JSON: %w", err)
				}
			} else {
				// Simple columns mode - parse columnsSchema
				var columns []interface{}
				if err := json.Unmarshal([]byte(columnsSchema), &columns); err != nil {
					return fmt.Errorf("invalid --columns-schema JSON: %w", err)
				}
				schemaMap = map[string]interface{}{
					"columnsSchema": columns,
				}
			}

			// Ensure id is set
			schemaMap["id"] = schemaId

			// Auto-fetch version if not provided and basicDataId is given
			if _, hasVersion := schemaMap["version"]; !hasVersion {
				if basicDataId == 0 {
					return fmt.Errorf("--basic-data-id is required when version is not provided")
				}
				// Query current schema to get version
				params := map[string]interface{}{
					"businessId":   basicDataId,
					"businessType": 2,
				}
				schemaResp, err := c.Get("/basicdata/schema/get", params)
				if err != nil {
					return fmt.Errorf("failed to fetch current schema: %w", err)
				}
				var schemaResult struct {
					Success bool `json:"success"`
					Result  struct {
						Version int `json:"version"`
					} `json:"result"`
				}
				if err := json.Unmarshal(schemaResp, &schemaResult); err != nil {
					return fmt.Errorf("failed to parse schema response: %w", err)
				}
				if !schemaResult.Success {
					return fmt.Errorf("failed to fetch current schema version")
				}
				schemaMap["version"] = schemaResult.Result.Version
			}

			resp, err := c.Post("/basicdata/schema/update", schemaMap)
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

	cmd.Flags().IntVar(&schemaId, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&basicDataId, "basic-data-id", 0, "Table ID (required for auto-version)")
	cmd.Flags().StringVar(&schema, "schema", "", "Complete schema JSON (optional)")
	cmd.Flags().StringVar(&columnsSchema, "columns-schema", "", "Columns schema JSON array (optional, simpler mode)")
	_ = cmd.MarkFlagRequired("schema-id")
	return cmd
}
