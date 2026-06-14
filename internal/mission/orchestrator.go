// SPDX-License-Identifier: MIT
// Purpose: Multi-agent research missions with explore and librarian agents.
// Docs: internal/mission/orchestrator.doc.md
package mission

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/OpenSIN-Code/web_search_bundle/internal/profiles"
	"github.com/OpenSIN-Code/web_search_bundle/internal/verify"
)

// Mission is a complete multi-agent research run.
type Mission struct {
	ID              string                     `json:"id"`
	Topic           string                     `json:"topic"`
	Profile         *profiles.Profile          `json:"profile"`
	StartedAt       time.Time                  `json:"started_at"`
	CompletedAt     time.Time                  `json:"completed_at,omitempty"`
	Status          string                     `json:"status"`
	ExploreReports  []ExploreReport            `json:"explore_reports"`
	LibrarianReport *LibrarianReport           `json:"librarian_report,omitempty"`
	Verification    *verify.VerificationReport `json:"verification,omitempty"`
	AllResults      []engines.Result           `json:"all_results"`
	Synthesis       string                     `json:"synthesis"`
	Errors          []string                   `json:"errors,omitempty"`
}

// ExploreReport is the output of one explore agent.
type ExploreReport struct {
	AgentID  int              `json:"agent_id"`
	Focus    string           `json:"focus"`
	Results  []engines.Result `json:"results"`
	Sources  []string         `json:"sources"`
	Duration time.Duration    `json:"duration"`
	Error    string           `json:"error,omitempty"`
}

// LibrarianReport is the synthesis output.
type LibrarianReport struct {
	Synthesis   string   `json:"synthesis"`
	KeyFindings []string `json:"key_findings"`
	Gaps        []string `json:"gaps,omitempty"`
	Confidence  float64  `json:"confidence"`
}

// Orchestrator runs multi-agent missions.
type Orchestrator struct {
	baseOrch *orchestrator.Orchestrator
}

// NewOrchestrator creates a mission orchestrator.
func NewOrchestrator(baseOrch *orchestrator.Orchestrator) *Orchestrator {
	return &Orchestrator{baseOrch: baseOrch}
}

// Run executes a complete mission.
func (o *Orchestrator) Run(ctx context.Context, topic string, profile *profiles.Profile) (*Mission, error) {
	mission := &Mission{
		ID:        generateMissionID(),
		Topic:     topic,
		Profile:   profile,
		StartedAt: time.Now(),
		Status:    "running",
	}

	fmt.Printf("🚀 Mission %s: launching %d explore agents\n", mission.ID[:8], profile.Agents.Explore.Count)
	reports := o.runExploreAgents(ctx, topic, profile)
	mission.ExploreReports = reports

	for _, r := range reports {
		mission.AllResults = append(mission.AllResults, r.Results...)
	}
	if len(mission.AllResults) == 0 {
		mission.Status = "failed"
		mission.Errors = append(mission.Errors, "no results from any explore agent")
		return mission, fmt.Errorf("no results from explore agents")
	}

	fmt.Printf("📚 Mission %s: running librarian agents\n", mission.ID[:8])
	librarianReport, err := o.runLibrarianAgents(ctx, topic, mission.AllResults, profile)
	if err != nil {
		mission.Errors = append(mission.Errors, "librarian: "+err.Error())
	} else {
		mission.LibrarianReport = librarianReport
		mission.Synthesis = librarianReport.Synthesis
	}

	verifyEngine := verify.NewEngine(&verify.CitationDiscipline{
		MinSourcesPerClaim:  profile.Verification.MinSourcesPerClaim,
		ConfidenceThreshold: profile.Verification.ConfidenceThreshold,
		FlagContested:       profile.Verification.FlagContested,
	})
	mission.Verification = verifyEngine.Verify(topic, mission.AllResults)

	mission.CompletedAt = time.Now()
	mission.Status = "completed"
	return mission, nil
}

func (o *Orchestrator) runExploreAgents(ctx context.Context, topic string, profile *profiles.Profile) []ExploreReport {
	reports := make([]ExploreReport, profile.Agents.Explore.Count)
	var wg sync.WaitGroup
	foci := distributeFocus(profile.Agents.Explore.Count, profile.Agents.Explore.FocusDistribution)

	timeout := 30 * time.Second
	if d, err := time.ParseDuration(profile.Agents.Explore.Timeout); err == nil {
		timeout = d
	}

	for i := 0; i < profile.Agents.Explore.Count; i++ {
		wg.Add(1)
		go func(idx int, focus string) {
			defer wg.Done()
			agentCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			start := time.Now()

			// Run base search on the topic.
			res, err := o.baseOrch.Search(agentCtx, topic)
			report := ExploreReport{AgentID: idx, Focus: focus, Duration: time.Since(start)}
			if err != nil {
				report.Error = err.Error()
			} else {
				report.Results = res.Results
			}
			reports[idx] = report
		}(i, foci[i])
	}
	wg.Wait()
	return reports
}

func (o *Orchestrator) runLibrarianAgents(ctx context.Context, topic string, results []engines.Result, profile *profiles.Profile) (*LibrarianReport, error) {
	var findings []string
	for _, r := range results {
		if r.Engagement > 100 {
			findings = append(findings, r.Title)
		}
	}
	if len(findings) > 10 {
		findings = findings[:10]
	}

	var engagement int
	for _, r := range results {
		engagement += r.Engagement
	}
	confidence := 0.7
	if engagement > 1000 {
		confidence = 0.85
	} else if engagement > 100 {
		confidence = 0.75
	}

	return &LibrarianReport{
		Synthesis:   fmt.Sprintf("Research synthesis for %s: %d results found across %d sources.", topic, len(results), len(profile.Sources.Required)),
		KeyFindings: findings,
		Confidence:  confidence,
	}, nil
}

func distributeFocus(n int, dist map[string]float64) []string {
	if len(dist) == 0 {
		dist = map[string]float64{"technical": 0.25, "community": 0.25, "market": 0.25, "creator": 0.25}
	}
	foci := make([]string, n)
	keys := make([]string, 0, len(dist))
	for k := range dist {
		keys = append(keys, k)
	}
	for i := 0; i < n; i++ {
		foci[i] = keys[i%len(keys)]
	}
	return foci
}

func generateMissionID() string {
	return fmt.Sprintf("mission-%d", time.Now().UnixNano())
}
