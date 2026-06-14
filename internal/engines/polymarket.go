// SPDX-License-Identifier: MIT
// Purpose: Polymarket CLOB API search engine.
// Docs: internal/engines/polymarket.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PolymarketEngine searches Polymarket prediction markets.
type PolymarketEngine struct {
	client *http.Client
}

// NewPolymarketEngine creates a Polymarket engine.
func NewPolymarketEngine() *PolymarketEngine {
	return &PolymarketEngine{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the engine name.
func (e *PolymarketEngine) Name() string { return "polymarket" }

// Search queries Polymarket active markets.
func (e *PolymarketEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://clob.polymarket.com/markets?active=true&closed=false&limit=%d", limit)

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
		return nil, fmt.Errorf("polymarket: %s", resp.Status)
	}

	var payload struct {
		Markets []struct {
			Question    string  `json:"question"`
			Slug        string  `json:"slug"`
			Description string  `json:"description"`
			Volume      float64 `json:"volume"`
			Volume24h   float64 `json:"volume24h"`
			EndDateISO  string  `json:"endDateISO"`
		} `json:"markets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []Result
	for _, m := range payload.Markets {
		if query != "" && !contains(m.Question+m.Description, query) {
			continue
		}
		results = append(results, Result{
			Title:      m.Question,
			URL:        "https://polymarket.com/event/" + m.Slug,
			Snippet:    truncate(m.Description, 300),
			Source:     "polymarket",
			Engagement: int(m.Volume24h),
			Timestamp:  parseTime(m.EndDateISO),
		})
	}

	return results, nil
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || containsInsensitive(haystack, needle)
}

func containsInsensitive(a, b string) bool {
	if len(b) > len(a) {
		return false
	}
	for i := 0; i <= len(a)-len(b); i++ {
		match := true
		for j := 0; j < len(b); j++ {
			if toLower(a[i+j]) != toLower(b[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
