// SPDX-License-Identifier: MIT
// Purpose: Bluesky AT Protocol search engine.
// Docs: internal/engines/bluesky.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// BlueskyEngine searches Bluesky via the public search API.
type BlueskyEngine struct {
	client *http.Client
}

// NewBlueskyEngine creates a Bluesky engine.
func NewBlueskyEngine() *BlueskyEngine {
	return &BlueskyEngine{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the engine name.
func (e *BlueskyEngine) Name() string { return "bluesky" }

// Search queries Bluesky.
func (e *BlueskyEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://search.bsky.app/search/posts?q=%s&limit=%d", url.QueryEscape(query), limit)

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
		return nil, fmt.Errorf("bluesky: %s", resp.Status)
	}

	var posts []struct {
		Post struct {
			Author struct {
				DisplayName string `json:"displayName"`
				Handle      string `json:"handle"`
			} `json:"author"`
			Record struct {
				Text      string    `json:"text"`
				CreatedAt time.Time `json:"createdAt"`
			} `json:"record"`
			URI string `json:"uri"`
		} `json:"post"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}

	var results []Result
	for _, p := range posts {
		results = append(results, Result{
			Title:     truncate(p.Post.Record.Text, 120),
			URL:       atURIToWeb(p.Post.URI),
			Snippet:   p.Post.Record.Text,
			Source:    "bluesky",
			Timestamp: p.Post.Record.CreatedAt,
			Author:    p.Post.Author.Handle,
		})
	}

	return results, nil
}

func atURIToWeb(uri string) string {
	if len(uri) < 5 || uri[:4] != "at://" {
		return uri
	}
	parts := uri[5:]
	return "https://bsky.app/profile/" + parts
}
