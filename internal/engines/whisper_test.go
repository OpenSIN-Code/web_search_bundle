// SPDX-License-Identifier: MIT
// Purpose: Hermetic tests for Whisper transcription helpers.
// Docs: whisper_test.doc.md
package engines

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestAudio(t *testing.T, dir string) string {
	t.Helper()
	f := filepath.Join(dir, "audio.wav")
	if err := os.WriteFile(f, []byte("RIFFfakeaudio"), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestTranscribeOversizedFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "audio.wav")
	// Create a 26 MB file to exceed the 25 MB limit.
	size := 26 * 1024 * 1024
	data := make([]byte, size)
	if err := os.WriteFile(f, data, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := transcribe(context.Background(), f, "groq")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostWhisperSuccess(t *testing.T) {
	dir := t.TempDir()
	f := writeTestAudio(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"text":"hello world"}`)
	}))
	defer srv.Close()

	text, err := postWhisper(context.Background(), f, "key", srv.URL, "whisper-large-v3")
	if err != nil {
		t.Fatalf("postWhisper error: %v", err)
	}
	if text != "hello world" {
		t.Errorf("text = %q, want hello world", text)
	}
}

func TestPostWhisperNon200(t *testing.T) {
	dir := t.TempDir()
	f := writeTestAudio(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	_, err := postWhisper(context.Background(), f, "key", srv.URL, "whisper-large-v3")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostWhisperInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	f := writeTestAudio(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "not json")
	}))
	defer srv.Close()

	_, err := postWhisper(context.Background(), f, "key", srv.URL, "whisper-large-v3")
	if err == nil {
		t.Fatal("expected error")
	}
}

// hostMappedTransport routes requests to httptest servers based on the original host.
type hostMappedTransport struct {
	base   http.RoundTripper
	groq   *url.URL
	openai *url.URL
}

func (h *hostMappedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Hostname() {
	case "api.groq.com":
		if h.groq == nil {
			return h.base.RoundTrip(req)
		}
		req.URL.Scheme = h.groq.Scheme
		req.URL.Host = h.groq.Host
		return h.base.RoundTrip(req)
	case "api.openai.com":
		if h.openai == nil {
			return h.base.RoundTrip(req)
		}
		req.URL.Scheme = h.openai.Scheme
		req.URL.Host = h.openai.Host
		return h.base.RoundTrip(req)
	}
	return h.base.RoundTrip(req)
}

// withMockTransport temporarily replaces http.DefaultTransport with a host mapper.
func withMockTransport(t *testing.T, groq, openai *httptest.Server, fn func()) {
	t.Helper()
	orig := http.DefaultTransport
	mapper := &hostMappedTransport{base: http.DefaultTransport}
	if groq != nil {
		u, err := url.Parse(groq.URL)
		if err != nil {
			t.Fatalf("parse groq url: %v", err)
		}
		mapper.groq = u
	}
	if openai != nil {
		u, err := url.Parse(openai.URL)
		if err != nil {
			t.Fatalf("parse openai url: %v", err)
		}
		mapper.openai = u
	}
	http.DefaultTransport = mapper
	defer func() { http.DefaultTransport = orig }()
	fn()
}

func TestTranscribeGroqWithKey(t *testing.T) {
	dir := t.TempDir()
	f := writeTestAudio(t, dir)

	groq := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"text":"groq result"}`)
	}))
	defer groq.Close()

	t.Setenv("GROQ_API_KEY", "key")
	withMockTransport(t, groq, nil, func() {
		text, err := transcribeGroq(context.Background(), f)
		if err != nil {
			t.Fatalf("transcribeGroq error: %v", err)
		}
		if text != "groq result" {
			t.Errorf("text = %q, want groq result", text)
		}
	})
}

func TestTranscribeGroqMissingKey(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "")
	_, err := transcribeGroq(context.Background(), "nonexistent.wav")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "GROQ_API_KEY not set") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTranscribeOpenAIWithKey(t *testing.T) {
	dir := t.TempDir()
	f := writeTestAudio(t, dir)

	openai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"text":"openai result"}`)
	}))
	defer openai.Close()

	t.Setenv("OPENAI_API_KEY", "key")
	withMockTransport(t, nil, openai, func() {
		text, err := transcribeOpenAI(context.Background(), f)
		if err != nil {
			t.Fatalf("transcribeOpenAI error: %v", err)
		}
		if text != "openai result" {
			t.Errorf("text = %q, want openai result", text)
		}
	})
}

func TestTranscribeOpenAIMissingKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, err := transcribeOpenAI(context.Background(), "nonexistent.wav")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY not set") {
		t.Errorf("unexpected error: %v", err)
	}
}

