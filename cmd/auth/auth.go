// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	storeToken        = auth.StoreToken
	storeTokenForAuth = auth.StoreTokenForAuth
	getToken          = auth.GetToken
	deleteToken       = auth.DeleteToken
	verifyTokenFunc   = verifyToken
)

// NewCmdAuth creates the auth command.
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Authenticate with Lingtong platform using API token (apk-xxx).",
	}

	cmd.AddCommand(newCmdAuthLogin(f))
	cmd.AddCommand(newCmdAuthStatus(f))
	cmd.AddCommand(newCmdAuthLogout(f))
	cmd.AddCommand(newCmdAuthList(f))
	cmd.AddCommand(newCmdAuthUse(f))
	cmd.AddCommand(newCmdAuthRemove(f))

	return cmd
}

func newCmdAuthLogin(f *cmdutil.Factory) *cobra.Command {
	var token string
	var fromEnv bool
	var label string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Lingtong platform",
		Long: `Login to Lingtong platform and store API token securely.

The token can be provided in three ways:
1. Via --token flag: lingtong-cli auth login --token apk-xxx
2. Via environment variable: lingtong-cli auth login --from-env
3. Interactive input: lingtong-cli auth login

Use --label to assign a name to this identity for multi-auth management.

EXAMPLES:
    lingtong-cli auth login
    lingtong-cli auth login --token apk-xxx
    lingtong-cli auth login --token apk-xxx --label company-a
    lingtong-cli auth login --from-env --label ci-token`,
		RunE: func(cmd *cobra.Command, args []string) error {
			host := f.Config.Host
			if host == "" {
				return fmt.Errorf("no host configured. Run `lingtong-cli config init` first")
			}

			// Get token from different sources
			if token == "" && fromEnv {
				token = os.Getenv("LINGTONG_API_TOKEN")
				if token == "" {
					return fmt.Errorf("LINGTONG_API_TOKEN environment variable not set")
				}
			}

			if token == "" {
				fmt.Print("API Token (apk-xxx): ")
				fmt.Scanln(&token)
			}

			if token == "" {
				return fmt.Errorf("token is required")
			}

			// Validate token format
			if !strings.HasPrefix(token, "apk-") {
				fmt.Println("Warning: Token should start with 'apk-'. Proceeding anyway...")
			}

			// Verify token by making a test request
			fmt.Print("Verifying token... ")
			err := verifyTokenFunc(host, token)
			if err != nil {
				fmt.Printf("Warning: Token verification failed: %v\n", err)
				fmt.Println("Token saved but may be invalid. You can try again with a valid token.")
			} else {
				fmt.Println("OK")
			}

			// Fetch tenant info (best-effort, non-fatal)
			tenantId := ""
			fmt.Print("Fetching tenant info... ")
			tenantId, err = fetchTenantId(host, token)
			if err != nil {
				fmt.Printf("skipped (%v)\n", err)
			} else {
				fmt.Printf("done (Tenant: %s)\n", tenantId)
			}

			// Determine label
			if label == "" {
				label = generateLabel(f.Config, tenantId, token)
			}

			// Compute token suffix for masked display
			tokenSuffix := ""
			if len(token) >= 4 {
				tokenSuffix = token[len(token)-4:]
			}

			// Store token in OS keychain (multi-auth format)
			profileName := f.EffectiveProfile()
			err = storeTokenForAuth(profileName, label, token)
			if err != nil {
				return fmt.Errorf("failed to store token: %w", err)
			}

			// Also store as default token for backward compat when no multi-auth
			if len(f.Config.Auths) == 0 && f.Config.CurrentAuth == "" {
				_ = storeToken(token)
			}

			// Cache auth identity in config
			identity := config.NewAuthIdentity(label, tokenSuffix, tenantId)
			f.Config.SetAuth(label, identity)
			f.Config.CurrentAuth = label

			if err := f.Config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Label: %s\n", label)
			fmt.Println("Login successful. Token stored securely in OS keychain.")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "API token (apk-xxx)")
	cmd.Flags().BoolVar(&fromEnv, "from-env", false, "Read token from LINGTONG_API_TOKEN environment variable")
	cmd.Flags().StringVar(&label, "label", "", "Label for this identity (auto-generated if empty)")
	return cmd
}

// generateLabel creates a unique label for the auth identity.
func generateLabel(cfg *config.Config, tenantId, token string) string {
	var base string
	if tenantId != "" {
		// Use last 6 chars of tenantId
		suffix := tenantId
		if len(suffix) > 6 {
			suffix = suffix[len(suffix)-6:]
		}
		base = "tenant-" + suffix
	} else {
		// Use last 4 chars of token
		suffix := token
		if len(suffix) > 4 {
			suffix = suffix[len(suffix)-4:]
		}
		base = "token-" + suffix
	}

	// Ensure uniqueness
	label := base
	for i := 2; cfg.GetAuth(label) != nil; i++ {
		label = fmt.Sprintf("%s-%d", base, i)
	}
	return label
}

func verifyToken(host, token string) error {
	// Simple verification by making a lightweight API call
	// We use the connector info endpoint as it doesn't require additional parameters
	c := client.NewClient(host, token)
	_, err := c.Get("/gw/ai/connector/info", map[string]interface{}{
		"connector": "kmerp",
		"env":       "test",
	})

	// We only care about authentication errors (401), not business errors
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "401") || strings.Contains(errStr, "无身份信息") {
			return fmt.Errorf("invalid token: authentication failed")
		}
		// Other errors (network, etc.) are not critical
		return nil
	}

	return nil
}

func newCmdAuthStatus(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use EffectiveAuth to respect --auth flag
			effectiveAuth := f.EffectiveAuth()
			if effectiveAuth != "" {
				identity := f.Config.ResolveAuth(effectiveAuth)
				if identity != nil {
					profileName := f.EffectiveProfile()
					token, err := auth.GetTokenForAuth(profileName, identity.Label)
					if err == nil && token != "" {
						printAuthStatus(f, identity, token, profileName)
						return nil
					}
				}
			}

			// Fallback to legacy single-token mode
			token, err := getToken()
			if err != nil {
				fmt.Println("Not authenticated. Run `lingtong-cli auth login` to login.")
				return nil
			}

			// Show token status (masked)
			maskedToken := maskToken(token)
			fmt.Printf("Authenticated: Yes\n")
			fmt.Printf("Token: %s\n", maskedToken)
			fmt.Printf("Host: %s\n", f.Config.Host)

			// Verify token validity
			if f.Config.Host != "" {
				fmt.Print("\nVerifying token... ")
				err := verifyTokenFunc(f.Config.Host, token)
				if err != nil {
					fmt.Printf("Invalid: %v\n", err)
					fmt.Println("Run `lingtong-cli auth login` to re-authenticate.")
				} else {
					fmt.Println("Valid")
				}
			}

			return nil
		},
	}
}

// printAuthStatus prints status for a multi-auth identity.
func printAuthStatus(f *cmdutil.Factory, identity *config.AuthIdentity, token, profileName string) {
	maskedToken := maskToken(token)
	fmt.Printf("Authenticated: Yes\n")
	fmt.Printf("Token: %s\n", maskedToken)
	fmt.Printf("Host: %s\n", f.Config.Host)
	fmt.Printf("Label: %s\n", identity.Label)

	if identity.TenantInfo != nil && identity.TenantInfo.TenantId != "" {
		fmt.Printf("\nTenant: %s\n", identity.TenantInfo.TenantId)
		fmt.Printf("Cached: %s\n", identity.TenantInfo.FetchedAt)
	}

	// Verify token validity
	if f.Config.Host != "" {
		fmt.Print("\nVerifying token... ")
		err := verifyTokenFunc(f.Config.Host, token)
		if err != nil {
			fmt.Printf("Invalid: %v\n", err)
			fmt.Println("Run `lingtong-cli auth login` to re-authenticate.")
		} else {
			fmt.Println("Valid")
		}
	}
}

func newCmdAuthLogout(f *cmdutil.Factory) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear stored token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				return logoutAll(f)
			}
			return logoutCurrent(f)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Logout from all identities")
	return cmd
}

// logoutCurrent removes the current auth identity (or legacy token if no multi-auth).
func logoutCurrent(f *cmdutil.Factory) error {
	profileName := f.EffectiveProfile()

	// Try multi-auth first
	if f.Config.CurrentAuth != "" {
		label := f.Config.CurrentAuth
		if err := auth.DeleteTokenForAuth(profileName, label); err != nil {
			// Ignore keychain errors (token may already be gone)
			_ = err
		}
		f.Config.RemoveAuth(label)
		f.Config.CurrentAuth = ""

		// Switch to another auth if available
		auths := f.Config.ListAuths()
		if len(auths) > 0 {
			f.Config.CurrentAuth = auths[0]
			// Update legacy token for backward compat
			if token, err := auth.GetTokenForAuth(profileName, auths[0]); err == nil {
				_ = storeToken(token)
			}
		}

		if err := f.Config.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Logged out from identity %q.\n", label)
		if len(auths) > 0 {
			fmt.Printf("Switched to identity %q.\n", auths[0])
		}
		return nil
	}

	// Legacy single-token mode
	err := deleteToken()
	if err != nil {
		return fmt.Errorf("failed to clear token: %w", err)
	}
	fmt.Println("Logged out successfully. Token removed from keychain.")
	return nil
}

// logoutAll removes all auth identities.
func logoutAll(f *cmdutil.Factory) error {
	profileName := f.EffectiveProfile()
	labels := f.Config.ListAuths()
	for _, label := range labels {
		_ = auth.DeleteTokenForAuth(profileName, label)
		f.Config.RemoveAuth(label)
	}
	f.Config.CurrentAuth = ""

	// Also clear legacy token
	_ = deleteToken()

	if err := f.Config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Logged out from all %d identities.\n", len(labels))
	return nil
}

// maskToken masks the token for display, showing only first 8 characters and a safe suffix.
func maskToken(token string) string {
	if len(token) <= 12 {
		return "****"
	}

	suffixLen := min(5, len(token)-10)
	return token[:8] + "..." + token[len(token)-suffixLen:]
}
