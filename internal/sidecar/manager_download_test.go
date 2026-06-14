// SPDX-License-Identifier: MIT
// Purpose: Unit tests for sidecar binary download and URL selection paths.
// Docs: manager_download_test.doc.md

package sidecar

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureBinary_DownloadSuccess(t *testing.T) {
	m := newTestManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("#!/bin/sh\necho downloaded\n"))
	}))
	defer server.Close()

	m.binaries["testbin"] = &Binary{
		Name: "testbin",
		DownloadURL: func(os, arch string) string {
			return server.URL
		},
	}

	path, err := m.EnsureBinary("testbin")
	if err != nil {
		t.Fatalf("EnsureBinary error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected downloaded file: %v", err)
	}
}

func TestExecute_DownloadedBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping executable download test on windows")
	}
	m := newTestManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("#!/bin/sh\necho ok\n"))
	}))
	defer server.Close()

	m.binaries["testbin"] = &Binary{
		Name: "testbin",
		DownloadURL: func(os, arch string) string {
			return server.URL
		},
	}

	out, err := m.Execute("testbin")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !strings.Contains(string(out), "ok") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestEnsureBinary_DownloadNotFound(t *testing.T) {
	m := newTestManager(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	m.binaries["testbin"] = &Binary{
		Name: "testbin",
		DownloadURL: func(os, arch string) string {
			return server.URL
		},
	}

	_, err := m.EnsureBinary("testbin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

func TestEnsureBinary_NoDownloadURL(t *testing.T) {
	m := newTestManager(t)
	m.binaries["testbin"] = &Binary{
		Name: "testbin",
		DownloadURL: func(os, arch string) string {
			return ""
		},
	}
	_, err := m.EnsureBinary("testbin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no download URL") {
		t.Errorf("expected no download URL error, got: %v", err)
	}
}

func TestEnsureBinary_CachedPath(t *testing.T) {
	m := newTestManager(t)
	binPath := filepath.Join(m.binDir, "cachedbin")
	if err := os.WriteFile(binPath, []byte("fake"), 0755); err != nil {
		t.Fatal(err)
	}
	m.binaries["cachedbin"] = &Binary{Name: "cachedbin"}
	got, err := m.EnsureBinary("cachedbin")
	if err != nil {
		t.Fatalf("EnsureBinary error: %v", err)
	}
	if got != binPath {
		t.Errorf("expected %s, got %s", binPath, got)
	}
}

func TestDownloadURLForYTDLPLinux(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["yt-dlp"]
	url := bin.DownloadURL("linux", "amd64")
	if !strings.Contains(url, "yt-dlp_linux") {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestDownloadURLForYTDLPWindows(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["yt-dlp"]
	url := bin.DownloadURL("windows", "amd64")
	if !strings.Contains(url, "yt-dlp.exe") {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestDownloadURLForYTDLPDarwinLegacy(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["yt-dlp"]
	url := bin.DownloadURL("darwin", "amd64")
	if !strings.Contains(url, "yt-dlp_macos_legacy") {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestDownloadURLForYTDLPDarwinARM64(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["yt-dlp"]
	url := bin.DownloadURL("darwin", "arm64")
	if !strings.Contains(url, "yt-dlp_macos") || strings.Contains(url, "legacy") {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestDownloadURLForYTDLPDarwinUnknown(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["yt-dlp"]
	url := bin.DownloadURL("darwin", "unknown")
	if !strings.Contains(url, "yt-dlp_macos_legacy") {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestDownloadURLForYTDLPUnknownOS(t *testing.T) {
	m := newTestManager(t)
	bin := m.binaries["yt-dlp"]
	url := bin.DownloadURL("freebsd", "amd64")
	if url != "" {
		t.Errorf("expected empty URL, got %s", url)
	}
}

func TestFindSystemFFmpeg_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping windows-specific path test")
	}
	// On Windows this test would exercise the windows path; we just ensure no panic.
	_, err := findSystemFFmpeg()
	t.Logf("findSystemFFmpeg on windows: %v", err)
}

func TestNewManager_MkdirAllError(t *testing.T) {
	// Create a file named "home" so MkdirAll cannot create a directory there.
	homeFile := filepath.Join(t.TempDir(), "home")
	if err := os.WriteFile(homeFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", homeFile)
	_, err := NewManager()
	if err == nil {
		t.Fatal("expected error when MkdirAll fails")
	}
}
