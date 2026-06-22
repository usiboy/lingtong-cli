// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"

	"github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/cmdutil"
	ltconfig "github.com/lingtong/cli/internal/config"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// NewCmdProfile creates the profile command group.
func NewCmdProfile(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
		Long: `Manage multiple configuration profiles for different environments
(e.g., dev, staging, prod).

EXAMPLES:
    # List all profiles
    lingtong-cli config profile list

    # Show current active profile
    lingtong-cli config profile current

    # Switch to a different profile
    lingtong-cli config profile use staging

    # Add a new profile
    lingtong-cli config profile add --name dev --host https://dev-api.lingtong.com

    # Remove a profile (requires --yes confirmation)
    lingtong-cli config profile remove --name dev --yes`,
	}

	cmd.AddCommand(newCmdProfileList(f))
	cmd.AddCommand(newCmdProfileCurrent(f))
	cmd.AddCommand(newCmdProfileUse(f))
	cmd.AddCommand(newCmdProfileAdd(f))
	cmd.AddCommand(newCmdProfileRemove(f))

	return cmd
}

// newCmdProfileList lists all available profiles.
func newCmdProfileList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration profiles",
		Long: `List all available configuration profiles with their host URLs.

EXAMPLES:
    lingtong-cli config profile list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles := f.Config.ListProfiles()

			result := map[string]interface{}{
				"current":  f.Config.CurrentProfile,
				"profiles": []map[string]interface{}{},
			}

			profileList := make([]map[string]interface{}, 0, len(profiles))
			for _, name := range profiles {
				p := f.Config.GetProfile(name)
				if p == nil {
					continue
				}
				profileList = append(profileList, map[string]interface{}{
					"name": name,
					"host": p.Host,
				})
			}
			result["profiles"] = profileList
			result["count"] = len(profileList)

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "config.profile.list")
			return w.Write(result)
		},
	}
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

// newCmdProfileCurrent shows the currently active profile.
func newCmdProfileCurrent(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the currently active profile",
		Long: `Show the name and details of the currently active profile.

EXAMPLES:
    lingtong-cli config profile current`,
		RunE: func(cmd *cobra.Command, args []string) error {
			current := f.Config.CurrentProfile
			if current == "" {
				// No profile set, using top-level config
				result := map[string]interface{}{
					"profile": "",
					"host":    f.Config.Host,
					"brand":   f.Config.Brand,
					"message": "Using top-level configuration (no profile active)",
				}
				format := output.Format(cmd.Flag("format").Value.String())
				w := f.NewWriter(format, "config.profile.current")
				return w.Write(result)
			}

			p := f.Config.GetProfile(current)
			if p == nil {
				return lterrors.NewValidationError(lterrors.SubtypeNotFound,
					"current profile %q not found", current).
					WithHint("run 'lingtong-cli config profile list' to see available profiles")
			}

			result := map[string]interface{}{
				"profile": current,
				"host":    p.Host,
				"brand":   p.Brand,
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "config.profile.current")
			return w.Write(result)
		},
	}
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

// newCmdProfileUse switches to a different profile.
func newCmdProfileUse(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <profile-name>",
		Short: "Switch to a different profile",
		Long: `Set the active configuration profile. This updates the CurrentProfile
field in the config file.

EXAMPLES:
    lingtong-cli config profile use staging
    lingtong-cli config profile use prod`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Verify the profile exists
			p := f.Config.GetProfile(profileName)
			if p == nil {
				return lterrors.NewValidationError(lterrors.SubtypeNotFound,
					"profile %q not found", profileName).
					WithHint("run 'lingtong-cli config profile list' to see available profiles")
			}

			// Update CurrentProfile
			f.Config.CurrentProfile = profileName
			if err := f.Config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			result := map[string]interface{}{
				"success": true,
				"profile": profileName,
				"host":    p.Host,
				"message": fmt.Sprintf("Switched to profile %q", profileName),
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "config.profile.use")
			return w.Write(result)
		},
	}
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")
	return cmd
}

// newCmdProfileAdd adds a new profile.
func newCmdProfileAdd(f *cmdutil.Factory) *cobra.Command {
	var name, host, brand, token string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new configuration profile",
		Long: `Add a new configuration profile with the specified name and host.

Each profile carries its own credentials. Pass --token to store an environment
-specific token in the OS keychain so that switching profiles also switches the
token used for API calls. Without --token the profile reuses the default token.

EXAMPLES:
    lingtong-cli config profile add --name dev --host https://dev-api.lingtong.com
    lingtong-cli config profile add --name prod --host https://api.lingtong.com --brand lingtong
    lingtong-cli config profile add --name prod --host https://api.lingtong.com --token apk-xxxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return lterrors.NewValidationError(lterrors.SubtypeMissingField,
					"--name is required").
					WithHint("usage: lingtong-cli config profile add --name <name> --host <url>")
			}
			if host == "" {
				return lterrors.NewValidationError(lterrors.SubtypeMissingField,
					"--host is required").
					WithHint("usage: lingtong-cli config profile add --name <name> --host <url>")
			}

			// Check if profile already exists
			if f.Config.GetProfile(name) != nil {
				return lterrors.NewValidationError(lterrors.SubtypeConflict,
					"profile %q already exists", name).
					WithHint("use 'lingtong-cli config profile use' to switch to it, or remove it first")
			}

			// Add the profile
			profile := ltconfig.Profile{
				Host:  host,
				Brand: brand,
			}
			f.Config.SetProfile(name, profile)

			if err := f.Config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			// Store the profile-scoped token in the keychain if provided.
			tokenStored := false
			if token != "" {
				if err := auth.StoreTokenForProfile(name, token); err != nil {
					return fmt.Errorf("profile saved but failed to store token: %w", err)
				}
				tokenStored = true
			}

			result := map[string]interface{}{
				"success":     true,
				"profile":     name,
				"host":        host,
				"tokenStored": tokenStored,
				"message":     fmt.Sprintf("Added profile %q", name),
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "config.profile.add")
			return w.Write(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Profile name (required)")
	cmd.Flags().StringVar(&host, "host", "", "API host URL (required)")
	cmd.Flags().StringVar(&brand, "brand", "", "Brand name (optional)")
	cmd.Flags().StringVar(&token, "token", "", "Profile-scoped API token, stored in the OS keychain (optional)")
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")

	return cmd
}

// newCmdProfileRemove removes a profile.
func newCmdProfileRemove(f *cmdutil.Factory) *cobra.Command {
	var name string
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a configuration profile",
		Long: `Remove a configuration profile by name. This is a destructive operation
and requires --yes confirmation.

EXAMPLES:
    lingtong-cli config profile remove --name dev --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return lterrors.NewValidationError(lterrors.SubtypeMissingField,
					"--name is required").
					WithHint("usage: lingtong-cli config profile remove --name <name> --yes")
			}

			// Confirm destructive operation
			if err := cmdutil.ConfirmDestructive(yes, "removing profile"); err != nil {
				return err
			}

			// Check if profile exists
			if f.Config.GetProfile(name) == nil {
				return lterrors.NewValidationError(lterrors.SubtypeNotFound,
					"profile %q not found", name)
			}

			// Check if removing current profile
			if f.Config.CurrentProfile == name {
				return lterrors.NewValidationError(lterrors.SubtypeConflict,
					"cannot remove active profile %q", name).
					WithHint("switch to another profile first with 'lingtong-cli config profile use'")
			}

			// Remove the profile
			f.Config.RemoveProfile(name)

			if err := f.Config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			// Best-effort cleanup of the profile-scoped token. A missing token
			// is fine (not every profile has one), so the error is ignored.
			_ = auth.DeleteTokenForProfile(name)

			result := map[string]interface{}{
				"success": true,
				"profile": name,
				"message": fmt.Sprintf("Removed profile %q", name),
			}

			format := output.Format(cmd.Flag("format").Value.String())
			w := f.NewWriter(format, "config.profile.remove")
			return w.Write(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Profile name to remove (required)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm destructive operation")
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")

	return cmd
}
