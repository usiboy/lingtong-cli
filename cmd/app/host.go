// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"strings"
)

const missingHostConfigMessage = "no host configured. Run `lingtong-cli config init --host <url>` first"

func requireHostConfigured(host string) error {
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf(missingHostConfigMessage)
	}
	return nil
}
