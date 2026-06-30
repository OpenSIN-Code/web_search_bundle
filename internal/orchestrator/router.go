// SPDX-License-Identifier: MIT
// Purpose: Cost-aware provider router that classifies queries and routes
// them to the cheapest capable engine, reducing API costs.
// Docs: internal/orchestrator/router.doc.md
package orchestrator

import (
	"strings"
)

// QueryType classifies a search query by intent.
type QueryType int

const (
	QueryTypeSimple QueryType = iota
	QueryTypeNews
	QueryTypeSocial
	QueryTypeTech
	QueryTypeResearch
	QueryTypeVideo
	QueryTypeDocs
)

// String returns the human-readable name of the query type.
func (q QueryType) String() string {
	switch q {
	case QueryTypeNews:
		return "news"
	case QueryTypeSocial:
		return "social"
	case QueryTypeTech:
		return "tech"
	case QueryTypeResearch:
		return "research"
	case QueryTypeVideo:
		return "video"
	case QueryTypeDocs:
		return "docs"
	default:
		return "simple"
	}
}

// RoutingDecision is the output of Route — which engines to query and how.
type RoutingDecision struct {
	Type        QueryType
	Engines     []string
	Reason      string
	MaxParallel int
}

// CostEstimate describes the per-engine cost profile for a routing decision.
type CostEstimate struct {
	Engine         string
	IsFree         bool
	CreditsPerCall int
}

// keywordSets maps query types to their triggering keywords (lowercased).
var keywordSets = []struct {
	Type     QueryType
	Keywords []string
}{
	{QueryTypeNews, []string{"news", "latest", "recent", "today", "breaking"}},
	{QueryTypeSocial, []string{"reddit", "social", "twitter", "engagement", "viral"}},
	{QueryTypeTech, []string{"github", "code", "programming", "api", "tech"}},
	{QueryTypeResearch, []string{"research", "compare", "analysis", "deep", "study"}},
	{QueryTypeVideo, []string{"video", "youtube", "watch"}},
	{QueryTypeDocs, []string{"docs", "documentation", "api reference", "how to use", "library", "framework", "sdk", "react", "nextjs", "next.js", "vue", "svelte", "prisma", "supabase", "tailwind", "django", "fastapi", "express", "nestjs", "typescript", "install", "import", "npm install", "pip install", "cargo add", "go get"}},
}

// ClassifyQuery inspects the query string and returns its QueryType.
// The first matching keyword set wins; order follows the declaration above.
func ClassifyQuery(query string) QueryType {
	lq := strings.ToLower(query)
	for _, ks := range keywordSets {
		for _, kw := range ks.Keywords {
			if strings.Contains(lq, kw) {
				return ks.Type
			}
		}
	}
	return QueryTypeSimple
}

// routingTable maps each QueryType to its preferred engines and parallelism.
var routingTable = map[QueryType]struct {
	Engines     []string
	Reason      string
	MaxParallel int
}{
	QueryTypeSimple: {
		Engines:     []string{"duckduckgo"},
		Reason:      "simple query: free DuckDuckGo sufficient",
		MaxParallel: 1,
	},
	QueryTypeNews: {
		Engines:     []string{"tavily", "brave"},
		Reason:      "news query: paid fast engines for freshness",
		MaxParallel: 2,
	},
	QueryTypeSocial: {
		Engines:     []string{"reddit", "hackernews"},
		Reason:      "social query: free social-signal engines",
		MaxParallel: 2,
	},
	QueryTypeTech: {
		Engines:     []string{"github", "brave"},
		Reason:      "tech query: GitHub code search + Brave web",
		MaxParallel: 2,
	},
	QueryTypeResearch: {
		Engines:     []string{"duckduckgo", "tavily", "brave", "reddit", "hackernews", "github", "youtube", "context7"},
		Reason:      "research query: broad fan-out across all engines",
		MaxParallel: 5,
	},
	QueryTypeVideo: {
		Engines:     []string{"youtube"},
		Reason:      "video query: YouTube only",
		MaxParallel: 1,
	},
	QueryTypeDocs: {
		Engines:     []string{"context7", "duckduckgo"},
		Reason:      "docs query: context7 for official docs + DuckDuckGo fallback",
		MaxParallel: 2,
	},
}

// costTable maps engine names to their cost profile.
var costTable = map[string]struct {
	IsFree         bool
	CreditsPerCall int
}{
	"duckduckgo":  {true, 0},
	"context7":    {true, 0},
	"reddit":      {true, 0},
	"hackernews":  {true, 0},
	"github":      {true, 0},
	"youtube":     {true, 0},
	"searxng":     {true, 0},
	"bluesky":     {true, 0},
	"x":           {true, 0},
	"polymarket":  {true, 0},
	"brave":       {false, 1},
	"tavily":      {false, 1},
	"serpapi":     {false, 1},
	"perplexity":  {false, 1},
}

// Route classifies the query and selects the best engines from the
// available set. If none of the recommended engines are available,
// it falls back to all available engines.
func Route(query string, availableEngines []string) RoutingDecision {
	qt := ClassifyQuery(query)
	entry, ok := routingTable[qt]
	if !ok {
		entry = routingTable[QueryTypeSimple]
	}

	available := make(map[string]bool, len(availableEngines))
	for _, e := range availableEngines {
		available[strings.ToLower(e)] = true
	}

	var selected []string
	for _, e := range entry.Engines {
		if available[strings.ToLower(e)] {
			selected = append(selected, e)
		}
	}

	if len(selected) == 0 {
		return RoutingDecision{
			Type:        qt,
			Engines:     availableEngines,
			Reason:      qt.String() + " query: recommended engines unavailable, falling back to all available",
			MaxParallel: len(availableEngines),
		}
	}

	return RoutingDecision{
		Type:        qt,
		Engines:     selected,
		Reason:      entry.Reason,
		MaxParallel: entry.MaxParallel,
	}
}

// EstimateCost returns per-engine cost estimates for a routing decision.
func EstimateCost(decision RoutingDecision) []CostEstimate {
	estimates := make([]CostEstimate, 0, len(decision.Engines))
	for _, e := range decision.Engines {
		ce, ok := costTable[strings.ToLower(e)]
		if !ok {
			ce = struct {
				IsFree         bool
				CreditsPerCall int
			}{false, 1}
		}
		estimates = append(estimates, CostEstimate{
			Engine:         e,
			IsFree:         ce.IsFree,
			CreditsPerCall: ce.CreditsPerCall,
		})
	}
	return estimates
}
