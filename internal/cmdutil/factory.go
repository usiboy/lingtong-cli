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
	Envelope   bool                  // global --envelope flag
	JqExpr     string                // global --jq flag
	Profile    string                // global --profile flag (overrides CurrentProfile)
	Notice     *output.Notice        // system notice for envelope injection
	NoticeChan <-chan *output.Notice // async notice fetch channel
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

	// Resolve profile: --profile flag > CurrentProfile > top-level config
	// The profile resolution is deferred to PersistentPreRun in root.go,
	// where we have access to the --profile flag value.

	return &Factory{
		Config: cfg,
		IOStreams: &output.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

// EffectiveProfile returns the profile that is actually in effect: the
// --profile flag when set, otherwise the persisted CurrentProfile, otherwise
// "" (the default profile / top-level config).
func (f *Factory) EffectiveProfile() string {
	if f.Profile != "" {
		return f.Profile
	}
	if f.Config != nil {
		return f.Config.CurrentProfile
	}
	return ""
}

// InstallHelpFunc wraps the default help function.
func InstallHelpFunc(root *cobra.Command) {
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
	})
}

// NewWriter creates a new output writer with the factory's OmitNull config.
// Optional commandPath sets the identity field in envelope mode.
// If a NoticeChan is available, the writer will wait for it before writing.
func (f *Factory) NewWriter(format output.Format, commandPath ...string) *output.Writer {
	identity := ""
	if len(commandPath) > 0 {
		identity = commandPath[0]
	}

	// Resolve notice without ever blocking command output: take it only if the
	// background fetch has already produced a result, otherwise skip it.
	if f.Notice == nil && f.NoticeChan != nil {
		select {
		case n := <-f.NoticeChan:
			f.Notice = n
		default:
			// Not ready yet — notices are non-critical, so don't wait.
		}
	}

	return output.NewWriterWithOptions(f.IOStreams, format,
		output.WithOmitNull(f.Config.OmitNull),
		output.WithEnvelope(f.Envelope, identity),
		output.WithJq(f.JqExpr),
		output.WithNotice(f.Notice),
	)
}
