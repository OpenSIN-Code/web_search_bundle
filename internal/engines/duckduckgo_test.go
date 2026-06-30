// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the DuckDuckGo free search engine.
// Docs: internal/engines/duckduckgo.doc.md
package engines

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// sampleDuckDuckGoHTML mirrors the structure DuckDuckGo's html endpoint returns:
// result blocks with result__a (title) and result__snippet anchors whose href
// encodes the real target in the uddg query parameter.
const sampleDuckDuckGoHTML = `<!DOCTYPE html>
<html><body>
<div class="result results_links results_links_deep web-result">
  <div class="links_main links_deep result__body">
    <h2 class="result__title">
      <a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgolang.org%2F&amp;rut=abc">
        The Go Programming Language
      </a>
    </h2>
    <a class="result__snippet" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgolang.org%2F&amp;rut=abc">
      Go is an open source programming language that makes it easy to build
      <b>simple</b>, reliable, and efficient software.
    </a>
  </div>
</div>
<div class="result results_links results_links_deep web-result">
  <div class="links_main links_deep result__body">
    <h2 class="result__title">
      <a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fpkg.go.dev%2F&amp;rut=def">
        Go Documentation - pkg.go.dev
      </a>
    </h2>
    <a class="result__snippet" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fpkg.go.dev%2F&amp;rut=def">
      Discover packages and modules for the Go programming language.
    </a>
  </div>
</div>
<div class="result results_links results_links_deep web-result">
  <div class="links_main links_deep result__body">
    <h2 class="result__title">
      <a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgo.dev%2Fblog%2F&amp;rut=ghi">
        The Go Blog &amp; News
      </a>
    </h2>
    <a class="result__snippet" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgo.dev%2Fblog%2F&amp;rut=ghi">
      Official blog with <b>updates</b>, tutorials, and release notes.
    </a>
  </div>
</div>
</body></html>`

func TestDuckDuckGoName(t *testing.T) {
	e := NewDuckDuckGoEngine()
	if got := e.Name(); got != "duckduckgo" {
		t.Errorf("Name() = %q, want %q", got, "duckduckgo")
	}
}

func TestDuckDuckGoParseHTML(t *testing.T) {
	results := parseDuckDuckGoHTML(sampleDuckDuckGoHTML, 10)
	if len(results) != 3 {
		t.Fatalf("parseDuckDuckGoHTML: got %d results, want 3", len(results))
	}

	want := []struct {
		title, url, snippet string
	}{
		{
			title:   "The Go Programming Language",
			url:     "https://golang.org/",
			snippet: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
		},
		{
			title:   "Go Documentation - pkg.go.dev",
			url:     "https://pkg.go.dev/",
			snippet: "Discover packages and modules for the Go programming language.",
		},
		{
			title:   "The Go Blog & News",
			url:     "https://go.dev/blog/",
			snippet: "Official blog with updates, tutorials, and release notes.",
		},
	}

	for i, w := range want {
		r := results[i]
		if r.Title != w.title {
			t.Errorf("result[%d].Title = %q, want %q", i, r.Title, w.title)
		}
		if r.URL != w.url {
			t.Errorf("result[%d].URL = %q, want %q", i, r.URL, w.url)
		}
		if r.Snippet != w.snippet {
			t.Errorf("result[%d].Snippet = %q, want %q", i, r.Snippet, w.snippet)
		}
		if r.Source != "duckduckgo" {
			t.Errorf("result[%d].Source = %q, want %q", i, r.Source, "duckduckgo")
		}
	}
}

func TestDuckDuckGoParseHTMLLimit(t *testing.T) {
	results := parseDuckDuckGoHTML(sampleDuckDuckGoHTML, 2)
	if len(results) != 2 {
		t.Fatalf("parseDuckDuckGoHTML(limit=2): got %d results, want 2", len(results))
	}
	if results[0].URL != "https://golang.org/" {
		t.Errorf("first result URL = %q, want https://golang.org/", results[0].URL)
	}
}

func TestDuckDuckGoParseHTMLEmpty(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{"empty string", ""},
		{"no result blocks", "<html><body><div>nothing here</div></body></html>"},
		{"malformed block no anchor", `<div class="result"><p>no link</p></div>`},
		{"result block missing href", `<div class="result"><a class="result__a">no href</a></div>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			results := parseDuckDuckGoHTML(c.html, 10)
			if len(results) != 0 {
				t.Errorf("parseDuckDuckGoHTML(%q): got %d results, want 0", c.name, len(results))
			}
		})
	}
}

func TestDecodeDuckDuckGoURL(t *testing.T) {
	cases := []struct {
		name string
		href string
		want string
	}{
		{
			name: "standard redirect",
			href: "//duckduckgo.com/l/?uddg=https%3A%2F%2Fgolang.org%2F&rut=abc",
			want: "https://golang.org/",
		},
		{
			name: "encoded path with query",
			href: "//duckduckgo.com/l/?uddg=https%3A%2F%2Fpkg.go.dev%2Fsearch%3Fq%3Dtest",
			want: "https://pkg.go.dev/search?q=test",
		},
		{
			name: "protocol-relative direct http",
			href: "https://example.com/page",
			want: "https://example.com/page",
		},
		{
			name: "empty href",
			href: "",
			want: "",
		},
		{
			name: "bare fragment",
			href: "#section",
			want: "",
		},
		{
			name: "uddg without scheme",
			href: "//duckduckgo.com/l/?uddg=example.com%2Fpage",
			want: "example.com/page",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := decodeDuckDuckGoURL(c.href)
			if got != c.want {
				t.Errorf("decodeDuckDuckGoURL(%q) = %q, want %q", c.href, got, c.want)
			}
		})
	}
}

func TestDuckDuckGoStripTags(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"plain text", "plain text"},
		{"<b>bold</b> word", "bold word"},
		{"a &amp; b &lt; c &gt; d &quot;e&quot; &#39;f&#39;", `a & b < c > d "e" 'f'`},
		{"  spaced  ", "spaced"},
		{"", ""},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := ddgStripTags(c.in)
			if got != c.want {
				t.Errorf("ddgStripTags(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestDuckDuckGoSearchHTMLEmptyQuery(t *testing.T) {
	e := NewDuckDuckGoEngine()
	if _, err := e.Search(context.Background(), "", 10); err == nil {
		t.Fatal("Search(empty query) expected error, got nil")
	}
}

func TestDuckDuckGoSearchViaTestServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("request method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "Mozilla/5.0") {
			t.Errorf("User-Agent = %q, want Mozilla prefix", got)
		}
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
			t.Errorf("Content-Type = %q, want form-urlencoded", ct)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "q=golang+testing") {
			t.Errorf("request body = %q, want q=golang+testing", string(body))
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleDuckDuckGoHTML))
	}))
	defer srv.Close()

	e := NewDuckDuckGoEngine()
	e.endpoint = srv.URL

	results, err := e.Search(context.Background(), "golang testing", 2)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Title != "The Go Programming Language" {
		t.Errorf("results[0].Title = %q, want 'The Go Programming Language'", results[0].Title)
	}
	if results[0].URL != "https://golang.org/" {
		t.Errorf("results[0].URL = %q, want https://golang.org/", results[0].URL)
	}
	for _, r := range results {
		if r.Source != "duckduckgo" {
			t.Errorf("result Source = %q, want duckduckgo", r.Source)
		}
	}
}

func TestDuckDuckGoSearchRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	e := NewDuckDuckGoEngine()
	e.endpoint = srv.URL

	_, err := e.Search(context.Background(), "test", 10)
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("expected rate-limited error, got: %v", err)
	}
}

func TestDuckDuckGoSearchServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	e := NewDuckDuckGoEngine()
	e.endpoint = srv.URL

	_, err := e.Search(context.Background(), "test", 10)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected 500 error, got: %v", err)
	}
}

func TestDuckDuckGoSearchEmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>no results</body></html>"))
	}))
	defer srv.Close()

	e := NewDuckDuckGoEngine()
	e.endpoint = srv.URL

	results, err := e.Search(context.Background(), "test", 10)
	if err != nil {
		t.Fatalf("Search error on empty body: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}
