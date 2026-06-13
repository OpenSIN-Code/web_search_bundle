// Purpose: Perplexity via OpenRouter API engine.
// Docs: internal/engines/perplexity.doc.md
package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// PerplexityEngine queries Perplexity Sonar via OpenRouter.
type PerplexityEngine struct {
	client *http.Client
	apiKey string
}

// PerplexityResult is a Perplexity-style answer.
type PerplexityResult struct {
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Snippet   string   `json:"snippet"`
	Answer    string   `json:"answer"`
	Citations []string `json:"citations"`
	Source    string   `json:"source"`
}

// NewPerplexityEngine creates a Perplexity engine using OPENROUTER_API_KEY.
func NewPerplexityEngine() *PerplexityEngine {
	return &PerplexityEngine{
		client: &http.Client{Timeout: 60 * time.Second},
		apiKey: os.Getenv("OPENROUTER_API_KEY"),
	}
}

// Name returns the engine name.
func (e *PerplexityEngine) Name() string { return "perplexity" }

// Search queries Perplexity Sonar via OpenRouter.
func (e *PerplexityEngine) Search(ctx context.Context, query string, numResults int) ([]PerplexityResult, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("perplexity: OPENROUTER_API_KEY not set")
	}

	payload := map[string]interface{}{
		"model": "perplexity/sonar",
		"messages": []map[string]string{
			{"role": "user", "content": query},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("HTTP-Referer", "https://opensin.ai")
	req.Header.Set("X-Title", "sin-websearch")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter: %s - %s", resp.Status, string(b))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("perplexity: no response")
	}

	return []PerplexityResult{{
		Title:   "Perplexity Sonar Answer",
		URL:     "https://perplexity.ai",
		Snippet: response.Choices[0].Message.Content,
		Answer:  response.Choices[0].Message.Content,
		Source:  "perplexity",
	}}, nil
}

// SearchResults adapts Perplexity results to the common Result type.
func (e *PerplexityEngine) SearchResults(ctx context.Context, query string, limit int) ([]Result, error) {
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
			Source:  "perplexity",
		})
	}
	return results, nil
}
