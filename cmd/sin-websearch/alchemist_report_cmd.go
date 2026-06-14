// Purpose: CLI command to show the alchemist report from the last run.
// Docs: alchemist_report_cmd.doc.md
package main

import (
	"fmt"
	"os"

	"github.com/OpenSIN-Code/web_search_bundle/internal/alchemist"
	"github.com/spf13/cobra"
)

func newAlchemistReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Show report from last run",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := os.Getwd()
			history, err := alchemist.NewHistory(repoPath)
			if err != nil {
				return err
			}
			defer history.Close()

			summary, err := history.Summary()
			if err != nil {
				return err
			}

			recent, err := history.Recent(20)
			if err != nil {
				return err
			}

			successRate := 0.0
			if val, ok := summary["success_rate"].(float64); ok {
				successRate = val
			}

			bestDelta := 0.0
			if val, ok := summary["best_delta"].(float64); ok {
				bestDelta = val
			}

			fmt.Println("🧪 Alchemist Last Run Report")
			fmt.Printf("============================\n\n")
			fmt.Printf("Total experiments: %v\n", summary["total_experiments"])
			fmt.Printf("✅ Committed:      %v\n", summary["committed"])
			fmt.Printf("❌ Discarded:      %v\n", summary["discarded"])
			fmt.Printf("⚠️  Errors:         %v\n", summary["errors"])
			fmt.Printf("Success rate:      %.1f%%\n", successRate*100)
			fmt.Printf("Best improvement:  %.4f\n", bestDelta)
			fmt.Printf("Total compute:     %v\n\n", summary["total_runtime"])

			fmt.Println("Recent experiments:")
			for _, e := range recent {
				fmt.Printf("  [%s] %s → %.4f\n", e.Decision, e.Hypothesis, e.MetricAfter)
			}

			return nil
		},
	}
}
