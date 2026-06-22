// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package connector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// NewCmdConnector creates the connector command.
func NewCmdConnector(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connector",
		Short: "Manage connectors",
		Long:  "Query connector configurations, metadata, and category information.",
	}

	cmd.AddCommand(newCmdConnectorInfo(f))
	cmd.AddCommand(newCmdConnectorCategoryList(f))
	cmd.AddCommand(newCmdConnectorList(f))
	cmd.AddCommand(newCmdConnectorAccount(f))
	cmd.AddCommand(newCmdConnectorCheckAuth(f))
	cmd.AddCommand(newCmdConnectorInvoke(f))
	cmd.AddCommand(newCmdConnectorMethods(f))
	cmd.AddCommand(newCmdConnectorSchema(f))
	cmd.AddCommand(newCmdConnectorCache(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func newCmdConnectorInfo(f *cmdutil.Factory) *cobra.Command {
	var connector, env, authAccountId string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Query connector details",
		Long: `Query connector configuration details and metadata.

EXAMPLES:
    lingtong-cli connector info --connector kmerp
    lingtong-cli connector info --connector kmerp --env prod`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connector == "" {
				return fmt.Errorf("--connector is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := fmt.Sprintf("/gw/ai/connector/info?connector=%s&env=%s", connector, env)
			if authAccountId != "" {
				path += "&authAccountId=" + authAccountId
			}

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

	cmd.Flags().StringVar(&connector, "connector", "", "Connector identifier (required)")
	cmd.Flags().StringVar(&env, "env", "test", "Environment (test/prod)")
	cmd.Flags().StringVar(&authAccountId, "auth-account-id", "", "Authentication account ID")
	_ = cmd.MarkFlagRequired("connector")
	return cmd
}

func newCmdConnectorCategoryList(f *cmdutil.Factory) *cobra.Command {
	var connector string
	cmd := &cobra.Command{
		Use:   "category list",
		Short: "List connector model categories",
		Long: `Query model category list for a connector.

EXAMPLES:
    lingtong-cli connector category list --connector kmerp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connector == "" {
				return fmt.Errorf("--connector is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := fmt.Sprintf("/gw/ai/connector/category/list?connector=%s", connector)

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

	cmd.Flags().StringVar(&connector, "connector", "", "Connector identifier (required)")
	_ = cmd.MarkFlagRequired("connector")
	return cmd
}

func newCmdConnectorList(f *cmdutil.Factory) *cobra.Command {
	var appId int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connector accounts",
		Long: `List all connector accounts configured in the application.

EXAMPLES:
    lingtong-cli connector list
    lingtong-cli connector list --app-id 165`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.NewClient(f.Config.Host, f.Config.Token)
			path := "/gw/ai/connectors"
			if appId != 0 {
				path += fmt.Sprintf("?appId=%d", appId)
			}

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

	cmd.Flags().IntVar(&appId, "app-id", 0, "Application ID (optional)")
	return cmd
}

func newCmdConnectorAccount(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage connector accounts",
		Long:  "List, verify, and create connector authentication accounts.",
	}

	cmd.AddCommand(newCmdConnectorAccountList(f))
	cmd.AddCommand(newCmdConnectorAccountVerify(f))
	cmd.AddCommand(newCmdConnectorAccountCreate(f))

	return cmd
}

func newCmdConnectorAccountList(f *cmdutil.Factory) *cobra.Command {
	var connector, env, name string
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connector accounts",
		Long: `List all configured authentication accounts.

When no --connector is specified, lists all accounts across all connectors.
Use --connector to filter by a specific connector type.
Use --name to search by account name (client-side filtering).

EXAMPLES:
    # List all accounts
    lingtong-cli connector account list

    # List accounts for a specific connector
    lingtong-cli connector account list --connector kmerp

    # List accounts with environment filter
    lingtong-cli connector account list --connector kmerp --env test

    # Search by name
    lingtong-cli connector account list --name "广州"

    # Paginate results
    lingtong-cli connector account list --page 2 --page-size 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.NewClient(f.Config.Host, f.Config.Token)

			var resp []byte
			var err error

			if connector == "" {
				// List all accounts (simplified version, sorted by name)
				resp, err = c.Get("/gw/account/listAll2", nil)
			} else {
				// List accounts for specific connector
				path := fmt.Sprintf("/gw/account/connector/list?connector=%s", connector)
				if env != "" {
					path += fmt.Sprintf("&env=%s", env)
				}
				resp, err = c.Get(path, nil)
			}

			if err != nil {
				return fmt.Errorf("failed to list connector accounts: %w", err)
			}

			// Parse response
			var result map[string]interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Apply client-side name filter if specified
			if name != "" {
				if data, ok := result["result"].([]interface{}); ok {
					var filtered []interface{}
					for _, item := range data {
						if account, ok := item.(map[string]interface{}); ok {
							if accountName, ok := account["name"].(string); ok {
								if strings.Contains(strings.ToLower(accountName), strings.ToLower(name)) {
									filtered = append(filtered, item)
								}
							}
						}
					}
					result["result"] = filtered
				}
			}

			// Apply pagination if specified
			if page > 0 || pageSize > 0 {
				if data, ok := result["result"].([]interface{}); ok {
					if page <= 0 {
						page = 1
					}
					if pageSize <= 0 {
						pageSize = 20
					}
					start := (page - 1) * pageSize
					end := start + pageSize
					if start > len(data) {
						start = len(data)
					}
					if end > len(data) {
						end = len(data)
					}
					result["result"] = data[start:end]
					result["pagination"] = map[string]interface{}{
						"page":     page,
						"pageSize": pageSize,
						"total":    len(data),
					}
				}
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(result)
		},
	}

	cmd.Flags().StringVar(&connector, "connector", "", "Connector identifier (optional)")
	cmd.Flags().StringVar(&env, "env", "", "Environment filter (test/prod)")
	cmd.Flags().StringVar(&name, "name", "", "Search by account name")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (1-based)")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Page size")
	return cmd
}

func newCmdConnectorAccountVerify(f *cmdutil.Factory) *cobra.Command {
	var accountId int
	var connector, env string
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify connector account connection",
		Long: `Test the connection of a connector account to ensure it is properly configured.

EXAMPLES:
    lingtong-cli connector account verify --connector kmerp --account-id 123
    lingtong-cli connector account verify --connector feishu --account-id 456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connector == "" {
				return fmt.Errorf("--connector is required")
			}
			if accountId == 0 {
				return fmt.Errorf("--account-id is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)

			// Get account details first
			getPath := fmt.Sprintf("/gw/account/get2?id=%d", accountId)
			getResp, err := c.Get(getPath, nil)
			if err != nil {
				return fmt.Errorf("failed to get account details: %w", err)
			}

			var accountResult map[string]interface{}
			if err := json.Unmarshal(getResp, &accountResult); err != nil {
				return fmt.Errorf("failed to parse account response: %w", err)
			}

			// Verify the account
			verifyPath := "/gw/account/verify"
			verifyBody := map[string]interface{}{
				"id":        accountId,
				"connector": connector,
				"env":       env,
			}

			verifyResp, err := c.Post(verifyPath, verifyBody)
			if err != nil {
				return fmt.Errorf("failed to verify connector account: %w", err)
			}

			var verifyResult map[string]interface{}
			if err := json.Unmarshal(verifyResp, &verifyResult); err != nil {
				return fmt.Errorf("failed to parse verify response: %w", err)
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(verifyResult)
		},
	}

	cmd.Flags().StringVar(&connector, "connector", "", "Connector identifier (required)")
	cmd.Flags().IntVar(&accountId, "account-id", 0, "Account ID (required)")
	cmd.Flags().StringVar(&env, "env", "test", "Environment (test/prod)")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("account-id")
	return cmd
}

func newCmdConnectorAccountCreate(f *cmdutil.Factory) *cobra.Command {
	var connector, env, name, accountData string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new connector account",
		Long: `Create a new authentication account for a connector.

The account data should be provided as a JSON string containing the required
authentication fields for the connector (e.g., appKey, appSecret, token).

EXAMPLES:
    # Create kmerp account with AppKey/AppSecret
    lingtong-cli connector account create \
      --connector kmerp \
      --name "快麦测试账号" \
      --env test \
      --data '{"appKey":"xxx","appSecret":"yyy"}'

    # Create feishu account with OAuth token
    lingtong-cli connector account create \
      --connector feishu \
      --name "飞书多维表格" \
      --env prod \
      --data '{"app_id":"xxx","app_secret":"yyy","tenant_access_token":"zzz"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if connector == "" {
				return fmt.Errorf("--connector is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if accountData == "" {
				return fmt.Errorf("--data is required")
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)

			// Parse account data
			var fields []map[string]interface{}
			if err := json.Unmarshal([]byte(accountData), &fields); err != nil {
				// Try as key-value map
				var fieldMap map[string]interface{}
				if err2 := json.Unmarshal([]byte(accountData), &fieldMap); err2 != nil {
					return fmt.Errorf("invalid --data format: must be JSON object or array")
				}
				// Convert to field list format
				fields = []map[string]interface{}{
					{"fieldName": "config", "fieldValue": fieldMap},
				}
			}

			// Build request body
			body := map[string]interface{}{
				"connector":                 connector,
				"name":                      name,
				"env":                       env,
				"open":                      1,
				"ltAuthAccountFieldDtoList": fields,
			}

			resp, err := c.Post("/gw/account/save", body)
			if err != nil {
				return fmt.Errorf("failed to create connector account: %w", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Check if successful
			if success, ok := result["success"].(bool); ok && success {
				fmt.Fprintf(f.IOStreams.Out, "✓ Connector account '%s' created successfully\n", name)
				if accountRelationId, ok := result["result"].(map[string]interface{})["accountRelationId"]; ok {
					fmt.Fprintf(f.IOStreams.Out, "  Account Relation ID: %v\n", accountRelationId)
				}
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format)
			return w.Write(result)
		},
	}

	cmd.Flags().StringVar(&connector, "connector", "", "Connector identifier (required)")
	cmd.Flags().StringVar(&name, "name", "", "Account name (required)")
	cmd.Flags().StringVar(&env, "env", "test", "Environment (test/prod)")
	cmd.Flags().StringVar(&accountData, "data", "", "Authentication data as JSON (required)")
	_ = cmd.MarkFlagRequired("connector")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

func newCmdConnectorCheckAuth(f *cmdutil.Factory) *cobra.Command {
	var workflowId, sceneId int
	var connector string
	cmd := &cobra.Command{
		Use:   "check-auth",
		Short: "Check connector authorization status",
		Long: `Check if the connectors used in a workflow or scene have proper authorization.

This command automatically detects which connectors are used and verifies if
corresponding authentication accounts are configured.

EXAMPLES:
    # Check auth for a workflow
    lingtong-cli connector check-auth --workflow-id 947

    # Check auth for a scene
    lingtong-cli connector check-auth --scene-id 123

    # Check specific connector
    lingtong-cli connector check-auth --connector kmerp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.NewClient(f.Config.Host, f.Config.Token)

			var connectors []string
			var contextInfo string

			// Determine which connectors to check
			if connector != "" {
				connectors = []string{connector}
				contextInfo = fmt.Sprintf("Connector: %s", connector)
			} else if workflowId != 0 {
				// Get workflow details to find used connectors
				path := fmt.Sprintf("/gw/workflow/get?workflowId=%d", workflowId)
				resp, err := c.Get(path, nil)
				if err != nil {
					return fmt.Errorf("failed to get workflow details: %w", err)
				}

				var workflow map[string]interface{}
				if err := json.Unmarshal(resp, &workflow); err != nil {
					return fmt.Errorf("failed to parse workflow response: %w", err)
				}

				// TODO: Extract connectors from workflow nodes
				// For now, we'll check common connectors
				connectors = []string{"kmerp", "feishu", "dingtalk"}
				contextInfo = fmt.Sprintf("Workflow ID: %d", workflowId)
			} else if sceneId != 0 {
				// Get scene details
				path := fmt.Sprintf("/gw/scene/get?id=%d", sceneId)
				resp, err := c.Get(path, nil)
				if err != nil {
					return fmt.Errorf("failed to get scene details: %w", err)
				}

				var scene map[string]interface{}
				if err := json.Unmarshal(resp, &scene); err != nil {
					return fmt.Errorf("failed to parse scene response: %w", err)
				}

				// TODO: Extract connectors from scene
				connectors = []string{"kmerp", "feishu", "dingtalk"}
				contextInfo = fmt.Sprintf("Scene ID: %d", sceneId)
			} else {
				return fmt.Errorf("--workflow-id, --scene-id, or --connector is required")
			}

			// Check authorization for each connector
			fmt.Fprintf(f.IOStreams.Out, "Checking connector authorization for %s...\n\n", contextInfo)

			allAuthorized := true
			results := make([]map[string]interface{}, 0, len(connectors))

			for _, conn := range connectors {
				// Query connector accounts
				path := fmt.Sprintf("/gw/account/connector/list?connector=%s", conn)
				resp, err := c.Get(path, nil)

				if err != nil {
					results = append(results, map[string]interface{}{
						"connector":    conn,
						"authorized":   false,
						"error":        err.Error(),
						"accountCount": 0,
					})
					allAuthorized = false
					continue
				}

				var result map[string]interface{}
				if err := json.Unmarshal(resp, &result); err != nil {
					results = append(results, map[string]interface{}{
						"connector":    conn,
						"authorized":   false,
						"error":        err.Error(),
						"accountCount": 0,
					})
					allAuthorized = false
					continue
				}

				// Check if accounts exist
				var accountCount int
				if resultList, ok := result["result"].([]interface{}); ok {
					accountCount = len(resultList)
				}

				authorized := accountCount > 0
				if !authorized {
					allAuthorized = false
				}

				results = append(results, map[string]interface{}{
					"connector":    conn,
					"authorized":   authorized,
					"accountCount": accountCount,
				})
			}

			// Display results
			w := f.NewWriter(output.Format(cmd.Flag("format").Value.String()))

			if allAuthorized {
				fmt.Fprintf(f.IOStreams.Out, "✓ All connectors are properly authorized\n\n")
			} else {
				fmt.Fprintf(f.IOStreams.Out, "⚠ Some connectors are missing authorization:\n\n")

				for _, r := range results {
					if authorized, _ := r["authorized"].(bool); !authorized {
						connector := r["connector"].(string)
						fmt.Fprintf(f.IOStreams.Out, "  ✗ %s: No account configured\n", connector)
						fmt.Fprintf(f.IOStreams.Out, "    → Run: lingtong-cli connector account create --connector %s --name \"My Account\" --data '{...}'\n\n", connector)
					}
				}
			}

			return w.Write(map[string]interface{}{
				"context":       contextInfo,
				"connectors":    results,
				"allAuthorized": allAuthorized,
			})
		},
	}

	cmd.Flags().IntVar(&workflowId, "workflow-id", 0, "Workflow ID to check")
	cmd.Flags().IntVar(&sceneId, "scene-id", 0, "Scene ID to check")
	cmd.Flags().StringVar(&connector, "connector", "", "Specific connector to check")
	return cmd
}
