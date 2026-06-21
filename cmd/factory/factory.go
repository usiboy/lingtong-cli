// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package factory provides hand-written commands for the connector factory
// (/factory/*): connector definitions, their HTTP interfaces ("methods"), and
// scripts. Destructive operations are confirmation-gated.
package factory

import (
	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdFactory creates the factory command group.
func NewCmdFactory(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "factory",
		Short: "Manage connector factories",
		Long: `Connector factory operations (/factory/*).

A factory defines a custom connector: its HTTP interfaces and scripts. Use the
sub-groups 'http' (interfaces) and 'script' to manage those.`,
	}

	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdGet(f))
	cmd.AddCommand(newCmdSave(f))
	cmd.AddCommand(newCmdDelete(f))
	cmd.AddCommand(newCmdHTTP(f))
	cmd.AddCommand(newCmdScript(f))

	cmd.PersistentFlags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

// ==================== factory list/get/save/delete ====================

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connector factories",
		Long: `List connector factories (paginated). Wraps GET /factory/list.

EXAMPLES:
    lingtong-cli factory list
    lingtong-cli factory list --page 2 --page-size 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/factory/list", map[string]interface{}{
				"pageNum":  page,
				"pageSize": pageSize,
			})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.list")
		},
	}
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	return cmd
}

func newCmdGet(f *cmdutil.Factory) *cobra.Command {
	var id int
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get connector factory details",
		Long: `Show details for one connector factory. Wraps GET /factory/get.

EXAMPLES:
    lingtong-cli factory get --id 12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/factory/get", map[string]interface{}{"id": id})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.get")
		},
	}
	cmd.Flags().IntVar(&id, "id", 0, "Factory ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newCmdSave(f *cmdutil.Factory) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Create or update a connector factory",
		Long: `Create or update a connector factory from a JSON definition.

Wraps POST /factory/save. Include an "id" field in --data to update an existing
factory; omit it to create a new one.

EXAMPLES:
    lingtong-cli factory save --data '{"name":"my-connector","connector":"mycon"}'
    lingtong-cli factory save --data '{"id":12,"name":"renamed"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(data != "", "data"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			body, err := cmdutil.ParseDataObject(data)
			if err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/factory/save", body)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.save")
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "Factory definition (JSON object, required)")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var id int
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a connector factory (requires --yes)",
		Long: `Delete a connector factory by ID. Wraps POST /factory/delete.

This is destructive and requires --yes.

EXAMPLES:
    lingtong-cli factory delete --id 12 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := cmdutil.ConfirmDestructive(yes, "delete factory"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/factory/delete", id)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.delete")
		},
	}
	cmd.Flags().IntVar(&id, "id", 0, "Factory ID (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm the destructive operation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// ==================== factory http (interfaces / "methods") ====================

func newCmdHTTP(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Manage factory HTTP interfaces",
		Long:  "HTTP interface (method) operations for a connector factory.",
	}
	cmd.AddCommand(newCmdHTTPList(f))
	cmd.AddCommand(newCmdHTTPGet(f))
	cmd.AddCommand(newCmdHTTPRun(f))
	cmd.AddCommand(newCmdHTTPDelete(f))
	return cmd
}

func newCmdHTTPList(f *cmdutil.Factory) *cobra.Command {
	var page, pageSize, factoryID int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HTTP interfaces",
		Long: `List a factory's HTTP interfaces (paginated). Wraps GET /factory/http/list.

EXAMPLES:
    lingtong-cli factory http list --factory-id 12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			params := map[string]interface{}{"pageNum": page, "pageSize": pageSize}
			if factoryID != 0 {
				params["factoryId"] = factoryID
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/factory/http/list", params)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.http.list")
		},
	}
	cmd.Flags().IntVar(&factoryID, "factory-id", 0, "Factory ID filter")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	return cmd
}

func newCmdHTTPGet(f *cmdutil.Factory) *cobra.Command {
	var id int
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an HTTP interface's details",
		Long: `Show one HTTP interface. Wraps GET /factory/http/get.

EXAMPLES:
    lingtong-cli factory http get --id 88`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/factory/http/get", map[string]interface{}{"id": id})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.http.get")
		},
	}
	cmd.Flags().IntVar(&id, "id", 0, "HTTP interface ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newCmdHTTPRun(f *cmdutil.Factory) *cobra.Command {
	var data string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run/test an HTTP interface",
		Long: `Execute an HTTP interface for testing. Wraps POST /factory/http/run.

--data is the run payload (JSON), typically including the interface id and
parameter values.

EXAMPLES:
    lingtong-cli factory http run --data '{"id":88,"params":{"page":1}}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(data != "", "data"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			body, err := cmdutil.ParseDataObject(data)
			if err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/factory/http/run", body)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.http.run")
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "Run payload (JSON object, required)")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

func newCmdHTTPDelete(f *cmdutil.Factory) *cobra.Command {
	var id int
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an HTTP interface (requires --yes)",
		Long: `Delete an HTTP interface by ID. Wraps POST /factory/http/delete.

This is destructive and requires --yes.

EXAMPLES:
    lingtong-cli factory http delete --id 88 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := cmdutil.ConfirmDestructive(yes, "delete HTTP interface"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/factory/http/delete", id)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.http.delete")
		},
	}
	cmd.Flags().IntVar(&id, "id", 0, "HTTP interface ID (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm the destructive operation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// ==================== factory script ====================

func newCmdScript(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "script",
		Short: "Manage factory scripts",
		Long:  "Script operations for a connector factory.",
	}
	cmd.AddCommand(newCmdScriptList(f))
	cmd.AddCommand(newCmdScriptGet(f))
	cmd.AddCommand(newCmdScriptDelete(f))
	return cmd
}

func newCmdScriptList(f *cmdutil.Factory) *cobra.Command {
	var factoryID int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a factory's scripts",
		Long: `List scripts for a factory. Wraps GET /factory/script/list.

EXAMPLES:
    lingtong-cli factory script list --factory-id 12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(factoryID != 0, "factory-id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/factory/script/list", map[string]interface{}{"factoryId": factoryID})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.script.list")
		},
	}
	cmd.Flags().IntVar(&factoryID, "factory-id", 0, "Factory ID (required)")
	_ = cmd.MarkFlagRequired("factory-id")
	return cmd
}

func newCmdScriptGet(f *cmdutil.Factory) *cobra.Command {
	var id int
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a script's details",
		Long: `Show one script. Wraps GET /factory/script/get.

EXAMPLES:
    lingtong-cli factory script get --id 55`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Get("/factory/script/get", map[string]interface{}{"id": id})
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.script.get")
		},
	}
	cmd.Flags().IntVar(&id, "id", 0, "Script ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newCmdScriptDelete(f *cmdutil.Factory) *cobra.Command {
	var id int
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a script (requires --yes)",
		Long: `Delete a script by ID. Wraps POST /factory/script/delete.

This is destructive and requires --yes.

EXAMPLES:
    lingtong-cli factory script delete --id 55 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireFlag(id != 0, "id"); err != nil {
				return err
			}
			if err := cmdutil.ConfirmDestructive(yes, "delete script"); err != nil {
				return err
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}
			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/factory/script/delete", id)
			if err != nil {
				return err
			}
			return f.WriteResponse(cmd, resp, "factory.script.delete")
		},
	}
	cmd.Flags().IntVar(&id, "id", 0, "Script ID (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm the destructive operation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
