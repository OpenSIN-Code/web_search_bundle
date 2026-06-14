// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for the sidecar manager.
// Docs: sidecar_test.doc.md

package sidecar

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m := &Manager{
		binDir:   t.TempDir(),
		binaries: make(map[string]*Binary),
	}
	m.registerYTDLP()
	m.registerScrapeCreators()
	m.registerFFmpeg()
	return m
}

func TestNewManager(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}
	if m.binDir == "" {
		t.Error("expected binDir set")
	}
	if len(m.binaries) == 0 {
		t.Error("expected binaries registered")
	}
	expectedDir := filepath.Join(home, ".sin-websearch", "bin")
	if m.binDir != expectedDir {
		t.Errorf("binDir = %q, want %q", m.binDir, expectedDir)
	}
}

func TestEnsureBinaryUnknown(t *testing.T) {
	m := newTestManager(t)
	_, err := m.EnsureBinary("unknown")
	if err == nil {
		t.Fatal("expected error for unknown binary")
	}
	if !strings.Contains(err.Error(), "unknown binary") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEnsureBinaryYTDLPCached(t *testing.T) {
	m := newTestManager(t)
	binPath := filepath.Join(m.binDir, "yt-dlp")
	script := "#!/bin/sh\necho fake yt-dlp\n"
	if err := os.WriteFile(binPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := m.EnsureBinary("yt-dlp")
	if err != nil {
		t.Fatalf("EnsureBinary error: %v", err)
	}
	if got != binPath {
		t.Errorf("expected %s, got %s", binPath, got)
	}
}

func TestExecuteYTDLPCached(t *testing.T) {
	m := newTestManager(t)
	binPath := filepath.Join(m.binDir, "yt-dlp")
	script := "#!/bin/sh\necho 'ok'\n"
	if runtime.GOOS == "windows" {
		binPath += ".exe"
		script = "@echo off\necho ok\n"
	}
	if err := os.WriteFile(binPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	out, err := m.Execute("yt-dlp", "--version")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !strings.Contains(string(out), "ok") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestExecuteUnknownBinary(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Execute("unknown")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDownloadURLForYTDLPMacOSARM64(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["yt-dlp"]
	url := bin.DownloadURL("darwin", "arm64")
	if !strings.Contains(url, "yt-dlp_macos") {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestDownloadURLForScrapeCreators(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["scrapecreators"]
	url := bin.DownloadURL("linux", "amd64")
	if !strings.Contains(url, "scrapecreators") {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestDownloadURLForFFmpeg(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["ffmpeg"]
	url := bin.DownloadURL("darwin", "arm64")
	if url != "" {
		t.Errorf("expected empty ffmpeg URL, got %s", url)
	}
}

func TestFindSystemFFmpeg(t *testing.T) {
	// On CI/standard Macs this may not exist; just ensure no panic.
	_, err := findSystemFFmpeg()
	if err != nil {
		t.Logf("ffmpeg not found: %v", err)
	}
}

func TestEnsureBinaryFFmpegSystem(t *testing.T) {
	m := newTestManager(t)
	// EnsureBinary for ffmpeg requires a system ffmpeg or a download URL.
	// On macOS without ffmpeg it will fail; accept either outcome.
	_, err := m.EnsureBinary("ffmpeg")
	if err != nil {
		if !strings.Contains(err.Error(), "install ffmpeg") && !strings.Contains(err.Error(), "download failed") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

