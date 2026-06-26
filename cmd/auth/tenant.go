// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"encoding/json"
	"fmt"

	"github.com/lingtong/cli/internal/client"
)

// fetchTenantId retrieves the tenant ID for the given token by calling the
// /tenant/config/get endpoint via proxy. This endpoint returns tenant
// configuration including tenantId.
//
// If the request fails or returns no tenant info, an error is returned.
// The caller should decide whether to treat this as fatal or continue
// with an unknown tenant.
func fetchTenantId(host, token string) (string, error) {
	c := client.NewClient(host, token)
	// Use proxy mode (default) to access /tenant/config/get

	resp, err := c.Get("/tenant/config/get", nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tenant info: %w", err)
	}

	// Parse response: { "success": true, "result": { "tenantId": "...", ... } }
	var result struct {
		Success bool `json:"success"`
		Result  struct {
			TenantId string `json:"tenantId"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("failed to parse tenant response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("tenant info request failed")
	}

	tenantId := result.Result.TenantId
	if tenantId == "" {
		return "", fmt.Errorf("empty tenantId in response")
	}

	return tenantId, nil
}
