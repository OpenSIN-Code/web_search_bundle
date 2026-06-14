// SPDX-License-Identifier: MIT
// Purpose: CLI command to run a single alchemist autonomous loop.
// Docs: alchemist_run_cmd.doc.md
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/alchemist"
	"github.com/spf13/cobra"
)

func newAlchemistRunCmd() *cobra.Command {
	var (
		budget         string
		runtime        string
		targetFile     string
		metric         string
		regex          string
		higher         bool
		runCmd         string
		safety         string
		maxExperiments int
		programFile    string
		noHooks        bool
		litRefresh     int
		litProfile     string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the autonomous loop",
		RunE: func(cmd *cobra.Command, args []string) error {
			budgetDur, err := time.ParseDuration(budget)
			if err != nil {
				return fmt.Errorf("invalid budget: %w", err)
			}

			var runtimeDur time.Duration
			if runtime != "" {
				runtimeDur, err = time.ParseDuration(runtime)
				if err != nil {
					return fmt.Errorf("invalid runtime: %w", err)
				}
			}

			repoPath, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg := alchemist.Config{
				RepoPath:               repoPath,
				ProgramFile:            programFile,
				TargetFile:             targetFile,
				MetricName:             metric,
				MetricRegex:            regex,
				HigherIsBetter:         higher,
				Budget:                 budgetDur,
				RunCmd:                 []string{"sh", "-c", runCmd},
				MaxExperiments:         maxExperiments,
				MaxRuntime:             runtimeDur,
				Safety:                 alchemist.SafetyMode(safety),
				HooksEnabled:           !noHooks,
				LiteratureRefreshEvery: litRefresh,
				LiteratureProfile:      litProfile,
			}

			daemon, err := alchemist.NewDaemon(cfg)
			if err != nil {
				return fmt.Errorf("init daemon: %w", err)
			}
			defer daemon.Close()

			// Graceful shutdown
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				fmt.Println("\n⏹  Shutting down gracefully...")
				cancel()
			}()

			report, err := daemon.Run(ctx)
			if err != nil {
				return err
			}

			// Print morning report
			md, err := report.RenderMarkdown()
			if err != nil {
				return err
			}

			// Save to file
			reportPath := filepath.Join(repoPath, ".sin-code",
				fmt.Sprintf("alchemist-report-%s.md", time.Now().Format("2006-01-02-1504")))
			_ = os.WriteFile(reportPath, []byte(md), 0644) // #nosec G306 — user-visible report

			fmt.Println("\n" + md)
			fmt.Printf("\n📄 Report saved to: %s\n", reportPath)

			return nil
		},
	}

	cmd.Flags().StringVar(&budget, "budget", "5m", "Time budget per experiment")
	cmd.Flags().StringVar(&runtime, "runtime", "", "Total runtime (e.g. 8h, 30m)")
	cmd.Flags().StringVar(&targetFile, "target", "train.py", "File to mutate")
	cmd.Flags().StringVar(&metric, "metric", "val_bpb", "Metric name")
	cmd.Flags().StringVar(&regex, "regex", `val_bpb:\s*([0-9\.]+)`, "Metric extraction regex")
	cmd.Flags().BoolVar(&higher, "higher-is-better", false, "Optimization direction")
	cmd.Flags().StringVar(&runCmd, "cmd", "", "Verification command (M3 gate)")
	cmd.Flags().StringVar(&safety, "safety", "auto-commit", "Safety mode: interactive|auto-commit|headless")
	cmd.Flags().IntVar(&maxExperiments, "max", 0, "Max experiments (0=unlimited)")
	cmd.Flags().StringVar(&programFile, "program", "program.md", "Path to program.md")
	cmd.Flags().BoolVar(&noHooks, "no-hooks", false, "Disable lifecycle hooks")
	cmd.Flags().IntVar(&litRefresh, "literature-refresh", 0, "Refresh hypotheses every N experiments (0=disabled)")
	cmd.Flags().StringVar(&litProfile, "literature-profile", "technical-deep-dive", "sin-websearch profile for literature refresh")

	_ = cmd.MarkFlagRequired("cmd")

	return cmd
}
