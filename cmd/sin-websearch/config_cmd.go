// SPDX-License-Identifier: MIT
// Purpose: Config command to inspect and initialize configuration.
// Docs: cmd/sin-websearch/config_cmd.doc.md
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and manage configuration",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show the default configuration path",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			fmt.Println(filepath.Join(home, ".config", "sin-websearch", "sin-websearch.yaml"))
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "example",
		Short: "Print an example configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			example := `# sin-websearch configuration
serpapi_keys:
  - "${SERPAPI_KEY_1}"
brave_api_key: "${BRAVE_API_KEY}"
openrouter_api_key: "${OPENROUTER_API_KEY}"
scrapecreators_api_key: "${SCRAPECREATORS_API_KEY}"
groq_api_key: "${GROQ_API_KEY}"
openai_api_key: "${OPENAI_API_KEY}"
http_port: 8787
mcp_port: 8788
searxng_urls:
  - "http://localhost:8080"
`
			fmt.Print(example)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "load",
		Short: "Try loading the configuration and report status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				fmt.Printf("Config load error: %v\n", err)
				return nil
			}
			fmt.Printf("Loaded config\nSerpAPI keys: %d\nBrave: %v\nOpenRouter: %v\nHTTP port: %d\n",
				len(cfg.SerpAPIKeys), cfg.BraveAPIKey != "", cfg.OpenRouterKey != "", cfg.HTTPPort)
			return nil
		},
	})
	return cmd
}
