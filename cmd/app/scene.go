// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// AppScene represents a scene within an application JSON file.
type AppScene struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	Trigger     string `json:"trigger,omitempty"`
}

// AppFile represents the top-level structure of an application JSON file.
type AppFile struct {
	AppName            string        `json:"appName"`
	AppConnectors      []string      `json:"appConnectors"`
	BasicDatas         []interface{} `json:"basicDatas"`
	Scenes             []interface{} `json:"scenes"`
	Workflows          []interface{} `json:"workflows"`
	WorkflowConnectors []interface{} `json:"workflowConnectors"`
}

func newCmdAppScene(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scene",
		Short: "Manage scenes within an application",
		Long:  "List, add, and remove scenes in an application JSON file.",
	}

	cmd.AddCommand(newCmdAppSceneList(f))
	cmd.AddCommand(newCmdAppSceneAdd(f))
	cmd.AddCommand(newCmdAppSceneRemove(f))

	return cmd
}

func newCmdAppSceneList(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all scenes in an application",
		Long: `List all scenes defined in an application JSON file.

EXAMPLES:
    lingtong-cli app scene list --file app.json
    lingtong-cli app scene list --file app.json --format table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}

			appFile, err := loadAppFile(file)
			if err != nil {
				return err
			}

			scenes := extractScenes(appFile.Scenes)

			formatFlag := cmd.Flag("format")
			format := output.Format("json")
			if formatFlag != nil {
				format = output.Format(formatFlag.Value.String())
			}
			w := output.NewWriter(f.IOStreams, format)
			return w.Write(scenes)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Application JSON file path (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newCmdAppSceneAdd(f *cmdutil.Factory) *cobra.Command {
	var file, name, source, target, outputFlag string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new scene to an application",
		Long: `Add a new scene to an application JSON file.

EXAMPLES:
    lingtong-cli app scene add --file app.json --name "Order Sync" --source kmerp --target kingDeeCloudStar --output app.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if source == "" {
				return fmt.Errorf("--source is required")
			}
			if target == "" {
				return fmt.Errorf("--target is required")
			}

			appFile, err := loadAppFile(file)
			if err != nil {
				return err
			}

			newScene := AppScene{
				Name:   name,
				Source: source,
				Target: target,
			}

			appFile.Scenes = append(appFile.Scenes, newScene)

			outFile := outputFlag
			if outFile == "" {
				outFile = file
			}

			if err := saveAppFile(outFile, appFile); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added scene %q to %s\n", name, outFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Application JSON file path (required)")
	cmd.Flags().StringVar(&name, "name", "", "Scene name (required)")
	cmd.Flags().StringVar(&source, "source", "", "Source connector name (required)")
	cmd.Flags().StringVar(&target, "target", "", "Target connector name (required)")
	cmd.Flags().StringVar(&outputFlag, "output", "", "Output file path (default: overwrite input file)")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("target")

	return cmd
}

func newCmdAppSceneRemove(f *cmdutil.Factory) *cobra.Command {
	var file, name, outputFlag string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a scene from an application",
		Long: `Remove a scene from an application JSON file by name.

EXAMPLES:
    lingtong-cli app scene remove --file app.json --name "Order Sync"
    lingtong-cli app scene remove --file app.json --name "Order Sync" --output updated.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			appFile, err := loadAppFile(file)
			if err != nil {
				return err
			}

			updated, removed := removeSceneByName(appFile.Scenes, name)
			if !removed {
				return fmt.Errorf("scene %q not found in application", name)
			}

			appFile.Scenes = updated

			outFile := outputFlag
			if outFile == "" {
				outFile = file
			}

			if err := saveAppFile(outFile, appFile); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed scene %q from %s\n", name, outFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Application JSON file path (required)")
	cmd.Flags().StringVar(&name, "name", "", "Scene name to remove (required)")
	cmd.Flags().StringVar(&outputFlag, "output", "", "Output file path (default: overwrite input file)")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// loadAppFile reads and parses an application JSON file.
func loadAppFile(path string) (*AppFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var appFile AppFile
	if err := json.Unmarshal(data, &appFile); err != nil {
		return nil, fmt.Errorf("invalid application JSON: %w", err)
	}

	return &appFile, nil
}

// saveAppFile writes an application struct to a JSON file.
func saveAppFile(path string, appFile *AppFile) error {
	formatted, err := json.MarshalIndent(appFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	if err := os.WriteFile(path, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// extractScenes converts raw scene data into typed AppScene structs for display.
func extractScenes(raw []interface{}) []AppScene {
	var scenes []AppScene
	for _, item := range raw {
		switch v := item.(type) {
		case map[string]interface{}:
			scene := AppScene{}
			if n, ok := v["name"].(string); ok {
				scene.Name = n
			}
			if d, ok := v["description"].(string); ok {
				scene.Description = d
			}
			if s, ok := v["source"].(string); ok {
				scene.Source = s
			}
			if t, ok := v["target"].(string); ok {
				scene.Target = t
			}
			if tr, ok := v["trigger"].(string); ok {
				scene.Trigger = tr
			}
			scenes = append(scenes, scene)
		case AppScene:
			scenes = append(scenes, v)
		}
	}
	return scenes
}

// removeSceneByName removes the first scene with the given name.
// Returns the updated slice and a boolean indicating if a scene was removed.
func removeSceneByName(scenes []interface{}, name string) ([]interface{}, bool) {
	for i, item := range scenes {
		switch v := item.(type) {
		case map[string]interface{}:
			if n, ok := v["name"].(string); ok && strings.EqualFold(n, name) {
				return append(scenes[:i], scenes[i+1:]...), true
			}
		case AppScene:
			if strings.EqualFold(v.Name, name) {
				return append(scenes[:i], scenes[i+1:]...), true
			}
		}
	}
	return scenes, false
}
