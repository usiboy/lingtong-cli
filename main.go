// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT
//
// lingtong-cli — Lingtong iPaaS CLI tool for AI Agents.
package main

import (
	"os"

	"github.com/lingtong/cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
