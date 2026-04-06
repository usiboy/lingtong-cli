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

// StoreToken stores the JWT token in OS keychain.
func StoreToken(token string) error {
	return keyring.Set(serviceName, userName, token)
}

// GetToken retrieves the JWT token from OS keychain.
func GetToken() (string, error) {
	return keyring.Get(serviceName, userName)
}

// DeleteToken removes the JWT token from OS keychain.
func DeleteToken() error {
	return keyring.Delete(serviceName, userName)
}
