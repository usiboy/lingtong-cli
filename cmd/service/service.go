// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package service provides OpenAPI-driven automatic command generation.
// It reads an OpenAPI specification and dynamically creates cobra commands
// for each API operation, achieving near-100% API coverage without manual
// command writing.
package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/openapi"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// RegisterServiceCommands registers all auto-generated service commands
// to the root command. Commands are grouped by path prefix (e.g., "scene",
// "basicdata", "factory").
func RegisterServiceCommands(root *cobra.Command, f *cmdutil.Factory, spec *openapi.Spec) {
	// Create the top-level "service" command
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Auto-generated API commands from OpenAPI spec",
		Long: `Dynamically generated commands covering all API endpoints.

These commands are auto-generated from the OpenAPI specification.
For high-frequency operations, prefer the dedicated commands
(e.g., 'scene list', 'connector info') which provide better UX.

EXAMPLES:
    lingtong-cli service scene list --pageNum 1 --pageSize 10
    lingtong-cli service basicdata save --body '{"name":"test"}'
    lingtong-cli service factory http-list --connector kmerp`,
	}
	serviceCmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")

	// Group operations by prefix
	groups := spec.GroupByPrefix()

	// Sort prefixes for deterministic output
	prefixes := make([]string, 0, len(groups))
	for prefix := range groups {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)

	// Build module commands
	for _, prefix := range prefixes {
		ops := groups[prefix]
		moduleCmd := buildModuleCommand(prefix, ops, f)
		serviceCmd.AddCommand(moduleCmd)
	}

	root.AddCommand(serviceCmd)
}

// buildModuleCommand creates a command group for a path prefix.
// e.g., "scene" -> service scene [list|get|create|...]
func buildModuleCommand(prefix string, ops []*openapi.Operation, f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   prefix,
		Short: fmt.Sprintf("API operations for /%s/*", prefix),
	}

	// Track command names to avoid duplicates
	seen := make(map[string]bool)

	for _, op := range ops {
		// Skip wildcard paths (e.g. /noco/datasource/app/{id}/**): they cannot
		// be represented as a fixed command and would send a literal path.
		if strings.Contains(op.Path, "*") {
			continue
		}

		cmdName := openapi.CommandName(op.Path, prefix)

		// Skip if we've already registered this command name
		if seen[cmdName] {
			// Append method suffix to disambiguate
			cmdName = cmdName + "-" + strings.ToLower(op.Method)
			if seen[cmdName] {
				continue
			}
		}
		seen[cmdName] = true

		opCmd := buildOperationCommand(op, cmdName, f)
		cmd.AddCommand(opCmd)
	}

	return cmd
}

// buildOperationCommand creates a cobra command for a single API operation.
func buildOperationCommand(op *openapi.Operation, cmdName string, f *cmdutil.Factory) *cobra.Command {
	// Build description
	desc := op.Summary
	if op.Description != "" && op.Description != op.Summary {
		desc += "\n\n" + op.Description
	}
	desc += fmt.Sprintf("\n\nAPI: %s %s", op.Method, op.Path)

	cmd := &cobra.Command{
		Use:   cmdName,
		Short: op.Summary,
		Long:  desc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeOperation(cmd, op, f)
		},
	}

	// Add flags from parameters
	addFlags(cmd, op)

	// Add format flag
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")

	return cmd
}

// addFlags adds cobra flags based on the operation's parameters.
func addFlags(cmd *cobra.Command, op *openapi.Operation) {
	for _, param := range op.Parameters {
		switch param.In {
		case "query":
			addQueryFlag(cmd, param)
		case "path":
			addPathFlag(cmd, param)
		case "body":
			addBodyFlag(cmd, param)
		}
	}
}

// addPathFlag adds a (required) flag for a path parameter such as
// /meta/model/{connector}/get, whose value is substituted into the path.
func addPathFlag(cmd *cobra.Command, param openapi.Parameter) {
	desc := param.Description
	if desc == "" {
		desc = "path parameter " + param.Name
	}
	if cmd.Flags().Lookup(param.Name) == nil {
		cmd.Flags().String(param.Name, "", desc)
	}
	// Path parameters are always required to form a valid URL.
	_ = cmd.MarkFlagRequired(param.Name)
}

// addQueryFlag adds a flag for a query parameter.
func addQueryFlag(cmd *cobra.Command, param openapi.Parameter) {
	name := param.Name
	desc := param.Description
	if desc == "" {
		desc = param.Name
	}

	switch param.Type {
	case "integer", "int32", "int64":
		cmd.Flags().Int(name, 0, desc)
	case "boolean":
		cmd.Flags().Bool(name, false, desc)
	case "array":
		cmd.Flags().StringSlice(name, nil, desc+" (comma-separated)")
	default:
		// string and others
		cmd.Flags().String(name, "", desc)
	}

	if param.Required {
		_ = cmd.MarkFlagRequired(name)
	}
}

// addBodyFlag adds a flag for a body parameter.
// For simple body params, we add individual flags.
// For complex body params, we add a --body JSON flag.
func addBodyFlag(cmd *cobra.Command, param openapi.Parameter) {
	// If the schema is a simple type, add a direct flag
	if param.Schema != nil && param.Schema.Type == "object" && len(param.Schema.Properties) > 0 {
		// Add flags for each property
		for propName, propSchema := range param.Schema.Properties {
			// Skip if flag already exists
			if cmd.Flags().Lookup(propName) != nil {
				continue
			}
			desc := fmt.Sprintf("Body field: %s", propName)
			switch propSchema.Type {
			case "integer", "int32", "int64":
				cmd.Flags().Int(propName, 0, desc)
			case "boolean":
				cmd.Flags().Bool(propName, false, desc)
			case "array":
				cmd.Flags().String(propName, "", desc+" (JSON array)")
			default:
				cmd.Flags().String(propName, "", desc)
			}
		}
	}

	// Always add a --body flag as a fallback for complex cases
	// But only if it doesn't already exist
	if cmd.Flags().Lookup("body") == nil {
		cmd.Flags().String("body", "", "Request body as JSON (overrides individual field flags)")
	}
}

// executeOperation executes an API operation.
func executeOperation(cmd *cobra.Command, op *openapi.Operation, f *cmdutil.Factory) error {
	c := client.NewClient(f.Config.Host, f.Config.Token)

	// Build path with query parameters
	path := op.Path
	params := make(map[string]interface{})

	// Substitute path parameters (e.g. {connector}) into the path.
	for _, param := range op.PathParams() {
		val, _ := cmd.Flags().GetString(param.Name)
		path = strings.ReplaceAll(path, "{"+param.Name+"}", url.PathEscape(val))
	}

	for _, param := range op.Parameters {
		if param.In == "query" {
			flag := cmd.Flags().Lookup(param.Name)
			if flag != nil && flag.Changed {
				switch param.Type {
				case "integer", "int32", "int64":
					val, _ := cmd.Flags().GetInt(param.Name)
					params[param.Name] = val
				case "boolean":
					val, _ := cmd.Flags().GetBool(param.Name)
					params[param.Name] = val
				case "array":
					val, _ := cmd.Flags().GetStringSlice(param.Name)
					params[param.Name] = val
				default:
					val, _ := cmd.Flags().GetString(param.Name)
					params[param.Name] = val
				}
			}
		}
	}

	// Build request body
	var body interface{}
	bodyFlag := cmd.Flags().Lookup("body")
	if bodyFlag != nil && bodyFlag.Changed {
		bodyStr, _ := cmd.Flags().GetString("body")
		if err := json.Unmarshal([]byte(bodyStr), &body); err != nil {
			return fmt.Errorf("invalid --body JSON: %w", err)
		}
	} else {
		// Try to build body from individual field flags
		bodyParams := op.BodyParams()
		if len(bodyParams) > 0 {
			bodyMap := make(map[string]interface{})
			for _, param := range bodyParams {
				if param.Schema != nil && param.Schema.Type == "object" {
					for propName, propSchema := range param.Schema.Properties {
						flag := cmd.Flags().Lookup(propName)
						if flag != nil && flag.Changed {
							switch propSchema.Type {
							case "integer", "int32", "int64":
								val, _ := cmd.Flags().GetInt(propName)
								bodyMap[propName] = val
							case "boolean":
								val, _ := cmd.Flags().GetBool(propName)
								bodyMap[propName] = val
							default:
								val, _ := cmd.Flags().GetString(propName)
								// Try to parse as JSON first
								var jsonVal interface{}
								if err := json.Unmarshal([]byte(val), &jsonVal); err == nil {
									bodyMap[propName] = jsonVal
								} else {
									bodyMap[propName] = val
								}
							}
						}
					}
				}
			}
			if len(bodyMap) > 0 {
				body = bodyMap
			}
		}
	}

	// Execute the request
	var resp []byte
	var err error

	switch op.Method {
	case "GET":
		resp, err = c.Get(path, params)
	case "POST":
		if len(params) > 0 {
			resp, err = c.Do("POST", path, params, body)
		} else {
			resp, err = c.Post(path, body)
		}
	case "PUT":
		resp, err = c.Put(path, body)
	case "DELETE":
		resp, err = c.Delete(path)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", op.Method)
	}

	if err != nil {
		return err
	}

	// Parse and output the response
	format := output.Format(cmd.Flag("format").Value.String())
	w := f.NewWriter(format)
	var data interface{}
	if err := json.Unmarshal(resp, &data); err != nil {
		return err
	}
	return w.Write(data)
}
