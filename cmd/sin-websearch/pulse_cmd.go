// Purpose: Pulse command for social sentiment search.
// Docs: cmd/sin-websearch/pulse_cmd.doc.md
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func newPulseCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "pulse [topic]",
		Short: "Social pulse search (engagement-focused)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orch, err := buildOrchestrator()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			res, err := orch.Pulse(ctx, args[0])
			if err != nil {
				return err
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			fmt.Printf("Pulse for %s\n\n", args[0])
			for _, r := range res.Results {
				fmt.Printf("[%s] %d ⬆  %s\n  %s\n\n", r.Source, r.Engagement, r.Title, r.URL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
