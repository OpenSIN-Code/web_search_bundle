// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the SQLite cache.
// Docs: cache_test.doc.md

package cache

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cache.db")
	c, err := New(path)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestHashKey(t *testing.T) {
	a := HashKey("query", []string{"reddit", "hackernews"})
	b := HashKey("query", []string{"hackernews", "reddit"})
	if a == b {
		t.Error("expected different hashes for different source order")
	}
	if a == "" {
		t.Error("expected non-empty hash")
	}
	c := HashKey("query", []string{"reddit", "hackernews"})
	if a != c {
		t.Error("expected same hash for same inputs")
	}
}

func TestCacheSetGet(t *testing.T) {
	c := newTestCache(t)
	payload := map[string]string{"title": "test"}
	if err := c.Set("key1", []string{"reddit"}, payload, time.Hour); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	data, found, err := c.Get("key1")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !found {
		t.Fatal("expected key to be found")
	}
	if !contains(string(data), "title") {
		t.Errorf("expected JSON payload, got %s", data)
	}
}

func TestCacheGetMissing(t *testing.T) {
	c := newTestCache(t)
	data, found, err := c.Get("missing")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if found {
		t.Error("expected key not found")
	}
	if data != nil {
		t.Errorf("expected nil data, got %s", data)
	}
}

func TestCacheGetExpired(t *testing.T) {
	c := newTestCache(t)
	if err := c.Set("expired", []string{"reddit"}, "value", -time.Hour); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	_, found, err := c.Get("expired")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if found {
		t.Error("expected expired key not found")
	}
}

func TestCacheSetVideo(t *testing.T) {
	c := newTestCache(t)
	payload := map[string]string{"url": "https://video.com"}
	if err := c.SetVideo("vkey", "https://video.com", payload, time.Hour); err != nil {
		t.Fatalf("SetVideo error: %v", err)
	}
	data, found, err := c.GetVideo("vkey")
	if err != nil {
		t.Fatalf("GetVideo error: %v", err)
	}
	if !found {
		t.Fatal("expected video key found")
	}
	if !contains(string(data), "url") {
		t.Errorf("expected JSON payload, got %s", data)
	}
}

func TestCacheGetVideoMissing(t *testing.T) {
	c := newTestCache(t)
	data, found, err := c.GetVideo("missing")
	if err != nil {
		t.Fatalf("GetVideo error: %v", err)
	}
	if found {
		t.Error("expected video key not found")
	}
	if data != nil {
		t.Errorf("expected nil data, got %s", data)
	}
}

func TestCacheStats(t *testing.T) {
	c := newTestCache(t)
	if err := c.Set("k1", []string{"reddit"}, "v1", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := c.SetVideo("k2", "url", "v2", time.Hour); err != nil {
		t.Fatal(err)
	}
	search, video, err := c.Stats()
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}
	if search != 1 {
		t.Errorf("search count = %d, want 1", search)
	}
	if video != 1 {
		t.Errorf("video count = %d, want 1", video)
	}
}

func TestCacheClear(t *testing.T) {
	c := newTestCache(t)
	if err := c.Set("k1", []string{"reddit"}, "v1", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := c.SetVideo("k2", "url", "v2", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear error: %v", err)
	}
	search, video, err := c.Stats()
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}
	if search != 0 || video != 0 {
		t.Errorf("expected empty cache, got %d search, %d video", search, video)
	}
}

func TestCacheCompact(t *testing.T) {
	c := newTestCache(t)
	if err := c.Set("old", []string{"reddit"}, "v", -time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := c.Set("new", []string{"reddit"}, "v", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := c.Compact(); err != nil {
		t.Fatalf("Compact error: %v", err)
	}
	search, _, err := c.Stats()
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}
	if search != 1 {
		t.Errorf("expected 1 entry after compact, got %d", search)
	}
}

func TestCacheString(t *testing.T) {
	c := newTestCache(t)
	if err := c.Set("k1", []string{"reddit"}, "v1", time.Hour); err != nil {
		t.Fatal(err)
	}
	s := c.String()
	if !contains(s, "1 search entries") {
		t.Errorf("unexpected String: %s", s)
	}
}

func TestCacheClose(t *testing.T) {
	c := newTestCache(t)
	if err := c.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
