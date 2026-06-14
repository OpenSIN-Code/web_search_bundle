// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for HTTP-based search engines using httptest.
// Docs: search_engines_test.doc.md

package engines

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
)

func TestRedditEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"children": []map[string]interface{}{
				{
					"data": map[string]interface{}{
						"title":       "Go 1.24",
						"url":         "https://go.dev",
						"selftext":    "hello world",
						"ups":         42,
						"subreddit":   "golang",
						"author":      "user",
						"created_utc": 1700000000,
						"permalink":   "/r/golang/comments/1/x",
					},
				},
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewRedditEngine()
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "go", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Title != "Go 1.24" {
		t.Errorf("unexpected title: %s", res[0].Title)
	}
	if res[0].URL != "https://go.dev" {
		t.Errorf("unexpected URL: %s", res[0].URL)
	}
}

func TestBlueskyEngineSearchWithMockClient(t *testing.T) {
	posts := []map[string]interface{}{
		{
			"post": map[string]interface{}{
				"author": map[string]interface{}{
					"displayName": "User",
					"handle":      "user.bsky.social",
				},
				"record": map[string]interface{}{
					"text":      "hello bluesky",
					"createdAt": "2024-01-01T00:00:00Z",
				},
				"uri": "at://user.bsky.social/post/123",
			},
		},
	}
	data, _ := json.Marshal(posts)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewBlueskyEngine()
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "hello", 5)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Author != "user.bsky.social" {
		t.Errorf("unexpected author: %s", res[0].Author)
	}
	if res[0].URL != "at://user.bsky.social/post/123" { // current behavior of atURIToWeb
		t.Errorf("unexpected URL: %s", res[0].URL)
	}
}

func TestHackerNewsEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"hits": []map[string]interface{}{
			{
				"title":      "Show HN",
				"url":        "https://example.com",
				"story_text": "story text",
				"points":     100,
				"author":     "user",
				"created_at": "2024-01-01T00:00:00Z",
				"objectID":   "123",
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewHackerNewsEngine()
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "show hn", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Title != "Show HN" {
		t.Errorf("unexpected title: %s", res[0].Title)
	}
}

func TestGitHubEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"full_name":        "org/repo",
				"html_url":         "https://github.com/org/repo",
				"description":      "repo description",
				"stargazers_count": 50,
				"updated_at":       "2024-01-01T00:00:00Z",
				"owner":            map[string]string{"login": "owner"},
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewGitHubEngine()
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "repo", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Engagement != 50 {
		t.Errorf("unexpected engagement: %d", res[0].Engagement)
	}
}

func TestBraveEngineSearchMissingKey(t *testing.T) {
	e := NewBraveEngine("")
	_, err := e.Search(context.Background(), "go", 1)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "api key not configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBraveEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"web": map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"title":       "Brave",
					"url":         "https://brave.com",
					"description": "desc",
					"profile":     map[string]string{"name": "Brave"},
				},
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewBraveEngine("key")
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "brave", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Title != "Brave" {
		t.Errorf("unexpected title: %s", res[0].Title)
	}
}

func TestPolymarketEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"markets": []map[string]interface{}{
			{
				"question":    "Will it rain?",
				"slug":        "will-it-rain",
				"description": "desc",
				"volume":      1000.0,
				"volume24h":   100.0,
				"endDateISO":  "2024-12-31T00:00:00Z",
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewPolymarketEngine()
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "rain", 10)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Title != "Will it rain?" {
		t.Errorf("unexpected title: %s", res[0].Title)
	}
}

func TestSearxNGEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"title":   "Result",
				"url":     "https://result.com",
				"content": "content",
				"engines": []string{"duckduckgo"},
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewSearxNGEngine()
	e.client = &http.Client{Transport: transport}
	e.baseURL = "http://localhost" // query path is appended to this by Search.

	res, err := e.Search(context.Background(), "query", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Engine != "duckduckgo" {
		t.Errorf("unexpected engine: %s", res[0].Engine)
	}

	adapted, err := e.SearchResults(context.Background(), "query", 1)
	if err != nil {
		t.Fatalf("SearchResults error: %v", err)
	}
	if len(adapted) != 1 {
		t.Fatalf("expected 1 adapted result, got %d", len(adapted))
	}
	if adapted[0].Source != "searxng" {
		t.Errorf("unexpected source: %s", adapted[0].Source)
	}
}

func TestPerplexityEngineSearchMissingKey(t *testing.T) {
	e := NewPerplexityEngine()
	_, err := e.Search(context.Background(), "go", 1)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "OPENROUTER_API_KEY not set") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPerplexityEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]string{"content": "answer"},
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewPerplexityEngine()
	e.apiKey = "key"
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "q", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Answer != "answer" {
		t.Errorf("unexpected answer: %s", res[0].Answer)
	}

	adapted, err := e.SearchResults(context.Background(), "q", 1)
	if err != nil {
		t.Fatalf("SearchResults error: %v", err)
	}
	if len(adapted) != 1 {
		t.Fatalf("expected 1 adapted result, got %d", len(adapted))
	}
}

func TestSerpAPIEngineSearchEmptyPool(t *testing.T) {
	e := NewSerpAPIEngine([]string{})
	_, err := e.Search(context.Background(), "go", 1)
	if err == nil {
		t.Fatal("expected error for empty pool")
	}
	if !strings.Contains(err.Error(), "no api keys configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSerpAPIEngineSearchWithMockClient(t *testing.T) {
	payload := map[string]interface{}{
		"organic_results": []map[string]interface{}{
			{
				"title":   "Result",
				"link":    "https://result.com",
				"snippet": "snippet",
			},
		},
	}
	data, _ := json.Marshal(payload)

	transport := &mockRoundTripper{body: string(data), status: http.StatusOK}
	e := NewSerpAPIEngine([]string{"key"})
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "q", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Source != "serpapi" {
		t.Errorf("unexpected source: %s", res[0].Source)
	}
}

func TestSerpAPIEngineSearchRateLimited(t *testing.T) {
	transport := &mockRoundTripper{body: "rate limited", status: http.StatusTooManyRequests}
	e := NewSerpAPIEngine([]string{"key"})
	e.client = &http.Client{Transport: transport}

	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestXTwitterEngineSearchNoAuth(t *testing.T) {
	e := NewXTwitterEngine()
	e.auth = "" // ensure no auth
	_, err := e.Search(context.Background(), "go", 1)
	if err == nil {
		t.Fatal("expected error for missing auth")
	}
	if !strings.Contains(err.Error(), "no browser session found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestXTwitterEngineParseAdaptiveResponse(t *testing.T) {
	payload := map[string]interface{}{
		"globalObjects": map[string]interface{}{
			"tweets": map[string]interface{}{
				"123": map[string]interface{}{
					"full_text":      "hello world",
					"created_at":     "Mon Jan 01 00:00:00 +0000 2024",
					"user_id_str":    "456",
					"retweet_count":  5,
					"favorite_count": 10,
				},
			},
			"users": map[string]interface{}{
				"456": map[string]string{
					"screen_name": "user",
				},
			},
		},
	}
	data, _ := json.Marshal(payload)
	e := NewXTwitterEngine()
	res, err := e.parseAdaptiveResponse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("parseAdaptiveResponse error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Author != "user" {
		t.Errorf("unexpected author: %s", res[0].Author)
	}
}

func TestYouTubeEngineSearchWithMockSidecar(t *testing.T) {
	// Use a temp HOME so the sidecar manager uses a hermetic bin directory.
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := filepath.Join(home, ".sin-websearch", "bin")
	if err := os.MkdirAll(binDir, 0750); err != nil {
		t.Fatal(err)
	}

	// Create a fake yt-dlp binary that prints canned JSON.
	fakeBin := filepath.Join(binDir, "yt-dlp")
	script := `#!/bin/sh
printf '%s\n' '{"title":"Video","webpage_url":"https://youtube.com/watch?v=1","view_count":100,"like_count":10,"channel":"Channel","automatic_captions":{"en":[{"ext":"vtt","url":"https://subs"}]}}'
`
	if err := os.WriteFile(fakeBin, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	sc, err := sidecar.NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	e := NewYouTubeEngine(sc)
	res, err := e.Search(context.Background(), "query", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Title != "Video" {
		t.Errorf("unexpected title: %s", res[0].Title)
	}
	if res[0].Views != 100 {
		t.Errorf("unexpected views: %d", res[0].Views)
	}

	adapted, err := e.SearchResults(context.Background(), "query", 1)
	if err != nil {
		t.Fatalf("SearchResults error: %v", err)
	}
	if len(adapted) != 1 {
		t.Fatalf("expected 1 adapted result, got %d", len(adapted))
	}
	if adapted[0].Source != "youtube" {
		t.Errorf("unexpected source: %s", adapted[0].Source)
	}
}

// mockRoundTripper is a test HTTP transport that returns a fixed response.
type mockRoundTripper struct {
	body      string
	status    int
	lastURL   string
	lastHost  string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.lastURL = req.URL.String()
	m.lastHost = req.URL.Host
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
