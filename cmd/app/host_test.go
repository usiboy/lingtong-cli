// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package app

import (
	"strings"
	"testing"
)

func assertAppMissingHostError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected missing host configuration error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected no host configured error, got %v", err)
	}
	if !strings.Contains(err.Error(), "lingtong-cli config init --host <url>") {
		t.Fatalf("expected config init guidance, got %v", err)
	}
	if strings.Contains(err.Error(), "unsupported protocol scheme") {
		t.Fatalf("expected friendly config error, got raw network error: %v", err)
	}
}
