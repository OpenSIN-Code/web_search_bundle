// Purpose: Benchmark configuration loading and helper functions.
// Docs: internal/config/config.doc.md
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkConfigItoa(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = itoa(i)
	}
}

func BenchmarkLoadKeysFromEnv(b *testing.B) {
	b.Setenv("SERPAPI_KEY", "test-key")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loadKeysFromEnv("SERPAPI_KEY", 8)
	}
}

func BenchmarkConfigLoad(b *testing.B) {
	tmpHome := b.TempDir()
	cfgDir := filepath.Join(tmpHome, ".config", "sin-websearch")
	if err := os.MkdirAll(cfgDir, 0750); err != nil {
		b.Fatal(err)
	}
	cfgPath := filepath.Join(cfgDir, "sin-websearch.yaml")
	content := []byte(`
serpapi_keys:
  - key1
  - key2
brave_api_key: brave-test
http_port: 9090
mcp_port: 9091
cache_path: /tmp/sin-websearch-bench.db
`)
	if err := os.WriteFile(cfgPath, content, 0600); err != nil {
		b.Fatal(err)
	}

	origHome := os.Getenv("HOME")
	b.Setenv("HOME", tmpHome)
	b.Cleanup(func() { os.Setenv("HOME", origHome) })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load()
	}
}

func BenchmarkConfigLoadDefaults(b *testing.B) {
	origHome := os.Getenv("HOME")
	b.Setenv("HOME", b.TempDir())
	b.Cleanup(func() { os.Setenv("HOME", origHome) })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load()
	}
}

func BenchmarkDefaultCachePath(b *testing.B) {
	origHome := os.Getenv("HOME")
	b.Setenv("HOME", b.TempDir())
	b.Cleanup(func() { os.Setenv("HOME", origHome) })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = defaultCachePath()
	}
}

func BenchmarkIsNotFound(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isNotFound(nil)
	}
}
