// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// NewCmdApi creates the api command.
func NewCmdApi(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api <method> <path>",
		Short: "Call any Lingtong API endpoint",
		Long: `Call any Lingtong API endpoint directly.

EXAMPLES:
    lingtong-cli api GET /gw/ai/connector/info?connector=kmerp
    lingtong-cli api POST /gw/ai/scene/create --data '{"name":"test"}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path := args[1]

			dataStr, _ := cmd.Flags().GetString("data")
			paramsStr, _ := cmd.Flags().GetString("params")
			format := output.Format(cmd.Flag("format").Value.String())

			c := client.NewClient(f.Config.Host, f.Config.Token)

			var resp []byte
			var err error

			switch method {
			case "GET":
				resp, err = c.Get(path, parseParams(paramsStr))
			case "POST":
				resp, err = c.Post(path, parseData(dataStr))
			case "PUT":
				resp, err = c.Put(path, parseData(dataStr))
			case "DELETE":
				resp, err = c.Delete(path)
			default:
				return fmt.Errorf("unsupported HTTP method: %s", method)
			}

			if err != nil {
				return err
			}

			w := output.NewWriter(f.IOStreams, format)
			var data interface{}
			if err := json.Unmarshal(resp, &data); err != nil {
				return err
			}
			return w.Write(data)
		},
	}

	cmd.Flags().String("data", "", "Request body (JSON)")
	cmd.Flags().String("params", "", "Query parameters (JSON)")
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

func parseData(dataStr string) interface{} {
	if dataStr == "" {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return dataStr
	}
	return data
}

func parseParams(paramsStr string) map[string]interface{} {
	if paramsStr == "" {
		return nil
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		return nil
	}
	return params
}
