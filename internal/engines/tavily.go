// SPDX-License-Identifier: MIT
// Purpose: Tavily search engine with 4-level depth tiering and include_answer support.
// Docs: internal/engines/tavily.doc.md
package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TavilyEngine queries the Tavily Search API with query-depth tiering.
type TavilyEngine struct {
	client       *http.Client
	apiKey       string
	defaultDepth string
}

// NewTavilyEngine creates a Tavily engine with the given API key.
func NewTavilyEngine(apiKey string) *TavilyEngine {
	return &TavilyEngine{
		client: &http.Client{Timeout: 15 * time.Second},
		apiKey: apiKey,
	}
}

// SetDefaultDepth overrides the auto-classified depth with a fixed value.
func (e *TavilyEngine) SetDefaultDepth(depth string) {
	e.defaultDepth = depth
}

// Name returns the engine name.
func (e *TavilyEngine) Name() string { return "tavily" }

// classifyDepth picks a Tavily search_depth based on query complexity.
//
// Tier ladder (checked in order):
//   - advanced: research / compare / analysis / deep / detailed  (2 credits)
//   - fast:     news / latest / recent / today                   (1 credit)
//   - ultra-fast: <5 words and no special keywords               (1 credit)
//   - basic:    everything else                                  (1 credit)
func classifyDepth(query string) string {
	q := strings.ToLower(query)

	for _, kw := range []string{"research", "compare", "analysis", "deep", "detailed"} {
		if strings.Contains(q, kw) {
			return "advanced"
		}
	}

	for _, kw := range []string{"news", "latest", "recent", "today"} {
		if strings.Contains(q, kw) {
			return "fast"
		}
	}

	if len(strings.Fields(query)) < 5 {
		return "ultra-fast"
	}

	return "basic"
}

// Search queries the Tavily Search API.
func (e *TavilyEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("tavily: api key not configured")
	}
	if limit == 0 {
		limit = 10
	}

	depth := e.defaultDepth
	if depth == "" {
		depth = classifyDepth(query)
	}

	payload := map[string]interface{}{
		"query":           query,
		"search_depth":    depth,
		"max_results":     limit,
		"include_answer":  "basic",
		"auto_parameters": true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("tavily: rate limited")
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tavily: %s - %s", resp.Status, string(b))
	}

	var response struct {
		Answer  string `json:"answer"`
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var results []Result

	if response.Answer != "" {
		results = append(results, Result{
			Title:   "Tavily Answer",
			URL:     "https://tavily.com",
			Snippet: response.Answer,
			Source:  "tavily_answer",
		})
	}

	for _, r := range response.Results {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
			Source:  "tavily",
			Score:   r.Score,
		})
	}

	return results, nil
}
