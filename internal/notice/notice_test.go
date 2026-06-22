// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package notice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// withIsolatedEnv points the cache at a temp dir and restores package vars.
func withIsolatedEnv(t *testing.T, version, url string, ttl time.Duration) {
	t.Helper()
	origURL, origTTL, origDir, origVer := CheckURL, CacheTTL, CacheDir, currentVersion
	t.Cleanup(func() {
		CheckURL, CacheTTL, CacheDir, currentVersion = origURL, origTTL, origDir, origVer
	})
	CacheDir = t.TempDir()
	CheckURL = url
	CacheTTL = ttl
	currentVersion = version
	// Ensure no stray env override leaks in.
	t.Setenv("LINGTONG_UPDATE_URL", "")
}

func TestFetchNotices_DevVersionSkips(t *testing.T) {
	withIsolatedEnv(t, "dev", "http://127.0.0.1:0", time.Hour)
	if n := FetchNotices(context.Background()); n != nil {
		t.Errorf("dev build should not produce a notice, got %v", n)
	}
}

func TestFetchNotices_UpdateAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v99.0.0"}`))
	}))
	defer srv.Close()
	withIsolatedEnv(t, "1.0.0", srv.URL, time.Hour)

	n := FetchNotices(context.Background())
	if n == nil || n.Update == nil {
		t.Fatalf("expected an update notice, got %v", n)
	}
	if n.Update.Version != "v99.0.0" {
		t.Errorf("Update.Version = %q, want v99.0.0", n.Update.Version)
	}
	if n.Update.URL == "" {
		t.Error("Update.URL should not be empty")
	}
}

func TestFetchNotices_UpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	defer srv.Close()
	withIsolatedEnv(t, "1.0.0", srv.URL, time.Hour)

	if n := FetchNotices(context.Background()); n != nil {
		t.Errorf("up-to-date build should produce no notice, got %v", n)
	}
}

func TestFetchNotices_Unreachable(t *testing.T) {
	withIsolatedEnv(t, "1.0.0", "http://127.0.0.1:0", time.Hour)
	if n := FetchNotices(context.Background()); n != nil {
		t.Errorf("unreachable endpoint should degrade to nil, got %v", n)
	}
}

func TestFetchNotices_UsesCache(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"tag_name":"v99.0.0"}`))
	}))
	defer srv.Close()
	withIsolatedEnv(t, "1.0.0", srv.URL, time.Hour)

	first := FetchNotices(context.Background())
	second := FetchNotices(context.Background())
	if first == nil || second == nil {
		t.Fatal("expected notices on both calls")
	}
	if hits != 1 {
		t.Errorf("expected the network to be hit once (cache), got %d hits", hits)
	}
}

func TestFetchNotices_CacheExpires(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"tag_name":"v99.0.0"}`))
	}))
	defer srv.Close()
	withIsolatedEnv(t, "1.0.0", srv.URL, time.Nanosecond) // immediately stale

	_ = FetchNotices(context.Background())
	time.Sleep(time.Millisecond)
	_ = FetchNotices(context.Background())
	if hits < 2 {
		t.Errorf("expired cache should trigger a refetch, got %d hits", hits)
	}
}

func TestFetchNotices_EnvOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v99.0.0"}`))
	}))
	defer srv.Close()
	withIsolatedEnv(t, "1.0.0", "http://127.0.0.1:0", time.Hour)
	t.Setenv("LINGTONG_UPDATE_URL", srv.URL)

	n := FetchNotices(context.Background())
	if n == nil || n.Update == nil {
		t.Fatalf("env override should be honored, got %v", n)
	}
}

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	origDir := CacheDir
	t.Cleanup(func() { CacheDir = origDir })
	CacheDir = dir

	writeCache(cacheEntry{CheckedAt: time.Now()})
	if _, err := os.Stat(filepath.Join(dir, cacheFileName)); err != nil {
		t.Fatalf("cache file not written: %v", err)
	}
	if _, ok := readCache(); !ok {
		t.Error("readCache should succeed after writeCache")
	}
}

func TestReadCache_Missing(t *testing.T) {
	origDir := CacheDir
	t.Cleanup(func() { CacheDir = origDir })
	CacheDir = t.TempDir() // empty
	if _, ok := readCache(); ok {
		t.Error("readCache should report not-ok when no file exists")
	}
}

func TestReadCache_Corrupt(t *testing.T) {
	origDir := CacheDir
	t.Cleanup(func() { CacheDir = origDir })
	dir := t.TempDir()
	CacheDir = dir
	if err := os.WriteFile(filepath.Join(dir, cacheFileName), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, ok := readCache(); ok {
		t.Error("readCache should report not-ok on corrupt JSON")
	}
}

func TestCachePath_DefaultsToHome(t *testing.T) {
	origDir := CacheDir
	t.Cleanup(func() { CacheDir = origDir })
	CacheDir = ""
	home := t.TempDir()
	t.Setenv("HOME", home)
	got := cachePath()
	want := filepath.Join(home, ".lingtong-cli", cacheFileName)
	if got != want {
		t.Errorf("cachePath() = %q, want %q", got, want)
	}
}
