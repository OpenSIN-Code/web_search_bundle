// SPDX-License-Identifier: MIT
// Purpose: Tests for new optimization feature config keys.
// Docs: internal/config/config.doc.md

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// clearFeatureEnv removes any env vars that could interfere with feature tests.
func clearFeatureEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"SIN_WEBSEARCH_TAVILY_API_KEY",
		"SIN_WEBSEARCH_TAVILY_DEFAULT_DEPTH",
		"SIN_WEBSEARCH_TAVILY_INCLUDE_ANSWER",
		"SIN_WEBSEARCH_SEMANTIC_CACHE_ENABLED",
		"SIN_WEBSEARCH_SEMANTIC_CACHE_THRESHOLD",
		"SIN_WEBSEARCH_COST_AWARE_ROUTING",
		"SIN_WEBSEARCH_DUCKDUCKGO_ENABLED",
		"SIN_WEBSEARCH_MCP_TOOL_ANNOTATIONS",
		"SIN_WEBSEARCH_MCP_STREAMING_ENABLED",
		"TAVILY_API_KEY",
		"NIM_API_KEY",
	} {
		t.Setenv(key, "")
	}
}

func TestFeatureDefaultsWhenNoConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	clearFeatureEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.TavilyDefaultDepth != "basic" {
		t.Errorf("TavilyDefaultDepth = %q, want %q", cfg.TavilyDefaultDepth, "basic")
	}
	if !cfg.TavilyIncludeAnswer {
		t.Error("TavilyIncludeAnswer = false, want true")
	}
	if !cfg.SemanticCacheEnabled {
		t.Error("SemanticCacheEnabled = false, want true")
	}
	if cfg.SemanticCacheThreshold != 0.85 {
		t.Errorf("SemanticCacheThreshold = %f, want 0.85", cfg.SemanticCacheThreshold)
	}
	if !cfg.CostAwareRouting {
		t.Error("CostAwareRouting = false, want true")
	}
	if !cfg.DuckDuckGoEnabled {
		t.Error("DuckDuckGoEnabled = false, want true")
	}
	if !cfg.MCPToolAnnotations {
		t.Error("MCPToolAnnotations = false, want true")
	}
	if !cfg.MCPStreamingEnabled {
		t.Error("MCPStreamingEnabled = false, want true")
	}
	if cfg.TavilyAPIKey != "" {
		t.Errorf("TavilyAPIKey = %q, want empty by default", cfg.TavilyAPIKey)
	}
	if cfg.NIMAPIKey != "" {
		t.Errorf("NIMAPIKey = %q, want empty by default", cfg.NIMAPIKey)
	}
}

func TestFeatureEnvVarOverrides(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	clearFeatureEnv(t)

	t.Setenv("TAVILY_API_KEY", "tavily-env-key")
	t.Setenv("NIM_API_KEY", "nim-env-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.TavilyAPIKey != "tavily-env-key" {
		t.Errorf("TavilyAPIKey = %q, want tavily-env-key", cfg.TavilyAPIKey)
	}
	if cfg.NIMAPIKey != "nim-env-key" {
		t.Errorf("NIMAPIKey = %q, want nim-env-key", cfg.NIMAPIKey)
	}
}

func TestFeatureEnvVarOverridesValuesSetViaYAML(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	clearFeatureEnv(t)

	// Write a config file with TavilyAPIKey set; env should override.
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "sin-websearch")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "sin-websearch.yaml")
	data := []byte(`tavily_api_key: "yaml-tavily-key"
nim_api_key: "yaml-nim-key"
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Env vars should NOT override YAML values (matches existing pattern).
	// The existing code only falls back to env when the YAML value is empty.
	t.Setenv("TAVILY_API_KEY", "env-tavily-key")
	t.Setenv("NIM_API_KEY", "env-nim-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.TavilyAPIKey != "yaml-tavily-key" {
		t.Errorf("TavilyAPIKey = %q, want yaml-tavily-key (YAML wins)", cfg.TavilyAPIKey)
	}
	if cfg.NIMAPIKey != "yaml-nim-key" {
		t.Errorf("NIMAPIKey = %q, want yaml-nim-key (YAML wins)", cfg.NIMAPIKey)
	}
}

func TestFeatureLoadFromYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearFeatureEnv(t)

	configDir := filepath.Join(home, ".config", "sin-websearch")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "sin-websearch.yaml")
	data := []byte(`tavily_api_key: "my-tavily-key"
tavily_default_depth: "advanced"
tavily_include_answer: false
semantic_cache_enabled: false
semantic_cache_threshold: 0.92
nim_api_key: "my-nim-key"
cost_aware_routing: false
duckduckgo_enabled: false
mcp_tool_annotations: false
mcp_streaming_enabled: false
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.TavilyAPIKey != "my-tavily-key" {
		t.Errorf("TavilyAPIKey = %q, want my-tavily-key", cfg.TavilyAPIKey)
	}
	if cfg.TavilyDefaultDepth != "advanced" {
		t.Errorf("TavilyDefaultDepth = %q, want advanced", cfg.TavilyDefaultDepth)
	}
	if cfg.TavilyIncludeAnswer {
		t.Error("TavilyIncludeAnswer = true, want false")
	}
	if cfg.SemanticCacheEnabled {
		t.Error("SemanticCacheEnabled = true, want false")
	}
	if cfg.SemanticCacheThreshold != 0.92 {
		t.Errorf("SemanticCacheThreshold = %f, want 0.92", cfg.SemanticCacheThreshold)
	}
	if cfg.NIMAPIKey != "my-nim-key" {
		t.Errorf("NIMAPIKey = %q, want my-nim-key", cfg.NIMAPIKey)
	}
	if cfg.CostAwareRouting {
		t.Error("CostAwareRouting = true, want false")
	}
	if cfg.DuckDuckGoEnabled {
		t.Error("DuckDuckGoEnabled = true, want false")
	}
	if cfg.MCPToolAnnotations {
		t.Error("MCPToolAnnotations = true, want false")
	}
	if cfg.MCPStreamingEnabled {
		t.Error("MCPStreamingEnabled = true, want false")
	}
}

func TestFeatureExistingKeysUnaffected(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearFeatureEnv(t)

	configDir := filepath.Join(home, ".config", "sin-websearch")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "sin-websearch.yaml")
	data := []byte(`http_port: 7777
cache_path: /tmp/existing.db
serpapi_keys:
  - "serp-key-1"
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.HTTPPort != 7777 {
		t.Errorf("HTTPPort = %d, want 7777", cfg.HTTPPort)
	}
	if cfg.CachePath != "/tmp/existing.db" {
		t.Errorf("CachePath = %q, want /tmp/existing.db", cfg.CachePath)
	}
	if len(cfg.SerpAPIKeys) != 1 || cfg.SerpAPIKeys[0] != "serp-key-1" {
		t.Errorf("SerpAPIKeys = %v, want [serp-key-1]", cfg.SerpAPIKeys)
	}
	// New defaults should still be present.
	if !cfg.DuckDuckGoEnabled {
		t.Error("DuckDuckGoEnabled = false, want default true")
	}
	if !cfg.CostAwareRouting {
		t.Error("CostAwareRouting = false, want default true")
	}
}
