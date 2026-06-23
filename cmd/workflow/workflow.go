// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// fetchWorkflowMeta retrieves the persisted metadata (appId/env/name/ts) for an
// existing workflow. The /workflow/update endpoint requires these fields plus
// the DSL as a `content` string; without them the node graph is silently
// dropped, which is the root cause of the buildFlowSource NPE on API-created
// workflows.
func fetchWorkflowMeta(c *client.Client, workflowId int) (map[string]interface{}, error) {
	resp, err := c.Get(fmt.Sprintf("/gw/workflow/get?workflowId=%d", workflowId), nil)
	if err != nil {
		return nil, err
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(resp, &parsed); err != nil {
		return nil, err
	}
	result, _ := parsed["result"].(map[string]interface{})
	if result == nil {
		return nil, fmt.Errorf("workflow %d not found", workflowId)
	}
	return result, nil
}

// tsToString normalises a workflow `ts` (version token) to the string form the
// update endpoint expects, regardless of how JSON decoded the number.
func tsToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	}
	return ""
}

// metaInt extracts an integer field (e.g. appId) from decoded JSON.
func metaInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		n, _ := t.Int64()
		return int(n)
	}
	return 0
}

// NewCmdWorkflow creates the workflow command.
func NewCmdWorkflow(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflows",
		Long:  "Create, publish, execute, and monitor workflows.",
	}

	// Existing commands
	cmd.AddCommand(newCmdWorkflowList(f))
	cmd.AddCommand(newCmdWorkflowExecute(f))
	cmd.AddCommand(newCmdWorkflowInfo(f))
	cmd.AddCommand(newCmdWorkflowLogs(f))

	// Phase 1: Core functionality
	cmd.AddCommand(newCmdWorkflowPublish(f))
	cmd.AddCommand(newCmdWorkflowVersions(f))
	cmd.AddCommand(newCmdWorkflowAPITest(f))

	// Phase 2: Full CRUD
	cmd.AddCommand(newCmdWorkflowCreate(f))
	cmd.AddCommand(newCmdWorkflowUpdate(f))
	cmd.AddCommand(newCmdWorkflowDelete(f))
	cmd.AddCommand(newCmdWorkflowAPIEnable(f))
	cmd.AddCommand(newCmdWorkflowAPIDisable(f))

	// Phase 3: Advanced lifecycle management
	cmd.AddCommand(newCmdWorkflowTemplate(f))
	cmd.AddCommand(newCmdWorkflowValidate(f))
	cmd.AddCommand(newCmdWorkflowVersion(f))
	cmd.AddCommand(newCmdWorkflowTest(f))
	cmd.AddCommand(newCmdWorkflowDoc(f))
	cmd.AddCommand(newCmdWorkflowDependency(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func newCmdWorkflowList(f *cmdutil.Factory) *cobra.Command {
	var appId int
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows in an application",
		Long: `List all workflows in an application.

EXAMPLES:
    lingtong-cli workflow list --app-id 165`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appId == 0 {
				return fmt.Errorf("--app-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			params := map[string]interface{}{
				"appId":    appId,
				"pageNum":  page,
				"pageSize": pageSize,
			}
			path := "/gw/workflow/list"

			resp, err := c.Get(path, params)
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

	cmd.Flags().IntVar(&appId, "app-id", 0, "Application ID (required)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "size", 10, "Page size")
	_ = cmd.MarkFlagRequired("app-id")
	return cmd
}

func newCmdWorkflowExecute(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var wait bool
	var paramsJson string
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a workflow and view execution results",
		Long: `Execute a workflow and optionally wait for completion to view results.

The execution flow:
1. Create execution task via /gw/workflow/debug/create
2. Execute workflow via /gw/workflow/debug/do (SSE streaming)
3. Query execution context via /gw/workflow/design/context

EXAMPLES:
    # Execute workflow without waiting
    lingtong-cli workflow execute --workflow-id 947

    # Execute workflow and wait for results
    lingtong-cli workflow execute --workflow-id 947 --wait

    # Execute workflow with parameters
    lingtong-cli workflow execute --workflow-id 947 --params '{"key":"value"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			c.DisableProxy()

			// Step 1: Create execution task
			createPath := "/gw/workflow/debug/create"
			createParams := map[string]interface{}{
				"workflowId": workflowId,
			}
			createResp, err := c.Do("POST", createPath, createParams, nil)
			if err != nil {
				return fmt.Errorf("failed to create execution task: %w", err)
			}

			var createResult map[string]interface{}
			if err := json.Unmarshal(createResp, &createResult); err != nil {
				return fmt.Errorf("failed to parse create response: %w", err)
			}

			// Extract receiptId from response
			result, ok := createResult["result"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid create response format")
			}

			receiptId, ok := result["receiptId"].(string)
			if !ok {
				return fmt.Errorf("receiptId not found in response")
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)

			// If not waiting, just return the receiptId
			if !wait {
				output := map[string]interface{}{
					"receiptId": receiptId,
					"message":   "Execution task created. Use 'workflow logs --receipt-id " + receiptId + "' to view results.",
				}
				return w.Write(output)
			}

			// Step 2: Execute workflow (SSE streaming)
			// Note: SSE execution is complex, for now we'll query the context after a delay
			// TODO: Implement proper SSE streaming in CLI
			fmt.Fprintf(f.IOStreams.Out, "Executing workflow (receiptId: %s)...\n", receiptId)
			fmt.Fprintf(f.IOStreams.Out, "Note: SSE streaming is not supported in CLI yet. Querying execution context...\n\n")

			// Step 3: Query execution context
			contextPath := "/gw/workflow/design/context"
			contextBody := map[string]interface{}{
				"taskId":   receiptId,
				"pageNum":  1,
				"pageSize": 100,
			}

			// Add custom parameters if provided
			if paramsJson != "" {
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(paramsJson), &params); err != nil {
					return fmt.Errorf("invalid params JSON: %w", err)
				}
				contextBody["params"] = params
			}

			contextResp, err := c.Post(contextPath, contextBody)
			if err != nil {
				return fmt.Errorf("failed to query execution context: %w", err)
			}

			var contextResult map[string]interface{}
			if err := json.Unmarshal(contextResp, &contextResult); err != nil {
				return fmt.Errorf("failed to parse context response: %w", err)
			}

			return w.Write(contextResult)
		},
	}

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for execution to complete and show results")
	cmd.Flags().StringVar(&paramsJson, "params", "", "Workflow parameters as JSON string")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowInfo(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Query workflow details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := fmt.Sprintf("/gw/workflow/get?workflowId=%d", workflowId)

			resp, err := c.Get(path, nil)
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

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowLogs(f *cmdutil.Factory) *cobra.Command {
	var receiptId string
	var pid string
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Query workflow execution logs",
		Long: `Query workflow execution context and logs by receipt ID.

The receipt ID is returned by the 'workflow execute' command.

EXAMPLES:
    # Query execution logs
    lingtong-cli workflow logs --receipt-id abc123

    # Query child context by pid
    lingtong-cli workflow logs --receipt-id abc123 --pid 456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if receiptId == "" {
				return fmt.Errorf("--receipt-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := "/gw/workflow/design/context"

			body := map[string]interface{}{
				"taskId":   receiptId,
				"pageNum":  page,
				"pageSize": pageSize,
			}

			if pid != "" {
				body["pid"] = pid
			}

			resp, err := c.Post(path, body)
			if err != nil {
				return fmt.Errorf("failed to query execution context: %w", err)
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

	cmd.Flags().StringVar(&receiptId, "receipt-id", "", "Execution receipt ID (required)")
	cmd.Flags().StringVar(&pid, "pid", "", "Parent ID for querying child context")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "size", 20, "Page size")
	_ = cmd.MarkFlagRequired("receipt-id")
	return cmd
}

func newCmdWorkflowPublish(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var version, memo string
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a workflow to create a versioned snapshot",
		Long: `Publish a workflow to create a versioned snapshot for production use.

EXAMPLES:
    # Publish with version and memo
    lingtong-cli workflow publish --workflow-id 100 --version "v1.0.0" --memo "Initial release"

    # Publish without version (auto-generated)
    lingtong-cli workflow publish --workflow-id 100`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := "/gw/workflow/publish/release"

			body := map[string]interface{}{
				"workflowId": workflowId,
			}

			if version != "" {
				body["version"] = version
			}
			if memo != "" {
				body["memo"] = memo
			}

			resp, err := c.Post(path, body)
			if err != nil {
				return fmt.Errorf("failed to publish workflow: %w", err)
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

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	cmd.Flags().StringVar(&version, "version", "", "Version number (optional, auto-generated if not provided)")
	cmd.Flags().StringVar(&memo, "memo", "", "Publish memo or description")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowVersions(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List all published versions of a workflow",
		Long: `List all published versions (snapshots) of a workflow.

EXAMPLES:
    # List all versions
    lingtong-cli workflow versions --workflow-id 100

    # List with pagination
    lingtong-cli workflow versions --workflow-id 100 --page 1 --size 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			params := map[string]interface{}{
				"workflowId": workflowId,
				"pageNum":    page,
				"pageSize":   pageSize,
			}
			path := "/gw/workflow/version/list"

			resp, err := c.Get(path, params)
			if err != nil {
				return fmt.Errorf("failed to fetch workflow versions: %w", err)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)

			// Extract list data for any format
			var result map[string]interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return err
			}

			// If we have result.snapshotList, try to format nicely
			if data, ok := result["result"].(map[string]interface{}); ok {
				if list, ok := data["snapshotList"].([]interface{}); ok && len(list) > 0 {
					if format == output.FormatTable || format == output.FormatPretty {
						fmt.Fprintf(f.IOStreams.Out, "%-15s %-30s %-25s %-10s\n", "VERSION", "MEMO", "PUBLISH_TIME", "STATUS")
						fmt.Fprintf(f.IOStreams.Out, "%-15s %-30s %-25s %-10s\n", "-------", "----", "------------", "------")
						for _, item := range list {
							if row, ok := item.(map[string]interface{}); ok {
								version := ""
								memo := ""
								publishTime := ""
								status := ""
								if v, ok := row["version"].(string); ok {
									version = v
								}
								if m, ok := row["memo"].(string); ok {
									memo = m
								}
								if pt, ok := row["publishTime"].(string); ok {
									publishTime = pt
								}
								if s, ok := row["status"].(string); ok {
									status = s
								}
								fmt.Fprintf(f.IOStreams.Out, "%-15s %-30s %-25s %-10s\n", version, memo, publishTime, status)
							}
						}
						return nil
					}
				}
			}

			// Default to JSON output
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "size", 20, "Page size")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowAPITest(f *cmdutil.Factory) *cobra.Command {
	var appTag string
	var paramsJson string
	var token string
	cmd := &cobra.Command{
		Use:   "api-test",
		Short: "Test a published workflow's open API endpoint",
		Long: `Test a published workflow by calling its open API endpoint directly.

This command uses direct HTTP mode (not proxy) to call the open API at /gw/{appTag}/workflows/run.

EXAMPLES:
    # Test with JSON parameters
    lingtong-cli workflow api-test --app-tag abc123 --params '{"key":"value"}'

    # Test with custom token
    lingtong-cli workflow api-test --app-tag abc123 --token "your-token" --params '{"key":"value"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appTag == "" {
				return fmt.Errorf("--app-tag is required")
			}
			if paramsJson == "" {
				return fmt.Errorf("--params is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			c.DisableProxy()

			// Override token if provided by creating a new client
			if token != "" {
				c = client.NewClient(f.Config.Host, token)
				c.DisableProxy()
			}

			path := fmt.Sprintf("/gw/%s/workflows/run", appTag)

			var params map[string]interface{}
			if err := json.Unmarshal([]byte(paramsJson), &params); err != nil {
				return fmt.Errorf("invalid params JSON: %w", err)
			}

			resp, err := c.Post(path, params)
			if err != nil {
				return fmt.Errorf("failed to call workflow API: %w", err)
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

	cmd.Flags().StringVar(&appTag, "app-tag", "", "App tag for the published workflow (required)")
	cmd.Flags().StringVar(&paramsJson, "params", "", "Input parameters as JSON string (required)")
	cmd.Flags().StringVar(&token, "token", "", "Bearer token for API call (overrides configured token)")
	_ = cmd.MarkFlagRequired("app-tag")
	_ = cmd.MarkFlagRequired("params")
	return cmd
}

func newCmdWorkflowCreate(f *cmdutil.Factory) *cobra.Command {
	var name, description, env, dslFile, dslString, template string
	var appId int
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workflow",
		Long: `Create a new workflow, optionally seeding its DSL.

Creation is two steps, matching the platform API: /workflow/create makes the
shell (appId/name/env) and returns the new workflow id; when DSL is supplied it
is then persisted through /workflow/update as the 'content' string so the full
node graph is saved.

Provide DSL via --dsl-file, --dsl-string, or a built-in --template.

EXAMPLES:
    # Create an empty workflow shell
    lingtong-cli workflow create --app-id 165 --name "My Workflow"

    # Create with DSL file
    lingtong-cli workflow create --app-id 165 --name "My Workflow" --dsl-file workflow.json

    # Create with a built-in template
    lingtong-cli workflow create --app-id 165 --name "Simple Workflow" --template simple`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if appId == 0 {
				return fmt.Errorf("--app-id is required")
			}

			// Get DSL from file, string, or template (optional).
			var dslContent string
			if dslFile != "" {
				data, err := readFile(dslFile)
				if err != nil {
					return fmt.Errorf("failed to read DSL file: %w", err)
				}
				dslContent = string(data)
			} else if dslString != "" {
				dslContent = dslString
			} else if template != "" {
				dslContent = getWorkflowTemplate(template, name, env)
				if dslContent == "" {
					return fmt.Errorf("unknown template: %s (available: simple, connector, order_sync, approval, data_pipeline, api_wrapper)", template)
				}
			}
			if dslContent != "" {
				var dsl map[string]interface{}
				if err := json.Unmarshal([]byte(dslContent), &dsl); err != nil {
					return fmt.Errorf("invalid DSL JSON: %w", err)
				}
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)

			// Step 1: create the shell.
			createBody := map[string]interface{}{
				"appId": appId,
				"name":  name,
			}
			if env != "" {
				createBody["env"] = env
			}
			if description != "" {
				createBody["memo"] = description
			}
			createResp, err := c.Post("/gw/workflow/create", createBody)
			if err != nil {
				return fmt.Errorf("failed to create workflow: %w", err)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)

			// No DSL: return the create response as-is.
			if dslContent == "" {
				var data interface{}
				if err := json.Unmarshal(createResp, &data); err != nil {
					return err
				}
				return w.Write(data)
			}

			// Step 2: persist the DSL via the config endpoint.
			newID, ts := newWorkflowIDFromCreate(createResp)
			if newID == 0 {
				return fmt.Errorf("workflow created but new id not found in response; DSL not persisted")
			}
			updateBody := map[string]interface{}{
				"workflowId": newID,
				"content":    dslContent,
				"name":       name,
				"env":        env,
				"appId":      appId,
				"ts":         ts,
				"forced":     true,
			}
			updateResp, err := c.Post("/gw/workflow/update", updateBody)
			if err != nil {
				return fmt.Errorf("workflow %d created but DSL update failed: %w", newID, err)
			}
			var data interface{}
			if err := json.Unmarshal(updateResp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Workflow name (required)")
	cmd.Flags().IntVar(&appId, "app-id", 0, "Application ID (required)")
	cmd.Flags().StringVar(&description, "description", "", "Workflow description")
	cmd.Flags().StringVar(&env, "env", "", "Environment (test/formal)")
	cmd.Flags().StringVar(&dslFile, "dsl-file", "", "Path to DSL JSON file")
	cmd.Flags().StringVar(&dslString, "dsl-string", "", "DSL JSON string")
	cmd.Flags().StringVar(&template, "template", "", "Built-in template (simple, connector, order_sync, approval, data_pipeline, api_wrapper)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// newWorkflowIDFromCreate extracts the new workflow id and ts from a
// /workflow/create response, tolerating the id living either directly on the
// result or on a nested workflow object.
func newWorkflowIDFromCreate(resp []byte) (int, string) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(resp, &parsed); err != nil {
		return 0, ""
	}
	result, _ := parsed["result"].(map[string]interface{})
	if result == nil {
		return 0, ""
	}
	if id := metaInt(result["id"]); id != 0 {
		return id, tsToString(result["ts"])
	}
	// Some responses nest the workflow under a typed object.
	for _, v := range result {
		if obj, ok := v.(map[string]interface{}); ok {
			if id := metaInt(obj["id"]); id != 0 {
				return id, tsToString(obj["ts"])
			}
		}
	}
	return 0, ""
}

func newCmdWorkflowUpdate(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var name, env, dslFile, dslString string
	var forced bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing workflow",
		Long: `Update an existing workflow's DSL or metadata.

You can provide DSL via --dsl-file or --dsl-string. The DSL is sent to the
config endpoint (/workflow/update) as the 'content' string together with the
workflow's appId/env/name/ts, so the full node graph is persisted. Updating
only the name routes to /workflow/update/basic.

IMPORTANT: pushing DSL through /workflow/update/basic (metadata only) silently
drops the node graph and causes a buildFlowSource NPE at execution time.

EXAMPLES:
    # Update with DSL file (persists the full node graph)
    lingtong-cli workflow update --workflow-id 100 --dsl-file workflow.json

    # Update name only
    lingtong-cli workflow update --workflow-id 100 --name "New Name"

    # Force update (ignore validation warnings)
    lingtong-cli workflow update --workflow-id 100 --dsl-file workflow.json --forced`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			// Get DSL from file or string (kept as a raw string: the API stores
			// the DSL as the 'content' string field, not a nested object).
			var dslContent string
			if dslFile != "" {
				data, err := readFile(dslFile)
				if err != nil {
					return fmt.Errorf("failed to read DSL file: %w", err)
				}
				dslContent = string(data)
			} else if dslString != "" {
				dslContent = dslString
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)

			// The update endpoints require appId/env/name/ts that live on the
			// persisted workflow, so fetch them and let flags override.
			meta, err := fetchWorkflowMeta(c, workflowId)
			if err != nil {
				return fmt.Errorf("failed to load workflow %d: %w", workflowId, err)
			}
			effName := name
			if effName == "" {
				effName, _ = meta["name"].(string)
			}
			effEnv := env
			if effEnv == "" {
				effEnv, _ = meta["env"].(string)
			}
			appId := metaInt(meta["appId"])

			var path string
			var body map[string]interface{}
			if dslContent != "" {
				// Validate the DSL parses before sending.
				var dsl map[string]interface{}
				if err := json.Unmarshal([]byte(dslContent), &dsl); err != nil {
					return fmt.Errorf("invalid DSL JSON: %w", err)
				}
				path = "/gw/workflow/update"
				body = map[string]interface{}{
					"workflowId": workflowId,
					"content":    dslContent,
					"name":       effName,
					"env":        effEnv,
					"appId":      appId,
					"ts":         tsToString(meta["ts"]),
					"forced":     forced,
				}
			} else {
				// Metadata-only update.
				path = "/gw/workflow/update/basic"
				body = map[string]interface{}{
					"workflowId": workflowId,
					"name":       effName,
					"env":        effEnv,
					"appId":      appId,
				}
			}

			resp, err := c.Post(path, body)
			if err != nil {
				return fmt.Errorf("failed to update workflow: %w", err)
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

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "New workflow name")
	cmd.Flags().StringVar(&env, "env", "", "Environment override (test/formal); defaults to the workflow's env")
	cmd.Flags().StringVar(&dslFile, "dsl-file", "", "Path to DSL JSON file")
	cmd.Flags().StringVar(&dslString, "dsl-string", "", "DSL JSON string")
	cmd.Flags().BoolVar(&forced, "forced", false, "Force update (ignore validation warnings)")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowDelete(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a workflow",
		Long: `Delete a workflow permanently.

EXAMPLES:
    # Delete with confirmation prompt
    lingtong-cli workflow delete --workflow-id 100

    # Delete without confirmation
    lingtong-cli workflow delete --workflow-id 100 --confirm`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			if !confirm {
				fmt.Fprintf(f.IOStreams.ErrOut, "Are you sure you want to delete workflow %d? [y/N]: ", workflowId)
				var response string
				fmt.Fscanln(f.IOStreams.In, &response)
				if response != "y" && response != "Y" {
					fmt.Fprintf(f.IOStreams.Out, "Cancelled.\n")
					return nil
				}
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := "/gw/workflow/delete"

			body := map[string]interface{}{
				"ids": []int{workflowId},
			}

			resp, err := c.Post(path, body)
			if err != nil {
				return fmt.Errorf("failed to delete workflow: %w", err)
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

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowAPIEnable(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	cmd := &cobra.Command{
		Use:   "api-enable",
		Short: "Enable open API access for a workflow",
		Long: `Enable open API access for a workflow, allowing it to be called via the open API endpoint.

EXAMPLES:
    lingtong-cli workflow api-enable --workflow-id 100`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := "/gw/workflow/api/open"

			body := map[string]interface{}{
				"workflowId": workflowId,
				"open":       1,
			}

			resp, err := c.Post(path, body)
			if err != nil {
				return fmt.Errorf("failed to enable API access: %w", err)
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

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowAPIDisable(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	cmd := &cobra.Command{
		Use:   "api-disable",
		Short: "Disable open API access for a workflow",
		Long: `Disable open API access for a workflow, preventing it from being called via the open API endpoint.

EXAMPLES:
    lingtong-cli workflow api-disable --workflow-id 100`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := "/gw/workflow/api/open"

			body := map[string]interface{}{
				"workflowId": workflowId,
				"open":       0,
			}

			resp, err := c.Post(path, body)
			if err != nil {
				return fmt.Errorf("failed to disable API access: %w", err)
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

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

// getWorkflowTemplate returns a built-in workflow template as JSON string.
func getWorkflowTemplate(templateName, name, env string) string {
	_ = name // Used for template customization
	_ = env
	switch templateName {
	case "simple":
		return `{
  "nodes": [
    {"id": "node_start", "type": "w_start", "name": "Start", "properties": {}},
    {"id": "node_end", "type": "w_end", "name": "End", "properties": {}}
  ],
  "edges": [
    {"id": "edge_1", "source": "node_start", "target": "node_end", "sourceHandle": "right", "targetHandle": "left"}
  ]
}`
	case "connector":
		return `{
  "nodes": [
    {"id": "node_start", "type": "w_start", "name": "Start", "properties": {}},
    {"id": "node_connector", "type": "w_connector", "name": "Connector", "properties": {"connectorId": "", "interfaceId": "", "authAccountId": ""}},
    {"id": "node_end", "type": "w_end", "name": "End", "properties": {}}
  ],
  "edges": [
    {"id": "edge_1", "source": "node_start", "target": "node_connector", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_2", "source": "node_connector", "target": "node_end", "sourceHandle": "right", "targetHandle": "left"}
  ]
}`
	case "order_sync":
		return `{
  "nodes": [
    {"id": "node_start", "type": "w_start", "name": "Start", "properties": {}},
    {"id": "node_query_orders", "type": "w_connector", "name": "Query Orders from ERP", "properties": {"connectorId": "", "interfaceId": "", "authAccountId": ""}},
    {"id": "node_transform", "type": "w_script", "name": "Transform Order Data", "properties": {"language": "javascript", "script": "// Transform order data\nvar orders = input.data;\nreturn orders.map(function(order) {\n  return {orderId: order.id, amount: order.total};\n});"}},
    {"id": "node_sync_target", "type": "w_connector", "name": "Sync to Target System", "properties": {"connectorId": "", "interfaceId": "", "authAccountId": ""}},
    {"id": "node_end", "type": "w_end", "name": "End", "properties": {}}
  ],
  "edges": [
    {"id": "edge_1", "source": "node_start", "target": "node_query_orders", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_2", "source": "node_query_orders", "target": "node_transform", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_3", "source": "node_transform", "target": "node_sync_target", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_4", "source": "node_sync_target", "target": "node_end", "sourceHandle": "right", "targetHandle": "left"}
  ]
}`
	case "approval":
		return `{
  "nodes": [
    {"id": "node_start", "type": "w_start", "name": "Start", "properties": {}},
    {"id": "node_check_amount", "type": "w_if", "name": "Check Amount", "properties": {"conditions": [{"field": "input.amount", "operator": ">", "value": 1000}]}},
    {"id": "node_manager_approval", "type": "w_connector", "name": "Manager Approval", "properties": {"connectorId": "", "interfaceId": "", "authAccountId": ""}},
    {"id": "node_end", "type": "w_end", "name": "End", "properties": {}}
  ],
  "edges": [
    {"id": "edge_1", "source": "node_start", "target": "node_check_amount", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_2", "source": "node_check_amount", "target": "node_manager_approval", "sourceHandle": "true", "targetHandle": "left"},
    {"id": "edge_3", "source": "node_check_amount", "target": "node_end", "sourceHandle": "false", "targetHandle": "left"},
    {"id": "edge_4", "source": "node_manager_approval", "target": "node_end", "sourceHandle": "right", "targetHandle": "left"}
  ]
}`
	case "data_pipeline":
		return `{
  "nodes": [
    {"id": "node_start", "type": "w_start", "name": "Start", "properties": {}},
    {"id": "node_fetch_data", "type": "w_connector", "name": "Fetch Source Data", "properties": {"connectorId": "", "interfaceId": "", "authAccountId": ""}},
    {"id": "node_data_split", "type": "w_dataSplit", "name": "Split Data", "properties": {"splitField": "items"}},
    {"id": "node_transform", "type": "w_script", "name": "Transform Each Item", "properties": {"language": "javascript", "script": "// Transform item\nvar item = input.item;\nreturn {id: item.id, name: item.name, price: item.price * 1.1};"}},
    {"id": "node_end", "type": "w_end", "name": "End", "properties": {}}
  ],
  "edges": [
    {"id": "edge_1", "source": "node_start", "target": "node_fetch_data", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_2", "source": "node_fetch_data", "target": "node_data_split", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_3", "source": "node_data_split", "target": "node_transform", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_4", "source": "node_transform", "target": "node_end", "sourceHandle": "right", "targetHandle": "left"}
  ]
}`
	case "api_wrapper":
		return `{
  "nodes": [
    {"id": "node_start", "type": "w_start", "name": "Start", "properties": {}},
    {"id": "node_validate_input", "type": "w_script", "name": "Validate Input", "properties": {"language": "javascript", "script": "// Validate input parameters\nvar input = input.params;\nif (!input.id) { throw new Error('Missing required field: id'); }\nreturn input;"}},
    {"id": "node_call_api", "type": "w_connector", "name": "Call External API", "properties": {"connectorId": "", "interfaceId": "", "authAccountId": ""}},
    {"id": "node_end", "type": "w_end", "name": "End", "properties": {}}
  ],
  "edges": [
    {"id": "edge_1", "source": "node_start", "target": "node_validate_input", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_2", "source": "node_validate_input", "target": "node_call_api", "sourceHandle": "right", "targetHandle": "left"},
    {"id": "edge_3", "source": "node_call_api", "target": "node_end", "sourceHandle": "right", "targetHandle": "left"}
  ]
}`
	default:
		return ""
	}
}

// readFile reads a file's content.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// getTemplateList returns all available workflow templates.
func getTemplateList() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "simple", "nodes": 2, "description": "Simple start to end workflow"},
		{"name": "connector", "nodes": 3, "description": "Start to connector to end workflow"},
		{"name": "order_sync", "nodes": 5, "description": "Order sync from ERP system"},
		{"name": "approval", "nodes": 4, "description": "Approval workflow with conditional branches"},
		{"name": "data_pipeline", "nodes": 5, "description": "Data pipeline with transformation"},
		{"name": "api_wrapper", "nodes": 4, "description": "Wrapper to expose connector as API"},
	}
}

// ============================================================================
// Phase 3: Template Commands
// ============================================================================

func newCmdWorkflowTemplate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "template", Short: "Manage workflow templates", Long: "List, view, and use templates."}
	cmd.AddCommand(newCmdWorkflowTemplateList(f))
	cmd.AddCommand(newCmdWorkflowTemplateShow(f))
	cmd.AddCommand(newCmdWorkflowTemplateUse(f))
	return cmd
}

func newCmdWorkflowTemplateList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available workflow templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			templates := getTemplateList()
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(templates)
		},
	}
	return cmd
}

func newCmdWorkflowTemplateShow(f *cmdutil.Factory) *cobra.Command {
	var templateName string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show template details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if templateName == "" {
				return fmt.Errorf("--name is required")
			}
			dsl := getWorkflowTemplate(templateName, "Example Workflow", "")
			if dsl == "" {
				return fmt.Errorf("template '%s' not found", templateName)
			}
			var template map[string]interface{}
			if err := json.Unmarshal([]byte(dsl), &template); err != nil {
				return fmt.Errorf("invalid template DSL: %w", err)
			}
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(template)
		},
	}
	cmd.Flags().StringVar(&templateName, "name", "", "Template name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newCmdWorkflowTemplateUse(f *cmdutil.Factory) *cobra.Command {
	var templateName, workflowName, outputFile string
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a template to create a DSL file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if templateName == "" {
				return fmt.Errorf("--name is required")
			}
			if outputFile == "" {
				return fmt.Errorf("--output is required")
			}
			dsl := getWorkflowTemplate(templateName, workflowName, "")
			if dsl == "" {
				return fmt.Errorf("template '%s' not found", templateName)
			}
			if err := os.WriteFile(outputFile, []byte(dsl), 0644); err != nil {
				return fmt.Errorf("failed to write DSL file: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Template '%s' saved to %s\n", templateName, outputFile)
			return nil
		},
	}
	cmd.Flags().StringVar(&templateName, "name", "", "Template name (required)")
	cmd.Flags().StringVar(&workflowName, "workflow-name", "Example Workflow", "Workflow name")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output DSL file (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

// ============================================================================
// Phase 3: Validate Command
// ============================================================================

func newCmdWorkflowValidate(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var dslFile string
	var strict bool
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate workflow DSL",
		RunE: func(cmd *cobra.Command, args []string) error {
			var dsl map[string]interface{}
			if workflowId != 0 {
				c := client.NewClient(f.Config.Host, f.Config.Token)
				path := fmt.Sprintf("/gw/workflow/get?workflowId=%d", workflowId)
				resp, err := c.Get(path, nil)
				if err != nil {
					return fmt.Errorf("failed to fetch workflow: %w", err)
				}
				var result map[string]interface{}
				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}
				if res, ok := result["result"].(map[string]interface{}); ok {
					if content, ok := res["content"].(string); ok {
						if err := json.Unmarshal([]byte(content), &dsl); err != nil {
							return fmt.Errorf("invalid DSL JSON: %w", err)
						}
					} else if dslData, ok := res["dsl"]; ok {
						if m, ok := dslData.(map[string]interface{}); ok {
							dsl = m
						}
					}
				}
			} else if dslFile != "" {
				data, err := readFile(dslFile)
				if err != nil {
					return fmt.Errorf("failed to read DSL file: %w", err)
				}
				if err := json.Unmarshal(data, &dsl); err != nil {
					return fmt.Errorf("invalid DSL JSON: %w", err)
				}
			} else {
				return fmt.Errorf("either --workflow-id or --dsl-file is required")
			}
			result := validateDSL(dsl, strict)
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(result)
		},
	}
	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID to validate")
	cmd.Flags().StringVar(&dslFile, "dsl-file", "", "DSL file to validate")
	cmd.Flags().BoolVar(&strict, "strict", false, "Include warnings")
	return cmd
}

func validateDSL(dsl map[string]interface{}, strict bool) map[string]interface{} {
	var errors []string
	var warnings []string
	nodesRaw, hasNodes := dsl["nodes"]
	edgesRaw, hasEdges := dsl["edges"]
	if !hasNodes {
		errors = append(errors, "Missing 'nodes' field")
	}
	if !hasEdges {
		errors = append(errors, "Missing 'edges' field")
	}
	if !hasNodes || !hasEdges {
		return map[string]interface{}{"valid": false, "errors": errors, "warnings": warnings, "nodeCount": 0, "edgeCount": 0}
	}
	nodes, nodesOk := nodesRaw.([]interface{})
	edges, edgesOk := edgesRaw.([]interface{})
	if !nodesOk {
		errors = append(errors, "'nodes' is not an array")
	}
	if !edgesOk {
		errors = append(errors, "'edges' is not an array")
	}
	if len(errors) > 0 {
		return map[string]interface{}{"valid": false, "errors": errors, "warnings": warnings, "nodeCount": 0, "edgeCount": 0}
	}
	nodeIds := make(map[string]bool)
	startCount := 0
	endCount := 0
	validNodeTypes := map[string]bool{
		"w_start": true, "w_end": true, "w_connector": true, "w_if": true, "w_switch": true,
		"w_dataSplit": true, "w_cycle": true, "w_joinPipeline": true, "w_script": true,
		"w_modePipe": true, "w_dataPush": true, "w_pushRecord": true,
	}
	for _, nodeRaw := range nodes {
		node, ok := nodeRaw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := node["id"].(string)
		nodeType, _ := node["type"].(string)
		if id != "" {
			nodeIds[id] = true
		}
		if nodeType == "w_start" {
			startCount++
		}
		if nodeType == "w_end" {
			endCount++
		}
		if nodeType != "" && !validNodeTypes[nodeType] {
			errors = append(errors, fmt.Sprintf("Invalid node type '%s'", nodeType))
		}
		if nodeType == "w_connector" || nodeType == "w_modePipe" {
			if data, ok := node["data"].(map[string]interface{}); ok {
				if _, has := data["connector"]; !has {
					errors = append(errors, fmt.Sprintf("Node '%s' missing 'connector'", id))
				}
			}
		}
		// Script nodes require an assertConfig at runtime; a workflow that omits
		// it publishes fine but fails execution with "断言配置不允许为null".
		if nodeType == "w_script" {
			data, _ := node["data"].(map[string]interface{})
			if data == nil || data["assertConfig"] == nil {
				warnings = append(warnings, fmt.Sprintf("Node '%s' (w_script) missing 'assertConfig' (required at runtime, e.g. {\"assertType\":\"throwException\"})", id))
			}
		}
	}
	if startCount == 0 {
		errors = append(errors, "No start node found")
	} else if startCount > 1 {
		errors = append(errors, fmt.Sprintf("Multiple start nodes (%d)", startCount))
	}
	if endCount == 0 {
		errors = append(errors, "No end node found")
	}
	for _, edgeRaw := range edges {
		edge, ok := edgeRaw.(map[string]interface{})
		if !ok {
			continue
		}
		source, _ := edge["source"].(string)
		target, _ := edge["target"].(string)
		if source != "" && !nodeIds[source] {
			errors = append(errors, fmt.Sprintf("Edge references non-existent source '%s'", source))
		}
		if target != "" && !nodeIds[target] {
			errors = append(errors, fmt.Sprintf("Edge references non-existent target '%s'", target))
		}
	}
	if strict {
		if len(nodes) > 20 {
			warnings = append(warnings, "Large workflow (>20 nodes), consider splitting")
		}
	}
	return map[string]interface{}{
		"valid": len(errors) == 0, "errors": errors, "warnings": warnings,
		"nodeCount": len(nodes), "edgeCount": len(edges),
	}
}

// ============================================================================
// Phase 3: Additional Commands (Version, Test, Doc, Dependency)
// ============================================================================

func newCmdWorkflowVersion(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "version", Short: "Manage workflow versions", Long: "Version management operations including rollback."}
	cmd.AddCommand(newCmdWorkflowVersionRollback(f))
	return cmd
}

func newCmdWorkflowVersionRollback(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var version string
	var dryRun, confirm bool
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback workflow to a previous version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}
			if version == "" {
				return fmt.Errorf("--version is required")
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			params := map[string]interface{}{"workflowId": workflowId, "pageNum": 1, "pageSize": 100}
			resp, err := c.Get("/gw/workflow/version/list", params)
			if err != nil {
				return fmt.Errorf("failed to fetch versions: %w", err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return err
			}
			var snapshotId interface{}
			if res, ok := result["result"].(map[string]interface{}); ok {
				if list, ok := res["snapshotList"].([]interface{}); ok {
					for _, item := range list {
						if ver, ok := item.(map[string]interface{}); ok {
							if ver["version"] == version {
								snapshotId = ver["id"]
								break
							}
						}
					}
				}
			}
			if snapshotId == nil {
				return fmt.Errorf("version '%s' not found", version)
			}
			if dryRun {
				return f.NewWriter(output.Format(cmd.Flag("format").Value.String())).Write(map[string]interface{}{
					"dryRun": true, "workflowId": workflowId, "targetVersion": version, "snapshotId": snapshotId,
				})
			}
			if !confirm {
				fmt.Fprintf(f.IOStreams.ErrOut, "Rollback workflow %d to version %s? [y/N]: ", workflowId, version)
				var response string
				fmt.Fscanln(f.IOStreams.In, &response)
				if response != "y" && response != "Y" {
					fmt.Fprintln(f.IOStreams.Out, "Cancelled.")
					return nil
				}
			}
			body := map[string]interface{}{"workflowId": workflowId, "version": version, "snapshotId": snapshotId}
			rollbackResp, err := c.Post("/gw/workflow/version/rollback", body)
			if err != nil {
				return fmt.Errorf("failed to rollback: %w", err)
			}
			var rollbackResult map[string]interface{}
			if err := json.Unmarshal(rollbackResp, &rollbackResult); err != nil {
				return err
			}
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(rollbackResult)
		},
	}
	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	cmd.Flags().StringVar(&version, "version", "", "Version to rollback to (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without executing")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("workflow-id")
	_ = cmd.MarkFlagRequired("version")
	return cmd
}

func newCmdWorkflowTest(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "test", Short: "Test workflow execution", Long: "Run workflow tests."}
	cmd.AddCommand(newCmdWorkflowTestRun(f))
	return cmd
}

func newCmdWorkflowTestRun(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var testDataFile, paramsJson string
	var verbose bool
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a workflow test",
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowId == 0 {
				return fmt.Errorf("--workflow-id is required")
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			var testData map[string]interface{}
			if testDataFile != "" {
				data, err := readFile(testDataFile)
				if err != nil {
					return fmt.Errorf("failed to read test data: %w", err)
				}
				if err := json.Unmarshal(data, &testData); err != nil {
					return fmt.Errorf("invalid test data JSON: %w", err)
				}
			}
			if paramsJson != "" {
				if testData == nil {
					testData = make(map[string]interface{})
				}
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(paramsJson), &params); err != nil {
					return fmt.Errorf("invalid params JSON: %w", err)
				}
				for k, v := range params {
					testData[k] = v
				}
			}
			createPath := "/gw/workflow/debug/create"
			createParams := map[string]interface{}{
				"workflowId": workflowId,
			}
			createResp, err := c.Do("POST", createPath, createParams, nil)
			if err != nil {
				return fmt.Errorf("failed to create debug task: %w", err)
			}
			var createResult map[string]interface{}
			if err := json.Unmarshal(createResp, &createResult); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			result, ok := createResult["result"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid response format")
			}
			receiptId, ok := result["receiptId"].(string)
			if !ok {
				return fmt.Errorf("receiptId not found")
			}
			if verbose {
				fmt.Fprintf(f.IOStreams.Out, "✓ Test started (receiptId: %s)\n", receiptId)
			}
			contextBody := map[string]interface{}{"taskId": receiptId, "pageNum": 1, "pageSize": 100}
			if testData != nil && len(testData) > 0 {
				contextBody["params"] = testData
			}
			contextResp, err := c.Post("/gw/workflow/design/context", contextBody)
			if err != nil {
				return fmt.Errorf("failed to query context: %w", err)
			}
			var contextResult map[string]interface{}
			if err := json.Unmarshal(contextResp, &contextResult); err != nil {
				return fmt.Errorf("failed to parse context: %w", err)
			}
			testReport := map[string]interface{}{
				"receiptId": receiptId, "workflowId": workflowId, "testData": testData, "result": contextResult,
			}
			if verbose {
				fmt.Fprintln(f.IOStreams.Out, "✓ Test completed")
			}
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(testReport)
		},
	}
	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID (required)")
	cmd.Flags().StringVar(&testDataFile, "test-data-file", "", "Test data JSON file")
	cmd.Flags().StringVar(&paramsJson, "params", "", "Test parameters as JSON")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	_ = cmd.MarkFlagRequired("workflow-id")
	return cmd
}

func newCmdWorkflowDoc(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "doc", Short: "Generate workflow documentation", Long: "Generate docs from workflow DSL."}
	cmd.AddCommand(newCmdWorkflowDocGenerate(f))
	return cmd
}

func newCmdWorkflowDocGenerate(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var dslFile, outputFile string
	var includeAPI, includeExamples bool
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate workflow documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			var workflow map[string]interface{}
			if workflowId != 0 {
				c := client.NewClient(f.Config.Host, f.Config.Token)
				path := fmt.Sprintf("/gw/workflow/get?workflowId=%d", workflowId)
				resp, err := c.Get(path, nil)
				if err != nil {
					return fmt.Errorf("failed to fetch workflow: %w", err)
				}
				var result map[string]interface{}
				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}
				if res, ok := result["result"].(map[string]interface{}); ok {
					workflow = res
				}
			} else if dslFile != "" {
				data, err := readFile(dslFile)
				if err != nil {
					return fmt.Errorf("failed to read DSL file: %w", err)
				}
				if err := json.Unmarshal(data, &workflow); err != nil {
					return fmt.Errorf("invalid DSL JSON: %w", err)
				}
			} else {
				return fmt.Errorf("either --workflow-id or --dsl-file is required")
			}
			doc := generateDoc(workflow, includeAPI, includeExamples)
			if outputFile != "" {
				if err := os.WriteFile(outputFile, []byte(doc), 0644); err != nil {
					return fmt.Errorf("failed to write doc: %w", err)
				}
				fmt.Fprintf(f.IOStreams.Out, "✓ Documentation saved to %s\n", outputFile)
			} else {
				fmt.Fprint(f.IOStreams.Out, doc)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID")
	cmd.Flags().StringVar(&dslFile, "dsl-file", "", "DSL file path")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "Output file path")
	cmd.Flags().BoolVar(&includeAPI, "include-api", false, "Include API docs")
	cmd.Flags().BoolVar(&includeExamples, "include-examples", false, "Include call examples")
	return cmd
}

func generateDoc(workflow map[string]interface{}, includeAPI, includeExamples bool) string {
	doc := "# Workflow Documentation\n\n"
	doc += "## Overview\n\n"
	if name, ok := workflow["name"].(string); ok {
		doc += fmt.Sprintf("**Name**: %s\n\n", name)
	}
	if id, ok := workflow["id"].(float64); ok {
		doc += fmt.Sprintf("**ID**: %.0f\n\n", id)
	}
	if env, ok := workflow["env"].(string); ok {
		doc += fmt.Sprintf("**Environment**: %s\n\n", env)
	}
	doc += "## Nodes\n\n"
	if nodes, ok := workflow["nodes"].([]interface{}); ok {
		for i, nodeRaw := range nodes {
			node, ok := nodeRaw.(map[string]interface{})
			if !ok {
				continue
			}
			nodeId, _ := node["id"].(string)
			nodeType, _ := node["type"].(string)
			nodeName := nodeId
			if data, ok := node["data"].(map[string]interface{}); ok {
				if title, ok := data["title"].(string); ok {
					nodeName = title
				}
			}
			doc += fmt.Sprintf("### %d. %s (%s)\n", i+1, nodeName, nodeId)
			doc += fmt.Sprintf("- **Type**: %s\n", nodeType)
			if data, ok := node["data"].(map[string]interface{}); ok {
				if conn, ok := data["connector"].(string); ok {
					doc += fmt.Sprintf("- **Connector**: %s\n", conn)
				}
			}
			doc += "\n"
		}
	}
	doc += "## Edges\n\n"
	if edges, ok := workflow["edges"].([]interface{}); ok {
		doc += "| Source | Target |\n|--------|--------|\n"
		for _, edgeRaw := range edges {
			edge, ok := edgeRaw.(map[string]interface{})
			if !ok {
				continue
			}
			source, _ := edge["source"].(string)
			target, _ := edge["target"].(string)
			doc += fmt.Sprintf("| %s | %s |\n", source, target)
		}
	}
	if includeAPI {
		doc += "\n## API\n\n"
		if openApi, ok := workflow["openApi"].(float64); ok && openApi == 1 {
			doc += "API access is enabled.\n"
		} else {
			doc += "API access is not enabled.\n"
		}
	}
	return doc
}

func newCmdWorkflowDependency(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "dependency", Short: "Analyze workflow dependencies", Long: "Parse and display dependencies."}
	cmd.AddCommand(newCmdWorkflowDependencyList(f))
	return cmd
}

func newCmdWorkflowDependencyList(f *cmdutil.Factory) *cobra.Command {
	var workflowId int
	var dslFile string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			var dsl map[string]interface{}
			if workflowId != 0 {
				c := client.NewClient(f.Config.Host, f.Config.Token)
				path := fmt.Sprintf("/gw/workflow/get?workflowId=%d", workflowId)
				resp, err := c.Get(path, nil)
				if err != nil {
					return fmt.Errorf("failed to fetch workflow: %w", err)
				}
				var result map[string]interface{}
				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}
				if res, ok := result["result"].(map[string]interface{}); ok {
					if content, ok := res["content"].(string); ok {
						if err := json.Unmarshal([]byte(content), &dsl); err != nil {
							return fmt.Errorf("invalid DSL JSON: %w", err)
						}
					} else if dslData, ok := res["dsl"]; ok {
						if m, ok := dslData.(map[string]interface{}); ok {
							dsl = m
						}
					}
				}
			} else if dslFile != "" {
				data, err := readFile(dslFile)
				if err != nil {
					return fmt.Errorf("failed to read DSL file: %w", err)
				}
				if err := json.Unmarshal(data, &dsl); err != nil {
					return fmt.Errorf("invalid DSL JSON: %w", err)
				}
			} else {
				return fmt.Errorf("either --workflow-id or --dsl-file is required")
			}
			deps := analyzeDeps(dsl)
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(deps)
		},
	}
	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID")
	cmd.Flags().StringVar(&dslFile, "dsl-file", "", "DSL file path")
	return cmd
}

func analyzeDeps(dsl map[string]interface{}) map[string]interface{} {
	var connectors []map[string]interface{}
	var scripts []map[string]interface{}
	if nodes, ok := dsl["nodes"].([]interface{}); ok {
		for _, nodeRaw := range nodes {
			node, ok := nodeRaw.(map[string]interface{})
			if !ok {
				continue
			}
			nodeType, _ := node["type"].(string)
			nodeId, _ := node["id"].(string)
			if nodeType == "w_connector" || nodeType == "w_modePipe" {
				if data, ok := node["data"].(map[string]interface{}); ok {
					conn := map[string]interface{}{
						"nodeId": nodeId, "nodeType": nodeType, "connector": data["connector"],
						"interfaceModelId": data["interfaceModelId"], "authAccountId": data["authAccountId"],
					}
					connectors = append(connectors, conn)
				}
			}
			if nodeType == "w_script" {
				if data, ok := node["data"].(map[string]interface{}); ok {
					if sc, ok := data["scriptConfig"].(map[string]interface{}); ok {
						scripts = append(scripts, map[string]interface{}{"nodeId": nodeId, "language": sc["language"]})
					}
				}
			}
		}
	}
	return map[string]interface{}{
		"connectors": connectors, "scripts": scripts,
		"connectorCount": len(connectors), "scriptCount": len(scripts),
	}
}
