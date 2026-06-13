// Purpose: Implement Git operations for Alchemist.
// Docs: git_ops.doc.md

package alchemist

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Snapshot captures git state for rollback
type Snapshot struct {
	Branch    string
	CommitSHA string
	Dirty     bool
}

// GitOps handles git operations with safety enforcement
type GitOps struct {
	repoPath   string
	workBranch string
	safety     SafetyMode
	mainBranch string
}

// NewGitOps creates a git operations handler
func NewGitOps(repoPath, workBranch string, safety SafetyMode) (*GitOps, error) {
	// Detect main branch
	main, err := detectMainBranch(repoPath)
	if err != nil {
		main = "main"
	}

	return &GitOps{
		repoPath:   repoPath,
		workBranch: workBranch,
		safety:     safety,
		mainBranch: main,
	}, nil
}

// CreateWorkBranch creates an isolated branch for experiments
func (g *GitOps) CreateWorkBranch(ctx context.Context) error {
	if g.safety == SafetyHeadless {
		// In headless mode, don't create branches (read-only)
		return nil
	}

	// Ensure clean working tree first
	if dirty, _ := g.isDirty(ctx); dirty {
		return fmt.Errorf("working tree is dirty; commit or stash before starting alchemist")
	}

	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", g.workBranch)
	cmd.Dir = g.repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create branch: %s: %w", string(out), err)
	}
	return nil
}

// ReturnToMainBranch switches back to the main branch (headless cleanup)
func (g *GitOps) ReturnToMainBranch(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", g.mainBranch)
	cmd.Dir = g.repoPath
	_ = cmd.Run()
	return nil
}

// Snapshot captures current state for rollback
func (g *GitOps) Snapshot(ctx context.Context) (*Snapshot, error) {
	branch, err := g.currentBranch(ctx)
	if err != nil {
		return nil, err
	}
	sha, err := g.currentSHA(ctx)
	if err != nil {
		return nil, err
	}
	dirty, _ := g.isDirty(ctx)
	return &Snapshot{
		Branch:    branch,
		CommitSHA: sha,
		Dirty:     dirty,
	}, nil
}

// Restore rolls back to a snapshot
func (g *GitOps) Restore(ctx context.Context, snap *Snapshot) error {
	if g.safety == SafetyHeadless {
		return nil // nothing to restore
	}

	// Hard reset to the snapshot commit
	cmd := exec.CommandContext(ctx, "git", "reset", "--hard", snap.CommitSHA)
	cmd.Dir = g.repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reset: %s: %w", string(out), err)
	}

	// Clean untracked files
	cmd = exec.CommandContext(ctx, "git", "clean", "-fd")
	cmd.Dir = g.repoPath
	_ = cmd.Run()

	return nil
}

// CommitIfImproved commits changes if metric improved
// Returns the new commit SHA
func (g *GitOps) CommitIfImproved(ctx context.Context, metricValue float64, message string) (string, error) {
	if g.safety == SafetyHeadless {
		return "", fmt.Errorf("headless mode: commits forbidden (M4 safety)")
	}

	// Stage all changes
	cmd := exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = g.repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("stage: %s: %w", string(out), err)
	}

	// Check if anything to commit
	cmd = exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = g.repoPath
	if err := cmd.Run(); err == nil {
		// No changes
		sha, _ := g.currentSHA(ctx)
		return sha, nil
	}

	// Commit with metric in message
	fullMsg := fmt.Sprintf("%s\n\nAlchemist: metric=%.4f\nSafety: %s", message, metricValue, g.safety)
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", fullMsg)
	cmd.Dir = g.repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("commit: %s: %w", string(out), err)
	}

	return g.currentSHA(ctx)
}

// Push would push to remote — ALWAYS forbidden in headless/auto-commit mode.
// Only interactive mode allows push, and only with explicit user confirmation.
func (g *GitOps) Push(ctx context.Context) error {
	if g.safety != SafetyInteractive {
		return fmt.Errorf("push forbidden in %s mode (safety invariant)", g.safety)
	}
	cmd := exec.CommandContext(ctx, "git", "push", "-u", "origin", g.workBranch)
	cmd.Dir = g.repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("push: %s: %w", string(out), err)
	}
	return nil
}

// Diff returns the diff between main and work branch
func (g *GitOps) Diff(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", g.mainBranch+"..."+g.workBranch)
	cmd.Dir = g.repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// LogRecentCommits returns last N commits on work branch
func (g *GitOps) LogRecentCommits(ctx context.Context, n int) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "--oneline", "-n", fmt.Sprintf("%d", n), g.workBranch)
	cmd.Dir = g.repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// Stats returns repository statistics
func (g *GitOps) Stats(ctx context.Context) (map[string]any, error) {
	commits, _ := g.countCommits(ctx)
	return map[string]any{
		"work_branch":   g.workBranch,
		"main_branch":   g.mainBranch,
		"new_commits":   commits,
		"safety_mode":   g.safety,
		"remote_pushed": false, // never pushed automatically
	}, nil
}

// --- helpers ---

func (g *GitOps) currentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *GitOps) currentSHA(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *GitOps) isDirty(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

func (g *GitOps) countCommits(ctx context.Context) (int, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-list", "--count",
		fmt.Sprintf("%s..%s", g.mainBranch, g.workBranch))
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var n int
	_, _ = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &n)
	return n, nil
}

func detectMainBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		// Fallback: try common names
		for _, name := range []string{"main", "master"} {
			check := exec.Command("git", "rev-parse", "--verify", name)
			check.Dir = repoPath
			if check.Run() == nil {
				return name, nil
			}
		}
		return "main", nil
	}
	// refs/remotes/origin/main → main
	ref := strings.TrimSpace(string(out))
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}
	return "main", nil
}
