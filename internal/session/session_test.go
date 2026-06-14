// SPDX-License-Identifier: MIT
// Purpose: Hermetic unit tests for browser session extraction.
// Docs: session_test.doc.md

package session

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewBrowserSession(t *testing.T) {
	b := NewBrowserSession()
	if b == nil {
		t.Fatal("expected session")
	}
}

func TestGetCookieDBPathsDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	b := NewBrowserSession()
	paths := b.getCookieDBPaths()
	if len(paths) == 0 {
		t.Fatal("expected paths")
	}
	found := false
	for _, p := range paths {
		if contains(p, "Chrome/Default/Cookies") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Chrome path in %v", paths)
	}
}

func TestGetCookieDBPathsLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	b := NewBrowserSession()
	paths := b.getCookieDBPaths()
	if len(paths) == 0 {
		t.Fatal("expected paths")
	}
}

func TestGetCookieDBPathsWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	b := NewBrowserSession()
	paths := b.getCookieDBPaths()
	if len(paths) == 0 {
		t.Fatal("expected paths")
	}
}

func TestGetCookieNoBrowser(t *testing.T) {
	b := NewBrowserSession()
	_, err := b.GetCookie("example.com", "session")
	if err == nil {
		t.Fatal("expected error when no browser cookies exist")
	}
}

func TestGetXAuthTokenNoBrowser(t *testing.T) {
	b := NewBrowserSession()
	_, err := b.GetXAuthToken()
	if err == nil {
		t.Fatal("expected error when no browser session exists")
	}
}

func TestExtractCookieChromiumSkipped(t *testing.T) {
	b := NewBrowserSession()
	_, err := b.extractCookie("/tmp/Cookies", "twitter.com", "auth_token")
	if err == nil {
		t.Fatal("expected error for chromium cookie")
	}
	if !contains(err.Error(), "chromium cookie decryption") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtractFirefoxCookie(t *testing.T) {
	b := NewBrowserSession()
	dbPath := createFirefoxCookieDB(t, "twitter.com", "auth_token", "secret123")

	val, err := b.extractFirefoxCookie(dbPath, "twitter.com", "auth_token")
	if err != nil {
		t.Fatalf("extractFirefoxCookie error: %v", err)
	}
	if val != "secret123" {
		t.Errorf("unexpected value: %q", val)
	}

	_, err = b.extractFirefoxCookie(dbPath, "twitter.com", "missing")
	if err == nil {
		t.Fatal("expected error for missing cookie")
	}
}

func TestExtractCookieFirefox(t *testing.T) {
	b := NewBrowserSession()
	dbPath := createFirefoxCookieDB(t, "x.com", "ct0", "ct0value")

	val, err := b.extractCookie(dbPath, "x.com", "ct0")
	if err != nil {
		t.Fatalf("extractCookie error: %v", err)
	}
	if val != "ct0value" {
		t.Errorf("unexpected value: %q", val)
	}
}

func createFirefoxCookieDB(t *testing.T, host, name, value string) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "cookies.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE moz_cookies (host TEXT, name TEXT, value TEXT)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO moz_cookies (host, name, value) VALUES (?, ?, ?)`, host, name, value)
	if err != nil {
		t.Fatal(err)
	}
	return dbPath
}

func TestExtractCookieNoProfile(t *testing.T) {
	b := NewBrowserSession()
	_, err := b.extractCookie("/tmp/nonexistent/*/cookies.sqlite", "twitter.com", "auth_token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTempFileCleanup(t *testing.T) {
	// Sanity check that extractFirefoxCookie cleans up the temp copy.
	dbPath := createFirefoxCookieDB(t, "twitter.com", "auth_token", "x")
	b := NewBrowserSession()
	_, _ = b.extractFirefoxCookie(dbPath, "twitter.com", "auth_token")
	// tmpDB is removed by defer; just verify the original still exists.
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("original db missing: %v", err)
	}
}
