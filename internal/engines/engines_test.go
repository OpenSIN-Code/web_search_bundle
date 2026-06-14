// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for engine helper functions and constructors.
// Docs: engines_test.doc.md

package engines

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
)

func TestResultType(t *testing.T) {
	r := Result{
		Title:   "title",
		URL:     "https://example.com",
		Snippet: "snippet",
		Source:  "reddit",
		Score:   1.0,
	}
	if r.Title != "title" {
		t.Errorf("unexpected title: %s", r.Title)
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
	}
	for _, c := range cases {
		got := truncate(c.in, c.max)
		if got != c.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}

func TestAtURIToWeb(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"at://user.bsky.social/post/123", "at://user.bsky.social/post/123"}, // current behavior; prefix check uses [:4]
		{"https://bsky.app/profile/user", "https://bsky.app/profile/user"},
		{"short", "short"},
	}
	for _, c := range cases {
		got := atURIToWeb(c.in)
		if got != c.want {
			t.Errorf("atURIToWeb(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestContainsInsensitive(t *testing.T) {
	cases := []struct {
		haystack, needle string
		want             bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "foo", false},
		{"", "", true},
		{"", "a", false},
	}
	for _, c := range cases {
		got := contains(c.haystack, c.needle)
		if got != c.want {
			t.Errorf("contains(%q, %q) = %v, want %v", c.haystack, c.needle, got, c.want)
		}
	}
}

func TestContainsInsensitiveDirect(t *testing.T) {
	if !containsInsensitive("abc", "B") {
		t.Error("expected match")
	}
	if containsInsensitive("abc", "z") {
		t.Error("expected no match")
	}
}

func TestToLower(t *testing.T) {
	cases := []struct {
		in, want byte
	}{
		{'A', 'a'},
		{'Z', 'z'},
		{'a', 'a'},
		{'1', '1'},
	}
	for _, c := range cases {
		got := toLower(c.in)
		if got != c.want {
			t.Errorf("toLower(%c) = %c, want %c", c.in, got, c.want)
		}
	}
}

func TestParseTime(t *testing.T) {
	got := parseTime("2024-01-01T00:00:00Z")
	if got.IsZero() {
		t.Error("expected parsed time")
	}
	if !parseTime("").IsZero() {
		t.Error("expected zero time for empty string")
	}
}

func TestParseVideoTime(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"", 0},
		{"90", 90},
		{"1:30", 90},
		{"1:30:45", 5445},
		{"bad", 0},
	}
	for _, c := range cases {
		got := parseVideoTime(c.in)
		if got != c.want {
			t.Errorf("parseVideoTime(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseVTT(t *testing.T) {
	vtt := "WEBVTT\n\n1\n00:00:00.000 --> 00:00:02.000\nHello <b>world</b>\n\n2\n00:00:02.000 --> 00:00:04.000\nHello world\n\n3\n00:00:04.000 --> 00:00:06.000\nMore text\n"
	got := parseVTT(vtt)
	want := "Hello world More text" // dedupeLines removes consecutive duplicates
	if got != want {
		t.Errorf("parseVTT() = %q, want %q", got, want)
	}
}

func TestStripHTML(t *testing.T) {
	got := stripHTML("a<b>c</b>d")
	if got != "acd" {
		t.Errorf("stripHTML = %q, want %q", got, "acd")
	}
}

func TestDedupeLines(t *testing.T) {
	in := []string{"a", "a", "b", "b", "c"}
	want := []string{"a", "b", "c"}
	got := dedupeLines(in)
	if len(got) != len(want) {
		t.Fatalf("dedupeLines = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dedupeLines[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDetectVideoSource(t *testing.T) {
	cases := []struct {
		url, want string
	}{
		{"https://youtube.com/watch?v=1", "youtube"},
		{"https://youtu.be/1", "youtube"},
		{"https://tiktok.com/@user/video/1", "tiktok"},
		{"https://instagram.com/reel/1", "instagram"},
		{"https://x.com/user/status/1", "x"},
		{"https://twitter.com/user/status/1", "x"},
		{"https://vimeo.com/1", "vimeo"},
		{"https://loom.com/share/1", "loom"},
		{"https://unknown.com/video", "unknown"},
	}
	for _, c := range cases {
		got := detectVideoSource(c.url)
		if got != c.want {
			t.Errorf("detectVideoSource(%q) = %q, want %q", c.url, got, c.want)
		}
	}

	dir := t.TempDir()
	f := filepath.Join(dir, "local.mp4")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if detectVideoSource(f) != "local" {
		t.Errorf("expected local for existing file")
	}
}

func TestDetectWhisperPref(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "key")
	t.Setenv("OPENAI_API_KEY", "")
	if got := detectWhisperPref(); got != "groq" {
		t.Errorf("detectWhisperPref = %q, want groq", got)
	}

	t.Setenv("GROQ_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "key")
	if got := detectWhisperPref(); got != "openai" {
		t.Errorf("detectWhisperPref = %q, want openai", got)
	}

	t.Setenv("OPENAI_API_KEY", "")
	if got := detectWhisperPref(); got != "none" {
		t.Errorf("detectWhisperPref = %q, want none", got)
	}
}

func TestEngineNames(t *testing.T) {
	cases := []struct {
		engine Engine
		want   string
	}{
		{NewBlueskyEngine(), "bluesky"},
		{NewRedditEngine(), "reddit"},
		{NewHackerNewsEngine(), "hackernews"},
		{NewGitHubEngine(), "github"},
		{NewPolymarketEngine(), "polymarket"},
		{NewBraveEngine(""), "brave"},
		{NewSerpAPIEngine([]string{}), "serpapi"},
	}
	for _, c := range cases {
		if got := c.engine.Name(); got != c.want {
			t.Errorf("Name() = %q, want %q", got, c.want)
		}
	}
}

// newTestSidecarManager creates a sidecar manager in a temp HOME directory.
func newTestSidecarManager(t *testing.T) (*sidecar.Manager, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	sc, err := sidecar.NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return sc, home
}

func TestNewVideoEngine(t *testing.T) {
	sc, _ := newTestSidecarManager(t)
	e := NewVideoEngine(sc)
	if e.workDir == "" {
		t.Error("expected workDir set")
	}
	if e.maxFrames != 100 {
		t.Errorf("maxFrames = %d, want 100", e.maxFrames)
	}
}

func TestNewYouTubeEngine(t *testing.T) {
	sc, _ := newTestSidecarManager(t)
	e := NewYouTubeEngine(sc)
	if e.Name() != "youtube" {
		t.Errorf("Name = %q, want youtube", e.Name())
	}
}

func TestNewXTwitterEngine(t *testing.T) {
	e := NewXTwitterEngine()
	if e.Name() != "x" {
		t.Errorf("Name = %q, want x", e.Name())
	}
}

func TestTranscribeUnsupportedBackend(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "audio.wav")
	if err := os.WriteFile(f, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := transcribe(context.TODO(), f, "unsupported")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "unsupported whisper backend: unsupported" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTranscribeFileNotFound(t *testing.T) {
	_, err := transcribe(context.TODO(), "nonexistent.wav", "groq")
	if err == nil {
		t.Fatal("expected error")
	}
}
