// Purpose: Resolve command for entity resolution.
// Docs: cmd/sin-websearch/resolve_cmd.doc.md
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/resolver"
	"github.com/spf13/cobra"
)

func newResolveCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "resolve [name]",
		Short: "Resolve a name or topic to platform handles",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := resolver.NewEntityResolver()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			entity, err := r.Resolve(ctx, args[0])
			if err != nil {
				return err
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(entity)
			}
			fmt.Printf("Query: %s\n", entity.Query)
			fmt.Printf("X: %v\nGitHub Users: %v\nGitHub Repos: %v\nSubreddits: %v\n",
				entity.XHandles, entity.GitHubUsers, entity.GitHubRepos, entity.Subreddits)
			fmt.Printf("Expanded Queries: %v\n", entity.ExpandQueries())
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
