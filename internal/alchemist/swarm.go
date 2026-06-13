// Purpose: Coordinate multi-strategy parallel Alchemist workers.
// Docs: swarm.doc.md

package alchemist

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SwarmConfig configures a multi-strategy parallel alchemist run.
type SwarmConfig struct {
	BaseConfig Config   // base alchemist config inherited by workers
	Strategies []string // strategy names (e.g. ["conservative", "aggressive"])
	MaxWorkers int      // concurrency limit (default: len(Strategies))
	FirstWin   bool     // if true, cancel others when first verified win occurs
	SharedDB   bool     // share SQLite state across workers
}

// SwarmWorker represents a single strategy's run.
type SwarmWorker struct {
	ID         int
	Strategy   Strategy
	Daemon     *Daemon
	Branch     string
	Results    []ExperimentRecord
	BestMetric float64
	Commits    int
	Discards   int
	Error      error
	StoppedAt  time.Time
}

// SwarmReport is the consolidated output of a swarm run.
type SwarmReport struct {
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Workers      []SwarmWorker
	Winner       *SwarmWorker
	TotalExp     int
	TotalCommits int
	TotalDiscard int
	BestMetric   float64
	BestStrategy string
	BestCommit   string
}

// Swarm runs multi-strategy parallel research.
type Swarm struct {
	cfg      SwarmConfig
	repoPath string
	history  *History
	strats   []Strategy
	logger   *slog.Logger
	winner   *SwarmWorker
	winnerMu sync.Mutex
}

// NewSwarm creates a swarm coordinator.
func NewSwarm(cfg SwarmConfig) (*Swarm, error) {
	if len(cfg.Strategies) == 0 {
		cfg.Strategies = []string{"conservative", "aggressive", "creative", "minimal"}
	}
	if cfg.MaxWorkers == 0 {
		cfg.MaxWorkers = len(cfg.Strategies)
	}

	var strats []Strategy
	for _, name := range cfg.Strategies {
		strats = append(strats, GetStrategy(name))
	}

	hist, err := NewHistory(cfg.BaseConfig.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("init shared history: %w", err)
	}

	logger := cfg.BaseConfig.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	return &Swarm{
		cfg:      cfg,
		repoPath: cfg.BaseConfig.RepoPath,
		history:  hist,
		strats:   strats,
		logger:   logger,
	}, nil
}

// Run executes the swarm. Each strategy gets its own work branch + daemon.
func (s *Swarm) Run(ctx context.Context) (*SwarmReport, error) {
	report := &SwarmReport{
		StartTime: time.Now(),
	}

	s.logger.Info("swarm starting",
		"strategies", s.cfg.Strategies,
		"workers", s.cfg.MaxWorkers,
		"first_win", s.cfg.FirstWin)

	workers := make([]SwarmWorker, len(s.strats))
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.cfg.MaxWorkers)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, strat := range s.strats {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, strat Strategy) {
			defer wg.Done()
			defer func() { <-sem }()

			worker := s.runWorker(runCtx, idx, strat)
			workers[idx] = worker

			if s.cfg.FirstWin && worker.Commits > 0 {
				s.winnerMu.Lock()
				if s.winner == nil || worker.BestMetric > s.winner.BestMetric {
					s.winner = &worker
					s.logger.Info("swarm: first winner declared",
						"strategy", strat.Name,
						"metric", worker.BestMetric)
					cancel()
				}
				s.winnerMu.Unlock()
			}
		}(i, strat)
	}

	wg.Wait()
	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.Workers = workers

	s.pickWinner(report)
	s.aggregateStats(report)

	return report, nil
}

func (s *Swarm) runWorker(ctx context.Context, id int, strat Strategy) SwarmWorker {
	worker := SwarmWorker{
		ID:       id,
		Strategy: strat,
		Branch:   fmt.Sprintf("alchemist/swarm/%s/%s", strat.Name, time.Now().Format("20060102-150405")),
	}

	workerCfg := s.cfg.BaseConfig
	workerCfg.WorkBranch = worker.Branch
	workerCfg.ProgramFile = fmt.Sprintf("program.%s.md", strat.Name)

	if err := s.ensureStrategyProgram(strat, workerCfg.ProgramFile); err != nil {
		worker.Error = fmt.Errorf("setup program.md: %w", err)
		return worker
	}

	daemon, err := NewDaemon(workerCfg)
	if err != nil {
		worker.Error = fmt.Errorf("init daemon: %w", err)
		return worker
	}
	defer daemon.Close()
	worker.Daemon = daemon

	morningReport, err := daemon.Run(ctx)
	if err != nil && ctx.Err() == nil {
		worker.Error = err
	}
	_ = morningReport

	records, _ := s.history.All()
	for _, r := range records {
		if r.Decision == "committed" {
			worker.Commits++
			if r.MetricAfter > worker.BestMetric {
				worker.BestMetric = r.MetricAfter
				worker.Results = append(worker.Results, r)
			}
		} else if r.Decision == "discarded" {
			worker.Discards++
		}
	}

	worker.StoppedAt = time.Now()

	s.logger.Info("swarm worker done",
		"strategy", strat.Name,
		"commits", worker.Commits,
		"discards", worker.Discards,
		"best_metric", worker.BestMetric)

	return worker
}

func (s *Swarm) ensureStrategyProgram(strat Strategy, filename string) error {
	path := filepath.Join(s.repoPath, filename)

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	basePath := filepath.Join(s.repoPath, s.cfg.BaseConfig.ProgramFile)
	base, err := os.ReadFile(basePath)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("<!-- Strategy: %s -->\n<!-- %s -->\n<!-- Risk: %.2f | Max mutation: %d lines -->\n\n",
		strat.Name, strat.Description, strat.RiskAppetite, strat.MaxMutation)

	content := header + string(base)
	content += fmt.Sprintf("\n\n## Strategy Overlay\n\n%s\n", strat.PromptOverlay)

	return os.WriteFile(path, []byte(content), 0644)
}

func (s *Swarm) pickWinner(report *SwarmReport) {
	var best *SwarmWorker
	for i := range report.Workers {
		w := &report.Workers[i]
		if w.Error != nil {
			continue
		}
		if w.Commits == 0 {
			continue
		}
		if best == nil || w.BestMetric > best.BestMetric {
			best = w
		}
	}
	report.Winner = best
	if best != nil {
		report.BestStrategy = best.Strategy.Name
		report.BestMetric = best.BestMetric
		if len(best.Results) > 0 {
			report.BestCommit = best.Results[len(best.Results)-1].CommitSHA
		}
	}
}

func (s *Swarm) aggregateStats(report *SwarmReport) {
	for _, w := range report.Workers {
		report.TotalExp += w.Commits + w.Discards
		report.TotalCommits += w.Commits
		report.TotalDiscard += w.Discards
	}
}

// Close releases shared resources.
func (s *Swarm) Close() error {
	return s.history.Close()
}

// RenderMarkdown formats the swarm report as Markdown.
func (r *SwarmReport) RenderMarkdown() string {
	var s string
	s += "# 🐝 Alchemist Swarm Report\n\n"
	s += fmt.Sprintf("**Run:** %s → %s (%s)\n",
		r.StartTime.Format("2006-01-02 15:04"),
		r.EndTime.Format("15:04"),
		r.Duration.Round(time.Second))
	s += fmt.Sprintf("**Workers:** %d strategies\n\n", len(r.Workers))

	s += "## 🏆 Winner\n\n"
	if r.Winner != nil {
		s += fmt.Sprintf("- **Strategy:** %s\n", r.Winner.Strategy.Name)
		s += fmt.Sprintf("- **Best metric:** %.4f\n", r.BestMetric)
		s += fmt.Sprintf("- **Commits:** %d / %d experiments\n",
			r.Winner.Commits, r.Winner.Commits+r.Winner.Discards)
		if r.BestCommit != "" {
			s += fmt.Sprintf("- **Commit:** `%s`\n", shortSHA(r.BestCommit))
		}
	} else {
		s += "_No winning strategy (no successful commits)._\n"
	}
	s += "\n"

	s += "## 📊 All Workers\n\n"
	s += "| Strategy | Commits | Discards | Best Metric | Branch |\n"
	s += "|---|---|---|---|---|\n"
	for _, w := range r.Workers {
		status := "✅"
		if w.Error != nil {
			status = "❌"
		} else if w.Commits == 0 {
			status = "⚪"
		}
		s += fmt.Sprintf("| %s %s | %d | %d | %.4f | `%s` |\n",
			status, w.Strategy.Name, w.Commits, w.Discards,
			w.BestMetric, w.Branch)
	}
	s += "\n"

	s += "## 📈 Aggregate Stats\n\n"
	s += fmt.Sprintf("- Total experiments: %d\n", r.TotalExp)
	s += fmt.Sprintf("- Total commits: %d\n", r.TotalCommits)
	s += fmt.Sprintf("- Total discards: %d\n", r.TotalDiscard)
	if r.TotalExp > 0 {
		s += fmt.Sprintf("- Overall success rate: %.1f%%\n",
			float64(r.TotalCommits)/float64(r.TotalExp)*100)
	}

	return s
}

