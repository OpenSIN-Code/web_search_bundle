// Purpose: Unit tests for the fan-out orchestrator.
// Docs: internal/orchestrator/orchestrator_test.doc.md
package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/cache"
	"github.com/OpenSIN-Code/web_search_bundle/internal/clustering"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/resolver"
)

// testEngine is a deterministic in-memory engine for unit tests.
type testEngine struct {
	name    string
	results []engines.Result
	err     error
	delay   time.Duration
}

func (e *testEngine) Name() string { return e.name }

func (e *testEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	if e.delay > 0 {
		select {
		case <-time.After(e.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
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

func makeTestEngine(name string, results []engines.Result) engines.Engine {
	return &testEngine{name: name, results: results}
}

// counterEngine counts how many times Search is invoked.
type counterEngine struct {
	name     string
	searches int
}

func (e *counterEngine) Name() string { return e.name }

func (e *counterEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	e.searches++
	return []engines.Result{{Title: "Result" + string(rune('0'+e.searches))}}, nil
}

func TestNew(t *testing.T) {
	engs := []engines.Engine{makeTestEngine("a", nil)}
	o := New(engs)
	if o == nil {
		t.Fatal("New returned nil")
	}
	if len(o.engineNames()) != 1 {
		t.Errorf("engineNames = %d, want 1", len(o.engineNames()))
	}
}

func TestNewWithCache(t *testing.T) {
	c, err := cache.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	o := NewWithCache(nil, c)
	if o.cache != c {
		t.Error("NewWithCache did not assign cache")
	}
}

func TestSearchCombinesResults(t *testing.T) {
	engs := []engines.Engine{
		makeTestEngine("reddit", []engines.Result{{Title: "Reddit Post", Engagement: 100}}),
		makeTestEngine("hackernews", []engines.Result{{Title: "HN Post", Engagement: 50}}),
	}
	o := New(engs)
	res, err := o.Search(context.Background(), "openclaw")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 2 {
		t.Errorf("Results = %d, want 2", len(res.Results))
	}
	if len(res.Clusters) == 0 {
		t.Error("expected at least one cluster")
	}
	if len(res.Errors) != 0 {
		t.Errorf("unexpected errors: %v", res.Errors)
	}
	if len(res.BestTakes) == 0 {
		t.Error("expected best takes")
	}
	if res.Entity == nil || res.Entity.Query != "openclaw" {
		t.Errorf("unexpected entity: %v", res.Entity)
	}
}

func TestSearchRecordsErrors(t *testing.T) {
	engs := []engines.Engine{
		makeTestEngine("reddit", []engines.Result{{Title: "Reddit Post"}}),
		&testEngine{name: "failing", err: errors.New("boom")},
	}
	o := New(engs)
	res, err := o.Search(context.Background(), "openclaw")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 {
		t.Errorf("Results = %d, want 1", len(res.Results))
	}
	if res.Errors["failing"] != "boom" {
		t.Errorf("expected failing error, got: %v", res.Errors)
	}
}

func TestSearchTimeout(t *testing.T) {
	engs := []engines.Engine{
		&testEngine{name: "slow", delay: 100 * time.Millisecond, err: context.DeadlineExceeded},
	}
	o := New(engs)
	res, err := o.Search(context.Background(), "openclaw")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Error("expected timeout error recorded")
	}
}

func TestSearchCacheStoresResult(t *testing.T) {
	c, err := cache.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	counter := &counterEngine{name: "counter"}
	o := NewWithCache([]engines.Engine{counter}, c)

	ctx := context.Background()
	res, err := o.Search(ctx, "cached topic")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("search results = %d, want 1", len(res.Results))
	}
	if counter.searches != 1 {
		t.Fatalf("counter = %d, want 1", counter.searches)
	}

	// Verify the cache actually contains a non-empty entry for the expected key.
	key := cache.HashKey("cached topic", []string{"counter"})
	cached, ok, err := c.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected cache entry for key %s", key)
	}
	if len(cached) == 0 {
		t.Fatal("expected non-empty cached payload")
	}
}

func TestSearchResultRoundTrip(t *testing.T) {
	res := &SearchResult{
		Entity: &resolver.ResolvedEntity{Query: "q"},
		Results: []engines.Result{
			{Title: "T", Source: "reddit", Engagement: 10},
		},
		Clusters: []clustering.Cluster{
			{ID: "id", Title: "T", Sources: []string{"reddit"}, Items: []clustering.ClusterItem{{Source: "reddit", Title: "T", Score: 0.5}}},
		},
		BestTakes: []string{"take"},
		Errors:    map[string]string{"e": "msg"},
	}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back SearchResult
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Entity == nil || back.Entity.Query != "q" {
		t.Errorf("entity mismatch: %v", back.Entity)
	}
	if len(back.Results) != 1 {
		t.Errorf("results mismatch: %v", back.Results)
	}
}

func TestCacheDirectRoundTrip(t *testing.T) {
	c, err := cache.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	res := &SearchResult{
		Results: []engines.Result{{Title: "Cached", Source: "counter"}},
	}
	key := cache.HashKey("cached topic", []string{"counter"})
	if err := c.Set(key, []string{"counter"}, res, time.Hour); err != nil {
		t.Fatal(err)
	}
	data, ok, err := c.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	var back SearchResult
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(back.Results) != 1 || back.Results[0].Title != "Cached" {
		t.Errorf("unexpected result: %v", back.Results)
	}
}

func TestCacheRoundTripWithResolvedEntity(t *testing.T) {
	c, err := cache.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// Simulate the exact Search result shape produced by the orchestrator.
	res := &SearchResult{
		Entity: &resolver.ResolvedEntity{Query: "cached topic"},
		Results: []engines.Result{
			{Title: "Cached Post", Source: "counter", Engagement: 10},
		},
		Clusters: []clustering.Cluster{
			{ID: "cachedpost", Title: "Cached Post", Sources: []string{"counter"}},
		},
		BestTakes: []string{"Cached Post"},
		Errors:    map[string]string{},
	}
	key := cache.HashKey("cached topic", []string{"counter"})
	if err := c.Set(key, []string{"counter"}, res, time.Hour); err != nil {
		t.Fatal(err)
	}
	data, ok, err := c.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	var back SearchResult
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(back.Results) != 1 || back.Results[0].Title != "Cached Post" {
		t.Errorf("unexpected result: %v", back.Results)
	}
}

func TestSearchStream(t *testing.T) {
	engs := []engines.Engine{
		makeTestEngine("reddit", []engines.Result{{Title: "Stream A"}}),
		makeTestEngine("hackernews", []engines.Result{{Title: "Stream B"}}),
	}
	o := New(engs)
	ch, err := o.SearchStream(context.Background(), "openclaw")
	if err != nil {
		t.Fatal(err)
	}
	var got []StreamResult
	for r := range ch {
		got = append(got, r)
	}
	if len(got) != 2 {
		t.Errorf("stream results = %d, want 2", len(got))
	}
}

func TestSearchStreamError(t *testing.T) {
	engs := []engines.Engine{
		makeTestEngine("reddit", []engines.Result{{Title: "Stream A"}}),
		&testEngine{name: "failing", err: errors.New("stream boom")},
	}
	o := New(engs)
	ch, err := o.SearchStream(context.Background(), "openclaw")
	if err != nil {
		t.Fatal(err)
	}
	var errors int
	var items int
	for r := range ch {
		if r.Error != "" {
			errors++
		}
		if len(r.Items) > 0 {
			items++
		}
	}
	if errors != 1 {
		t.Errorf("errors = %d, want 1", errors)
	}
	if items != 1 {
		t.Errorf("items = %d, want 1", items)
	}
}

func TestPulse(t *testing.T) {
	engs := []engines.Engine{
		makeTestEngine("reddit", []engines.Result{{Title: "Pulse", Engagement: 42}}),
	}
	o := New(engs)
	res, err := o.Pulse(context.Background(), "openclaw")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("Results = %d, want 1", len(res.Results))
	}
	if res.Results[0].Score != 42 {
		t.Errorf("Pulse Score = %v, want 42", res.Results[0].Score)
	}
}

func TestEngineNames(t *testing.T) {
	engs := []engines.Engine{
		makeTestEngine("a", nil),
		makeTestEngine("b", nil),
	}
	o := New(engs)
	names := o.engineNames()
	if len(names) != 2 {
		t.Fatalf("names = %d, want 2", len(names))
	}
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("names = %v", names)
	}
}

func TestPrintSummary(t *testing.T) {
	engs := []engines.Engine{
		makeTestEngine("reddit", []engines.Result{{Title: "Post"}}),
	}
	o := New(engs)
	res, err := o.Search(context.Background(), "openclaw")
	if err != nil {
		t.Fatal(err)
	}

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	o.PrintSummary(res)
	_ = w.Close()
	os.Stderr = oldStderr

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "Resolved entity") {
		t.Errorf("expected summary output, got: %s", out)
	}
	if !strings.Contains(string(out), "Total results") {
		t.Errorf("expected total results, got: %s", out)
	}
}

func TestPrintSummaryWithErrors(t *testing.T) {
	res := &SearchResult{
		Entity: &resolver.ResolvedEntity{Query: "q"},
		Errors: map[string]string{"reddit": "down"},
	}
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	New(nil).PrintSummary(res)
	_ = w.Close()
	os.Stderr = oldStderr

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "down") {
		t.Errorf("expected error in summary, got: %s", out)
	}
}

func TestSearchUsesTopicQueryWhenResolverEmpty(t *testing.T) {
	engs := []engines.Engine{
		&testEngine{
			name:    "recorder",
			results: []engines.Result{{Title: "Result"}},
		},
	}
	o := New(engs)
	// Topic with no known entity mappings.
	res, err := o.Search(context.Background(), "xyz-unknown-topic")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 {
		t.Errorf("Results = %d, want 1", len(res.Results))
	}
}
