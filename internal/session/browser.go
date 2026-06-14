// SPDX-License-Identifier: MIT
// Purpose: Extract browser cookies for authenticated sessions (X/Twitter, etc.).
// Docs: internal/session/browser.doc.md
package session

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "modernc.org/sqlite"
)

// BrowserSession extracts cookies from local browsers.
type BrowserSession struct{}

// NewBrowserSession creates a new browser session extractor.
func NewBrowserSession() *BrowserSession {
	return &BrowserSession{}
}

// GetCookie tries to find a cookie for a domain in supported browsers.
func (b *BrowserSession) GetCookie(domain, name string) (string, error) {
	paths := b.getCookieDBPaths()
	for _, dbPath := range paths {
		if _, err := os.Stat(dbPath); err != nil {
			continue
		}
		val, err := b.extractCookie(dbPath, domain, name)
		if err == nil && val != "" {
			return val, nil
		}
	}
	return "", fmt.Errorf("no cookie %s found for %s", name, domain)
}

// GetXAuthToken returns the X/Twitter auth_token from a local browser.
func (b *BrowserSession) GetXAuthToken() (string, error) {
	for _, domain := range []string{"twitter.com", "x.com"} {
		for _, name := range []string{"auth_token", "ct0"} {
			val, err := b.GetCookie(domain, name)
			if err == nil {
				return val, nil
			}
		}
	}
	return "", fmt.Errorf("no X/Twitter session found in any browser")
}

func (b *BrowserSession) getCookieDBPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library/Application Support/Google/Chrome/Default/Cookies"),
			filepath.Join(home, "Library/Application Support/BraveSoftware/Brave-Browser/Default/Cookies"),
			filepath.Join(home, "Library/Application Support/Microsoft Edge/Default/Cookies"),
			filepath.Join(home, "Library/Application Support/Firefox/Profiles/*/cookies.sqlite"),
		}
	case "linux":
		return []string{
			filepath.Join(home, ".config/google-chrome/Default/Cookies"),
			filepath.Join(home, ".config/BraveSoftware/Brave-Browser/Default/Cookies"),
			filepath.Join(home, ".mozilla/firefox/*/cookies.sqlite"),
		}
	case "windows":
		appdata := os.Getenv("LOCALAPPDATA")
		return []string{
			filepath.Join(appdata, "Google/Chrome/User Data/Default/Cookies"),
			filepath.Join(appdata, "BraveSoftware/Brave-Browser/User Data/Default/Cookies"),
			filepath.Join(appdata, "Microsoft/Edge/User Data/Default/Cookies"),
		}
	}
	return nil
}

func (b *BrowserSession) extractCookie(dbPath, domain, name string) (string, error) {
	// Firefox uses a sqlite file; Chromium uses an encrypted SQLite file.
	// This is a best-effort implementation for SQLite-based cookies.
	if !strings.HasSuffix(dbPath, ".sqlite") {
		// Chromium cookies are encrypted with OS-specific key; skip here.
		return "", fmt.Errorf("chromium cookie decryption not implemented")
	}

	entries, err := filepath.Glob(dbPath)
	if err != nil || len(entries) == 0 {
		return "", fmt.Errorf("no firefox profile found")
	}

	for _, path := range entries {
		val, err := b.extractFirefoxCookie(path, domain, name)
		if err == nil && val != "" {
			return val, nil
		}
	}
	return "", fmt.Errorf("cookie not found in firefox")
}

func (b *BrowserSession) extractFirefoxCookie(dbPath, domain, name string) (string, error) {
	tmpDB := filepath.Join(os.TempDir(), "sin_cookies_copy.sqlite")
	defer os.Remove(tmpDB)

	data, err := os.ReadFile(dbPath) // #nosec G304 dbPath is an OS/browser cookie path from getCookieDBPaths
	if err != nil {
		return "", err
	}
	// tmpDB is a hardcoded filename inside os.TempDir(); not user-controlled.
	if err := os.WriteFile(tmpDB, data, 0600); err != nil { // #nosec G703 G306
		return "", err
	}

	db, err := sql.Open("sqlite", tmpDB)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var value string
	query := `SELECT value FROM moz_cookies WHERE host LIKE ? AND name = ? LIMIT 1`
	err = db.QueryRow(query, "%"+domain, name).Scan(&value)
	return value, err
}
