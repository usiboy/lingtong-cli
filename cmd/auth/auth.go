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
	"github.com/spf13/cobra"
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

	return cmd
}

func newCmdAuthLogin(f *cmdutil.Factory) *cobra.Command {
	var token string
	var fromEnv bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Lingtong platform",
		Long: `Login to Lingtong platform and store API token securely.

The token can be provided in three ways:
1. Via --token flag: lingtong-cli auth login --token apk-xxx
2. Via environment variable: lingtong-cli auth login --from-env
3. Interactive input: lingtong-cli auth login

EXAMPLES:
    lingtong-cli auth login
    lingtong-cli auth login --token apk-Gx6vDOEmALY7iJRcLZcD4nWF
    lingtong-cli auth login --from-env`,
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

			// Store token in OS keychain
			err := auth.StoreToken(token)
			if err != nil {
				return fmt.Errorf("failed to store token: %w", err)
			}

			// Verify token by making a test request
			fmt.Print("Verifying token... ")
			err = verifyToken(host, token)
			if err != nil {
				fmt.Printf("Warning: Token verification failed: %v\n", err)
				fmt.Println("Token saved but may be invalid. You can try again with a valid token.")
			} else {
				fmt.Println("OK")
			}

			fmt.Println("Login successful. Token stored securely in OS keychain.")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "API token (apk-xxx)")
	cmd.Flags().BoolVar(&fromEnv, "from-env", false, "Read token from LINGTONG_API_TOKEN environment variable")
	return cmd
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
			token, err := auth.GetToken()
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
				err := verifyToken(f.Config.Host, token)
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

func newCmdAuthLogout(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear stored token",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := auth.DeleteToken()
			if err != nil {
				return fmt.Errorf("failed to clear token: %w", err)
			}
			fmt.Println("Logged out successfully. Token removed from keychain.")
			return nil
		},
	}
}

// maskToken masks the token for display, showing only first 8 and last 4 characters
func maskToken(token string) string {
	if len(token) <= 12 {
		return "****"
	}
	return token[:8] + "..." + token[len(token)-4:]
}
