// Purpose: CLI commands for the Alchemist autonomous loop.
// Docs: cmd/sin-websearch/alchemist_cmd.doc.md

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

func newAlchemistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alchemist",
		Short: "Autonomous Karpathy-style research loops",
		Long: `Run autonomous experiments with fixed-budget loops.
The agent modifies a target file, runs a verification command (M3 gate),
measures a metric, commits if improved, resets if not. Repeats overnight.

Safety invariants (from SIN-Code v3.5.0):
  • no gate → no daemon (M3: verification cmd required)
  • headless → no git changes (M4: ask=deny enforced)
  • budget exhausted → hook summons the human

Examples:
  # Overnight run (8 hours, auto-commit locally, never push)
  sin-websearch alchemist run --budget 5m --runtime 8h --cmd "go test -bench=."

  # Headless mode (log only, no git changes — safest)
  sin-websearch alchemist run --safety headless --budget 5m

  # View morning report from last run
  sin-websearch alchemist report

  # List experiment history
  sin-websearch alchemist history --limit 20

  # Initialize program.md template
  sin-websearch alchemist init`,
	}

	cmd.AddCommand(newAlchemistRunCmd())
	cmd.AddCommand(newAlchemistSwarmCmd())
	cmd.AddCommand(newAlchemistReportCmd())
	cmd.AddCommand(newAlchemistHistoryCmd())
	cmd.AddCommand(newAlchemistInitCmd())

	return cmd
}

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
			_ = os.WriteFile(reportPath, []byte(md), 0644)

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

func newAlchemistInitCmd() *cobra.Command {
	var template string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize program.md template",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := os.Getwd()
			path := filepath.Join(repoPath, "program.md")

			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("program.md already exists at %s", path)
			}

			content := getProgramTemplate(template)
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}

			fmt.Printf("✓ Created program.md at %s\n", path)
			fmt.Println("\nEdit the hypotheses queue, then run:")
			fmt.Println("  sin-websearch alchemist run --cmd \"your verification command\"")
			return nil
		},
	}
	cmd.Flags().StringVar(&template, "template", "go", "Template: go|python|ml")
	return cmd
}

func newAlchemistSwarmCmd() *cobra.Command {
	var (
		strategies    []string
		maxWorkers    int
		budget        string
		runtime       string
		targetFile    string
		metric        string
		regex         string
		higher        bool
		runCmd        string
		firstWin      bool
		noLiterature  bool
		litRefresh    int
		litProfile    string
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
			_ = os.MkdirAll(filepath.Dir(reportPath), 0755)
			_ = os.WriteFile(reportPath, []byte(md), 0644)
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

func getProgramTemplate(kind string) string {
	switch kind {
	case "python", "ml":
		return `# program.md — Autonomous Research Loop

## The Setup
- **Target File:** ` + "`train.py`" + ` (agent modifies this)
- **Immutable Files:** ` + "`prepare.py`" + `, ` + "`evaluate.py`" + ` (DO NOT MODIFY)
- **Metric:** ` + "`val_bpb`" + ` (lower is better)
- **Time Budget:** 5 minutes per experiment (wall clock)

## Rules
1. Formulate hypothesis from the queue below.
2. Modify ONLY the target file.
3. Run verification command.
4. If metric improves → commit. Otherwise → reset.
5. Update Learnings section after every run.

## Hypothesis Queue
- [ ] Try GELU activation instead of ReLU
- [ ] Add layer normalization before attention
- [ ] Reduce vocab_size to 4096
- [ ] Increase DEPTH to 12

## Learnings (agent updates this)
`

	default: // "go"
		return `# program.md — Autonomous Go Performance Research

## The Setup
- **Target File:** ` + "`internal/dispatcher/dispatcher.go`" + `
- **Immutable Files:** ` + "`*_test.go`" + ` files (DO NOT MODIFY)
- **Metric:** ` + "`ops_per_sec`" + ` from ` + "`go test -bench=.`" + ` (higher is better)
- **Time Budget:** 5 minutes per experiment

## Rules
1. Pick next hypothesis from queue.
2. Modify ONLY the target file.
3. Run ` + "`go test -bench=. -run=^$ -count=3`" + `
4. If ops/sec improves → commit. Otherwise → reset.
5. Append learning below.

## Hypothesis Queue
- [ ] Replace sync.Mutex with sync.RWMutex in hot path
- [ ] Use sync.Pool for request buffers
- [ ] Pre-allocate slices with known capacity
- [ ] Switch from map[string]X to map[int]X with interned keys
- [ ] Batch channel sends (10 at a time)

## Learnings (agent updates this)
`
	}
}
