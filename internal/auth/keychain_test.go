// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeychainUser(t *testing.T) {
	tests := []struct {
		profile string
		want    string
	}{
		{"", "default"},
		{"dev", "profile:dev"},
		{"prod", "profile:prod"},
		{"staging env", "profile:staging env"},
	}
	for _, tt := range tests {
		if got := keychainUser(tt.profile); got != tt.want {
			t.Errorf("keychainUser(%q) = %q, want %q", tt.profile, got, tt.want)
		}
	}
}

func TestKeychainUser_DefaultAndNamedDiffer(t *testing.T) {
	if keychainUser("") == keychainUser("dev") {
		t.Error("default and named profiles must map to distinct keychain accounts")
	}
}

func TestDefaultTokenRoundTrip(t *testing.T) {
	keyring.MockInit()

	if err := StoreToken("tok-default"); err != nil {
		t.Fatalf("StoreToken error = %v", err)
	}
	got, err := GetToken()
	if err != nil || got != "tok-default" {
		t.Fatalf("GetToken = (%q, %v), want tok-default", got, err)
	}
	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken error = %v", err)
	}
	if _, err := GetToken(); err == nil {
		t.Error("GetToken after delete should error")
	}
}

func TestProfileTokenIsolation(t *testing.T) {
	keyring.MockInit()

	if err := StoreTokenForProfile("dev", "tok-dev"); err != nil {
		t.Fatalf("store dev error = %v", err)
	}
	if err := StoreTokenForProfile("prod", "tok-prod"); err != nil {
		t.Fatalf("store prod error = %v", err)
	}
	if err := StoreToken("tok-default"); err != nil {
		t.Fatalf("store default error = %v", err)
	}

	// Each profile reads back its own token.
	if got, _ := GetTokenForProfile("dev"); got != "tok-dev" {
		t.Errorf("dev token = %q, want tok-dev", got)
	}
	if got, _ := GetTokenForProfile("prod"); got != "tok-prod" {
		t.Errorf("prod token = %q, want tok-prod", got)
	}
	if got, _ := GetToken(); got != "tok-default" {
		t.Errorf("default token = %q, want tok-default", got)
	}

	// Deleting one profile leaves the others intact.
	if err := DeleteTokenForProfile("dev"); err != nil {
		t.Fatalf("delete dev error = %v", err)
	}
	if _, err := GetTokenForProfile("dev"); err == nil {
		t.Error("dev token should be gone after delete")
	}
	if got, _ := GetTokenForProfile("prod"); got != "tok-prod" {
		t.Errorf("prod token should survive dev deletion, got %q", got)
	}
}
