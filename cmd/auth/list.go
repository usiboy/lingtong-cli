// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdAuthList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all authenticated identities",
		RunE: func(cmd *cobra.Command, args []string) error {
			labels := f.Config.ListAuths()
			if len(labels) == 0 {
				// Check legacy token
				token, err := getToken()
				if err != nil || token == "" {
					fmt.Println("No authenticated identities. Run `lingtong-cli auth login` to login.")
					return nil
				}
				fmt.Println("Authenticated with legacy token:")
				fmt.Printf("  Token: %s\n", maskToken(token))
				fmt.Printf("  Host:  %s\n", f.Config.Host)
				fmt.Println("\nRun `lingtong-cli auth login --label <name>` to migrate to multi-auth.")
				return nil
			}

			profileName := f.EffectiveProfile()
			currentAuth := f.Config.CurrentAuth

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  LABEL\tTOKEN\tTENANT\tCREATED")

			for _, label := range labels {
				identity := f.Config.GetAuth(label)
				if identity == nil {
					continue
				}

				marker := "  "
				if label == currentAuth {
					marker = "* "
				}

				// Masked token display
				tokenDisplay := "apk-..." + identity.TokenSuffix
				if identity.TokenSuffix == "" {
					tokenDisplay = "apk-...****"
				}

				// Tenant display
				tenantDisplay := "(unknown)"
				if identity.TenantInfo != nil && identity.TenantInfo.TenantId != "" {
					tenantDisplay = identity.TenantInfo.TenantId
				}

				// Created date (just the date part)
				created := identity.CreatedAt
				if len(created) >= 10 {
					created = created[:10]
				}

				fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\n", marker, label, tokenDisplay, tenantDisplay, created)
			}
			w.Flush()

			// Verify current token is accessible
			if currentAuth != "" {
				if _, err := auth.GetTokenForAuth(profileName, currentAuth); err != nil {
					fmt.Printf("\nWarning: Current identity %q token not found in keychain.\n", currentAuth)
				}
			}

			return nil
		},
	}
}
