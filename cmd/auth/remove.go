// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"fmt"

	"github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdAuthRemove(f *cmdutil.Factory) *cobra.Command {
	var label string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an auth identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" {
				return fmt.Errorf("--label is required")
			}

			// Check if identity exists
			identity := f.Config.GetAuth(label)
			if identity == nil {
				return fmt.Errorf("identity %q not found", label)
			}

			// Delete from keychain
			profileName := f.EffectiveProfile()
			if err := auth.DeleteTokenForAuth(profileName, label); err != nil {
				// Non-fatal: token may already be gone
				fmt.Printf("Warning: failed to remove token from keychain: %v\n", err)
			}

			// Remove from config
			f.Config.RemoveAuth(label)

			// If this was the current auth, switch to another or clear
			if f.Config.CurrentAuth == label {
				f.Config.CurrentAuth = ""
				auths := f.Config.ListAuths()
				if len(auths) > 0 {
					f.Config.CurrentAuth = auths[0]
					// Update legacy token
					if token, err := auth.GetTokenForAuth(profileName, auths[0]); err == nil {
						_ = auth.StoreToken(token)
					}
					fmt.Printf("Switched to identity %q.\n", auths[0])
				} else {
					// No more identities, clear legacy token
					_ = auth.DeleteToken()
				}
			}

			if err := f.Config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Removed identity %q.\n", label)
			_ = identity // suppress unused warning
			return nil
		},
	}

	cmd.Flags().StringVar(&label, "label", "", "Label of the identity to remove")
	return cmd
}
