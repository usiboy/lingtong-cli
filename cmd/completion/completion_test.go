// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package completion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func newTestFactory() *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{Host: "https://test.example.com"},
		IOStreams: &output.IOStreams{
			In:     &bytes.Buffer{},
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	}
}

func newTestRootCmd(f *cmdutil.Factory) *cobra.Command {
	root := &cobra.Command{
		Use:           "lingtong-cli",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(NewCmdCompletion(f))
	return root
}

func TestCompletionBash(t *testing.T) {
	f := newTestFactory()
	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	root := newTestRootCmd(f)
	root.SetArgs([]string{"completion", "bash"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "bash") && !strings.Contains(output, "completion") {
		t.Errorf("bash completion output seems empty or invalid: %d bytes", len(output))
	}
}

func TestCompletionZsh(t *testing.T) {
	f := newTestFactory()
	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	root := newTestRootCmd(f)
	root.SetArgs([]string{"completion", "zsh"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("zsh completion output is empty")
	}
}

func TestCompletionFish(t *testing.T) {
	f := newTestFactory()
	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	root := newTestRootCmd(f)
	root.SetArgs([]string{"completion", "fish"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("fish completion output is empty")
	}
}

func TestCompletionPowershell(t *testing.T) {
	f := newTestFactory()
	var buf bytes.Buffer
	f.IOStreams.Out = &buf
	root := newTestRootCmd(f)
	root.SetArgs([]string{"completion", "powershell"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("powershell completion output is empty")
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	f := newTestFactory()
	root := newTestRootCmd(f)
	root.SetArgs([]string{"completion", "invalid"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid shell, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error = %q, want 'unsupported shell'", err.Error())
	}
}

func TestCompletionNoArgs(t *testing.T) {
	f := newTestFactory()
	root := newTestRootCmd(f)
	root.SetArgs([]string{"completion"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing args, got nil")
	}
}
