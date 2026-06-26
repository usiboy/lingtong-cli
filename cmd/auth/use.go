// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"fmt"

	"github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdAuthUse(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <label>",
		Short: "Switch to a different auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			// Check if identity exists
			identity := f.Config.GetAuth(label)
			if identity == nil {
				return fmt.Errorf("identity %q not found. Run `lingtong-cli auth list` to see available identities", label)
			}

			// Verify token exists in keychain
			profileName := f.EffectiveProfile()
			token, err := auth.GetTokenForAuth(profileName, label)
			if err != nil {
				return fmt.Errorf("token for identity %q not found in keychain", label)
			}

			// Update current auth
			f.Config.CurrentAuth = label

			// Also update legacy token for backward compat
			if err := auth.StoreToken(token); err != nil {
				// Non-fatal: legacy token update failed
				_ = err
			}

			if err := f.Config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Switched to identity %q.\n", label)
			if identity.TenantInfo != nil && identity.TenantInfo.TenantId != "" {
				fmt.Printf("Tenant: %s\n", identity.TenantInfo.TenantId)
			}

			return nil
		},
	}

	return cmd
}
