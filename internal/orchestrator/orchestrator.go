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
	engines      []engines.Engine
	resolver     *resolver.EntityResolver
	judge        *judge.HumorJudge
	clusterer    *clustering.Clusterer
	cache        cache.CacheInterface
	routingEnabled bool
}

// New creates an orchestrator with the given engines.
func New(engines []engines.Engine) *Orchestrator {
	return &Orchestrator{
		engines:   engines,
		resolver:  resolver.NewEntityResolver(),
		judge:     judge.NewHumorJudge(),
		clusterer: clustering.NewClusterer(),
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
func (o *Orchestrator) Search(ctx context.Context, topic string) (*SearchResult, error) {
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
			var result SearchResult
			if err := json.Unmarshal(cached, &result); err == nil {
				return &result, nil
			}
		}
	}

	// Cost-aware routing: when enabled, filter engines to the recommended subset.
	selectedEngines := o.engines
	if o.routingEnabled {
		decision := Route(topic, o.engineNames())
		selectedEngines = o.filterEngines(decision.Engines)
		fmt.Fprintf(os.Stderr, "websearch: routing \"%s\" → %s (%d engines: %v)\n",
			topic, decision.Type, len(selectedEngines), decision.Engines)
	}

	var allItems []clustering.ClusterItem
	var allResults []engines.Result
	var mu sync.Mutex
	errs := make(map[string]string)

	var wg sync.WaitGroup
	for _, engine := range selectedEngines {
		wg.Add(1)
		go func(e engines.Engine) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			res, err := e.Search(ctx, queries[0], 10)
			if err != nil {
				mu.Lock()
				errs[e.Name()] = err.Error()
				mu.Unlock()
				return
			}

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
func (o *Orchestrator) SearchStream(ctx context.Context, topic string) (<-chan StreamResult, error) {
	entity, err := o.resolver.Resolve(ctx, topic)
	if err != nil {
		return nil, err
	}
	queries := entity.ExpandQueries()
	if len(queries) == 0 {
		queries = []string{topic}
	}

	out := make(chan StreamResult, len(o.engines))
	var wg sync.WaitGroup
	for _, engine := range o.engines {
		wg.Add(1)
		go func(e engines.Engine) {
			defer wg.Done()
			engCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			res, err := e.Search(engCtx, queries[0], 10)
			if err != nil {
				select {
				case out <- StreamResult{Source: e.Name(), Error: err.Error()}:
				case <-ctx.Done():
				}
				return
			}
			select {
			case out <- StreamResult{Source: e.Name(), Items: res}:
			case <-ctx.Done():
			}
		}(engine)
	}
	go func() {
		wg.Wait()
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
