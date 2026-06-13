// Purpose: Factory wiring engines and orchestrator for CLI commands.
// Docs: cmd/sin-websearch/factory.doc.md
package main

import (
	"context"

	"github.com/OpenSIN-Code/web_search_bundle/internal/cache"
	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/orchestrator"
)

// resultEngine wraps an engine that exposes a SearchResults method.
type resultEngine struct {
	name   string
	search func(ctx context.Context, query string, limit int) ([]engines.Result, error)
}

func (e *resultEngine) Search(ctx context.Context, query string, limit int) ([]engines.Result, error) {
	return e.search(ctx, query, limit)
}

func (e *resultEngine) Name() string { return e.name }

// buildOrchestrator creates an orchestrator with all available engines.
func buildOrchestrator() (*orchestrator.Orchestrator, error) {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	var engList []engines.Engine
	engList = append(engList, engines.NewRedditEngine())
	engList = append(engList, engines.NewHackerNewsEngine())
	engList = append(engList, engines.NewGitHubEngine())
	engList = append(engList, engines.NewPolymarketEngine())
	engList = append(engList, engines.NewBraveEngine(cfg.BraveAPIKey))
	engList = append(engList, engines.NewBlueskyEngine())
	engList = append(engList, engines.NewXTwitterEngine())

	searx := engines.NewSearxNGEngine()
	engList = append(engList, &resultEngine{name: searx.Name(), search: searx.SearchResults})

	perplexity := engines.NewPerplexityEngine()
	engList = append(engList, &resultEngine{name: perplexity.Name(), search: perplexity.SearchResults})

	engList = append(engList, engines.NewSerpAPIEngine(cfg.SerpAPIKeys))

	c, err := cache.New(cfg.CachePath)
	if err != nil {
		return orchestrator.New(engList), nil
	}
	return orchestrator.NewWithCache(engList, c), nil
}
