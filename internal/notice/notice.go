// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package notice provides the Notice system for injecting system-level
// notifications into CLI output envelopes. Today it surfaces CLI update
// availability; the Notice struct also carries an Announcement field reserved
// for platform-pushed messages once a server endpoint exists.
//
// Fetching is best-effort and must never slow a command down: callers run it on
// a goroutine and inject the result only if it is ready (see cmdutil.Factory).
// To avoid hammering the network on every invocation, results are cached on
// disk for CacheTTL.
package notice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lingtong/cli/internal/build"
	"github.com/lingtong/cli/internal/output"
	"github.com/lingtong/cli/internal/version"
)

const defaultTimeout = 2 * time.Second

// defaultCheckURL is the release endpoint consulted for update availability.
const defaultCheckURL = "https://api.github.com/repos/lingtong/cli/releases/latest"

// Tunables exposed as package variables so tests can redirect them and so the
// check endpoint can be overridden via the LINGTONG_UPDATE_URL env var.
var (
	// CheckURL is the endpoint queried for the latest release.
	CheckURL = defaultCheckURL
	// CacheTTL is how long a fetched (or empty) result is reused before the
	// network is consulted again.
	CacheTTL = 24 * time.Hour
	// CacheDir overrides where the cache file is written; empty resolves to
	// ~/.lingtong-cli.
	CacheDir = ""
	// currentVersion is the running CLI version; a variable so tests can set it.
	currentVersion = build.Version
)

// cacheFileName is the on-disk cache for the last notice check.
const cacheFileName = "notice-cache.json"

// cacheEntry is the persisted shape of a notice check.
type cacheEntry struct {
	CheckedAt time.Time      `json:"checkedAt"`
	Notice    *output.Notice `json:"notice,omitempty"`
}

// FetchNotices returns the current system notice, or nil when there is nothing
// to report. It reads from the disk cache when fresh and otherwise performs a
// short, bounded network check and refreshes the cache. All failures degrade to
// nil — notices are non-critical.
func FetchNotices(ctx context.Context) *output.Notice {
	if entry, ok := readCache(); ok && time.Since(entry.CheckedAt) < CacheTTL {
		return entry.Notice
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	notice := &output.Notice{}
	if update := checkForUpdates(ctx); update != nil {
		notice.Update = update
	}

	if notice.Update == nil && notice.Announcement == "" {
		notice = nil
	}

	writeCache(cacheEntry{CheckedAt: time.Now(), Notice: notice})
	return notice
}

// checkForUpdates checks whether a newer CLI version is available.
func checkForUpdates(ctx context.Context) *output.UpdateNotice {
	if currentVersion == "" || currentVersion == "dev" {
		return nil
	}

	url := CheckURL
	if env := os.Getenv("LINGTONG_UPDATE_URL"); env != "" {
		url = env
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	resp, err := (&http.Client{Timeout: defaultTimeout}).Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}
	if release.TagName == "" {
		return nil
	}

	if version.Compare(currentVersion, release.TagName) < 0 {
		return &output.UpdateNotice{
			Version: release.TagName,
			URL:     fmt.Sprintf("https://github.com/lingtong/cli/releases/tag/%s", release.TagName),
		}
	}
	return nil
}

// cachePath returns the cache file location, or "" when it cannot be resolved.
func cachePath() string {
	dir := CacheDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".lingtong-cli")
	}
	return filepath.Join(dir, cacheFileName)
}

// readCache loads the cached entry; ok is false when absent or unreadable.
func readCache() (cacheEntry, bool) {
	path := cachePath()
	if path == "" {
		return cacheEntry{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cacheEntry{}, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, false
	}
	return entry, true
}

// writeCache persists the entry, best-effort (errors are ignored).
func writeCache(entry cacheEntry) {
	path := cachePath()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
}
