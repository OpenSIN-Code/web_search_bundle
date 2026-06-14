// SPDX-License-Identifier: MIT
// Purpose: GitHub REST search engine for repositories, issues, and users.
// Docs: internal/engines/github.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// GitHubEngine searches GitHub via the REST API.
type GitHubEngine struct {
	client *http.Client
	token  string
}

// NewGitHubEngine creates a GitHub engine. Token is optional but raises rate limits.
func NewGitHubEngine() *GitHubEngine {
	return &GitHubEngine{
		client: &http.Client{Timeout: 10 * time.Second},
		token:  os.Getenv("GITHUB_TOKEN"),
	}
}

// Name returns the engine name.
func (e *GitHubEngine) Name() string { return "github" }

// Search queries GitHub repositories and issues.
func (e *GitHubEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&per_page=%d", url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if e.token != "" {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github: %s", resp.Status)
	}

	var payload struct {
		Items []struct {
			Name        string `json:"full_name"`
			URL         string `json:"html_url"`
			Description string `json:"description"`
			Stars       int    `json:"stargazers_count"`
			Owner       struct {
				Login string `json:"login"`
			} `json:"owner"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []Result
	for _, item := range payload.Items {
		results = append(results, Result{
			Title:      item.Name,
			URL:        item.URL,
			Snippet:    truncate(item.Description, 300),
			Source:     "github",
			Engagement: item.Stars,
			Timestamp:  item.UpdatedAt,
			Author:     item.Owner.Login,
		})
	}

	return results, nil
}
