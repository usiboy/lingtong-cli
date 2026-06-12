// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// DiffReport holds the differences between two application exports.
type DiffReport struct {
	AddedScenes       []string
	RemovedScenes     []string
	ModifiedScenes    []string
	AddedConnectors   []string
	RemovedConnectors []string
	AddedBasicDatas   []string
	RemovedBasicDatas []string
}

func newCmdAppDiff(f *cmdutil.Factory) *cobra.Command {
	var fileA, fileB string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two application configuration files",
		Long: `Compare two application export JSON files and show differences.

Shows added/removed/modified scenes, connectors, and basicDatas.

EXAMPLES:
    lingtong-cli app diff --file-a app-v1.json --file-b app-v2.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fileA == "" {
				return fmt.Errorf("--file-a is required")
			}
			if fileB == "" {
				return fmt.Errorf("--file-b is required")
			}

			dataA, err := os.ReadFile(fileA)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fileA, err)
			}

			dataB, err := os.ReadFile(fileB)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fileB, err)
			}

			var appA, appB AppExport
			if err := json.Unmarshal(dataA, &appA); err != nil {
				return fmt.Errorf("invalid JSON in %s: %w", fileA, err)
			}
			if err := json.Unmarshal(dataB, &appB); err != nil {
				return fmt.Errorf("invalid JSON in %s: %w", fileB, err)
			}

			report := diffApps(&appA, &appB)
			printDiffReport(cmd, &report, fileA, fileB)
			return nil
		},
	}

	cmd.Flags().StringVar(&fileA, "file-a", "", "First application JSON file (required)")
	cmd.Flags().StringVar(&fileB, "file-b", "", "Second application JSON file (required)")
	_ = cmd.MarkFlagRequired("file-a")
	_ = cmd.MarkFlagRequired("file-b")

	return cmd
}

// diffApps compares two application exports and returns a diff report.
func diffApps(a, b *AppExport) DiffReport {
	report := DiffReport{}

	// Compare scenes
	sceneMapA := make(map[string]bool)
	for _, s := range a.Scenes {
		sceneMapA[s.Name] = true
	}
	sceneMapB := make(map[string]bool)
	for _, s := range b.Scenes {
		sceneMapB[s.Name] = true
	}

	for name := range sceneMapA {
		if !sceneMapB[name] {
			report.RemovedScenes = append(report.RemovedScenes, name)
		}
	}
	for name := range sceneMapB {
		if !sceneMapA[name] {
			report.AddedScenes = append(report.AddedScenes, name)
		}
	}

	// Compare connectors
	connMapA := make(map[string]bool)
	for _, c := range a.AppConnectors {
		connMapA[c] = true
	}
	connMapB := make(map[string]bool)
	for _, c := range b.AppConnectors {
		connMapB[c] = true
	}

	for name := range connMapA {
		if !connMapB[name] {
			report.RemovedConnectors = append(report.RemovedConnectors, name)
		}
	}
	for name := range connMapB {
		if !connMapA[name] {
			report.AddedConnectors = append(report.AddedConnectors, name)
		}
	}

	// Compare basicDatas count (simple comparison)
	if len(a.BasicDatas) != len(b.BasicDatas) {
		diff := len(b.BasicDatas) - len(a.BasicDatas)
		if diff > 0 {
			for i := 0; i < diff; i++ {
				report.AddedBasicDatas = append(report.AddedBasicDatas, fmt.Sprintf("basicData[%d]", len(a.BasicDatas)+i))
			}
		} else {
			for i := 0; i < -diff; i++ {
				report.RemovedBasicDatas = append(report.RemovedBasicDatas, fmt.Sprintf("basicData[%d]", len(b.BasicDatas)+i))
			}
		}
	}

	// Sort for consistent output
	sort.Strings(report.AddedScenes)
	sort.Strings(report.RemovedScenes)
	sort.Strings(report.ModifiedScenes)
	sort.Strings(report.AddedConnectors)
	sort.Strings(report.RemovedConnectors)
	sort.Strings(report.AddedBasicDatas)
	sort.Strings(report.RemovedBasicDatas)

	return report
}

// printDiffReport outputs the diff report to stdout.
func printDiffReport(cmd *cobra.Command, report *DiffReport, fileA, fileB string) {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Comparing: %s -> %s\n", fileA, fileB)
	fmt.Fprintln(out, strings.Repeat("-", 50))

	hasChanges := false

	if len(report.AddedScenes) > 0 {
		hasChanges = true
		fmt.Fprintf(out, "Added scenes (%d):\n", len(report.AddedScenes))
		for _, s := range report.AddedScenes {
			fmt.Fprintf(out, "  + %s\n", s)
		}
		fmt.Fprintln(out)
	}

	if len(report.RemovedScenes) > 0 {
		hasChanges = true
		fmt.Fprintf(out, "Removed scenes (%d):\n", len(report.RemovedScenes))
		for _, s := range report.RemovedScenes {
			fmt.Fprintf(out, "  - %s\n", s)
		}
		fmt.Fprintln(out)
	}

	if len(report.ModifiedScenes) > 0 {
		hasChanges = true
		fmt.Fprintf(out, "Modified scenes (%d):\n", len(report.ModifiedScenes))
		for _, s := range report.ModifiedScenes {
			fmt.Fprintf(out, "  ~ %s\n", s)
		}
		fmt.Fprintln(out)
	}

	if len(report.AddedConnectors) > 0 {
		hasChanges = true
		fmt.Fprintf(out, "Added connectors (%d):\n", len(report.AddedConnectors))
		for _, c := range report.AddedConnectors {
			fmt.Fprintf(out, "  + %s\n", c)
		}
		fmt.Fprintln(out)
	}

	if len(report.RemovedConnectors) > 0 {
		hasChanges = true
		fmt.Fprintf(out, "Removed connectors (%d):\n", len(report.RemovedConnectors))
		for _, c := range report.RemovedConnectors {
			fmt.Fprintf(out, "  - %s\n", c)
		}
		fmt.Fprintln(out)
	}

	if len(report.AddedBasicDatas) > 0 || len(report.RemovedBasicDatas) > 0 {
		hasChanges = true
		if len(report.AddedBasicDatas) > 0 {
			fmt.Fprintf(out, "Added basicDatas (%d):\n", len(report.AddedBasicDatas))
			for _, b := range report.AddedBasicDatas {
				fmt.Fprintf(out, "  + %s\n", b)
			}
			fmt.Fprintln(out)
		}
		if len(report.RemovedBasicDatas) > 0 {
			fmt.Fprintf(out, "Removed basicDatas (%d):\n", len(report.RemovedBasicDatas))
			for _, b := range report.RemovedBasicDatas {
				fmt.Fprintf(out, "  - %s\n", b)
			}
			fmt.Fprintln(out)
		}
	}

	if !hasChanges {
		fmt.Fprintln(out, "No differences found.")
	}
}
