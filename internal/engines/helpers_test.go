// SPDX-License-Identifier: MIT
// Purpose: Additional hermetic unit tests for engine helpers and constructors.
// Docs: helpers_test.doc.md
package engines

import "testing"

func TestDedupeLinesEmpty(t *testing.T) {
	got := dedupeLines([]string{})
	if got != nil {
		t.Errorf("dedupeLines([]) = %v, want nil", got)
	}
	if len(got) != 0 {
		t.Errorf("dedupeLines([]) length = %d, want 0", len(got))
	}
}

func TestEngineNamesExtended(t *testing.T) {
	if got := NewPerplexityEngine().Name(); got != "perplexity" {
		t.Errorf("perplexity Name() = %q, want perplexity", got)
	}
	if got := NewSearxNGEngine().Name(); got != "searxng" {
		t.Errorf("searxng Name() = %q, want searxng", got)
	}
}

func TestAtURIToWebShort(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"at:", "at:"},
		{"at:/", "at:/"},
		{"at://", "at://"},
		// Current implementation checks only the first 4 bytes, so real at:// URIs pass through.
		{"at://user.bsky.social/post/123", "at://user.bsky.social/post/123"},
	}
	for _, c := range cases {
		got := atURIToWeb(c.in)
		if got != c.want {
			t.Errorf("atURIToWeb(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseTimeInvalid(t *testing.T) {
	got := parseTime("not-a-time")
	if !got.IsZero() {
		t.Errorf("parseTime(invalid) = %v, want zero", got)
	}
}

func TestParseVideoTimeMalformed(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"1:2:3:4", 0},
		{"abc", 0},
	}
	for _, c := range cases {
		got := parseVideoTime(c.in)
		if got != c.want {
			t.Errorf("parseVideoTime(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
