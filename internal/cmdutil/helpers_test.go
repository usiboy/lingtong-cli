// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package cmdutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/config"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func TestRequireHostConfigured(t *testing.T) {
	if err := RequireHostConfigured("https://x"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err := RequireHostConfigured("  ")
	if lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("category = %v, want validation", lterrors.CategoryOf(err))
	}
}

func TestRequireFlag(t *testing.T) {
	if err := RequireFlag(true, "x"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err := RequireFlag(false, "scene-id")
	if err == nil || !strings.Contains(err.Error(), "--scene-id is required") {
		t.Errorf("unexpected: %v", err)
	}
	if lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("category = %v, want validation", lterrors.CategoryOf(err))
	}
}

func TestConfirmDestructive(t *testing.T) {
	if err := ConfirmDestructive(true, "delete x"); err != nil {
		t.Errorf("unexpected error with --yes: %v", err)
	}
	err := ConfirmDestructive(false, "delete x")
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Errorf("category = %v, want confirmation", lterrors.CategoryOf(err))
	}
	if output.ExitCodeOf(err) != output.ExitConfirmationRequired {
		t.Errorf("exit = %d, want %d", output.ExitCodeOf(err), output.ExitConfirmationRequired)
	}
}

func TestParseDataObject(t *testing.T) {
	m, err := ParseDataObject(`{"a":1}`)
	if err != nil || m["a"].(float64) != 1 {
		t.Errorf("parse = %v, %v", m, err)
	}
	if _, err := ParseDataObject("{bad"); lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error for bad JSON, got %v", err)
	}
}

func TestWriteResponse(t *testing.T) {
	var buf bytes.Buffer
	f := &Factory{
		Config:    &config.Config{},
		IOStreams: &output.IOStreams{Out: &buf},
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().String("format", "json", "")
	if err := f.WriteResponse(cmd, []byte(`{"ok":true}`), "factory.list"); err != nil {
		t.Fatalf("WriteResponse error = %v", err)
	}
	if !strings.Contains(buf.String(), "ok") {
		t.Errorf("output missing payload: %s", buf.String())
	}

	// Invalid JSON should surface an error.
	if err := f.WriteResponse(cmd, []byte("{bad"), "x"); err == nil {
		t.Error("expected error for invalid response JSON")
	}
}

func TestInstallHelpFuncAndNewWriter(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	InstallHelpFunc(root) // must not panic and must keep a help func
	if root.HelpFunc() == nil {
		t.Error("help func is nil after InstallHelpFunc")
	}

	var buf bytes.Buffer
	f := &Factory{Config: &config.Config{}, IOStreams: &output.IOStreams{Out: &buf}}
	w := f.NewWriter(output.FormatJSON, "x.y")
	if w == nil {
		t.Fatal("NewWriter returned nil")
	}
	if err := w.Write(map[string]interface{}{"k": "v"}); err != nil {
		t.Fatalf("Write error = %v", err)
	}
	if !strings.Contains(buf.String(), "\"k\"") {
		t.Errorf("output missing payload: %s", buf.String())
	}
}

func TestNewDefault(t *testing.T) {
	f := NewDefault()
	if f == nil || f.Config == nil || f.IOStreams == nil {
		t.Fatal("NewDefault returned an incomplete factory")
	}
	if f.IOStreams.Out == nil || f.IOStreams.In == nil {
		t.Error("NewDefault did not wire IO streams")
	}
}

func TestEffectiveProfile(t *testing.T) {
	// --profile flag wins.
	f := &Factory{Profile: "flagged", Config: &config.Config{CurrentProfile: "persisted"}}
	if got := f.EffectiveProfile(); got != "flagged" {
		t.Errorf("EffectiveProfile = %q, want flagged", got)
	}

	// Falls back to persisted CurrentProfile.
	f = &Factory{Config: &config.Config{CurrentProfile: "persisted"}}
	if got := f.EffectiveProfile(); got != "persisted" {
		t.Errorf("EffectiveProfile = %q, want persisted", got)
	}

	// Empty when nothing set.
	f = &Factory{Config: &config.Config{}}
	if got := f.EffectiveProfile(); got != "" {
		t.Errorf("EffectiveProfile = %q, want empty", got)
	}

	// Safe with nil config.
	f = &Factory{}
	if got := f.EffectiveProfile(); got != "" {
		t.Errorf("EffectiveProfile with nil config = %q, want empty", got)
	}
}

func TestNewWriterNoticeReady(t *testing.T) {
	var buf bytes.Buffer
	ch := make(chan *output.Notice, 1)
	ch <- &output.Notice{Announcement: "hello"}
	f := &Factory{
		Config:     &config.Config{},
		IOStreams:  &output.IOStreams{Out: &buf},
		Envelope:   true,
		NoticeChan: ch,
	}
	// Notice is ready on the channel, so it should be consumed.
	_ = f.NewWriter(output.FormatJSON, "x.y")
	if f.Notice == nil || f.Notice.Announcement != "hello" {
		t.Errorf("expected ready notice to be consumed, got %v", f.Notice)
	}
}

func TestNewWriterNoticeNotReady(t *testing.T) {
	var buf bytes.Buffer
	ch := make(chan *output.Notice, 1) // empty: not ready
	f := &Factory{
		Config:     &config.Config{},
		IOStreams:  &output.IOStreams{Out: &buf},
		Envelope:   true,
		NoticeChan: ch,
	}
	// Must not block; notice stays nil.
	_ = f.NewWriter(output.FormatJSON, "x.y")
	if f.Notice != nil {
		t.Errorf("expected no notice when channel not ready, got %v", f.Notice)
	}
}
