// SPDX-License-Identifier: MIT
// Purpose: DuckDuckGo free HTML search engine — no API key required.
// Docs: internal/engines/duckduckgo.doc.md
package engines

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ddgEndpoint is the free, keyless HTML search endpoint.
const ddgEndpoint = "https://html.duckduckgo.com/html/"

// ddgUserAgent is a realistic desktop browser string to avoid bot blocks.
const ddgUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// DuckDuckGoEngine queries the free DuckDuckGo HTML endpoint with no API key.
type DuckDuckGoEngine struct {
	client   *http.Client
	endpoint string // overridable for tests; defaults to ddgEndpoint
}

// NewDuckDuckGoEngine creates a DuckDuckGo engine with a 10s timeout.
func NewDuckDuckGoEngine() *DuckDuckGoEngine {
	return &DuckDuckGoEngine{
		client:   &http.Client{Timeout: 10 * time.Second},
		endpoint: ddgEndpoint,
	}
}

// Name returns the engine identifier.
func (e *DuckDuckGoEngine) Name() string { return "duckduckgo" }

// Search queries DuckDuckGo's free HTML endpoint and parses results with regex.
func (e *DuckDuckGoEngine) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("duckduckgo: empty query")
	}
	if limit <= 0 {
		limit = 10
	}

	form := url.Values{}
	form.Set("q", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", ddgUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("duckduckgo: rate limited")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("duckduckgo: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseDuckDuckGoHTML(string(body), limit), nil
}

// parseDuckDuckGoHTML extracts search results from DuckDuckGo's HTML response.
// It finds all result__a (title) and result__snippet anchors and zips them by
// document order, capped at limit. Malformed or missing anchors are skipped so
// a partial format change never fails the whole engine.
func parseDuckDuckGoHTML(html string, limit int) []Result {
	titles := ddgTitleRE.FindAllStringSubmatch(html, -1)
	if len(titles) == 0 {
		return nil
	}
	snippets := ddgSnippetRE.FindAllStringSubmatch(html, -1)

	results := make([]Result, 0, len(titles))
	for i, m := range titles {
		if len(results) >= limit {
			break
		}
		if len(m) < 3 || m[2] == "" {
			continue
		}
		realURL := decodeDuckDuckGoURL(m[1])
		if realURL == "" {
			continue
		}
		snippet := ""
		if i < len(snippets) && len(snippets[i]) >= 3 {
			snippet = ddgStripTags(snippets[i][2])
		}
		results = append(results, Result{
			Title:   ddgStripTags(m[2]),
			URL:     realURL,
			Snippet: snippet,
			Source:  "duckduckgo",
		})
	}

	return results
}

// decodeDuckDuckGoURL extracts the real target URL from a DuckDuckGo redirect
// link of the form "//duckduckgo.com/l/?uddg=<encoded>&rut=...". If the href is
// already a direct URL (no uddg parameter), it is returned cleaned up.
func decodeDuckDuckGoURL(href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	// DuckDuckGo HTML uses protocol-relative URLs (//duckduckgo.com/...).
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}
	// Parse and look for the uddg query parameter holding the real target.
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	if target := u.Query().Get("uddg"); target != "" {
		if decoded, err := url.QueryUnescape(target); err == nil {
			return decoded
		}
		return target
	}
	// Some results link directly; only accept http(s) to avoid junk anchors.
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	return ""
}

// ddgStripTags removes any residual HTML tags, decodes common entities, and
// collapses internal whitespace so titles/snippets are single-line plain text.
func ddgStripTags(s string) string {
	s = ddgTagRE.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = ddgWhitespaceRE.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// Pre-compiled regexes. DuckDuckGo's HTML wraps each result in anchors with
// distinctive class tokens; we match those directly rather than trying to
// capture enclosing div blocks (RE2 has no lookahead support).
var (
	// Title anchor: <a class="result__a" href="HREF">TITLE</a>
	ddgTitleRE = regexp.MustCompile(`(?s)<a[^>]*class="[^"]*\bresult__a\b[^"]*"[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)

	// Snippet anchor: <a class="result__snippet" href="HREF">SNIPPET</a>
	ddgSnippetRE = regexp.MustCompile(`(?s)<a[^>]*class="[^"]*\bresult__snippet\b[^"]*"[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)

	// Any remaining HTML tag, for stripping residual markup in text.
	ddgTagRE = regexp.MustCompile(`<[^>]*>`)

	// Runs of whitespace (spaces, tabs, newlines) to collapse into one space.
	ddgWhitespaceRE = regexp.MustCompile(`\s+`)
)
