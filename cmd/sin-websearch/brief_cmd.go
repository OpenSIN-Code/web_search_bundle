// Purpose: Brief command to generate a research summary.
// Docs: cmd/sin-websearch/brief_cmd.doc.md
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newBriefCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brief [topic]",
		Short: "Generate a research briefing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orch, err := buildOrchestrator()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			res, err := orch.Search(ctx, args[0])
			if err != nil {
				return err
			}
			fmt.Printf("# Brief: %s\n\n", args[0])
			fmt.Printf("Entity: %s\n", res.Entity.Query)
			fmt.Printf("Clusters: %d | Results: %d\n\n", len(res.Clusters), len(res.Results))
			fmt.Println("## Best Takes")
			for _, take := range res.BestTakes {
				fmt.Printf("- %s\n", take)
			}
			fmt.Println("\n## Top Results")
			for i, r := range res.Results {
				if i >= 10 {
					break
				}
				fmt.Printf("- [%s] %s\n", r.Source, r.Title)
			}
			return nil
		},
	}
	return cmd
}
