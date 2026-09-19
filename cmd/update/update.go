// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package update provides the `update` command for checking and performing
// CLI version updates. It detects the installation method and outputs
// the appropriate update command.
package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/lingtong/cli/internal/build"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/lingtong/cli/internal/version"
	"github.com/spf13/cobra"
)

// defaultCheckURL is the default URL to check for the latest version.
const defaultCheckURL = "https://api.github.com/repos/usiboy/lingtong-cli/releases/latest"

// httpTimeout for version check requests.
const httpTimeout = 5 * time.Second

// CheckURL is the release endpoint queried for the latest version. It is a
// variable so tests can redirect it; it can also be overridden at runtime via
// the LINGTONG_UPDATE_URL environment variable.
var CheckURL = defaultCheckURL

// NewCmdUpdate creates the update command.
func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var checkOnly bool
	var targetVersion string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for and apply CLI updates",
		Long: `Check if a newer version of lingtong-cli is available and display
update instructions.

The command detects your installation method (npm, go install, or binary)
and provides the appropriate update command.

EXAMPLES:
    # Check for updates without applying
    lingtong-cli update --check

    # Show update instructions
    lingtong-cli update

    # Check for a specific version
    lingtong-cli update --version v1.3.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentVersion := build.Version
			if currentVersion == "" || currentVersion == "dev" {
				currentVersion = "dev"
			}

			// If target version is specified, use it directly
			if targetVersion != "" {
				result := buildUpdateResult(currentVersion, targetVersion, false)
				return outputResult(cmd, f, result)
			}

			// Check for latest version
			latest, err := fetchLatestVersion(resolveCheckURL())
			if err != nil {
				// Graceful degradation: show manual check instructions
				result := map[string]interface{}{
					"currentVersion": currentVersion,
					"checkError":     err.Error(),
					"message":        "Could not check for updates. Visit the releases page manually.",
					"releaseURL":     "https://github.com/usiboy/lingtong-cli/releases",
				}
				return outputResult(cmd, f, result)
			}

			needsUpdate := version.Compare(currentVersion, latest) < 0
			result := buildUpdateResult(currentVersion, latest, needsUpdate)

			if checkOnly {
				result["checkOnly"] = true
			}

			return outputResult(cmd, f, result)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't show install instructions")
	cmd.Flags().StringVar(&targetVersion, "version", "", "Target version to update to (e.g., v1.3.0)")
	cmd.Flags().String("format", "json", "Output format: json, table, pretty")

	return cmd
}

// fetchLatestVersion fetches the latest release version from the given URL.
func fetchLatestVersion(url string) (string, error) {
	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("version check returned status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no version tag found in release info")
	}

	return release.TagName, nil
}

// resolveCheckURL returns the effective release-check URL, honoring the
// LINGTONG_UPDATE_URL environment override.
func resolveCheckURL() string {
	if env := os.Getenv("LINGTONG_UPDATE_URL"); env != "" {
		return env
	}
	return CheckURL
}

// buildUpdateResult constructs the update result map.
func buildUpdateResult(current, latest string, needsUpdate bool) map[string]interface{} {
	result := map[string]interface{}{
		"currentVersion": current,
		"latestVersion":  latest,
		"upToDate":       !needsUpdate,
	}

	if needsUpdate {
		result["message"] = fmt.Sprintf("A new version is available: %s (current: %s)", latest, current)
		result["installInstructions"] = getInstallInstructions(latest)
		result["releaseURL"] = fmt.Sprintf("https://github.com/usiboy/lingtong-cli/releases/tag/%s", latest)
	} else {
		result["message"] = fmt.Sprintf("You are up to date (version %s)", current)
	}

	return result
}

// getInstallInstructions returns platform-specific update instructions.
func getInstallInstructions(version string) []map[string]string {
	osName := runtime.GOOS
	instructions := []map[string]string{}

	// npm installation
	instructions = append(instructions, map[string]string{
		"method":  "npm",
		"command": "npm install -g @lingtong-cli/cli",
	})

	// go install
	instructions = append(instructions, map[string]string{
		"method":  "go install",
		"command": fmt.Sprintf("go install github.com/lingtong/cli@%s", version),
	})

	// Binary download
	downloadURL := fmt.Sprintf("https://github.com/usiboy/lingtong-cli/releases/download/%s/lingtong-cli-%s-%s",
		version, osName, runtime.GOARCH)
	instructions = append(instructions, map[string]string{
		"method":  "binary",
		"command": fmt.Sprintf("curl -L %s -o lingtong-cli && chmod +x lingtong-cli", downloadURL),
	})

	return instructions
}

// outputResult writes the result through the factory's writer.
func outputResult(cmd *cobra.Command, f *cmdutil.Factory, result map[string]interface{}) error {
	format := output.Format(cmd.Flag("format").Value.String())
	w := f.NewWriter(format, "update")
	return w.Write(result)
}
