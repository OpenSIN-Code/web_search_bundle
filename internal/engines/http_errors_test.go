// SPDX-License-Identifier: MIT
// Purpose: Hermetic tests for HTTP engine error branches.
// Docs: http_errors_test.doc.md
package engines

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBlueskySearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusInternalServerError}
	e := NewBlueskyEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bluesky") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBlueskySearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewBlueskyEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBlueskySearchTransportError(t *testing.T) {
	transport := &errorRoundTripper{err: errors.New("network down")}
	e := NewBlueskyEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBraveSearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusForbidden}
	e := NewBraveEngine("key")
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "brave") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBraveSearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewBraveEngine("key")
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGitHubSearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusUnauthorized}
	e := NewGitHubEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "github") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGitHubSearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewGitHubEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGitHubSearchWithToken(t *testing.T) {
	transport := &mockRoundTripperWithRequest{body: `{"items":[]}`, status: http.StatusOK}
	e := NewGitHubEngine()
	e.token = "token"
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if !strings.Contains(transport.lastReq.Header.Get("Authorization"), "Bearer token") {
		t.Errorf("expected token header, got %q", transport.lastReq.Header.Get("Authorization"))
	}
}

func TestHackerNewsSearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusBadRequest}
	e := NewHackerNewsEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "hackernews") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHackerNewsSearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewHackerNewsEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPerplexitySearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusPaymentRequired}
	e := NewPerplexityEngine()
	e.apiKey = "key"
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "openrouter") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPerplexitySearchEmptyChoices(t *testing.T) {
	transport := &mockRoundTripper{body: `{"choices":[]}`, status: http.StatusOK}
	e := NewPerplexityEngine()
	e.apiKey = "key"
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no response") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPerplexitySearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewPerplexityEngine()
	e.apiKey = "key"
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPerplexitySearchResultsError(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusForbidden}
	e := NewPerplexityEngine()
	e.apiKey = "key"
	e.client = &http.Client{Transport: transport}
	_, err := e.SearchResults(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPolymarketSearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusServiceUnavailable}
	e := NewPolymarketEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "polymarket") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPolymarketSearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewPolymarketEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPolymarketSearchFilterQuery(t *testing.T) {
	transport := &mockRoundTripper{body: `{"markets":[{"question":"Will it rain?","slug":"rain","description":"desc","volume":1,"volume24h":1,"endDateISO":"2024-01-01T00:00:00Z"},{"question":"Other market","slug":"other","description":"x","volume":1,"volume24h":1,"endDateISO":"2024-01-01T00:00:00Z"}]}`, status: http.StatusOK}
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

func TestRedditSearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusTooManyRequests}
	e := NewRedditEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "reddit") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRedditSearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewRedditEngine()
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRedditSearchSelfPostURL(t *testing.T) {
	payload := `{"data":{"children":[{"data":{"title":"Self post","url":"","selftext":"text","ups":5,"subreddit":"test","author":"u","created_utc":1700000000,"permalink":"/r/test/comments/1/x"}}]}}`
	transport := &mockRoundTripper{body: payload, status: http.StatusOK}
	e := NewRedditEngine()
	e.client = &http.Client{Transport: transport}
	res, err := e.Search(context.Background(), "q", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	want := "https://www.reddit.com/r/test/comments/1/x"
	if res[0].URL != want {
		t.Errorf("URL = %q, want %q", res[0].URL, want)
	}
}

func TestSearxNGSearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusBadGateway}
	e := NewSearxNGEngine()
	e.client = &http.Client{Transport: transport}
	e.baseURL = "http://localhost"
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "searxng") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSearxNGSearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewSearxNGEngine()
	e.client = &http.Client{Transport: transport}
	e.baseURL = "http://localhost"
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearxNGSearchResultsError(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusBadGateway}
	e := NewSearxNGEngine()
	e.client = &http.Client{Transport: transport}
	e.baseURL = "http://localhost"
	_, err := e.SearchResults(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSerpAPISearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusBadRequest}
	e := NewSerpAPIEngine([]string{"key"})
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "serpapi") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSerpAPISearchInvalidJSON(t *testing.T) {
	transport := &mockRoundTripper{body: "not json", status: http.StatusOK}
	e := NewSerpAPIEngine([]string{"key"})
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestXTwitterSearchNon200(t *testing.T) {
	transport := &mockRoundTripper{body: "error", status: http.StatusUnauthorized}
	e := NewXTwitterEngine()
	e.auth = "token"
	e.client = &http.Client{Transport: transport}
	_, err := e.Search(context.Background(), "q", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "x:") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestXTwitterSearchWithAuth(t *testing.T) {
	payload := `{"globalObjects":{"tweets":{"123":{"full_text":"hello","created_at":"Mon Jan 01 00:00:00 +0000 2024","user_id_str":"456","retweet_count":1,"favorite_count":2}},"users":{"456":{"screen_name":"user"}}}}`
	transport := &mockRoundTripper{body: payload, status: http.StatusOK}
	e := NewXTwitterEngine()
	e.auth = "token"
	e.client = &http.Client{Transport: transport}
	res, err := e.Search(context.Background(), "q", 1)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Author != "user" {
		t.Errorf("author = %q, want user", res[0].Author)
	}
}

func TestXTwitterParseAdaptiveResponseInvalidJSON(t *testing.T) {
	e := NewXTwitterEngine()
	_, err := e.parseAdaptiveResponse(strings.NewReader("not json"))
	if err == nil {
		t.Fatal("expected error")
	}
}

// errorRoundTripper fails every request with a fixed error.
type errorRoundTripper struct {
	err error
}

func (e *errorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, e.err
}

// mockRoundTripperWithRequest captures the last request for header inspection.
type mockRoundTripperWithRequest struct {
	body    string
	status  int
	lastReq *http.Request
}

func (m *mockRoundTripperWithRequest) RoundTrip(req *http.Request) (*http.Response, error) {
	m.lastReq = req
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

