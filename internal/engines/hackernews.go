// Purpose: Hacker News Algolia search engine.
// Docs: internal/engines/hackernews.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// HackerNewsEngine queries the HN Algolia API.
type HackerNewsEngine struct {
	client *http.Client
}

// NewHackerNewsEngine creates a new HN engine.
func NewHackerNewsEngine() *HackerNewsEngine {
	return &HackerNewsEngine{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the engine name.
func (e *HackerNewsEngine) Name() string { return "hackernews" }

// Search queries HN Algolia.
func (e *HackerNewsEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://hn.algolia.com/api/v1/search?query=%s&hitsPerPage=%d", url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hackernews: %s", resp.Status)
	}

	var payload struct {
		Hits []struct {
			Title     string    `json:"title"`
			URL       string    `json:"url"`
			StoryText string    `json:"story_text"`
			Points    int       `json:"points"`
			Author    string    `json:"author"`
			CreatedAt time.Time `json:"created_at"`
			ObjectID  string    `json:"objectID"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []Result
	for _, hit := range payload.Hits {
		link := hit.URL
		if link == "" {
			link = "https://news.ycombinator.com/item?id=" + hit.ObjectID
		}
		results = append(results, Result{
			Title:      hit.Title,
			URL:        link,
			Snippet:    truncate(hit.StoryText, 300),
			Source:     "hackernews",
			Engagement: hit.Points,
			Timestamp:  hit.CreatedAt,
			Author:     hit.Author,
		})
	}

	return results, nil
}
