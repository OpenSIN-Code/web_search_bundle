// Purpose: Secrets command for Infisical and env-based secret loading.
// Docs: cmd/sin-websearch/secrets_cmd.doc.md
package main

import (
	"fmt"

	"github.com/OpenSIN-Code/web_search_bundle/internal/secrets"
	"github.com/spf13/cobra"
)

func newSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage secrets (Infisical + env vars)",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show secrets provider status",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := secrets.NewInfisicalClient()
			status := c.Status()
			fmt.Printf("Infisical available: %v\nCLI path: %v\nProject ID: %v\nEnvironment: %v\n",
				status["available"], status["cli_path"], status["project_id"], status["env"])
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "load",
		Short: "Load all secrets (redacted output)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := secrets.NewInfisicalClient()
			json, err := c.ExportAsJSON()
			if err != nil {
				return err
			}
			fmt.Println(json)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Verify required secrets are available",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := secrets.NewInfisicalClient()
			keys := c.LoadSerpAPIKeys()
			fmt.Printf("SerpAPI keys loaded: %d\n", len(keys))
			return nil
		},
	})
	return cmd
}
