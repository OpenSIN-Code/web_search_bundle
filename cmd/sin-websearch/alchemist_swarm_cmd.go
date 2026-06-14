// Purpose: CLI command to run a multi-strategy alchemist swarm.
// Docs: alchemist_swarm_cmd.doc.md
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/alchemist"
	"github.com/spf13/cobra"
)

func newAlchemistSwarmCmd() *cobra.Command {
	var (
		strategies   []string
		maxWorkers   int
		budget       string
		runtime      string
		targetFile   string
		metric       string
		regex        string
		higher       bool
		runCmd       string
		firstWin     bool
		noLiterature bool
		litRefresh   int
		litProfile   string
	)

	cmd := &cobra.Command{
		Use:   "swarm",
		Short: "Run multi-strategy parallel alchemist swarm",
		Long: `Launch multiple alchemist workers with different strategies in parallel.
Each strategy gets its own isolated work branch. The best-performing strategy wins.

Available strategies:
  conservative      Minimal changes, low risk (1 function, <20 lines)
  aggressive        Large refactors, high risk (restructure allowed)
  creative          Unconventional approaches, cross-domain inspiration
  minimal           Control group (1-5 line changes)
  literature-driven Hypotheses from sin-websearch SOTA scan

Examples:
  # Default 4-strategy swarm overnight
  sin-websearch alchemist swarm --runtime 8h --cmd "go test -bench=."

  # Custom strategies + first-win (fast mode)
  sin-websearch alchemist swarm --strategies conservative,creative --first-win --cmd "..."

  # With literature refresh every 5 experiments per worker
  sin-websearch alchemist swarm --literature-refresh 5 --cmd "..."`,
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

			available := alchemist.StrategyNames()
			for _, s := range strategies {
				found := false
				for _, a := range available {
					if s == a {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("unknown strategy '%s'. Available: %v", s, available)
				}
			}

			baseCfg := alchemist.Config{
				RepoPath:               repoPath,
				TargetFile:             targetFile,
				MetricName:             metric,
				MetricRegex:            regex,
				HigherIsBetter:         higher,
				Budget:                 budgetDur,
				RunCmd:                 []string{"sh", "-c", runCmd},
				MaxRuntime:             runtimeDur,
				Safety:                 alchemist.SafetyAutoCommit,
				LiteratureRefreshEvery: litRefresh,
				LiteratureProfile:      litProfile,
			}
			if noLiterature {
				baseCfg.LiteratureRefreshEvery = 0
			}

			swarmCfg := alchemist.SwarmConfig{
				BaseConfig: baseCfg,
				Strategies: strategies,
				MaxWorkers: maxWorkers,
				FirstWin:   firstWin,
				SharedDB:   true,
			}

			swarm, err := alchemist.NewSwarm(swarmCfg)
			if err != nil {
				return fmt.Errorf("init swarm: %w", err)
			}
			defer swarm.Close()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				fmt.Println("\n⏹  Shutting down swarm gracefully...")
				cancel()
			}()

			fmt.Printf("🐝 Starting alchemist swarm\n")
			fmt.Printf("   Strategies: %s\n", strings.Join(strategies, ", "))
			fmt.Printf("   Max workers: %d\n", maxWorkers)
			fmt.Printf("   Budget/exp: %s\n", budgetDur)
			fmt.Printf("   First-win: %v\n", firstWin)
			if !noLiterature {
				fmt.Printf("   Literature refresh: every %d experiments\n", litRefresh)
			}
			fmt.Println()

			report, err := swarm.Run(ctx)
			if err != nil {
				return err
			}

			md := report.RenderMarkdown()
			fmt.Println("\n" + md)

			reportPath := filepath.Join(repoPath, ".sin-code",
				fmt.Sprintf("swarm-report-%s.md", time.Now().Format("2006-01-02-1504")))
			_ = os.MkdirAll(filepath.Dir(reportPath), 0750)
			_ = os.WriteFile(reportPath, []byte(md), 0644) // #nosec G306 — user-visible report
			fmt.Printf("📄 Report saved to: %s\n", reportPath)

			if report.Winner != nil {
				fmt.Printf("\n🏆 Winner: strategy=%s metric=%.4f\n",
					report.BestStrategy, report.BestMetric)
				fmt.Printf("   Review branch: git log %s\n", report.Winner.Branch)
				fmt.Printf("   Merge manually: git merge %s\n", report.Winner.Branch)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&strategies, "strategies",
		[]string{"conservative", "aggressive", "creative", "minimal"},
		"Strategies to run in parallel")
	cmd.Flags().IntVar(&maxWorkers, "workers", 0, "Max concurrent workers (default: all)")
	cmd.Flags().StringVar(&budget, "budget", "5m", "Time budget per experiment")
	cmd.Flags().StringVar(&runtime, "runtime", "", "Total runtime (e.g. 8h)")
	cmd.Flags().StringVar(&targetFile, "target", "train.py", "File to mutate")
	cmd.Flags().StringVar(&metric, "metric", "val_bpb", "Metric name")
	cmd.Flags().StringVar(&regex, "regex", `val_bpb:\s*([0-9\.]+)`, "Metric regex")
	cmd.Flags().BoolVar(&higher, "higher-is-better", false, "Optimization direction")
	cmd.Flags().StringVar(&runCmd, "cmd", "", "Verification command (M3 gate)")
	cmd.Flags().BoolVar(&firstWin, "first-win", false, "Cancel others on first verified win")
	cmd.Flags().BoolVar(&noLiterature, "no-literature", false, "Disable sin-websearch refresh")
	cmd.Flags().IntVar(&litRefresh, "literature-refresh", 10, "Refresh every N experiments")
	cmd.Flags().StringVar(&litProfile, "literature-profile", "technical-deep-dive", "sin-websearch profile for literature refresh")

	_ = cmd.MarkFlagRequired("cmd")

	return cmd
}
