// SPDX-License-Identifier: MIT
// Purpose: Central configuration management for sin-websearch.
// Docs: internal/config/config.doc.md
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Version is the current build version of sin-websearch. It is set at build
// time via ldflags (e.g. -X config.Version=1.2.3) and exposed by the /health
// endpoint.
var Version = "dev"

// Config holds the parsed application configuration.
type Config struct {
	SerpAPIKeys           []string          `mapstructure:"serpapi_keys"`
	BraveAPIKey           string            `mapstructure:"brave_api_key"`
	OpenRouterKey         string            `mapstructure:"openrouter_api_key"`
	ScrapeCreatorsKey     string            `mapstructure:"scrapecreators_api_key"`
	GroqAPIKey            string            `mapstructure:"groq_api_key"`
	OpenAIAPIKey          string            `mapstructure:"openai_api_key"`
	CachePath             string            `mapstructure:"cache_path"`
	HTTPPort              int               `mapstructure:"http_port"`
	MCPPort               int               `mapstructure:"mcp_port"`
	Token                 string            `mapstructure:"token"`
	SearxNGURLs           []string          `mapstructure:"searxng_urls"`
	Defaults              map[string]string `mapstructure:"defaults"`
	RateLimitRPS          float64           `mapstructure:"rate_limit_rps"`
	RateLimitBurst        int               `mapstructure:"rate_limit_burst"`
	DisableRequestLogging bool              `mapstructure:"disable_request_logging"`
}

// Load reads the configuration from the default paths and environment variables.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("sin-websearch")
	v.SetConfigType("yaml")

	// Search config paths
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "sin-websearch"))
		v.AddConfigPath(filepath.Join(home, ".sin-websearch"))
	}
	v.AddConfigPath(".")

	// Defaults
	v.SetDefault("cache_path", defaultCachePath())
	v.SetDefault("http_port", 8787)
	v.SetDefault("mcp_port", 8788)
	v.SetDefault("searxng_urls", []string{})
	v.SetDefault("defaults", map[string]string{})
	v.SetDefault("rate_limit_rps", 10.0)
	v.SetDefault("rate_limit_burst", 20)
	v.SetDefault("disable_request_logging", false)

	v.SetEnvPrefix("SIN_WEBSEARCH")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil && !isNotFound(err) {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Backwards-compatible env var loading for keys.
	cfg.SerpAPIKeys = append(cfg.SerpAPIKeys, loadKeysFromEnv("SERPAPI_KEY", 8)...)
	if cfg.BraveAPIKey == "" {
		cfg.BraveAPIKey = os.Getenv("BRAVE_API_KEY")
	}
	if cfg.OpenRouterKey == "" {
		cfg.OpenRouterKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if cfg.ScrapeCreatorsKey == "" {
		cfg.ScrapeCreatorsKey = os.Getenv("SCRAPECREATORS_API_KEY")
	}
	if cfg.GroqAPIKey == "" {
		cfg.GroqAPIKey = os.Getenv("GROQ_API_KEY")
	}
	if cfg.OpenAIAPIKey == "" {
		cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.Token == "" {
		cfg.Token = os.Getenv("SIN_WEBSEARCH_TOKEN")
	}

	return &cfg, nil
}

func defaultCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "sin-websearch.db"
	}
	return filepath.Join(home, ".local", "share", "sin-websearch", "sin-websearch.db")
}

func isNotFound(err error) bool {
	_, ok := err.(viper.ConfigFileNotFoundError)
	return ok
}

func loadKeysFromEnv(prefix string, max int) []string {
	var keys []string
	for i := 1; i <= max; i++ {
		k := os.Getenv(prefix)
		if k != "" {
			keys = append(keys, k)
			break
		}
		k = os.Getenv(prefix + "_" + itoa(i))
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

func itoa(i int) string {
	// Minimal int-to-string for env loading without strconv in hot path.
	if i == 0 {
		return "0"
	}
	var digits []byte
	n := i
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// MustLoad loads configuration and panics on error. Useful for CLI commands.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
