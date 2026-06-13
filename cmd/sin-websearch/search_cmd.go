// Purpose: Search command for multi-source web search.
// Docs: cmd/sin-websearch/search_cmd.doc.md
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Multi-source web search",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orch, err := buildOrchestrator()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			res, err := orch.Search(ctx, args[0])
			if err != nil {
				return err
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			for _, r := range res.Results {
				fmt.Printf("[%s] %s\n  %s\n  %s\n\n", r.Source, r.Title, r.URL, r.Snippet)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
