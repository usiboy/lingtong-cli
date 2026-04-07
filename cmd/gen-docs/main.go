// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/lingtong/cli/cmd"
	"github.com/lingtong/cli/internal/cmdutil"
)

// gen-docs is a utility command to generate markdown documentation for all CLI commands.
// Usage: go run cmd/gen-docs/main.go [output-dir]
func main() {
	outputDir := "./doc/commands"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create the root command
	f := cmdutil.NewDefault()
	rootCmd := cmd.NewRootCommand(f)

	// Disable some root command features for doc generation
	rootCmd.DisableAutoGenTag = true

	// Generate markdown documentation
	fmt.Printf("Generating markdown documentation to %s...\n", outputDir)

	// Generate doc for the root command and all subcommands
	if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating markdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Documentation generated successfully!")
	fmt.Printf("Output directory: %s\n", outputDir)
	fmt.Println("\nGenerated files:")

	// List generated files
	files, err := os.ReadDir(outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading output directory: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err == nil {
				fmt.Printf("  - %s (%d bytes)\n", file.Name(), info.Size())
			}
		}
	}
}
