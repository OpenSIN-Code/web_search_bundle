// Purpose: Implement Karpathy-style fixed-budget autonomous research loop.
// Docs: loop.doc.md

package experiment

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// Config defines the boundaries of the autonomous loop
type Config struct {
	TargetFile     string        // e.g., "train.py" or "poc.go"
	MetricName     string        // e.g., "val_bpb" or "ops_per_sec"
	MetricRegex    string        // Regex to extract metric from stdout
	HigherIsBetter bool          // Optimization direction
	Budget         time.Duration // Hard wall-clock limit (e.g., 5 * time.Minute)
	RunCmd         []string      // e.g., ["uv", "run", "train.py"] or ["go", "test", "-bench=."]
}

// Result captures the outcome of a single budget-constrained run
type Result struct {
	Duration    time.Duration
	MetricValue float64
	Stdout      string
	Stderr      string
	TimedOut    bool
	Error       error
}

// Loop manages the autonomous experimentation cycle
type Loop struct {
	cfg      Config
	baseline float64
	re       *regexp.Regexp
}

// NewLoop creates a new experiment enforcer
func NewLoop(cfg Config) (*Loop, error) {
	re, err := regexp.Compile(cfg.MetricRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid metric regex: %w", err)
	}
	return &Loop{cfg: cfg, re: re}, nil
}

// Run executes the target command with a strict time budget
func (l *Loop) Run(ctx context.Context) (*Result, error) {
	// Enforce the Karpathy rule: Fixed time budget, regardless of hardware
	ctx, cancel := context.WithTimeout(ctx, l.cfg.Budget)
	defer cancel()

	cmd := exec.CommandContext(ctx, l.cfg.RunCmd[0], l.cfg.RunCmd[1:]...) // #nosec G204 — experiment intentionally runs user-defined verification command

	start := time.Now()
	outBytes, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &Result{
		Duration: duration,
		Stdout:   string(outBytes),
	}

	// Check for timeout (the budget was exhausted)
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		// In Karpathy's model, timeout is a valid state (it just means it didn't finish)
		// We extract whatever partial metric we can, or mark as failed.
	} else if err != nil {
		result.Error = fmt.Errorf("run failed: %w", err)
		return result, nil
	}

	// Extract metric
	matches := l.re.FindStringSubmatch(result.Stdout)
	if len(matches) > 1 {
		val, parseErr := strconv.ParseFloat(matches[1], 64)
		if parseErr == nil {
			result.MetricValue = val
		}
	}

	return result, nil
}

// Evaluate decides if the experiment should be kept (committed) or discarded
func (l *Loop) Evaluate(r *Result) (bool, string) {
	if r.Error != nil || r.TimedOut {
		return false, "discarded: error or timeout"
	}
	if r.MetricValue == 0 {
		return false, "discarded: no metric extracted"
	}

	if l.baseline == 0 {
		l.baseline = r.MetricValue
		return true, "accepted: new baseline established"
	}

	var improved bool
	if l.cfg.HigherIsBetter {
		improved = r.MetricValue > l.baseline
	} else {
		improved = r.MetricValue < l.baseline
	}

	if improved {
		l.baseline = r.MetricValue
		return true, fmt.Sprintf("accepted: new best %s = %.4f", l.cfg.MetricName, r.MetricValue)
	}

	return false, fmt.Sprintf("discarded: metric %.4f did not beat baseline %.4f", r.MetricValue, l.baseline)
}
