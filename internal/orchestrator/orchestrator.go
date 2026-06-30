// SPDX-License-Identifier: MIT
// Purpose: Parallel fan-out orchestrator across all search engines.
// Docs: internal/orchestrator/orchestrator.doc.md
package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/cache"
	"github.com/OpenSIN-Code/web_search_bundle/internal/clustering"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/judge"
	"github.com/OpenSIN-Code/web_search_bundle/internal/resolver"
)

// Orchestrator coordinates searches across multiple sources.
type Orchestrator struct {
	engines        []engines.Engine
	resolver       *resolver.EntityResolver
	judge          *judge.HumorJudge
	clusterer      *clustering.Clusterer
	cache          cache.CacheInterface
	routingEnabled bool
	stats          *StatsRegistry
	singleflight   singleFlightGroup
}

// New creates an orchestrator with the given engines.
func New(engines []engines.Engine) *Orchestrator {
	return &Orchestrator{
		engines:      engines,
		resolver:     resolver.NewEntityResolver(),
		judge:        judge.NewHumorJudge(),
		clusterer:    clustering.NewClusterer(),
		stats:        NewStatsRegistry(),
	}
}

// NewWithCache creates an orchestrator with a persistent cache.
func NewWithCache(engines []engines.Engine, c *cache.Cache) *Orchestrator {
	o := New(engines)
	o.cache = cache.NewCacheAdapter(c)
	return o
}

// NewWithCacheInterface creates an orchestrator with any CacheInterface
// (e.g. SemanticCache). This is the preferred constructor when semantic
// caching is enabled.
func NewWithCacheInterface(engines []engines.Engine, c cache.CacheInterface) *Orchestrator {
	o := New(engines)
	o.cache = c
	return o
}

// SetRoutingEnabled toggles cost-aware provider routing. When true, Search
// classifies the query and only fans out to the recommended engines instead
// of all engines. When false (default), all engines are queried in parallel.
func (o *Orchestrator) SetRoutingEnabled(enabled bool) {
	o.routingEnabled = enabled
}

// Stats returns the stats registry for health reporting.
func (o *Orchestrator) Stats() *StatsRegistry {
	return o.stats
}

// SearchResult is the combined output of a fan-out search.
type SearchResult struct {
	Entity    *resolver.ResolvedEntity `json:"entity,omitempty"`
	Clusters  []clustering.Cluster     `json:"clusters,omitempty"`
	Results   []engines.Result         `json:"results,omitempty"`
	BestTakes []string                 `json:"best_takes,omitempty"`
	Errors    map[string]string        `json:"errors,omitempty"`
}

// StreamResult is a single result emitted during streaming search.
type StreamResult struct {
	Source string           `json:"source"`
	Items  []engines.Result `json:"items"`
	Error  string           `json:"error,omitempty"`
}

// Search runs a fan-out query across all configured engines.
// Uses singleflight to coalesce identical concurrent queries into one engine call.
func (o *Orchestrator) Search(ctx context.Context, topic string) (*SearchResult, error) {
	return o.singleflight.Do(topic, func() (*SearchResult, error) {
		return o.searchInternal(ctx, topic)
	})
}

func (o *Orchestrator) searchInternal(ctx context.Context, topic string) (*SearchResult, error) {
	entity, err := o.resolver.Resolve(ctx, topic)
	if err != nil {
		return nil, err
	}

	queries := entity.ExpandQueries()
	if len(queries) == 0 {
		queries = []string{topic}
	}

	// Cache lookup
	if o.cache != nil {
		cached, ok, err := o.cache.Get(topic, o.engineNames())
		if err == nil && ok {
			o.stats.RecordCacheHit("orchestrator")
			var result SearchResult
			if err := json.Unmarshal(cached, &result); err == nil {
				return &result, nil
			}
		}
	}

	// Cost-aware routing: when enabled, filter engines to the recommended subset.
	selectedEngines := o.engines
	maxParallel := len(selectedEngines)
	if o.routingEnabled {
		decision := Route(topic, o.engineNames())
		selectedEngines = o.filterEngines(decision.Engines)
		maxParallel = decision.MaxParallel
		if maxParallel < 1 {
			maxParallel = 1
		}
		fmt.Fprintf(os.Stderr, "websearch: routing \"%s\" → %s (%d engines, maxParallel=%d)\n",
			topic, decision.Type, len(selectedEngines), maxParallel)
	}

	// Semaphore to enforce MaxParallel from router.
	sema := make(chan struct{}, maxParallel)

	var allItems []clustering.ClusterItem
	var allResults []engines.Result
	var mu sync.Mutex
	errs := make(map[string]string)

	var wg sync.WaitGroup
	for _, engine := range selectedEngines {
		wg.Add(1)
		go func(e engines.Engine) {
			defer wg.Done()

			// Acquire semaphore slot.
			select {
			case sema <- struct{}{}:
				defer func() { <-sema }()
			case <-ctx.Done():
				return
			}

			o.stats.RecordRequest(e.Name())
			start := time.Now()

			engCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			res, err := e.Search(engCtx, queries[0], 10)
			latency := time.Since(start)

			if err != nil {
				o.stats.RecordFailure(e.Name())
				mu.Lock()
				errs[e.Name()] = err.Error()
				mu.Unlock()
				return
			}

			o.stats.RecordSuccess(e.Name(), latency)

			mu.Lock()
			for _, r := range res {
				score := o.judge.ScoreResult(r.Title, r.Engagement, 0)
				allItems = append(allItems, clustering.ClusterItem{
					Source: r.Source,
					Title:  r.Title,
					URL:    r.URL,
					Score:  score.Virality*0.5 + score.Humor*0.5,
				})
				allResults = append(allResults, r)
			}
			mu.Unlock()
		}(engine)
	}
	wg.Wait()

	// Deduplicate results by URL, summing scores from multiple engines.
	allResults = dedupResults(allResults)

	clusters := o.clusterer.Cluster(allItems)

	var quotes []struct {
		Text    string
		Upvotes int
	}
	for _, item := range allItems {
		quotes = append(quotes, struct {
			Text    string
			Upvotes int
		}{Text: item.Title, Upvotes: int(item.Score * 1000)})
	}
	bestTakes := o.judge.BestTakes(quotes, 5)

	result := &SearchResult{
		Entity:    entity,
		Clusters:  clusters,
		Results:   allResults,
		BestTakes: bestTakes,
		Errors:    errs,
	}

	// Cache store
	if o.cache != nil {
		data, err := json.Marshal(result)
		if err == nil {
			_ = o.cache.Set(topic, o.engineNames(), data, 24*time.Hour)
		}
	}

	return result, nil
}

func (o *Orchestrator) engineNames() []string {
	var names []string
	for _, e := range o.engines {
		names = append(names, e.Name())
	}
	return names
}

// filterEngines returns only the engines whose names appear in the filter list.
func (o *Orchestrator) filterEngines(names []string) []engines.Engine {
	if len(names) == 0 {
		return o.engines
	}
	wanted := make(map[string]bool, len(names))
	for _, n := range names {
		wanted[n] = true
	}
	var filtered []engines.Engine
	for _, e := range o.engines {
		if wanted[e.Name()] {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// dedupResults removes duplicate results by URL, summing scores when the same
// URL appears across multiple engines. Results without a URL are kept as-is.
func dedupResults(results []engines.Result) []engines.Result {
	if len(results) == 0 {
		return results
	}
	seen := make(map[string]int, len(results)) // URL → index in output
	var out []engines.Result

	for _, r := range results {
		if r.URL == "" {
			out = append(out, r)
			continue
		}
		if idx, ok := seen[r.URL]; ok {
			// Merge: sum scores, keep the higher-engagement version.
			out[idx].Score += r.Score
			if r.Engagement > out[idx].Engagement {
				out[idx].Engagement = r.Engagement
			}
			// Append source if different.
			if r.Source != out[idx].Source {
				out[idx].Source = out[idx].Source + "," + r.Source
			}
		} else {
			seen[r.URL] = len(out)
			out = append(out, r)
		}
	}
	return out
}

// Pulse runs a quick social-pulse search focused on engagement.
func (o *Orchestrator) Pulse(ctx context.Context, topic string) (*SearchResult, error) {
	res, err := o.Search(ctx, topic)
	if err != nil {
		return nil, err
	}
	// Boost engagement scoring.
	for i := range res.Results {
		res.Results[i].Score = float64(res.Results[i].Engagement)
	}
	return res, nil
}

// SearchStream runs a fan-out search and streams results per engine as they arrive.
// Results are cached after all engines complete, identical to Search().
func (o *Orchestrator) SearchStream(ctx context.Context, topic string) (<-chan StreamResult, error) {
	// Check cache first — if hit, emit from cache and return immediately.
	if o.cache != nil {
		cached, ok, err := o.cache.Get(topic, o.engineNames())
		if err == nil && ok {
			o.stats.RecordCacheHit("orchestrator")
			var result SearchResult
			if err := json.Unmarshal(cached, &result); err == nil {
				out := make(chan StreamResult, 1)
				out <- StreamResult{Source: "cache", Items: result.Results}
				close(out)
				return out, nil
			}
		}
	}

	entity, err := o.resolver.Resolve(ctx, topic)
	if err != nil {
		return nil, err
	}
	queries := entity.ExpandQueries()
	if len(queries) == 0 {
		queries = []string{topic}
	}

	// Determine selected engines and maxParallel.
	selectedEngines := o.engines
	maxParallel := len(selectedEngines)
	if o.routingEnabled {
		decision := Route(topic, o.engineNames())
		selectedEngines = o.filterEngines(decision.Engines)
		maxParallel = decision.MaxParallel
		if maxParallel < 1 {
			maxParallel = 1
		}
	}

	sema := make(chan struct{}, maxParallel)
	out := make(chan StreamResult, len(selectedEngines))

	var allResults []engines.Result
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, engine := range selectedEngines {
		wg.Add(1)
		go func(e engines.Engine) {
			defer wg.Done()

			select {
			case sema <- struct{}{}:
				defer func() { <-sema }()
			case <-ctx.Done():
				return
			}

			o.stats.RecordRequest(e.Name())
			start := time.Now()

			engCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			res, err := e.Search(engCtx, queries[0], 10)
			latency := time.Since(start)

			if err != nil {
				o.stats.RecordFailure(e.Name())
				select {
				case out <- StreamResult{Source: e.Name(), Error: err.Error()}:
				case <-ctx.Done():
				}
				return
			}

			o.stats.RecordSuccess(e.Name(), latency)

			mu.Lock()
			allResults = append(allResults, res...)
			mu.Unlock()

			select {
			case out <- StreamResult{Source: e.Name(), Items: res}:
			case <-ctx.Done():
			}
		}(engine)
	}

	go func() {
		wg.Wait()
		// Cache the combined results after all engines complete.
		if o.cache != nil && len(allResults) > 0 {
			deduped := dedupResults(allResults)
			result := &SearchResult{
				Entity:  entity,
				Results: deduped,
			}
			if data, err := json.Marshal(result); err == nil {
				_ = o.cache.Set(topic, o.engineNames(), data, 24*time.Hour)
			}
		}
		close(out)
	}()

	return out, nil
}

// PrintSummary prints a human-readable summary to stderr.
func (o *Orchestrator) PrintSummary(sr *SearchResult) {
	fmt.Fprintf(os.Stderr, "Resolved entity: %s\n", sr.Entity.Query)
	fmt.Fprintf(os.Stderr, "Sources: %d\n", len(sr.Clusters))
	fmt.Fprintf(os.Stderr, "Total results: %d\n", len(sr.Results))
	if len(sr.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Errors:\n")
		for k, v := range sr.Errors {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", k, v)
		}
	}
}
