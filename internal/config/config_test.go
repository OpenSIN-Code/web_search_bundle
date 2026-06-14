// Purpose: Hermetic unit tests for configuration loading.
// Docs: config_test.doc.md

package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestDefaultCachePath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got := defaultCachePath()
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %s", got)
	}
	if !contains(got, "sin-websearch") {
		t.Errorf("expected path to contain sin-websearch, got %s", got)
	}
}

func TestDefaultCachePathNoHome(t *testing.T) {
	// Simulate no home by setting HOME to empty.
	t.Setenv("HOME", "")
	got := defaultCachePath()
	if got != "sin-websearch.db" {
		t.Errorf("expected fallback path, got %s", got)
	}
}

func TestIsNotFound(t *testing.T) {
	if !isNotFound(viper.ConfigFileNotFoundError{}) {
		t.Error("expected true for ConfigFileNotFoundError")
	}
	if isNotFound(errors.New("other error")) {
		t.Error("expected false for generic error")
	}
}

func TestItoa(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{123, "123"},
	}
	for _, c := range cases {
		if got := itoa(c.in); got != c.want {
			t.Errorf("itoa(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLoadKeysFromEnv(t *testing.T) {
	t.Setenv("TEST_KEY", "first")
	t.Setenv("TEST_KEY_1", "one")
	t.Setenv("TEST_KEY_2", "two")
	keys := loadKeysFromEnv("TEST_KEY", 4)
	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}
	if keys[0] != "first" {
		t.Errorf("expected first, got %s", keys[0])
	}

	t.Setenv("TEST_KEY", "")
	keys = loadKeysFromEnv("TEST_KEY", 4)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "one" || keys[1] != "two" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set HOME to a temp dir so no config file is found.
	t.Setenv("HOME", t.TempDir())
	// Clear any pre-existing env vars that could interfere.
	for _, key := range []string{"SIN_WEBSEARCH_HTTP_PORT", "SIN_WEBSEARCH_BRAVE_API_KEY", "SERPAPI_KEY", "BRAVE_API_KEY"} {
		t.Setenv(key, "")
	}

	// The codebase primarily relies on the legacy env var names for keys.
	t.Setenv("SIN_WEBSEARCH_HTTP_PORT", "9999")
	t.Setenv("BRAVE_API_KEY", "brave-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.HTTPPort != 9999 {
		t.Errorf("HTTPPort = %d, want 9999", cfg.HTTPPort)
	}
	if cfg.BraveAPIKey != "brave-key" {
		t.Errorf("BraveAPIKey = %q, want brave-key", cfg.BraveAPIKey)
	}
	if cfg.MCPPort != 8788 {
		t.Errorf("MCPPort = %d, want default 8788", cfg.MCPPort)
	}
}

func TestLoadFromFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, key := range []string{"SIN_WEBSEARCH_HTTP_PORT", "SIN_WEBSEARCH_CACHE_PATH", "SERPAPI_KEY"} {
		t.Setenv(key, "")
	}

	configDir := filepath.Join(home, ".config", "sin-websearch")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "sin-websearch.yaml")
	data := []byte("http_port: 7777\ncache_path: /tmp/test.db\n")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.HTTPPort != 7777 {
		t.Errorf("HTTPPort = %d, want 7777", cfg.HTTPPort)
	}
	if cfg.CachePath != "/tmp/test.db" {
		t.Errorf("CachePath = %q, want /tmp/test.db", cfg.CachePath)
	}
}

func TestLoadBackwardsCompatibleEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, key := range []string{"SIN_WEBSEARCH_BRAVE_API_KEY", "BRAVE_API_KEY", "SERPAPI_KEY"} {
		t.Setenv(key, "")
	}

	t.Setenv("BRAVE_API_KEY", "legacy-brave")
	t.Setenv("SERPAPI_KEY", "legacy-serp")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.BraveAPIKey != "legacy-brave" {
		t.Errorf("BraveAPIKey = %q, want legacy-brave", cfg.BraveAPIKey)
	}
	if len(cfg.SerpAPIKeys) == 0 || cfg.SerpAPIKeys[0] != "legacy-serp" {
		t.Errorf("SerpAPIKeys = %v, want legacy-serp", cfg.SerpAPIKeys)
	}
}

func TestMustLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := MustLoad()
	if cfg == nil {
		t.Fatal("expected config")
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
