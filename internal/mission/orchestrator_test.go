// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the multi-agent mission orchestrator.
// Docs: internal/mission/orchestrator_test.doc.md
package mission

import (
	"context"
	"os"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/OpenSIN-Code/web_search_bundle/internal/profiles"
)

// suppressTestOutput redirects stdout/stderr to /dev/null for the test duration.
func suppressTestOutput(t *testing.T) func() {
	t.Helper()
	oldStdout, oldStderr := os.Stdout, os.Stderr
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = null, null
	return func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
		_ = null.Close()
	}
}

// testEngine is a deterministic in-memory engine for unit tests.
type testEngine struct {
	name    string
	results []engines.Result
	err     error
}

func (e *testEngine) Name() string { return e.name }

func (e *testEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	if e.err != nil {
		return nil, e.err
	}
	res := make([]engines.Result, len(e.results))
	copy(res, e.results)
	for i := range res {
		res[i].Source = e.name
	}
	return res, nil
}

func testProfile(count int) *profiles.Profile {
	return &profiles.Profile{
		Name: "test",
		Agents: profiles.AgentConfig{
			Explore: profiles.ExploreAgentConfig{
				Count:    count,
				Parallel: true,
				Timeout:  "5s",
				FocusDistribution: map[string]float64{
					"technical": 0.25,
					"community": 0.25,
					"market":    0.25,
					"creator":   0.25,
				},
			},
			Librarian: profiles.LibrarianAgentConfig{Count: 2, Tasks: []string{"synthesize"}},
		},
		Sources:      profiles.SourceConfig{Required: []string{"reddit", "github"}},
		Output:       profiles.OutputConfig{Primary: "json"},
		Verification: profiles.VerificationConfig{MinSourcesPerClaim: 1, ConfidenceThreshold: 0.5, FlagContested: true},
	}
}

func TestNewOrchestrator(t *testing.T) {
	base := orchestrator.New(nil)
	mo := NewOrchestrator(base)
	if mo == nil {
		t.Fatal("NewOrchestrator returned nil")
	}
	if mo.baseOrch != base {
		t.Error("NewOrchestrator did not store base orchestrator")
	}
}

func TestDistributeFocusDefault(t *testing.T) {
	got := distributeFocus(4, nil)
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}
	want := map[string]int{"technical": 1, "community": 1, "market": 1, "creator": 1}
	counts := make(map[string]int)
	for _, f := range got {
		counts[f]++
	}
	for k, v := range want {
		if counts[k] != v {
			t.Errorf("focus %q count = %d, want %d", k, counts[k], v)
		}
	}
}

func TestDistributeFocusCustom(t *testing.T) {
	dist := map[string]float64{"a": 0.5, "b": 0.5}
	got := distributeFocus(5, dist)
	if len(got) != 5 {
		t.Fatalf("len = %d, want 5", len(got))
	}
	counts := make(map[string]int)
	for _, f := range got {
		counts[f]++
	}
	if counts["a"] != 3 {
		t.Errorf("a count = %d, want 3", counts["a"])
	}
	if counts["b"] != 2 {
		t.Errorf("b count = %d, want 2", counts["b"])
	}
}

func TestGenerateMissionID(t *testing.T) {
	id := generateMissionID()
	if id == "" {
		t.Fatal("expected non-empty mission id")
	}
	if id[:8] != "mission-" {
		t.Errorf("id prefix = %q, want mission-", id[:8])
	}
}

func TestRunLibrarianAgents(t *testing.T) {
	base := orchestrator.New(nil)
	mo := NewOrchestrator(base)
	profile := testProfile(2)
	results := []engines.Result{
		{Title: "hot finding", Engagement: 200},
		{Title: "warm finding", Engagement: 150},
		{Title: "cold finding", Engagement: 50},
	}
	report, err := mo.runLibrarianAgents(context.Background(), "topic", results, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected report")
	}
	if len(report.KeyFindings) != 2 {
		t.Errorf("findings = %d, want 2", len(report.KeyFindings))
	}
	if report.Confidence != 0.75 {
		t.Errorf("confidence = %v, want 0.75", report.Confidence)
	}
	if report.Synthesis == "" {
		t.Error("expected non-empty synthesis")
	}
}

func TestRunLibrarianAgentsHighEngagement(t *testing.T) {
	base := orchestrator.New(nil)
	mo := NewOrchestrator(base)
	profile := testProfile(2)
	results := make([]engines.Result, 20)
	for i := range results {
		results[i] = engines.Result{Title: "finding", Engagement: 101}
	}
	report, err := mo.runLibrarianAgents(context.Background(), "topic", results, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Confidence != 0.85 {
		t.Errorf("confidence = %v, want 0.85", report.Confidence)
	}
	if len(report.KeyFindings) != 10 {
		t.Errorf("findings = %d, want 10", len(report.KeyFindings))
	}
}

func TestRunExploreAgents(t *testing.T) {
	defer suppressTestOutput(t)()
	base := orchestrator.New([]engines.Engine{&testEngine{name: "a", results: []engines.Result{{Title: "r1"}}}})
	mo := NewOrchestrator(base)
	profile := testProfile(2)
	reports := mo.runExploreAgents(context.Background(), "topic", profile)
	if len(reports) != 2 {
		t.Fatalf("reports = %d, want 2", len(reports))
	}
	for i, r := range reports {
		if r.AgentID != i {
			t.Errorf("report[%d].AgentID = %d, want %d", i, r.AgentID, i)
		}
		if len(r.Results) != 1 {
			t.Errorf("report[%d].Results = %d, want 1", i, len(r.Results))
		}
	}
}

func TestRunExploreAgentsNoResults(t *testing.T) {
	defer suppressTestOutput(t)()
	base := orchestrator.New([]engines.Engine{&testEngine{name: "a", results: nil}})
	mo := NewOrchestrator(base)
	profile := testProfile(1)
	reports := mo.runExploreAgents(context.Background(), "topic", profile)
	if len(reports) != 1 {
		t.Fatalf("reports = %d, want 1", len(reports))
	}
	if len(reports[0].Results) != 0 {
		t.Errorf("results = %d, want 0", len(reports[0].Results))
	}
}

func TestRunCompletes(t *testing.T) {
	defer suppressTestOutput(t)()
	base := orchestrator.New([]engines.Engine{&testEngine{name: "a", results: []engines.Result{{Title: "r1", Snippet: "r1 has 2 features"}}}})
	mo := NewOrchestrator(base)
	profile := testProfile(2)
	mission, err := mo.Run(context.Background(), "test topic", profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mission.Status != "completed" {
		t.Errorf("status = %q, want completed", mission.Status)
	}
	if len(mission.ExploreReports) != 2 {
		t.Errorf("explore reports = %d, want 2", len(mission.ExploreReports))
	}
	if mission.LibrarianReport == nil {
		t.Error("expected librarian report")
	}
	if mission.Verification == nil {
		t.Error("expected verification report")
	}
	if mission.Synthesis == "" {
		t.Error("expected non-empty synthesis")
	}
}

func TestRunNoResultsFails(t *testing.T) {
	defer suppressTestOutput(t)()
	base := orchestrator.New([]engines.Engine{&testEngine{name: "a", results: nil}})
	mo := NewOrchestrator(base)
	profile := testProfile(1)
	mission, err := mo.Run(context.Background(), "empty topic", profile)
	if err == nil {
		t.Fatal("expected error for no results")
	}
	if mission.Status != "failed" {
		t.Errorf("status = %q, want failed", mission.Status)
	}
	if len(mission.Errors) == 0 {
		t.Error("expected recorded errors")
	}
}
