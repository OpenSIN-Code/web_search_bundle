// Purpose: SearxNG proxy engine for privacy-focused search.
// Docs: internal/engines/searxng.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// SearxNGEngine queries a SearxNG instance.
type SearxNGEngine struct {
	client  *http.Client
	baseURL string
}

// SearxNGResult is a result from a SearxNG instance.
type SearxNGResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Engine  string `json:"engine"`
	Source  string `json:"source"`
}

// NewSearxNGEngine creates a SearxNG engine. Reads SEARXNG_URL from env.
func NewSearxNGEngine() *SearxNGEngine {
	baseURL := os.Getenv("SEARXNG_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &SearxNGEngine{
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
	}
}

// Name returns the engine name.
func (e *SearxNGEngine) Name() string { return "searxng" }

// Search queries SearxNG.
func (e *SearxNGEngine) Search(ctx context.Context, query string, numResults int) ([]SearxNGResult, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("categories", "general")
	params.Add("time_range", "month")
	params.Add("pageno", "1")

	reqURL := fmt.Sprintf("%s/search?%s", e.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("searxng unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("searxng: %s - %s", resp.Status, string(body))
	}

	var response struct {
		Results []struct {
			Title   string   `json:"title"`
			URL     string   `json:"url"`
			Content string   `json:"content"`
			Engines []string `json:"engines"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var results []SearxNGResult
	for i, r := range response.Results {
		if i >= numResults {
			break
		}
		engine := "searxng"
		if len(r.Engines) > 0 {
			engine = r.Engines[0]
		}
		results = append(results, SearxNGResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
			Engine:  engine,
			Source:  "searxng",
		})
	}

	return results, nil
}

// SearchResults adapts SearxNG results to the common Result type.
func (e *SearxNGEngine) SearchResults(ctx context.Context, query string, limit int) ([]Result, error) {
	res, err := e.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	var results []Result
	for _, r := range res {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Snippet,
			Source:  "searxng",
		})
	}
	return results, nil
}
