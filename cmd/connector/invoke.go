// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package connector

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func newCmdConnectorInvoke(f *cmdutil.Factory) *cobra.Command {
	var connectorName, method, authAccount, env, bodyJson string
	var force bool

	cmd := &cobra.Command{
		Use:   "invoke",
		Short: "Invoke a connector method through the API proxy",
		Long: `Invoke a connector method through the API proxy without creating a workflow.

The method name is dot-separated, e.g. "erp.warehouse.list.query".
Use 'connector methods' to list available methods for a connector.
Built-in connectors may omit --auth-account; the server validates authorization.
The response preserves result.receiptStatus and result.responseBody.

EXAMPLES:
    lingtong-cli connector invoke --connector kmerp --method "erp.warehouse.list.query" --auth-account "广州力人服饰" --env test --body '{}'
    lingtong-cli connector invoke --connector kmerp --method "erp.trade.list.query" --auth-account "广州力人服饰" --env test --body '{"pageSize":100}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connectorName == "" {
				return fmt.Errorf("--connector is required")
			}
			if method == "" {
				return fmt.Errorf("--method is required")
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
			if body == nil {
				return fmt.Errorf("invalid --body JSON: must be a non-null object")
			}

			if force {
				fmt.Fprintln(f.IOStreams.ErrOut, "Warning: --force is deprecated and ignored; connector invoke no longer uses workflows or cache.")
			}

			c := client.NewClientWithTimeout(f.Config.Host, f.Config.Token, 120*time.Second)
			invokeBody := map[string]interface{}{
				"connector":   connectorName,
				"method":      method,
				"authName":    authAccount,
				"requestBody": body,
				"env":         env,
			}
			resp, err := c.Post("/gw/connector/invoke", invokeBody)
			if err != nil {
				return fmt.Errorf("failed to invoke connector: %w", err)
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

	cmd.Flags().StringVar(&connectorName, "connector", "", "Connector name (required)")
	cmd.Flags().StringVar(&method, "method", "", "Connector method name, e.g. erp.warehouse.list.query (required)")
	cmd.Flags().StringVar(&authAccount, "auth-account", "", "Auth account name (required for non-built-in connectors)")
	cmd.Flags().StringVar(&env, "env", "prod", "Environment (default: prod)")
	cmd.Flags().StringVar(&bodyJson, "body", "", "Request body as JSON (default: {})")
	cmd.Flags().BoolVar(&force, "force", false, "Deprecated: ignored; invocation no longer uses workflows or cache")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("method")
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
		Short: "Manage legacy connector workflow cache",
		Long:  "View or clear legacy local workflow cache. Connector invoke no longer uses it; platform workflows are not deleted.",
	}
	cmd.AddCommand(newCmdConnectorCacheList(f))
	cmd.AddCommand(newCmdConnectorCacheClear(f))
	return cmd
}

func newCmdConnectorCacheList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List legacy cached connector workflows",
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
		Short: "Clear legacy local connector workflow cache",
		Long: `Clear all legacy local workflow cache entries or a specific one. Platform workflows are not deleted.

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
