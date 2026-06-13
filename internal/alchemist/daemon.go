// Purpose: Implement the autonomous research daemon loop.
// Docs: daemon.doc.md

// Package alchemist implements Karpathy-style autonomous research loops
// with git automation, SQLite history, and morning reports.
// Respects SIN-Code safety invariants: M3 (verification gate) and M4 (permissions).
package alchemist

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/experiment"
)

// SafetyMode controls the daemon's autonomy level
type SafetyMode string

const (
	SafetyInteractive SafetyMode = "interactive" // ask human before commit
	SafetyAutoCommit  SafetyMode = "auto-commit" // commit locally, no push
	SafetyHeadless    SafetyMode = "headless"    // log only, no git changes (M4 safe)
)

// Config is the full daemon configuration
type Config struct {
	RepoPath       string // git repo to work in
	ProgramFile    string // default: "program.md"
	TargetFile     string // file the agent mutates
	MetricName     string // e.g., "val_bpb", "ops_per_sec"
	MetricRegex    string // regex to extract metric
	HigherIsBetter bool
	Budget         time.Duration // per-experiment wall-clock budget
	RunCmd         []string      // command to execute
	MaxExperiments int           // 0 = unlimited
	MaxRuntime     time.Duration // 0 = unlimited (overnight mode)
	Safety         SafetyMode
	WorkBranch     string // default: "alchemist/<timestamp>"
	HooksEnabled   bool
	Logger         *slog.Logger

	// Literature-driven hypothesis refresh.
	LiteratureRefreshEvery int    // 0 = disabled, default 10
	LiteratureProfile      string // research profile (default: technical-deep-dive)
}

// ExperimentRecord is a single logged experiment
type ExperimentRecord struct {
	ID            int64         `json:"id"`
	Timestamp     time.Time     `json:"timestamp"`
	Hypothesis    string        `json:"hypothesis"`
	MetricBefore  float64       `json:"metric_before"`
	MetricAfter   float64       `json:"metric_after"`
	Delta         float64       `json:"delta"`
	Duration      time.Duration `json:"duration"`
	Decision      string        `json:"decision"` // "committed", "discarded", "timeout", "error"
	CommitSHA     string        `json:"commit_sha,omitempty"`
	StdoutSnippet string        `json:"stdout_snippet,omitempty"`
}

// Daemon is the autonomous research loop
type Daemon struct {
	cfg             Config
	history         *History
	git             *GitOps
	program         *ProgramMD
	literature      *LiteratureLoader
	baseline        float64
	mu              sync.Mutex
	experimentCount int
	startTime       time.Time
	logger          *slog.Logger
}

// NewDaemon creates a new autonomous research daemon
func NewDaemon(cfg Config) (*Daemon, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	if cfg.ProgramFile == "" {
		cfg.ProgramFile = "program.md"
	}
	if cfg.WorkBranch == "" {
		cfg.WorkBranch = fmt.Sprintf("alchemist/%s", time.Now().Format("20060102-150405"))
	}
	if cfg.Safety == "" {
		cfg.Safety = SafetyAutoCommit
	}
	if cfg.LiteratureProfile == "" {
		cfg.LiteratureProfile = "technical-deep-dive"
	}

	git, err := NewGitOps(cfg.RepoPath, cfg.WorkBranch, cfg.Safety)
	if err != nil {
		return nil, fmt.Errorf("init git: %w", err)
	}

	prog, err := LoadProgramMD(filepath.Join(cfg.RepoPath, cfg.ProgramFile))
	if err != nil {
		return nil, fmt.Errorf("load program.md: %w", err)
	}

	loader := NewLiteratureLoader(cfg.RepoPath)
	loader.SetProfile(cfg.LiteratureProfile)
	loader.SetRefreshEvery(cfg.LiteratureRefreshEvery)

	return &Daemon{
		cfg:        cfg,
		git:        git,
		program:    prog,
		literature: loader,
		startTime:  time.Now(),
		logger:     cfg.Logger,
	}, nil
}

// initHistory creates the SQLite history store on the work branch.
func (d *Daemon) initHistory() error {
	if d.history != nil {
		return nil
	}
	hist, err := NewHistory(d.cfg.RepoPath)
	if err != nil {
		return fmt.Errorf("init history: %w", err)
	}
	d.history = hist
	return nil
}

// Run executes the autonomous loop until budget/limits are hit or context cancelled
func (d *Daemon) Run(ctx context.Context) (*MorningReport, error) {
	d.logger.Info("alchemist daemon starting",
		"safety", d.cfg.Safety,
		"budget_per_exp", d.cfg.Budget,
		"max_experiments", d.cfg.MaxExperiments,
		"max_runtime", d.cfg.MaxRuntime)

	// Safety: verify the gate (M3) is configured
	if len(d.cfg.RunCmd) == 0 {
		return nil, fmt.Errorf("safety invariant violated: no run command configured (M3 gate required)")
	}

	// Create work branch first so the history DB lives on the work branch.
	if err := d.git.CreateWorkBranch(ctx); err != nil {
		return nil, fmt.Errorf("create work branch: %w", err)
	}
	defer func() {
		if d.cfg.Safety == SafetyHeadless {
			// In headless mode, return to original branch, don't keep work branch
			d.git.ReturnToMainBranch(ctx)
		}
	}()

	if err := d.initHistory(); err != nil {
		return nil, err
	}

	// Hook: SessionStart
	if d.cfg.HooksEnabled {
		d.fireHook("alchemist_start", map[string]any{
			"branch": d.cfg.WorkBranch,
			"safety": d.cfg.Safety,
		})
	}

	loopCtx := ctx
	if d.cfg.MaxRuntime > 0 {
		var cancel context.CancelFunc
		loopCtx, cancel = context.WithTimeout(ctx, d.cfg.MaxRuntime)
		defer cancel()
	}

	expCfg := experiment.Config{
		TargetFile:     d.cfg.TargetFile,
		MetricName:     d.cfg.MetricName,
		MetricRegex:    d.cfg.MetricRegex,
		HigherIsBetter: d.cfg.HigherIsBetter,
		Budget:         d.cfg.Budget,
		RunCmd:         d.cfg.RunCmd,
	}

	expLoop, err := experiment.NewLoop(expCfg)
	if err != nil {
		return nil, fmt.Errorf("init experiment loop: %w", err)
	}

	// Main loop
	for {
		select {
		case <-loopCtx.Done():
			d.logger.Info("daemon stopping: context cancelled or budget exhausted")
			return d.generateMorningReport()
		default:
		}

		// Check experiment limit
		if d.cfg.MaxExperiments > 0 && d.experimentCount >= d.cfg.MaxExperiments {
			d.logger.Info("daemon stopping: experiment limit reached", "count", d.experimentCount)
			return d.generateMorningReport()
		}

		// Pick next hypothesis from program.md queue
		hypothesis := d.program.NextHypothesis()
		if hypothesis == "" {
			d.logger.Info("no more hypotheses in queue, stopping")
			return d.generateMorningReport()
		}

		d.logger.Info("starting experiment",
			"n", d.experimentCount+1,
			"hypothesis", hypothesis)

		// Snapshot current state
		snapshot, err := d.git.Snapshot(ctx)
		if err != nil {
			d.logger.Error("snapshot failed", "err", err)
			return d.generateMorningReport()
		}

		// Run the experiment (fixed budget)
		result, err := expLoop.Run(loopCtx)
		if err != nil {
			d.logger.Error("run failed", "err", err)
			d.git.Restore(ctx, snapshot)
			d.logRecord(hypothesis, 0, 0, 0, "error", "")
			continue
		}

		// Evaluate
		kept, reason := expLoop.Evaluate(result)
		d.experimentCount++

		var decision, commitSHA string
		if kept {
			decision = "committed"
			if d.cfg.Safety == SafetyHeadless {
				// In headless mode: log only, don't touch git (M4: ask=deny)
				decision = "headless-kept"
				d.logger.Info("headless mode: would commit but safety denies",
					"metric", result.MetricValue, "reason", reason)
			} else {
				msg := fmt.Sprintf("exp: %s (%s)", hypothesis, reason)
				sha, err := d.git.CommitIfImproved(ctx, result.MetricValue, msg)
				if err != nil {
					d.logger.Error("commit failed", "err", err)
					d.git.Restore(ctx, snapshot)
					decision = "commit-failed"
				} else {
					commitSHA = sha
					d.baseline = result.MetricValue
				}
			}
		} else {
			decision = "discarded"
			d.logger.Info("discarding", "reason", reason)
			if err := d.git.Restore(ctx, snapshot); err != nil {
				d.logger.Error("restore failed", "err", err)
			}
		}

		// Log to SQLite
		delta := result.MetricValue - d.baseline
		if d.baseline == 0 {
			delta = 0
		}
		d.logRecord(hypothesis, d.baseline, result.MetricValue, delta, decision, commitSHA)

		// Update program.md learnings
		d.program.AppendLearning(fmt.Sprintf("Run %d [%s]: %s → %s",
			d.experimentCount, decision, hypothesis, reason))

		// Periodic literature refresh
		if d.literature != nil && d.literature.ShouldRefresh(d.experimentCount) {
			d.logger.Info("triggering literature refresh",
				"experiment_count", d.experimentCount)
			litCtx, litCancel := context.WithTimeout(ctx, 5*time.Minute)
			result, err := d.literature.Refresh(litCtx, d.cfg.TargetFile+" optimization")
			litCancel()
			if err != nil {
				d.logger.Warn("literature refresh failed", "err", err)
			} else if result != nil {
				if err := d.literature.InjectIntoProgramMD(d.program, result); err != nil {
					d.logger.Warn("inject hypotheses failed", "err", err)
				} else {
					d.logger.Info("literature refresh injected",
						"new_hypotheses", len(result.NewHypotheses),
						"verified_claims", len(result.VerifiedClaims))
				}
			}
		}

		// Hook: VerifyPass / VerifyFail
		if d.cfg.HooksEnabled {
			eventName := "alchemist_discard"
			if kept {
				eventName = "alchemist_commit"
			}
			d.fireHook(eventName, map[string]any{
				"hypothesis": hypothesis,
				"metric":     result.MetricValue,
				"decision":   decision,
				"commit_sha": commitSHA,
			})
		}
	}
}

func (d *Daemon) logRecord(hypothesis string, before, after, delta float64, decision, sha string) {
	if d.history == nil {
		return
	}
	record := ExperimentRecord{
		Timestamp:    time.Now(),
		Hypothesis:   hypothesis,
		MetricBefore: before,
		MetricAfter:  after,
		Delta:        delta,
		Decision:     decision,
		CommitSHA:    sha,
	}
	if err := d.history.Insert(record); err != nil {
		d.logger.Error("history insert failed", "err", err)
	}
}

func (d *Daemon) fireHook(event string, data map[string]any) {
	// Integration with SIN-Code hooks system
	d.logger.Debug("hook fired", "event", event, "data", data)
}

// Close cleans up daemon resources
func (d *Daemon) Close() error {
	if d.history == nil {
		return nil
	}
	return d.history.Close()
}
