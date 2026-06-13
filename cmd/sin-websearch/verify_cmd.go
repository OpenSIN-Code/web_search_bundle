// Purpose: Verify command for claim verification.
// Docs: cmd/sin-websearch/verify_cmd.doc.md
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/verify"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [topic]",
		Short: "Verify claims across search results",
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
			engine := verify.NewEngine(verify.DefaultDiscipline())
			report := engine.Verify(args[0], res.Results)
			fmt.Print(report.FormatText())
			return nil
		},
	}
	return cmd
}
