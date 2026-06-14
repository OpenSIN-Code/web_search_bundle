// SPDX-License-Identifier: MIT
// Purpose: Unit tests for video analysis, briefing, and prompt handlers.
// Docs: video_test.doc.md

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installFakeSidecar pre-populates a fake yt-dlp and ffmpeg so the video
// engine fails during analysis instead of downloading real binaries.
func installFakeSidecar(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	binDir := filepath.Join(home, ".sin-websearch", "bin")
	if err := os.MkdirAll(binDir, 0750); err != nil {
		t.Fatalf("mkdir sidecar bin: %v", err)
	}

	fakeBin := func(name string) {
		path := filepath.Join(binDir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
			t.Fatalf("write fake %s: %v", name, err)
		}
	}
	fakeBin("yt-dlp")
	fakeBin("ffmpeg")
}

func TestHandleWatchInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/watch", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleWatch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "invalid request") {
		t.Errorf("expected invalid request error, got %s", rr.Body.String())
	}
}

func TestHandleWatchSidecarError(t *testing.T) {
	// Make the sidecar manager fail to create its bin directory.
	t.Setenv("HOME", "/dev/null")

	s := NewHTTPServer(nil, nil)
	body, _ := json.Marshal(map[string]string{"url": "https://example.com/video.mp4"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/watch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleWatch(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "sidecar") {
		t.Errorf("expected sidecar error, got %s", rr.Body.String())
	}
}

func TestHandleWatchEngineError(t *testing.T) {
	installFakeSidecar(t)

	s := NewHTTPServer(nil, nil)
	body, _ := json.Marshal(map[string]string{"url": "https://example.com/video.mp4"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/watch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleWatch(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "watch:") {
		t.Errorf("expected watch error, got %s", rr.Body.String())
	}
}

func TestHandleVideoBriefInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vbrief", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleVideoBrief(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleVideoBriefMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vbrief", nil)
	rr := httptest.NewRecorder()

	s.handleVideoBrief(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleVideoBriefSidecarError(t *testing.T) {
	t.Setenv("HOME", "/dev/null")

	s := NewHTTPServer(nil, nil)
	body, _ := json.Marshal(map[string]string{"url": "https://example.com/video.mp4"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vbrief", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleVideoBrief(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "sidecar") {
		t.Errorf("expected sidecar error, got %s", rr.Body.String())
	}
}

func TestHandleVideoBriefEngineError(t *testing.T) {
	installFakeSidecar(t)

	s := NewHTTPServer(nil, nil)
	body, _ := json.Marshal(map[string]string{"url": "https://example.com/video.mp4"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vbrief", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleVideoBrief(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "watch:") {
		t.Errorf("expected watch error, got %s", rr.Body.String())
	}
}

func TestHandleVideoPromptInvalidJSON(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vprompt", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleVideoPrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleVideoPromptMethodNotAllowed(t *testing.T) {
	s := NewHTTPServer(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vprompt", nil)
	rr := httptest.NewRecorder()

	s.handleVideoPrompt(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleVideoPromptSidecarError(t *testing.T) {
	t.Setenv("HOME", "/dev/null")

	s := NewHTTPServer(nil, nil)
	body, _ := json.Marshal(map[string]string{"url": "https://example.com/video.mp4"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vprompt", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleVideoPrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "sidecar") {
		t.Errorf("expected sidecar error, got %s", rr.Body.String())
	}
}

func TestHandleVideoPromptEngineError(t *testing.T) {
	installFakeSidecar(t)

	s := NewHTTPServer(nil, nil)
	body, _ := json.Marshal(map[string]string{"url": "https://example.com/video.mp4"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vprompt", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handleVideoPrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "watch:") {
		t.Errorf("expected watch error, got %s", rr.Body.String())
	}
}
