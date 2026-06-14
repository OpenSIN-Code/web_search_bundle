// SPDX-License-Identifier: MIT
// Purpose: Smoke tests for the Swarm multi-strategy coordinator.
// Docs: swarm_test.doc.md

package alchemist

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupGitRepo creates a temporary git repo with an initial commit.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "program.md"), []byte("# Program\n\n## Hypothesis Queue\n\n- [ ] Smoke test hypothesis\n\n## Learnings\n\n- Initial\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "train.py"), []byte("print('metric: 0.5')\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "-A")
	run("git", "commit", "-m", "initial")
	return dir
}

func TestNewSwarmDefaults(t *testing.T) {
	cfg := SwarmConfig{
		BaseConfig: Config{RepoPath: setupGitRepo(t)},
	}
	defer func() {
		_ = cfg.BaseConfig // no-op to keep reference alive
	}()

	swarm, err := NewSwarm(cfg)
	if err != nil {
		t.Fatalf("NewSwarm failed: %v", err)
	}
	defer swarm.Close()

	if len(swarm.strats) != 4 {
		t.Errorf("expected 4 default strategies, got %d", len(swarm.strats))
	}
	if swarm.cfg.MaxWorkers != 4 {
		t.Errorf("expected 4 workers, got %d", swarm.cfg.MaxWorkers)
	}
}

func TestNewSwarmCustomStrategies(t *testing.T) {
	cfg := SwarmConfig{
		BaseConfig: Config{RepoPath: setupGitRepo(t)},
		Strategies: []string{"minimal", "creative"},
		MaxWorkers: 2,
	}
	swarm, err := NewSwarm(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer swarm.Close()

	if len(swarm.strats) != 2 {
		t.Errorf("expected 2 strategies, got %d", len(swarm.strats))
	}
	if swarm.cfg.MaxWorkers != 2 {
		t.Errorf("expected 2 workers, got %d", swarm.cfg.MaxWorkers)
	}
}

func TestSwarmEnsureStrategyProgram(t *testing.T) {
	repo := setupGitRepo(t)
	swarm, err := NewSwarm(SwarmConfig{
		BaseConfig: Config{RepoPath: repo, ProgramFile: "program.md"},
		Strategies: []string{"minimal"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer swarm.Close()

	path := filepath.Join(repo, "program.minimal.md")
	if err := swarm.ensureStrategyProgram(swarm.strats[0], "program.minimal.md"); err != nil {
		t.Fatalf("ensureStrategyProgram failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !containsAll(content, []string{"Strategy: minimal", "Risk:", "Max mutation:"}) {
		t.Errorf("program.minimal.md missing strategy metadata: %s", content)
	}
	if !containsAll(content, []string{"single-line", "control group"}) {
		t.Errorf("program.minimal.md missing strategy overlay: %s", content)
	}

}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}

func TestSwarmReportRenderMarkdown(t *testing.T) {
	report := SwarmReport{
		StartTime:    time.Now().Add(-5 * time.Minute),
		EndTime:      time.Now(),
		Workers:      []SwarmWorker{},
		TotalExp:     0,
		TotalCommits: 0,
		TotalDiscard: 0,
	}
	md := report.RenderMarkdown()
	if !containsAll(md, []string{"Alchemist Swarm Report", "No winning strategy"}) {
		t.Errorf("render missing expected sections: %s", md)
	}

	report.Workers = []SwarmWorker{
		{
			Strategy:   GetStrategy("minimal"),
			Commits:    1,
			Discards:   0,
			BestMetric: 0.7,
			Branch:     "alchemist/swarm/minimal/20260101-000000",
		},
		{
			Strategy:   GetStrategy("creative"),
			Commits:    0,
			Discards:   1,
			BestMetric: 0,
			Branch:     "alchemist/swarm/creative/20260101-000001",
			Error:      nil,
		},
	}
	report.Winner = &report.Workers[0]
	report.BestMetric = 0.7
	report.BestStrategy = "minimal"
	md = report.RenderMarkdown()
	if !containsAll(md, []string{"minimal", "creative", "0.7000", "Winner"}) {
		t.Errorf("render missing expected winner sections: %s", md)
	}
}

func TestSwarmRunHeadless(t *testing.T) {
	repo := setupGitRepo(t)

	cfg := SwarmConfig{
		BaseConfig: Config{
			RepoPath:       repo,
			ProgramFile:    "program.md",
			TargetFile:     "train.py",
			MetricName:     "metric",
			MetricRegex:    `metric:\s*([0-9\.]+)`,
			HigherIsBetter: true,
			Budget:         5 * time.Second,
			RunCmd:         []string{"sh", "-c", "echo 'metric: 0.8' > train.py && echo 'metric: 0.8'"},
			MaxExperiments: 1,
			MaxRuntime:     10 * time.Second,
			Safety:         SafetyAutoCommit,
		},
		Strategies: []string{"minimal"},
		MaxWorkers: 1,
	}

	swarm, err := NewSwarm(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer swarm.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := swarm.Run(ctx)
	if err != nil {
		t.Fatalf("swarm run failed: %v", err)
	}

	if len(report.Workers) != 1 {
		t.Fatalf("expected 1 worker, got %d", len(report.Workers))
	}
	if report.Workers[0].Error != nil {
		t.Fatalf("worker error: %v", report.Workers[0].Error)
	}
	// Auto-commit mode should record the experiment as committed.
	if report.Workers[0].Commits != 1 {
		t.Errorf("expected 1 commit, got %d", report.Workers[0].Commits)
	}
	if report.TotalExp < 1 {
		t.Errorf("expected at least 1 experiment, got %d", report.TotalExp)
	}
}

