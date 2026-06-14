// SPDX-License-Identifier: MIT
// Purpose: Benchmark multi-agent mission orchestration hot paths.
// Docs: internal/mission/orchestrator.doc.md
package mission

import (
	"context"
	"os"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
	"github.com/OpenSIN-Code/web_search_bundle/internal/profiles"
)

// suppressOutput redirects stdout/stderr to /dev/null for the benchmark duration.
func suppressOutput(b *testing.B) func() {
	oldStdout, oldStderr := os.Stdout, os.Stderr
	null, err := os.Open(os.DevNull)
	if err != nil {
		b.Fatal(err)
	}
	os.Stdout, os.Stderr = null, null
	return func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
		_ = null.Close()
	}
}

// benchEngine is a fast in-memory engine for benchmarks.
type benchEngine struct {
	name string
}

func (e *benchEngine) Name() string { return e.name }

func (e *benchEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	if limit == 0 {
		limit = 10
	}
	res := make([]engines.Result, limit)
	for i := 0; i < limit; i++ {
		res[i] = engines.Result{
			Title:      query + " result " + string(rune('a'+i%26)),
			URL:        "https://example.com/" + e.name + "/" + string(rune('a'+i%26)),
			Source:     e.name,
			Engagement: (i + 1) * 50,
		}
	}
	return res, nil
}

func makeBenchEngines(n int) []engines.Engine {
	engs := make([]engines.Engine, n)
	for i := 0; i < n; i++ {
		engs[i] = &benchEngine{name: string(rune('a' + i%26))}
	}
	return engs
}

func benchmarkProfile(count int) *profiles.Profile {
	return &profiles.Profile{
		Name: "benchmark",
		Agents: profiles.AgentConfig{
			Explore: profiles.ExploreAgentConfig{
				Count:              count,
				Parallel:           true,
				Timeout:            "5s",
				MaxResultsPerAgent: 10,
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
		Verification: profiles.VerificationConfig{MinSourcesPerClaim: 1, ConfidenceThreshold: 0.5},
	}
}

func BenchmarkMissionRun4(b *testing.B) {
	defer suppressOutput(b)()
	base := orchestrator.New(makeBenchEngines(2))
	mo := NewOrchestrator(base)
	profile := benchmarkProfile(4)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mo.Run(ctx, "benchmark topic", profile)
	}
}

func BenchmarkMissionRun8(b *testing.B) {
	defer suppressOutput(b)()
	base := orchestrator.New(makeBenchEngines(2))
	mo := NewOrchestrator(base)
	profile := benchmarkProfile(8)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mo.Run(ctx, "benchmark topic", profile)
	}
}

func BenchmarkRunExploreAgents(b *testing.B) {
	defer suppressOutput(b)()
	base := orchestrator.New(makeBenchEngines(2))
	mo := NewOrchestrator(base)
	profile := benchmarkProfile(8)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mo.runExploreAgents(ctx, "benchmark topic", profile)
	}
}

func BenchmarkRunLibrarianAgents(b *testing.B) {
	base := orchestrator.New(makeBenchEngines(2))
	mo := NewOrchestrator(base)
	profile := benchmarkProfile(4)
	results := make([]engines.Result, 100)
	for i := 0; i < 100; i++ {
		results[i] = engines.Result{Title: "finding " + string(rune('a'+i%26)), Engagement: (i + 1) * 10}
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mo.runLibrarianAgents(ctx, "benchmark topic", results, profile)
	}
}

func BenchmarkDistributeFocus(b *testing.B) {
	dist := map[string]float64{"technical": 0.25, "community": 0.25, "market": 0.25, "creator": 0.25}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = distributeFocus(8, dist)
	}
}

func BenchmarkGenerateMissionID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateMissionID()
	}
}

func BenchmarkMissionRunTimeout(b *testing.B) {
	defer suppressOutput(b)()
	base := orchestrator.New(makeBenchEngines(2))
	mo := NewOrchestrator(base)
	profile := benchmarkProfile(4)
	profile.Agents.Explore.Timeout = "100ms"
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mo.Run(ctx, "benchmark topic", profile)
	}
}
