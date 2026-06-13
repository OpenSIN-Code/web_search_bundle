// Purpose: SerpAPI search engine with rotating key pool.
// Docs: internal/engines/serpapi.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/pool"
)

// SerpAPIEngine queries Google via SerpAPI with key rotation.
type SerpAPIEngine struct {
	client *http.Client
	pool   *pool.KeyPool
}

// NewSerpAPIEngine creates a SerpAPI engine with the given keys.
func NewSerpAPIEngine(keys []string) *SerpAPIEngine {
	return &SerpAPIEngine{
		client: &http.Client{Timeout: 15 * time.Second},
		pool:   pool.New(keys),
	}
}

// Name returns the engine name.
func (e *SerpAPIEngine) Name() string { return "serpapi" }

// Search queries SerpAPI Google.
func (e *SerpAPIEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if e.pool.IsEmpty() {
		return nil, fmt.Errorf("serpapi: no api keys configured")
	}
	if limit == 0 {
		limit = 10
	}

	key, err := e.pool.Next()
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(
		"https://serpapi.com/search?q=%s&num=%d&api_key=%s&source=sin-websearch",
		url.QueryEscape(query), limit, url.QueryEscape(key),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		e.pool.Ban(key, 5*time.Minute)
		return nil, fmt.Errorf("serpapi: rate limited")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("serpapi: %s", resp.Status)
	}

	var payload struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []Result
	for _, r := range payload.OrganicResults {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.Link,
			Snippet: r.Snippet,
			Source:  "serpapi",
		})
	}

	return results, nil
}
