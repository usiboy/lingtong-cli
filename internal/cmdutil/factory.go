// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package cmdutil

import (
	"os"

	"github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

// Factory provides dependencies to commands.
type Factory struct {
	Config     *config.Config
	IOStreams  *output.IOStreams
}

// NewDefault creates a factory with default values.
func NewDefault() *Factory {
	cfg, _ := config.Load()
	
	// Support token from environment variable for E2E testing
	if token := os.Getenv("LINGTONG_TOKEN"); token != "" {
		cfg.Token = token
	}
	
	// Load token from OS keychain if not set via environment
	if cfg.Token == "" {
		if token, err := auth.GetToken(); err == nil {
			cfg.Token = token
		}
	}
	
	return &Factory{
		Config:    cfg,
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

// InstallHelpFunc wraps the default help function.
func InstallHelpFunc(root *cobra.Command) {
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
	})
}

// NewWriter creates a new output writer with the factory's OmitNull config.
func (f *Factory) NewWriter(format output.Format) *output.Writer {
	return output.NewWriterWithOpts(f.IOStreams, format, f.Config.OmitNull)
}
