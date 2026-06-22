// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "lingtong-cli"
	userName    = "default"
)

// StoreToken stores the JWT token for the default profile in the OS keychain.
func StoreToken(token string) error {
	return StoreTokenForProfile("", token)
}

// GetToken retrieves the JWT token for the default profile from the OS keychain.
func GetToken() (string, error) {
	return GetTokenForProfile("")
}

// DeleteToken removes the JWT token for the default profile from the OS keychain.
func DeleteToken() error {
	return DeleteTokenForProfile("")
}

// keychainUser returns the keychain account name for a profile. The empty
// profile maps to the historical "default" account so existing tokens keep
// working; named profiles get an isolated "profile:<name>" account so each
// environment carries its own credentials.
func keychainUser(profile string) string {
	if profile == "" {
		return userName
	}
	return "profile:" + profile
}

// StoreTokenForProfile stores the JWT token for a named profile.
func StoreTokenForProfile(profile, token string) error {
	return keyring.Set(serviceName, keychainUser(profile), token)
}

// GetTokenForProfile retrieves the JWT token for a named profile.
func GetTokenForProfile(profile string) (string, error) {
	return keyring.Get(serviceName, keychainUser(profile))
}

// DeleteTokenForProfile removes the JWT token for a named profile.
func DeleteTokenForProfile(profile string) error {
	return keyring.Delete(serviceName, keychainUser(profile))
}
