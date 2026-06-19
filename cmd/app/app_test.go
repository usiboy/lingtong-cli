// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNewCmdApp_HelpIncludesReadOnlyQueryCommands(t *testing.T) {
	cmd := NewCmdApp(newTestFactory("http://example.com"))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected help error: %v", err)
	}

	help := out.String()
	if !strings.Contains(help, "List, get, export, import, validate, scaffold, diff, and scene") {
		t.Errorf("expected app help to describe all app operations, got:\n%s", help)
	}

	for _, command := range []string{
		"list",
		"get",
		"export",
		"import",
		"validate",
		"scaffold",
		"diff",
		"scene",
	} {
		if !strings.Contains(help, "\n  "+command) {
			t.Errorf("expected app help to include %q command entry, got:\n%s", command, help)
		}
	}
}
