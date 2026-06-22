// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package schema provides the `schema` command for browsing API documentation
// from the embedded OpenAPI specification. It helps AI Agents and humans
// discover available API endpoints, their parameters, and types.
package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/openapi"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// NewCmdSchema creates the schema command for browsing API documentation.
func NewCmdSchema(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema [command] [args]",
		Short: "Browse API documentation from OpenAPI spec",
		Long: `Browse the embedded OpenAPI specification to discover API endpoints,
their parameters, types, and descriptions.

EXAMPLES:
    # List all API paths (grouped by prefix)
    lingtong-cli schema list

    # Show detailed schema for a specific path
    lingtong-cli schema path /scene/list

    # Show all operations for a module
    lingtong-cli schema module scene

    # Search operations by keyword
    lingtong-cli schema search "场景"`,
	}

	cmd.AddCommand(newCmdSchemaList(f))
	cmd.AddCommand(newCmdSchemaPath(f))
	cmd.AddCommand(newCmdSchemaModule(f))
	cmd.AddCommand(newCmdSchemaSearch(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

// loadSpec loads the OpenAPI spec for schema commands using the shared resolver
// (env override → ~/.lingtong-cli/openapi.json → embedded), so `schema`,
// `service`, and `doctor` always describe the same API.
func loadSpec() (*openapi.Spec, error) {
	spec, err := openapi.LoadSpec()
	if err != nil {
		return nil, err
	}
	if spec == nil {
		return nil, fmt.Errorf("no OpenAPI spec available (embedded spec not found)")
	}
	return spec, nil
}

// ==================== schema list ====================

func newCmdSchemaList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all API paths grouped by module",
		Long: `List all available API paths from the OpenAPI spec, grouped by their
first-level path prefix (module).

EXAMPLES:
    lingtong-cli schema list
    lingtong-cli schema list --format pretty`,
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := loadSpec()
			if err != nil {
				return err
			}

			groups := spec.GroupByPrefix()

			// Build summary per module
			modules := make([]map[string]interface{}, 0, len(groups))
			for prefix, ops := range groups {
				modules = append(modules, map[string]interface{}{
					"module":         prefix,
					"operationCount": len(ops),
					"paths":          uniquePaths(ops),
				})
			}

			// Sort by module name for deterministic output
			sort.Slice(modules, func(i, j int) bool {
				return modules[i]["module"].(string) < modules[j]["module"].(string)
			})

			result := map[string]interface{}{
				"totalPaths":      len(spec.Paths),
				"totalOperations": countOperations(spec),
				"modules":         modules,
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "schema.list")
			return w.Write(result)
		},
	}
	return cmd
}

// ==================== schema path ====================

func newCmdSchemaPath(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path <api-path>",
		Short: "Show detailed schema for a specific API path",
		Long: `Show the full schema for a specific API path, including all operations
(GET/POST/PUT/DELETE), their parameters, types, and descriptions.

EXAMPLES:
    lingtong-cli schema path /scene/list
    lingtong-cli schema path /basicdata/save`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := loadSpec()
			if err != nil {
				return err
			}

			apiPath := args[0]
			// Normalize: ensure leading slash
			if !strings.HasPrefix(apiPath, "/") {
				apiPath = "/" + apiPath
			}

			pathItem, ok := spec.Paths[apiPath]
			if !ok {
				// Try fuzzy match
				matches := fuzzyMatchPaths(spec, apiPath)
				if len(matches) > 0 {
					return fmt.Errorf("path %q not found. Did you mean: %s",
						apiPath, strings.Join(matches, ", "))
				}
				return fmt.Errorf("path %q not found in OpenAPI spec", apiPath)
			}

			result := buildPathDetail(apiPath, &pathItem)

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "schema.path")
			return w.Write(result)
		},
	}
	return cmd
}

// ==================== schema module ====================

func newCmdSchemaModule(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module <module-name>",
		Short: "Show all operations for a module",
		Long: `Show all API operations belonging to a specific module (path prefix).

EXAMPLES:
    lingtong-cli schema module scene
    lingtong-cli schema module basicdata`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := loadSpec()
			if err != nil {
				return err
			}

			moduleName := args[0]
			groups := spec.GroupByPrefix()

			ops, ok := groups[moduleName]
			if !ok {
				// List available modules
				available := make([]string, 0, len(groups))
				for k := range groups {
					available = append(available, k)
				}
				sort.Strings(available)
				return fmt.Errorf("module %q not found. Available modules: %s",
					moduleName, strings.Join(available, ", "))
			}

			operations := make([]map[string]interface{}, 0, len(ops))
			for _, op := range ops {
				operations = append(operations, map[string]interface{}{
					"path":           op.Path,
					"method":         op.Method,
					"summary":        op.Summary,
					"paramCount":     len(op.Parameters),
					"requiredParams": countRequired(op.Parameters),
				})
			}

			result := map[string]interface{}{
				"module":     moduleName,
				"operations": operations,
				"totalCount": len(operations),
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "schema.module")
			return w.Write(result)
		},
	}
	return cmd
}

// ==================== schema search ====================

func newCmdSchemaSearch(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <keyword>",
		Short: "Search API operations by keyword",
		Long: `Search API operations by keyword in summary, description, or path.

EXAMPLES:
    lingtong-cli schema search "场景"
    lingtong-cli schema search "workflow"
    lingtong-cli schema search "list"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := loadSpec()
			if err != nil {
				return err
			}

			keyword := strings.ToLower(args[0])
			results := searchOperations(spec, keyword)

			result := map[string]interface{}{
				"keyword":    args[0],
				"matches":    results,
				"matchCount": len(results),
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "schema.search")
			return w.Write(result)
		},
	}
	return cmd
}

// ==================== helpers ====================

func uniquePaths(ops []*openapi.Operation) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, op := range ops {
		if !seen[op.Path] {
			seen[op.Path] = true
			paths = append(paths, op.Path)
		}
	}
	sort.Strings(paths)
	return paths
}

func countOperations(spec *openapi.Spec) int {
	count := 0
	for _, item := range spec.Paths {
		count += len(item.Operations())
	}
	return count
}

func countRequired(params []openapi.Parameter) int {
	n := 0
	for _, p := range params {
		if p.Required {
			n++
		}
	}
	return n
}

func buildPathDetail(path string, item *openapi.PathItem) map[string]interface{} {
	ops := item.Operations()
	operations := make([]map[string]interface{}, 0, len(ops))

	for _, op := range ops {
		params := make([]map[string]interface{}, 0, len(op.Parameters))
		for _, p := range op.Parameters {
			param := map[string]interface{}{
				"name":     p.Name,
				"in":       p.In,
				"type":     p.Type,
				"required": p.Required,
			}
			if p.Description != "" {
				param["description"] = p.Description
			}
			if len(p.Enum) > 0 {
				param["enum"] = p.Enum
			}
			params = append(params, param)
		}

		opDetail := map[string]interface{}{
			"method":     op.Method,
			"summary":    op.Summary,
			"parameters": params,
		}
		if op.Description != "" && op.Description != op.Summary {
			opDetail["description"] = op.Description
		}
		if len(op.Tags) > 0 {
			opDetail["tags"] = op.Tags
		}
		operations = append(operations, opDetail)
	}

	return map[string]interface{}{
		"path":       path,
		"operations": operations,
	}
}

func fuzzyMatchPaths(spec *openapi.Spec, query string) []string {
	var matches []string
	query = strings.ToLower(query)
	for path := range spec.Paths {
		if strings.Contains(strings.ToLower(path), query) {
			matches = append(matches, path)
		}
	}
	sort.Strings(matches)
	if len(matches) > 5 {
		matches = matches[:5]
	}
	return matches
}

func searchOperations(spec *openapi.Spec, keyword string) []map[string]interface{} {
	var results []map[string]interface{}

	for path, item := range spec.Paths {
		for _, op := range item.Operations() {
			summary := strings.ToLower(op.Summary)
			desc := strings.ToLower(op.Description)
			pathLower := strings.ToLower(path)

			if strings.Contains(summary, keyword) ||
				strings.Contains(desc, keyword) ||
				strings.Contains(pathLower, keyword) {

				result := map[string]interface{}{
					"path":    path,
					"method":  op.Method,
					"summary": op.Summary,
				}
				if op.Description != "" && op.Description != op.Summary {
					result["description"] = op.Description
				}
				results = append(results, result)
			}
		}
	}

	// Sort by path for deterministic output
	sort.Slice(results, func(i, j int) bool {
		pi, _ := results[i]["path"].(string)
		pj, _ := results[j]["path"].(string)
		return pi < pj
	})

	return results
}
