// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the Tavily search engine.
// Docs: tavily_test.doc.md
package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTavilyClassifyDepth(t *testing.T) {
	cases := []struct {
		query string
		want  string
	}{
		// advanced keywords
		{"research papers on climate", "advanced"},
		{"compare golang vs rust performance", "advanced"},
		{"deep analysis of market trends", "advanced"},
		{"detailed report on quantum computing", "advanced"},
		// fast keywords
		{"latest news on AI", "fast"},
		{"recent developments in tech", "fast"},
		{"today's weather", "fast"},
		{"news headlines", "fast"},
		// ultra-fast: <5 words, no special keywords
		{"golang", "ultra-fast"},
		{"hello world", "ultra-fast"},
		{"a b c d", "ultra-fast"},
		// basic: >=5 words, no special keywords
		{"this is a longer query with many words", "basic"},
		{"how to configure nginx reverse proxy", "basic"},
	}

	for _, c := range cases {
		got := classifyDepth(c.query)
		if got != c.want {
			t.Errorf("classifyDepth(%q) = %q, want %q", c.query, got, c.want)
		}
	}
}

func TestTavilyEngineName(t *testing.T) {
	e := NewTavilyEngine("key")
	if got := e.Name(); got != "tavily" {
		t.Errorf("Name() = %q, want tavily", got)
	}
}

func TestTavilyEngineSearchMissingKey(t *testing.T) {
	e := NewTavilyEngine("")
	_, err := e.Search(context.Background(), "query", 5)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "api key not configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTavilyEngineSearchQueryConstruction(t *testing.T) {
	respPayload := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"title":   "Result",
				"url":     "https://example.com",
				"content": "content",
				"score":   0.95,
			},
		},
	}
	data, _ := json.Marshal(respPayload)

	transport := &tavilyBodyTransport{body: string(data), status: http.StatusOK}
	e := NewTavilyEngine("test-key")
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "research climate change", 5)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}

	// Verify bearer auth header.
	if transport.gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want 'Bearer test-key'", transport.gotAuth)
	}

	// Verify POST body fields.
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(transport.gotBody), &body); err != nil {
		t.Fatalf("could not parse request body: %v", err)
	}
	if body["query"] != "research climate change" {
		t.Errorf("body query = %v, want 'research climate change'", body["query"])
	}
	if body["search_depth"] != "advanced" {
		t.Errorf("body search_depth = %v, want 'advanced' (query contains 'research')", body["search_depth"])
	}
	if body["max_results"] != float64(5) {
		t.Errorf("body max_results = %v, want 5", body["max_results"])
	}
	if body["include_answer"] != "basic" {
		t.Errorf("body include_answer = %v, want 'basic'", body["include_answer"])
	}
	if body["auto_parameters"] != true {
		t.Errorf("body auto_parameters = %v, want true", body["auto_parameters"])
	}
}

func TestTavilyEngineSearchAnswerExtraction(t *testing.T) {
	respPayload := map[string]interface{}{
		"answer": "The answer is 42.",
		"results": []map[string]interface{}{
			{
				"title":   "Result",
				"url":     "https://example.com",
				"content": "content",
				"score":   0.9,
			},
		},
	}
	data, _ := json.Marshal(respPayload)

	transport := &tavilyBodyTransport{body: string(data), status: http.StatusOK}
	e := NewTavilyEngine("key")
	e.client = &http.Client{Transport: transport}

	res, err := e.Search(context.Background(), "what is the answer", 5)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results (answer + 1 organic), got %d", len(res))
	}

	// Answer is prepended.
	if res[0].Source != "tavily_answer" {
		t.Errorf("first result source = %q, want 'tavily_answer'", res[0].Source)
	}
	if res[0].Snippet != "The answer is 42." {
		t.Errorf("first result snippet = %q, want 'The answer is 42.'", res[0].Snippet)
	}
	if res[0].Title != "Tavily Answer" {
		t.Errorf("first result title = %q, want 'Tavily Answer'", res[0].Title)
	}

	// Organic result follows.
	if res[1].Source != "tavily" {
		t.Errorf("second result source = %q, want 'tavily'", res[1].Source)
	}
	if res[1].Title != "Result" {
		t.Errorf("second result title = %q, want 'Result'", res[1].Title)
	}
	if res[1].Score != 0.9 {
		t.Errorf("second result score = %v, want 0.9", res[1].Score)
	}
}

func TestTavilyEngineSearchDefaultDepthOverride(t *testing.T) {
	respPayload := map[string]interface{}{
		"results": []map[string]interface{}{},
	}
	data, _ := json.Marshal(respPayload)

	transport := &tavilyBodyTransport{body: string(data), status: http.StatusOK}
	e := NewTavilyEngine("key")
	e.defaultDepth = "advanced"
	e.client = &http.Client{Transport: transport}

	if _, err := e.Search(context.Background(), "hello", 5); err != nil {
		t.Fatalf("Search error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(transport.gotBody), &body); err != nil {
		t.Fatalf("could not parse request body: %v", err)
	}
	// "hello" would normally classify as ultra-fast, but defaultDepth overrides.
	if body["search_depth"] != "advanced" {
		t.Errorf("defaultDepth override: search_depth = %v, want 'advanced'", body["search_depth"])
	}
}

func TestTavilyEngineSearchRateLimited(t *testing.T) {
	transport := &tavilyBodyTransport{body: "rate limited", status: http.StatusTooManyRequests}
	e := NewTavilyEngine("key")
	e.client = &http.Client{Transport: transport}

	_, err := e.Search(context.Background(), "query", 5)
	if err == nil {
		t.Fatal("expected rate-limit error")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTavilyEngineSearchLimitDefault(t *testing.T) {
	respPayload := map[string]interface{}{
		"results": []map[string]interface{}{},
	}
	data, _ := json.Marshal(respPayload)

	transport := &tavilyBodyTransport{body: string(data), status: http.StatusOK}
	e := NewTavilyEngine("key")
	e.client = &http.Client{Transport: transport}

	if _, err := e.Search(context.Background(), "hello", 0); err != nil {
		t.Fatalf("Search error: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(transport.gotBody), &body); err != nil {
		t.Fatalf("could not parse request body: %v", err)
	}
	if body["max_results"] != float64(10) {
		t.Errorf("max_results with limit=0 = %v, want 10 (default)", body["max_results"])
	}
}

// tavilyBodyTransport is a test HTTP transport that captures the POST body
// and Authorization header so tests can assert query construction.
type tavilyBodyTransport struct {
	body    string
	status  int
	gotBody string
	gotAuth string
}

func (m *tavilyBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		m.gotBody = string(b)
		req.Body = io.NopCloser(bytes.NewReader(b))
	}
	m.gotAuth = req.Header.Get("Authorization")
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
