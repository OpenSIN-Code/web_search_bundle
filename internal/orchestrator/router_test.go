// SPDX-License-Identifier: MIT
// Purpose: Unit tests for the cost-aware provider router.
// Docs: internal/orchestrator/router_test.doc.md
package orchestrator

import (
	"sort"
	"testing"
)

func TestRouter(t *testing.T) {
	t.Run("ClassifyQuery", testRouterClassifyQuery)
	t.Run("ClassifyQueryCaseInsensitive", testRouterClassifyQueryCaseInsensitive)
	t.Run("RouteSimple", testRouterRouteSimple)
	t.Run("RouteNews", testRouterRouteNews)
	t.Run("RouteSocial", testRouterRouteSocial)
	t.Run("RouteTech", testRouterRouteTech)
	t.Run("RouteResearch", testRouterRouteResearch)
	t.Run("RouteVideo", testRouterRouteVideo)
	t.Run("RouteFallbackAllAvailable", testRouterRouteFallbackAllAvailable)
	t.Run("RouteFallbackPartialMatch", testRouterRouteFallbackPartialMatch)
	t.Run("RouteFallbackSocialPartialMatch", testRouterRouteFallbackSocialPartialMatch)
	t.Run("RouteEmptyAvailable", testRouterRouteEmptyAvailable)
	t.Run("RouteCaseInsensitiveEngineMatch", testRouterRouteCaseInsensitiveEngineMatch)
	t.Run("RouteReasonPopulated", testRouterRouteReasonPopulated)
	t.Run("EstimateCost", testRouterEstimateCost)
	t.Run("EstimateCostUnknownEngine", testRouterEstimateCostUnknownEngine)
	t.Run("EstimateCostEmpty", testRouterEstimateCostEmpty)
	t.Run("EstimateCostForRouteResult", testRouterEstimateCostForRouteResult)
}

func testRouterClassifyQuery(t *testing.T) {
	tests := []struct {
		query string
		want  QueryType
	}{
		// News keywords
		{"latest AI news", QueryTypeNews},
		{"breaking story today", QueryTypeNews},
		{"recent developments in quantum", QueryTypeNews},
		// Social keywords
		{"reddit discussion about AI", QueryTypeSocial},
		{"twitter viral thread", QueryTypeSocial},
		{"social engagement metrics", QueryTypeSocial},
		// Tech keywords
		{"github code search", QueryTypeTech},
		{"programming best practices", QueryTypeTech},
		{"api documentation", QueryTypeTech},
		// Research keywords
		{"research paper on LLMs", QueryTypeResearch},
		{"compare Rust vs Go", QueryTypeResearch},
		{"deep analysis of the market", QueryTypeResearch},
		// Video keywords
		{"youtube video tutorial", QueryTypeVideo},
		{"watch this clip", QueryTypeVideo},
		// Simple (default)
		{"what is the weather", QueryTypeSimple},
		{"pizza near me", QueryTypeSimple},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			got := ClassifyQuery(tc.query)
			if got != tc.want {
				t.Errorf("ClassifyQuery(%q) = %s, want %s", tc.query, got, tc.want)
			}
		})
	}
}

func testRouterClassifyQueryCaseInsensitive(t *testing.T) {
	if got := ClassifyQuery("LATEST NEWS TODAY"); got != QueryTypeNews {
		t.Errorf("ClassifyQuery uppercase = %s, want news", got)
	}
	if got := ClassifyQuery("GitHub CODE"); got != QueryTypeTech {
		t.Errorf("ClassifyQuery mixed = %s, want tech", got)
	}
}

func testRouterRouteSimple(t *testing.T) {
	avail := []string{"duckduckgo", "brave", "reddit", "hackernews", "github", "youtube", "tavily"}
	dec := Route("what is the weather", avail)
	if dec.Type != QueryTypeSimple {
		t.Errorf("Type = %s, want simple", dec.Type)
	}
	if !sliceEqual(dec.Engines, []string{"duckduckgo"}) {
		t.Errorf("Engines = %v, want [duckduckgo]", dec.Engines)
	}
	if dec.MaxParallel != 1 {
		t.Errorf("MaxParallel = %d, want 1", dec.MaxParallel)
	}
}

func testRouterRouteNews(t *testing.T) {
	avail := []string{"duckduckgo", "brave", "tavily", "reddit", "hackernews", "github", "youtube"}
	dec := Route("breaking news today", avail)
	if dec.Type != QueryTypeNews {
		t.Errorf("Type = %s, want news", dec.Type)
	}
	if !sliceEqualSet(dec.Engines, []string{"tavily", "brave"}) {
		t.Errorf("Engines = %v, want {tavily, brave}", dec.Engines)
	}
	if dec.MaxParallel != 2 {
		t.Errorf("MaxParallel = %d, want 2", dec.MaxParallel)
	}
}

func testRouterRouteSocial(t *testing.T) {
	avail := []string{"duckduckgo", "brave", "reddit", "hackernews", "github", "youtube"}
	dec := Route("reddit viral thread", avail)
	if dec.Type != QueryTypeSocial {
		t.Errorf("Type = %s, want social", dec.Type)
	}
	if !sliceEqualSet(dec.Engines, []string{"reddit", "hackernews"}) {
		t.Errorf("Engines = %v, want {reddit, hackernews}", dec.Engines)
	}
	if dec.MaxParallel != 2 {
		t.Errorf("MaxParallel = %d, want 2", dec.MaxParallel)
	}
}

func testRouterRouteTech(t *testing.T) {
	avail := []string{"duckduckgo", "brave", "reddit", "hackernews", "github", "youtube"}
	dec := Route("github code search", avail)
	if dec.Type != QueryTypeTech {
		t.Errorf("Type = %s, want tech", dec.Type)
	}
	if !sliceEqualSet(dec.Engines, []string{"github", "brave"}) {
		t.Errorf("Engines = %v, want {github, brave}", dec.Engines)
	}
	if dec.MaxParallel != 2 {
		t.Errorf("MaxParallel = %d, want 2", dec.MaxParallel)
	}
}

func testRouterRouteResearch(t *testing.T) {
	avail := []string{"duckduckgo", "brave", "tavily", "reddit", "hackernews", "github", "youtube"}
	dec := Route("deep research analysis", avail)
	if dec.Type != QueryTypeResearch {
		t.Errorf("Type = %s, want research", dec.Type)
	}
	if dec.MaxParallel != 5 {
		t.Errorf("MaxParallel = %d, want 5", dec.MaxParallel)
	}
	if len(dec.Engines) < 3 {
		t.Errorf("expected broad fan-out, got %d engines", len(dec.Engines))
	}
}

func testRouterRouteVideo(t *testing.T) {
	avail := []string{"duckduckgo", "brave", "reddit", "hackernews", "github", "youtube"}
	dec := Route("watch youtube video", avail)
	if dec.Type != QueryTypeVideo {
		t.Errorf("Type = %s, want video", dec.Type)
	}
	if !sliceEqual(dec.Engines, []string{"youtube"}) {
		t.Errorf("Engines = %v, want [youtube]", dec.Engines)
	}
	if dec.MaxParallel != 1 {
		t.Errorf("MaxParallel = %d, want 1", dec.MaxParallel)
	}
}

func testRouterRouteFallbackAllAvailable(t *testing.T) {
	avail := []string{"searxng", "perplexity", "x"}
	dec := Route("what is the weather", avail)
	if dec.Type != QueryTypeSimple {
		t.Errorf("Type = %s, want simple", dec.Type)
	}
	if !sliceEqualSet(dec.Engines, avail) {
		t.Errorf("Engines = %v, want fallback to all available %v", dec.Engines, avail)
	}
	if dec.MaxParallel != len(avail) {
		t.Errorf("MaxParallel = %d, want %d", dec.MaxParallel, len(avail))
	}
}

func testRouterRouteFallbackPartialMatch(t *testing.T) {
	avail := []string{"brave", "searxng"}
	dec := Route("breaking news today", avail)
	if dec.Type != QueryTypeNews {
		t.Errorf("Type = %s, want news", dec.Type)
	}
	if !sliceEqualSet(dec.Engines, []string{"brave"}) {
		t.Errorf("Engines = %v, want [brave] (partial match)", dec.Engines)
	}
	if dec.MaxParallel != 2 {
		t.Errorf("MaxParallel = %d, want 2 (keep table value for partial)", dec.MaxParallel)
	}
}

func testRouterRouteFallbackSocialPartialMatch(t *testing.T) {
	avail := []string{"reddit", "searxng"}
	dec := Route("viral reddit post", avail)
	if dec.Type != QueryTypeSocial {
		t.Errorf("Type = %s, want social", dec.Type)
	}
	if !sliceEqual(dec.Engines, []string{"reddit"}) {
		t.Errorf("Engines = %v, want [reddit]", dec.Engines)
	}
}

func testRouterRouteEmptyAvailable(t *testing.T) {
	dec := Route("what is the weather", nil)
	if dec.Type != QueryTypeSimple {
		t.Errorf("Type = %s, want simple", dec.Type)
	}
	if len(dec.Engines) != 0 {
		t.Errorf("Engines = %v, want empty", dec.Engines)
	}
}

func testRouterRouteCaseInsensitiveEngineMatch(t *testing.T) {
	avail := []string{"DuckDuckGo", "Reddit", "GitHub"}
	dec := Route("what is the weather", avail)
	if !sliceEqualLower(dec.Engines, []string{"DuckDuckGo"}) {
		t.Errorf("Engines = %v, want [DuckDuckGo]", dec.Engines)
	}
	dec = Route("reddit viral thread", avail)
	if !sliceEqualLower(dec.Engines, []string{"Reddit"}) {
		t.Errorf("Engines = %v, want [Reddit]", dec.Engines)
	}
}

func testRouterRouteReasonPopulated(t *testing.T) {
	avail := []string{"duckduckgo"}
	dec := Route("hello world", avail)
	if dec.Reason == "" {
		t.Error("expected non-empty Reason")
	}
}

func testRouterEstimateCost(t *testing.T) {
	dec := RoutingDecision{
		Engines: []string{"duckduckgo", "brave", "reddit", "tavily", "youtube"},
	}
	estimates := EstimateCost(dec)
	if len(estimates) != 5 {
		t.Fatalf("EstimateCost returned %d, want 5", len(estimates))
	}

	want := map[string]CostEstimate{
		"duckduckgo": {Engine: "duckduckgo", IsFree: true, CreditsPerCall: 0},
		"brave":      {Engine: "brave", IsFree: false, CreditsPerCall: 1},
		"reddit":     {Engine: "reddit", IsFree: true, CreditsPerCall: 0},
		"tavily":     {Engine: "tavily", IsFree: false, CreditsPerCall: 1},
		"youtube":    {Engine: "youtube", IsFree: true, CreditsPerCall: 0},
	}
	for _, est := range estimates {
		exp, ok := want[est.Engine]
		if !ok {
			t.Errorf("unexpected engine in estimates: %s", est.Engine)
			continue
		}
		if est.IsFree != exp.IsFree {
			t.Errorf("%s IsFree = %v, want %v", est.Engine, est.IsFree, exp.IsFree)
		}
		if est.CreditsPerCall != exp.CreditsPerCall {
			t.Errorf("%s CreditsPerCall = %d, want %d", est.Engine, est.CreditsPerCall, exp.CreditsPerCall)
		}
	}
}

func testRouterEstimateCostUnknownEngine(t *testing.T) {
	dec := RoutingDecision{
		Engines: []string{"some-unknown-engine"},
	}
	estimates := EstimateCost(dec)
	if len(estimates) != 1 {
		t.Fatalf("EstimateCost returned %d, want 1", len(estimates))
	}
	if estimates[0].IsFree {
		t.Error("unknown engine should default to paid")
	}
	if estimates[0].CreditsPerCall != 1 {
		t.Errorf("unknown engine CreditsPerCall = %d, want 1", estimates[0].CreditsPerCall)
	}
}

func testRouterEstimateCostEmpty(t *testing.T) {
	dec := RoutingDecision{}
	estimates := EstimateCost(dec)
	if len(estimates) != 0 {
		t.Errorf("EstimateCost returned %d, want 0", len(estimates))
	}
}

func testRouterEstimateCostForRouteResult(t *testing.T) {
	avail := []string{"duckduckgo", "brave", "reddit", "hackernews", "github", "youtube", "tavily"}
	dec := Route("github code search", avail)
	estimates := EstimateCost(dec)

	var freeCount, paidCount int
	for _, est := range estimates {
		if est.IsFree {
			freeCount++
		} else {
			paidCount++
		}
	}
	if freeCount == 0 {
		t.Error("expected at least one free engine in tech route")
	}
	if paidCount == 0 {
		t.Error("expected at least one paid engine in tech route")
	}
}

// --- helpers ---

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sliceEqualSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sa := append([]string{}, a...)
	sb := append([]string{}, b...)
	sort.Strings(sa)
	sort.Strings(sb)
	return sliceEqual(sa, sb)
}

func sliceEqualLower(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !stringsEqualFold(a[i], b[i]) {
			return false
		}
	}
	return true
}

func stringsEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
