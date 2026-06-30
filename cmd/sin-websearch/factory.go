// SPDX-License-Identifier: MIT
// Purpose: Factory wiring engines and orchestrator for CLI commands.
// Docs: cmd/sin-websearch/factory.doc.md
package main

import (
	"context"
	"fmt"
	"os"

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

	// Tavily engine (depth-tiered, include_answer) — wired when API key present.
	if cfg.TavilyAPIKey != "" {
		te := engines.NewTavilyEngine(cfg.TavilyAPIKey)
		if cfg.TavilyDefaultDepth != "" {
			te.SetDefaultDepth(cfg.TavilyDefaultDepth)
		}
		engList = append(engList, te)
		fmt.Fprintf(os.Stderr, "websearch: tavily engine enabled (depth=%s)\n", cfg.TavilyDefaultDepth)
	}

	// DuckDuckGo engine (free, keyless) — wired when enabled (default true).
	if cfg.DuckDuckGoEnabled {
		engList = append(engList, engines.NewDuckDuckGoEngine())
		fmt.Fprintf(os.Stderr, "websearch: duckduckgo engine enabled (free)\n")
	}

	c, err := cache.New(cfg.CachePath)
	if err != nil {
		return orchestrator.New(engList), nil
	}

	// Wrap with semantic cache when enabled (default true).
	if cfg.SemanticCacheEnabled {
		embedder := cache.NewEmbedder()
		sc := cache.NewSemanticCache(c, embedder)
		if cfg.SemanticCacheThreshold > 0 && cfg.SemanticCacheThreshold != 0.85 {
			sc.SetThreshold(cfg.SemanticCacheThreshold)
		}
		fmt.Fprintf(os.Stderr, "websearch: semantic cache enabled (threshold=%.2f)\n", cfg.SemanticCacheThreshold)
		return orchestrator.NewWithCache(engList, c), nil
	}

	return orchestrator.NewWithCache(engList, c), nil
}
