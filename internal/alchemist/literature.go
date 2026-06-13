// Purpose: Refresh hypotheses from sin-websearch literature scans.
// Docs: literature.doc.md

package alchemist

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// LiteratureLoader uses sin-websearch to refresh hypotheses from SOTA research.
type LiteratureLoader struct {
	repoPath        string
	sinWebsearchBin string
	profile         string
	refreshEvery    int
	timeout         time.Duration
}

// LiteratureResult captures new hypotheses + learnings from a refresh.
type LiteratureResult struct {
	Timestamp        time.Time `json:"timestamp"`
	Topic            string    `json:"topic"`
	NewHypotheses    []string  `json:"new_hypotheses"`
	VerifiedClaims   []string  `json:"verified_claims"`
	ContestedClaims  []string  `json:"contested_claims"`
	TopSources       []string  `json:"top_sources"`
	TotalClaimsFound int       `json:"total_claims"`
	Error            string    `json:"error,omitempty"`
}

// NewLiteratureLoader creates a loader with sensible defaults.
func NewLiteratureLoader(repoPath string) *LiteratureLoader {
	return &LiteratureLoader{
		repoPath:        repoPath,
		sinWebsearchBin: detectSinWebsearch(),
		profile:         "technical-deep-dive",
		refreshEvery:    10,
		timeout:         5 * time.Minute,
	}
}

// SetRefreshEvery configures the refresh cadence (0 = disabled).
func (l *LiteratureLoader) SetRefreshEvery(n int) {
	l.refreshEvery = n
}

// SetProfile configures the research profile.
func (l *LiteratureLoader) SetProfile(p string) {
	l.profile = p
}

// ShouldRefresh returns true if it's time to refresh based on experiment count.
func (l *LiteratureLoader) ShouldRefresh(experimentCount int) bool {
	if l.refreshEvery == 0 {
		return false
	}
	return experimentCount > 0 && experimentCount%l.refreshEvery == 0
}

// Refresh runs a sin-websearch mission on the given topic and returns findings.
func (l *LiteratureLoader) Refresh(ctx context.Context, topic string) (*LiteratureResult, error) {
	result := &LiteratureResult{
		Timestamp: time.Now(),
		Topic:     topic,
	}

	if l.sinWebsearchBin == "" {
		result.Error = "sin-websearch binary not found on PATH"
		return result, fmt.Errorf(result.Error)
	}

	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	args := []string{"mission", topic, "--profile", l.profile, "--json"}
	cmd := exec.CommandContext(ctx, l.sinWebsearchBin, args...)
	out, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("sin-websearch failed: %v", err)
		return result, err
	}

	var missionResult struct {
		Topic      string `json:"topic"`
		Status     string `json:"status"`
		AllResults []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Source  string `json:"source"`
			Snippet string `json:"snippet"`
			Upvotes int    `json:"upvotes"`
		} `json:"all_results"`
		Synthesis string `json:"synthesis"`
		Verification *struct {
			TotalClaims     int `json:"total_claims"`
			Verified        int `json:"verified"`
			Contested       int `json:"contested"`
			StrongClaims    []struct {
				Text       string  `json:"text"`
				Confidence float64 `json:"confidence"`
			} `json:"strong_claims"`
			ContestedClaims []struct {
				Text       string  `json:"text"`
				Confidence float64 `json:"confidence"`
			} `json:"contested_claims"`
		} `json:"verification"`
	}

	if err := json.Unmarshal(out, &missionResult); err != nil {
		result.Error = fmt.Sprintf("parse mission output: %v", err)
		return result, err
	}

	if missionResult.Verification != nil {
		result.TotalClaimsFound = missionResult.Verification.TotalClaims
		for _, c := range missionResult.Verification.StrongClaims {
			if c.Confidence >= 0.7 {
				h := fmt.Sprintf("Implement: %s (sin-websearch conf=%.2f)",
					truncateClaim(c.Text, 80), c.Confidence)
				result.NewHypotheses = append(result.NewHypotheses, h)
				result.VerifiedClaims = append(result.VerifiedClaims, c.Text)
			}
		}
		for _, c := range missionResult.Verification.ContestedClaims {
			result.ContestedClaims = append(result.ContestedClaims, c.Text)
		}
	}

	// Top sources by engagement (simple bubble sort for top 5).
	type sourceStat struct {
		title  string
		source string
		engage int
	}
	var top []sourceStat
	for _, r := range missionResult.AllResults {
		top = append(top, sourceStat{r.Title, r.Source, r.Upvotes})
	}
	for i := 0; i < len(top) && i < 5; i++ {
		for j := i + 1; j < len(top); j++ {
			if top[j].engage > top[i].engage {
				top[i], top[j] = top[j], top[i]
			}
		}
	}
	for i := 0; i < len(top) && i < 5; i++ {
		result.TopSources = append(result.TopSources,
			fmt.Sprintf("[%s] %s", top[i].source, truncateClaim(top[i].title, 60)))
	}

	if len(result.NewHypotheses) > 10 {
		result.NewHypotheses = result.NewHypotheses[:10]
	}

	return result, nil
}

// InjectIntoProgramMD adds new hypotheses to program.md's queue.
func (l *LiteratureLoader) InjectIntoProgramMD(prog *ProgramMD, result *LiteratureResult) error {
	if result == nil || len(result.NewHypotheses) == 0 {
		return nil
	}

	marker := fmt.Sprintf("📚 Literature refresh at %s: %d new hypotheses from %d verified claims",
		result.Timestamp.Format("15:04"),
		len(result.NewHypotheses),
		result.TotalClaimsFound)
	prog.AppendLearning(marker)

	for _, h := range result.NewHypotheses {
		prog.AddHypothesis(h)
	}

	return nil
}

// detectSinWebsearch finds the sin-websearch binary on PATH.
func detectSinWebsearch() string {
	for _, name := range []string{"sin-websearch", "sinwebsearch"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

func truncateClaim(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
