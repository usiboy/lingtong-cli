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

// StoreTokenForAuth stores a token for a named auth identity within a profile.
// profile: profile name ("" for default)
// label: auth identity label (e.g. "company-a")
func StoreTokenForAuth(profile, label, token string) error {
	return keyring.Set(serviceName, authKeychainUser(profile, label), token)
}

// GetTokenForAuth retrieves a token for a named auth identity within a profile.
func GetTokenForAuth(profile, label string) (string, error) {
	return keyring.Get(serviceName, authKeychainUser(profile, label))
}

// DeleteTokenForAuth removes a token for a named auth identity within a profile.
func DeleteTokenForAuth(profile, label string) error {
	return keyring.Delete(serviceName, authKeychainUser(profile, label))
}

// authKeychainUser returns the keychain key for a profile+auth combination.
//
//	default profile + no label  → "default"              (backward compat)
//	default profile + label     → "auth:<label>"
//	named profile + no label    → "profile:<name>"        (backward compat)
//	named profile + label       → "profile:<name>:auth:<label>"
func authKeychainUser(profile, label string) string {
	if profile == "" && label == "" {
		return userName // "default"
	}
	if profile == "" {
		return "auth:" + label
	}
	if label == "" {
		return "profile:" + profile
	}
	return "profile:" + profile + ":auth:" + label
}
