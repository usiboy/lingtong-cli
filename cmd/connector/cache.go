// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package connector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConnectorCache stores cached connector invoke workflow info.
type ConnectorCache struct {
	// Universal: 通用工作流缓存（按环境区分，最多2个：test/prod）
	Universal map[string]*ConnectorCacheEntry `json:"universal"`
	// Legacy: 兼容旧版缓存（迁移期保留）
	Legacy map[string]*ConnectorCacheEntry `json:"legacy,omitempty"`
}

// ConnectorCacheEntry stores a single cached universal workflow.
type ConnectorCacheEntry struct {
	WorkflowId int    `json:"workflowId"`
	AppTag     string `json:"appTag"`
	Env        string `json:"env"`
}

// legacyCacheEntry is used for backward-compatible migration from old cache format.
type legacyCacheEntry struct {
	WorkflowId  int    `json:"workflowId"`
	AppTag      string `json:"appTag"`
	Connector   string `json:"connector"`
	Method      string `json:"method"`
	AuthAccount string `json:"authAccount"`
	Env         string `json:"env"`
}

// legacyCache represents the old cache file format.
type legacyCache struct {
	Entries map[string]*legacyCacheEntry `json:"entries"`
}

// cacheFilePath returns the path to the connector cache file.
func cacheFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".lingtong-cli", "connector-cache.json"), nil
}

// loadCache loads the connector cache from file, with backward-compatible migration.
func loadCache() (*ConnectorCache, error) {
	path, err := cacheFilePath()
	if err != nil {
		return emptyCache(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return emptyCache(), nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Try new format first
	var cache ConnectorCache
	if err := json.Unmarshal(data, &cache); err == nil && (cache.Universal != nil || cache.Legacy != nil) {
		if cache.Universal == nil {
			cache.Universal = make(map[string]*ConnectorCacheEntry)
		}
		if cache.Legacy == nil {
			cache.Legacy = make(map[string]*ConnectorCacheEntry)
		}
		return &cache, nil
	}

	// Migrate from old format: {"entries": {...}}
	var oldCache legacyCache
	if err := json.Unmarshal(data, &oldCache); err == nil && oldCache.Entries != nil {
		legacy := make(map[string]*ConnectorCacheEntry)
		for key, entry := range oldCache.Entries {
			legacy[key] = &ConnectorCacheEntry{
				WorkflowId: entry.WorkflowId,
				AppTag:     entry.AppTag,
				Env:        entry.Env,
			}
		}
		return &ConnectorCache{
			Universal: make(map[string]*ConnectorCacheEntry),
			Legacy:    legacy,
		}, nil
	}

	return emptyCache(), nil
}

// saveCache saves the connector cache to file.
func saveCache(cache *ConnectorCache) error {
	path, err := cacheFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize cache: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// emptyCache returns a new empty cache.
func emptyCache() *ConnectorCache {
	return &ConnectorCache{
		Universal: make(map[string]*ConnectorCacheEntry),
		Legacy:    make(map[string]*ConnectorCacheEntry),
	}
}

// universalCacheKey generates a cache key for universal workflow (by env only).
func universalCacheKey(env string) string {
	return env
}
