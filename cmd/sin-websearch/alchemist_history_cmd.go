// Purpose: CLI command to list alchemist experiment history.
// Docs: alchemist_history_cmd.doc.md
package main

import (
	"fmt"
	"os"

	"github.com/OpenSIN-Code/web_search_bundle/internal/alchemist"
	"github.com/spf13/cobra"
)

func newAlchemistHistoryCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show experiment history",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := os.Getwd()
			history, err := alchemist.NewHistory(repoPath)
			if err != nil {
				return err
			}
			defer history.Close()

			records, err := history.Recent(limit)
			if err != nil {
				return err
			}

			fmt.Printf("%-20s %-10s %-40s %-10s %-10s\n",
				"TIME", "DECISION", "HYPOTHESIS", "METRIC", "DELTA")
			fmt.Println(string(make([]byte, 100)))
			for _, r := range records {
				hyp := r.Hypothesis
				if len(hyp) > 40 {
					hyp = hyp[:37] + "..."
				}
				fmt.Printf("%-20s %-10s %-40s %-10.4f %-+10.4f\n",
					r.Timestamp.Format("01-02 15:04"),
					r.Decision, hyp, r.MetricAfter, r.Delta)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Max records")
	return cmd
}
