// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package connector

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// NumberCompressor implements the same compress algorithm as the backend.
// It encodes (tenantId, workflowId) into a short base62 string used as appTag in Open API URLs.
var charSet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func numberCompress(tenantId string, flowId int) string {
	tBig := new(big.Int)
	tBig.SetString(tenantId, 10)
	fBig := big.NewInt(int64(flowId))
	base := big.NewInt(62)
	zero := big.NewInt(0)

	compressNum := func(n *big.Int) string {
		if n.Sign() == 0 {
			return ""
		}
		var result []byte
		for n.Cmp(zero) > 0 {
			rem := new(big.Int)
			n.DivMod(n, base, rem)
			result = append([]byte{charSet[rem.Int64()]}, result...)
		}
		return string(result)
	}

	compressedTenant := compressNum(tBig)
	compressedId := compressNum(fBig)
	// Prepend length marker
	marker := byte(len(compressedTenant))
	return string(charSet[marker]) + compressedTenant + compressedId
}

// buildConnectorInvokeDSL builds the universal workflow DSL for connector invocation.
// The workflow accepts an `args` JSON parameter containing connector, method, authAccount, body.
// The script node dynamically reads these from args and calls AppInvoker.invoke().
//
// Workflow structure: Start → Script → End
// This follows the same pattern as workflow ID 1002.
func buildConnectorInvokeDSL() map[string]interface{} {
	script := `var connector = context.get("connector");
var method = context.get("method");
var authAccount = context.get("authAccount");
var body = context.get("body") || {};
var env = context.get("env");
if (!connector || !method) {
    throw "connector and method are required";
}
context.put("_env", env);
var result = AppInvoker.invoke(context, method, authAccount, body);
return result;`

	return map[string]interface{}{
		"environment": "formal",
		"name":        "连接器调用",
		"viewport":    map[string]interface{}{"x": 52, "y": 11, "zoom": 1},
		"nodes": []map[string]interface{}{
			{
				"id": "w_start_first", "type": "w_start",
				"dragging": false, "width": 256, "height": 92,
				"position":          map[string]int{"x": 100, "y": 100},
				"positionAbsolute": map[string]int{"x": 100, "y": 100},
				"selected":         false,
				"data": map[string]interface{}{
					"title": "开始", "desc": "", "disabled": false,
					"outputVariables": []map[string]interface{}{
						{"variable": "connector", "variableAttr": map[string]interface{}{
							"dataType": "text", "label": "连接器名称", "required": true,
						}},
						{"variable": "method", "variableAttr": map[string]interface{}{
							"dataType": "text", "label": "方法名", "required": true,
						}},
						{"variable": "authAccount", "variableAttr": map[string]interface{}{
							"dataType": "text", "label": "认证账户名称", "required": true,
						}},
						{"variable": "body", "variableAttr": map[string]interface{}{
							"dataType": "json", "label": "请求体",
						}},
						{"variable": "env", "variableAttr": map[string]interface{}{
							"dataType": "text", "label": "环境", "required": true,
						}},
					},
				},
			},
			{
				"id": "w_script_invoke", "type": "w_script",
				"dragging": false, "width": 256, "height": 58,
				"position":          map[string]int{"x": 416, "y": 100},
				"positionAbsolute": map[string]int{"x": 416, "y": 100},
				"selected":         false,
				"data": map[string]interface{}{
					"title": "代码执行", "desc": "通用连接器调用 - AppInvoker.invoke",
					"disabled": false, "pids": []string{"w_start_first"},
					"inputVariables": []map[string]interface{}{
						{"variable": "connector", "variableAttr": map[string]interface{}{"value": "$w_start_first.connector", "dataType": "text"}},
						{"variable": "method", "variableAttr": map[string]interface{}{"value": "$w_start_first.method", "dataType": "text"}},
						{"variable": "authAccount", "variableAttr": map[string]interface{}{"value": "$w_start_first.authAccount", "dataType": "text"}},
						{"variable": "body", "variableAttr": map[string]interface{}{"value": "$w_start_first.body", "dataType": "json"}},
						{"variable": "env", "variableAttr": map[string]interface{}{"value": "$w_start_first.env", "dataType": "text"}},
					},
					"scriptConfig": map[string]interface{}{"language": "javascript", "script": script},
					"outputVariables": []map[string]interface{}{
						{"variable": "result", "variableAttr": map[string]interface{}{"dataType": "json", "schema": ""}},
					},
					"assertConfig": map[string]interface{}{"assertType": "throwException"},
				},
			},
			{
				"id": "w_end_result", "type": "w_end",
				"dragging": false, "width": 256, "height": 92,
				"position":          map[string]int{"x": 732, "y": 100},
				"positionAbsolute": map[string]int{"x": 732, "y": 100},
				"selected":         false,
				"data": map[string]interface{}{
					"title": "结束", "desc": "", "disabled": false,
					"pids": []string{"w_script_invoke"},
					"outputVariables": []map[string]interface{}{
						{"variable": "result", "variableAttr": map[string]interface{}{
							"dataType": "json",
							"value":    "$w_script_invoke.result",
							"nodeId":   "w_end_result",
							"schemaObj": map[string]interface{}{"fullPath": "", "schemaNodeVariable": "result", "schemaNodeId": "w_script_invoke"},
						}},
					},
				},
			},
		},
		"edges": []map[string]interface{}{
			{"id": "e1", "source": "w_start_first", "target": "w_script_invoke", "sourceHandle": "source", "targetHandle": "target", "type": "rounded-corner"},
			{"id": "e2", "source": "w_script_invoke", "target": "w_end_result", "sourceHandle": "source", "targetHandle": "target", "type": "rounded-corner"},
		},
	}
}

// ensureConnectorInvokeWorkflow creates or retrieves a cached universal workflow for connector invocation.
// Only one workflow per environment is needed since connector/method/authAccount are passed as runtime parameters.
func ensureConnectorInvokeWorkflow(f *cmdutil.Factory, env string) (*ConnectorCacheEntry, error) {
	key := universalCacheKey(env)

	// Check cache first
	cache, err := loadCache()
	if err != nil {
		return nil, fmt.Errorf("failed to load cache: %w", err)
	}
	if entry, ok := cache.Universal[key]; ok {
		return entry, nil
	}

	c := client.NewClient(f.Config.Host, f.Config.Token)

	// Step 1: Create the workflow using /workflow/create
	timestamp := time.Now().Format("150405")
	workflowName := fmt.Sprintf("[CLI] 通用连接器调用 (%s) %s", env, timestamp)

	createBody := map[string]interface{}{
		"appId": 165,
		"name":  workflowName,
		"memo":  fmt.Sprintf("CLI自动创建的通用连接器调用工作流\n环境: %s\n⚠️ 此工作流由CLI自动管理，请勿手动修改", env),
		"env":   env,
	}

	createResp, err := c.Post("/gw/workflow/create", createBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	var createResult map[string]interface{}
	if err := json.Unmarshal(createResp, &createResult); err != nil {
		return nil, fmt.Errorf("failed to parse create response: %w", err)
	}

	workflowId := 0
	ts := int64(0)
	if result, ok := createResult["result"].(map[string]interface{}); ok {
		if id, ok := result["id"].(float64); ok {
			workflowId = int(id)
		}
		if t, ok := result["ts"].(float64); ok {
			ts = int64(t)
		}
	}
	if workflowId == 0 {
		return nil, fmt.Errorf("failed to get workflow ID from response: %s", string(createResp))
	}

	// Step 2: Update workflow with DSL content using /workflow/update
	dsl := buildConnectorInvokeDSL()
	dsl["workId"] = workflowId
	dslJson, _ := json.Marshal(dsl)

	updateBody := map[string]interface{}{
		"appId":      165,
		"workflowId": workflowId,
		"name":       workflowName,
		"content":    string(dslJson),
		"env":        env,
		"ts":         ts,
		"forced":     true,
	}

	_, err = c.Post("/gw/workflow/update", updateBody)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow with DSL: %w", err)
	}

	// Step 3: Get workflow info to obtain tenantId
	tenantId, err := getWorkflowTenantId(c, workflowId)
	if err != nil {
		return nil, err
	}

	// Step 4: Enable Open API
	openBody := map[string]interface{}{"workflowId": workflowId, "open": 1}
	_, err = c.Post("/gw/workflow/api/open", openBody)
	if err != nil {
		return nil, fmt.Errorf("failed to enable Open API: %w", err)
	}

	// Step 5: Publish the workflow
	publishBody := map[string]interface{}{
		"workflowId": workflowId,
		"version":    "v1.0.0",
		"memo":       "CLI自动发布",
	}
	_, err = c.Post("/gw/workflow/publish/release", publishBody)
	if err != nil {
		return nil, fmt.Errorf("failed to publish workflow: %w", err)
	}

	// Step 6: Generate appTag using NumberCompressor algorithm
	appTag := numberCompress(tenantId, workflowId)

	// Step 7: Cache the result (universal - one per env)
	entry := &ConnectorCacheEntry{
		WorkflowId: workflowId,
		AppTag:     appTag,
		Env:        env,
	}
	cache.Universal[key] = entry
	if err := saveCache(cache); err != nil {
		return nil, fmt.Errorf("failed to save cache: %w", err)
	}

	return entry, nil
}

// getWorkflowTenantId extracts the tenantId from a workflow's info.
func getWorkflowTenantId(c *client.Client, workflowId int) (string, error) {
	resp, err := c.Get(fmt.Sprintf("/gw/workflow/get?workflowId=%d", workflowId), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow info: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}
	if res, ok := result["result"].(map[string]interface{}); ok {
		if tid, ok := res["tenantId"].(string); ok && tid != "" {
			return tid, nil
		}
	}
	return "", fmt.Errorf("unable to determine tenantId from workflow info")
}

func newCmdConnectorInvoke(f *cmdutil.Factory) *cobra.Command {
	var connectorName, method, authAccount, env, bodyJson string
	var force bool

	cmd := &cobra.Command{
		Use:   "invoke",
		Short: "Invoke a connector method via universal workflow",
		Long: `Invoke a connector method by using a shared universal workflow.

The command will:
1. Check cache for an existing universal workflow (by env)
2. If not cached, create one generic workflow (Start → Script → End) with AppInvoker
3. Pass connector/method/authAccount/body as workflow input parameters
4. Invoke the workflow and return the result
5. Cache the workflow info for future calls (same env reuses the same workflow)

The methodName format is dot-separated, e.g. "erp.warehouse.list.query".
Use 'connector methods' to list available methods for a connector.

EXAMPLES:
    # Invoke warehouse query
    lingtong-cli connector invoke --connector kmerp --method "erp.warehouse.list.query" --auth-account "广州力人服饰" --env test --body '{}'

    # Invoke with parameters
    lingtong-cli connector invoke --connector kmerp --method "erp.trade.list.query" --auth-account "广州力人服饰" --env test --body '{"pageSize":10}'

    # Force recreate universal workflow (ignore cache)
    lingtong-cli connector invoke --connector kmerp --method "erp.warehouse.list.query" --auth-account "广州力人服饰" --env test --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connectorName == "" {
				return fmt.Errorf("--connector is required")
			}
			if method == "" {
				return fmt.Errorf("--method is required")
			}
			if authAccount == "" {
				return fmt.Errorf("--auth-account is required")
			}
			if env == "" {
				env = "prod"
			}

			var body map[string]interface{}
			if bodyJson != "" {
				if err := json.Unmarshal([]byte(bodyJson), &body); err != nil {
					return fmt.Errorf("invalid --body JSON: %w", err)
				}
			} else {
				body = make(map[string]interface{})
			}

			// If force, remove from cache
			if force {
				cache, err := loadCache()
				if err == nil {
					delete(cache.Universal, universalCacheKey(env))
					saveCache(cache)
				}
			}

			// Ensure universal workflow exists
			entry, err := ensureConnectorInvokeWorkflow(f, env)
			if err != nil {
				return err
			}

			// Invoke the workflow via Open API (direct mode, not proxy)
			// Use longer timeout since workflow execution may take time
			c := client.NewClientWithTimeout(f.Config.Host, f.Config.Token, 120*time.Second)
			c.DisableProxy()

			invokePath := fmt.Sprintf("/gw/%s/workflows/run", entry.AppTag)
			invokeBody := map[string]interface{}{
				"connector":   connectorName,
				"method":      method,
				"authAccount": authAccount,
				"body":        body,
				"env":         env,
			}

			resp, err := c.Post(invokePath, invokeBody)
			if err != nil {
				return fmt.Errorf("failed to invoke workflow: %w", err)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}

			// Add metadata and clean up redundant fields
			if result, ok := data.(map[string]interface{}); ok {
				// Remove redundant `body` field from result (it's a JSON string
				// duplicating the parsed fields like `list`, `success`, etc.)
				if innerResult, ok := result["result"].(map[string]interface{}); ok {
					delete(innerResult, "body")
				}
				result["_metadata"] = map[string]interface{}{
					"workflowId":  entry.WorkflowId,
					"appTag":      entry.AppTag,
					"connector":   connectorName,
					"method":      method,
					"authAccount": authAccount,
					"env":         env,
					"universal":   true,
				}
				return w.Write(result)
			}
			return w.Write(data)
		},
	}

	cmd.Flags().StringVar(&connectorName, "connector", "", "Connector name (required)")
	cmd.Flags().StringVar(&method, "method", "", "Connector method name, e.g. erp.warehouse.list.query (required)")
	cmd.Flags().StringVar(&authAccount, "auth-account", "", "Auth account name (required)")
	cmd.Flags().StringVar(&env, "env", "prod", "Environment (default: prod)")
	cmd.Flags().StringVar(&bodyJson, "body", "", "Request body as JSON (default: {})")
	cmd.Flags().BoolVar(&force, "force", false, "Force recreate universal workflow (ignore cache)")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("method")
	_ = cmd.MarkFlagRequired("auth-account")
	return cmd
}

// newCmdConnectorMethods lists available methods for a connector.
func newCmdConnectorMethods(f *cmdutil.Factory) *cobra.Command {
	var connectorName string
	var appId, authAccountId int

	cmd := &cobra.Command{
		Use:   "methods",
		Short: "List available methods for a connector",
		Long: `List all available methods for a connector with their metadata.

EXAMPLES:
    # List methods for kmerp connector
    lingtong-cli connector methods --connector kmerp --app-id 165 --auth-account-id 296`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connectorName == "" {
				return fmt.Errorf("--connector is required")
			}
			if appId == 0 {
				return fmt.Errorf("--app-id is required")
			}
			if authAccountId == 0 {
				return fmt.Errorf("--auth-account-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			params := map[string]interface{}{
				"appId":              appId,
				"connector":          connectorName,
				"authAccountId":      authAccountId,
				"interfaceModelType": "query",
			}

			resp, err := c.Get("/gw/model/info/category/interface", params)
			if err != nil {
				return fmt.Errorf("failed to list methods: %w", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return err
			}

			// Extract methods from categories
			methods := make([]map[string]interface{}, 0)
			if categories, ok := result["result"].([]interface{}); ok {
				for _, cat := range categories {
					if catMap, ok := cat.(map[string]interface{}); ok {
						if metaList, ok := catMap["modelMetaInfoList"].([]interface{}); ok {
							for _, meta := range metaList {
								if metaMap, ok := meta.(map[string]interface{}); ok {
									if method, ok := metaMap["method"].(string); ok && method != "" {
										methods = append(methods, map[string]interface{}{
											"method":   method,
											"title":    metaMap["title"],
											"business": metaMap["business"],
											"catId":    catMap["name"],
											"modelId":  metaMap["id"],
										})
									}
								}
							}
						}
					}
				}
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(map[string]interface{}{
				"count":   len(methods),
				"methods": methods,
			})
		},
	}

	cmd.Flags().StringVar(&connectorName, "connector", "", "Connector name (required)")
	cmd.Flags().IntVar(&appId, "app-id", 0, "Application ID (required)")
	cmd.Flags().IntVar(&authAccountId, "auth-account-id", 0, "Auth account ID (required)")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("app-id")
	_ = cmd.MarkFlagRequired("auth-account-id")
	return cmd
}

// newCmdConnectorSchema gets request/response field schema for a connector method.
func newCmdConnectorSchema(f *cmdutil.Factory) *cobra.Command {
	var connectorName, method string
	var authAccountId int

	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Get request/response field schema for a connector method",
		Long: `Get the request and response field schema for a connector method.

This command returns the field definitions (name, type, title) for both
request parameters and response fields of a connector method.

EXAMPLES:
    # Get schema for warehouse query method
    lingtong-cli connector schema --connector kmerp --method "erp.warehouse.list.query" --auth-account-id 296

    # Get schema and use it to construct invoke body
    lingtong-cli connector schema --connector kmerp --method "erp.warehouse.list.query" --auth-account-id 296 --format pretty`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connectorName == "" {
				return fmt.Errorf("--connector is required")
			}
			if method == "" {
				return fmt.Errorf("--method is required")
			}
			if authAccountId == 0 {
				return fmt.Errorf("--auth-account-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)

			// Step 1: Get modelId from method name
			params := map[string]interface{}{
				"appId":              165, // Default app ID
				"connector":          connectorName,
				"authAccountId":      authAccountId,
				"interfaceModelType": "query",
			}

			resp, err := c.Get("/gw/model/info/category/interface", params)
			if err != nil {
				return fmt.Errorf("failed to get model info: %w", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return err
			}

			// Find modelId for the method
			modelId := 0
			if categories, ok := result["result"].([]interface{}); ok {
				for _, cat := range categories {
					if catMap, ok := cat.(map[string]interface{}); ok {
						if metaList, ok := catMap["modelMetaInfoList"].([]interface{}); ok {
							for _, meta := range metaList {
								if metaMap, ok := meta.(map[string]interface{}); ok {
									if m, ok := metaMap["method"].(string); ok && m == method {
										if id, ok := metaMap["id"].(float64); ok {
											modelId = int(id)
										}
										break
									}
								}
							}
						}
					}
				}
			}

			if modelId == 0 {
				return fmt.Errorf("method '%s' not found for connector '%s'", method, connectorName)
			}

			// Step 2: Get request and response field schema
			schemaParams := map[string]interface{}{
				"connector":     connectorName,
				"authAccountId": authAccountId,
				"interfaceId":   modelId,
				"modelId":       modelId,
			}

			schemaResp, err := c.Get("/gw/meta/field/reqAndResp/list", schemaParams)
			if err != nil {
				return fmt.Errorf("failed to get field schema: %w", err)
			}

			var schemaResult map[string]interface{}
			if err := json.Unmarshal(schemaResp, &schemaResult); err != nil {
				return err
			}

			// Extract and format the result
			res := schemaResult["result"]
			if resMap, ok := res.(map[string]interface{}); ok {
				// Add metadata
				resMap["_metadata"] = map[string]interface{}{
					"connector": connectorName,
					"method":    method,
					"modelId":   modelId,
				}

				format := output.Format(cmd.Flag("format").Value.String())
				w := f.NewWriter(format)
				return w.Write(resMap)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(schemaResult)
		},
	}

	cmd.Flags().StringVar(&connectorName, "connector", "", "Connector name (required)")
	cmd.Flags().StringVar(&method, "method", "", "Method name (required)")
	cmd.Flags().IntVar(&authAccountId, "auth-account-id", 0, "Auth account ID (required)")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("method")
	_ = cmd.MarkFlagRequired("auth-account-id")
	return cmd
}

// newCmdConnectorCache manages the connector invoke cache.
func newCmdConnectorCache(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage connector invoke cache",
		Long:  "View or clear cached connector invoke workflows.",
	}
	cmd.AddCommand(newCmdConnectorCacheList(f))
	cmd.AddCommand(newCmdConnectorCacheClear(f))
	return cmd
}

func newCmdConnectorCacheList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cached connector invoke workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := loadCache()
			if err != nil {
				return err
			}
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)

			result := map[string]interface{}{}

			// Universal entries
			universalEntries := make([]map[string]interface{}, 0)
			for key, entry := range cache.Universal {
				universalEntries = append(universalEntries, map[string]interface{}{
					"key":        key,
					"workflowId": entry.WorkflowId,
					"appTag":     entry.AppTag,
					"env":        entry.Env,
					"type":       "universal",
				})
			}
			result["universal"] = universalEntries

			// Legacy entries (if any)
			if len(cache.Legacy) > 0 {
				legacyEntries := make([]map[string]interface{}, 0)
				for key, entry := range cache.Legacy {
					legacyEntries = append(legacyEntries, map[string]interface{}{
						"key":        key,
						"workflowId": entry.WorkflowId,
						"appTag":     entry.AppTag,
						"env":        entry.Env,
						"type":       "legacy",
					})
				}
				result["legacy"] = legacyEntries
			}

			return w.Write(result)
		},
	}
	return cmd
}

func newCmdConnectorCacheClear(f *cmdutil.Factory) *cobra.Command {
	var key string
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear cached connector invoke workflows",
		Long: `Clear all cached workflows or a specific one.

EXAMPLES:
    lingtong-cli connector cache clear
    lingtong-cli connector cache clear --key "test"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := loadCache()
			if err != nil {
				return err
			}
			if key != "" {
				if _, ok := cache.Universal[key]; !ok {
					return fmt.Errorf("cache key not found: %s", key)
				}
				delete(cache.Universal, key)
			} else {
				cache.Universal = make(map[string]*ConnectorCacheEntry)
				cache.Legacy = make(map[string]*ConnectorCacheEntry)
			}
			if err := saveCache(cache); err != nil {
				return err
			}
			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			msg := "All cached workflows cleared"
			if key != "" {
				msg = fmt.Sprintf("Cache entry '%s' cleared", key)
			}
			return w.Write(map[string]interface{}{"success": true, "message": msg})
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Specific cache key to clear (clears all if not specified)")
	return cmd
}
