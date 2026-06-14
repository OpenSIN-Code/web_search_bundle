// SPDX-License-Identifier: MIT
// Purpose: Brave Search API engine.
// Docs: internal/engines/brave.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// BraveEngine queries Brave Search.
type BraveEngine struct {
	client *http.Client
	apiKey string
}

// NewBraveEngine creates a Brave engine with the given API key.
func NewBraveEngine(apiKey string) *BraveEngine {
	return &BraveEngine{
		client: &http.Client{Timeout: 10 * time.Second},
		apiKey: apiKey,
	}
}

// Name returns the engine name.
func (e *BraveEngine) Name() string { return "brave" }

// Search queries Brave Search.
func (e *BraveEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("brave: api key not configured")
	}
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d", url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("brave: %s", resp.Status)
	}

	var payload struct {
		Web struct {
			Results []struct {
				Title   string `json:"title"`
				URL     string `json:"url"`
				Desc    string `json:"description"`
				Profile struct {
					Name string `json:"name"`
				} `json:"profile"`
			} `json:"results"`
		} `json:"web"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []Result
	for _, r := range payload.Web.Results {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Desc,
			Source:  "brave",
			Author:  r.Profile.Name,
		})
	}

	return results, nil
}
