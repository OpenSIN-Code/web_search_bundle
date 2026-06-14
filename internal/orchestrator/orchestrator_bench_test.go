// Purpose: Benchmark the fan-out search orchestrator hot paths.
// Docs: internal/orchestrator/orchestrator.doc.md
package orchestrator

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/cache"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
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
			Snippet:    "snippet " + string(rune('a'+i%26)),
			Source:     e.name,
			Engagement: (i + 1) * 50,
			Score:      float64(i + 1),
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

func BenchmarkOrchestratorSearch1(b *testing.B) {
	o := New(makeBenchEngines(1))
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = o.Search(ctx, "benchmark topic")
	}
}

func BenchmarkOrchestratorSearch4(b *testing.B) {
	o := New(makeBenchEngines(4))
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = o.Search(ctx, "benchmark topic")
	}
}

func BenchmarkOrchestratorSearch8(b *testing.B) {
	o := New(makeBenchEngines(8))
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = o.Search(ctx, "benchmark topic")
	}
}

func BenchmarkOrchestratorPulse4(b *testing.B) {
	o := New(makeBenchEngines(4))
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = o.Pulse(ctx, "benchmark topic")
	}
}

func BenchmarkOrchestratorSearchStream4(b *testing.B) {
	o := New(makeBenchEngines(4))
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch, _ := o.SearchStream(ctx, "benchmark topic")
		for range ch {
		}
	}
}

func BenchmarkOrchestratorSearchWithCache(b *testing.B) {
	c, err := cache.New(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()
	o := NewWithCache(makeBenchEngines(2), c)
	ctx := context.Background()
	// Prime cache.
	_, _ = o.Search(ctx, "cached topic")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = o.Search(ctx, "cached topic")
	}
}

func BenchmarkOrchestratorPrintSummary(b *testing.B) {
	defer suppressOutput(b)()
	o := New(makeBenchEngines(2))
	ctx := context.Background()
	res, err := o.Search(ctx, "benchmark topic")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.PrintSummary(res)
	}
}

func BenchmarkEngineNames(b *testing.B) {
	o := New(makeBenchEngines(8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = o.engineNames()
	}
}

func BenchmarkOrchestratorSearchTimeout(b *testing.B) {
	engs := []engines.Engine{
		&slowBenchEngine{name: "slow", delay: 100 * time.Millisecond},
	}
	o := New(engs)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = o.Search(ctx, "benchmark topic")
	}
}

// slowBenchEngine always sleeps longer than the orchestrator timeout.
type slowBenchEngine struct {
	name  string
	delay time.Duration
}

func (e *slowBenchEngine) Name() string { return e.name }

func (e *slowBenchEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	select {
	case <-time.After(e.delay):
		return nil, ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
