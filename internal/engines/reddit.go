// SPDX-License-Identifier: MIT
// Purpose: Reddit JSON search engine (no API key required).
// Docs: internal/engines/reddit.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// RedditEngine searches public Reddit listings.
type RedditEngine struct {
	client *http.Client
}

// NewRedditEngine creates a new Reddit engine.
func NewRedditEngine() *RedditEngine {
	return &RedditEngine{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the engine name.
func (e *RedditEngine) Name() string { return "reddit" }

// Search queries Reddit for the given query.
func (e *RedditEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d", url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "sin-websearch/1.0")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("reddit: %s", resp.Status)
	}

	var payload struct {
		Data struct {
			Children []struct {
				Data struct {
					Title     string  `json:"title"`
					URL       string  `json:"url"`
					Selftext  string  `json:"selftext"`
					Ups       int     `json:"ups"`
					Subreddit string  `json:"subreddit"`
					Author    string  `json:"author"`
					Created   float64 `json:"created_utc"`
					Permalink string  `json:"permalink"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []Result
	for _, child := range payload.Data.Children {
		d := child.Data
		link := d.URL
		if link == "" || link[0] == '/' {
			link = "https://www.reddit.com" + d.Permalink
		}
		results = append(results, Result{
			Title:      d.Title,
			URL:        link,
			Snippet:    truncate(d.Selftext, 300),
			Source:     "reddit",
			Engagement: d.Ups,
			Timestamp:  time.Unix(int64(d.Created), 0),
			Author:     d.Author,
			Subreddit:  d.Subreddit,
		})
	}

	return results, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
