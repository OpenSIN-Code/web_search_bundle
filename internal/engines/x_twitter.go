// Purpose: X/Twitter search engine using browser session cookies.
// Docs: internal/engines/x_twitter.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/session"
)

// XTwitterEngine searches X/Twitter using a logged-in browser session.
type XTwitterEngine struct {
	client *http.Client
	auth   string
}

// NewXTwitterEngine creates an X/Twitter engine using local browser cookies.
func NewXTwitterEngine() *XTwitterEngine {
	browser := session.NewBrowserSession()
	token, _ := browser.GetXAuthToken()
	return &XTwitterEngine{
		client: &http.Client{Timeout: 15 * time.Second},
		auth:   token,
	}
}

// Name returns the engine name.
func (e *XTwitterEngine) Name() string { return "x" }

// Search queries X/Twitter via the public search endpoint (requires auth).
func (e *XTwitterEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if e.auth == "" {
		return nil, fmt.Errorf("x: no browser session found; log in via Chrome/Firefox/Brave")
	}
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://x.com/i/api/2/search/adaptive.json?q=%s&count=%d&result_filter=recent",
		url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("Cookie", "auth_token="+e.auth)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("x: %s - %s", resp.Status, string(body))
	}

	// X's adaptive.json returns deeply nested tweets; parse the raw JSON minimally.
	return e.parseAdaptiveResponse(resp.Body)
}

func (e *XTwitterEngine) parseAdaptiveResponse(body io.Reader) ([]Result, error) {
	// Simplified parsing: extract top-level tweet objects.
	var payload struct {
		GlobalObjects struct {
			Tweets map[string]struct {
				FullText      string `json:"full_text"`
				CreatedAt     string `json:"created_at"`
				UserIDStr     string `json:"user_id_str"`
				RetweetCount  int    `json:"retweet_count"`
				FavoriteCount int    `json:"favorite_count"`
			} `json:"tweets"`
			Users map[string]struct {
				ScreenName string `json:"screen_name"`
			} `json:"users"`
		} `json:"globalObjects"`
	}

	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []Result
	for id, tweet := range payload.GlobalObjects.Tweets {
		user := payload.GlobalObjects.Users[tweet.UserIDStr]
		results = append(results, Result{
			Title:      truncate(tweet.FullText, 120),
			URL:        "https://x.com/" + user.ScreenName + "/status/" + id,
			Snippet:    tweet.FullText,
			Source:     "x",
			Engagement: tweet.RetweetCount + tweet.FavoriteCount,
			Author:     user.ScreenName,
		})
	}
	return results, nil
}
