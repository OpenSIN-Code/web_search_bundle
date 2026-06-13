// Purpose: Manage optional external binaries (yt-dlp, scrapecreators, ffmpeg).
// Docs: internal/sidecar/manager.doc.md
package sidecar

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// Manager downloads and executes external binaries on demand.
type Manager struct {
	binDir   string
	mu       sync.Mutex
	binaries map[string]*Binary
}

// Binary describes a downloadable external tool.
type Binary struct {
	Name        string
	Version     string
	DownloadURL func(os, arch string) string
}

// NewManager creates a sidecar manager in the user's home directory.
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	binDir := filepath.Join(home, ".sin-websearch", "bin")
	if err := os.MkdirAll(binDir, 0750); err != nil {
		return nil, err
	}

	m := &Manager{
		binDir:   binDir,
		binaries: make(map[string]*Binary),
	}

	m.registerYTDLP()
	m.registerScrapeCreators()
	m.registerFFmpeg()

	return m, nil
}

func (m *Manager) registerYTDLP() {
	m.binaries["yt-dlp"] = &Binary{
		Name:    "yt-dlp",
		Version: "2026.06.01",
		DownloadURL: func(os, arch string) string {
			switch {
			case os == "darwin" && arch == "arm64":
				return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
			case os == "darwin":
				return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos_legacy"
			case os == "linux":
				return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux"
			case os == "windows":
				return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
			default:
				return ""
			}
		},
	}
}

func (m *Manager) registerScrapeCreators() {
	m.binaries["scrapecreators"] = &Binary{
		Name:    "scrapecreators",
		Version: "1.2.0",
		DownloadURL: func(os, arch string) string {
			return fmt.Sprintf("https://cli.scrapecreators.com/%s-%s/scrapecreators", os, arch)
		},
	}
}

func (m *Manager) registerFFmpeg() {
	m.binaries["ffmpeg"] = &Binary{
		Name:    "ffmpeg",
		Version: "7.0",
		DownloadURL: func(os, arch string) string {
			// These are placeholder URLs; on macOS/Linux users are asked to install via
			// brew/apt. The manager prefers a system binary if available.
			return ""
		},
	}
}

// EnsureBinary returns the path to the requested binary, downloading it if needed.
func (m *Manager) EnsureBinary(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bin, ok := m.binaries[name]
	if !ok {
		return "", fmt.Errorf("unknown binary: %s", name)
	}

	binPath := filepath.Join(m.binDir, name)
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	// Prefer system ffmpeg when available.
	if name == "ffmpeg" {
		if sys, err := findSystemFFmpeg(); err == nil {
			_ = os.Symlink(sys, binPath)
			return binPath, nil
		}
	}

	url := bin.DownloadURL(runtime.GOOS, runtime.GOARCH)
	if url == "" {
		return "", fmt.Errorf("no download URL for %s/%s; install %s manually", runtime.GOOS, runtime.GOARCH, name)
	}

	fmt.Printf("⬇️  Downloading %s...\n", name)
	resp, err := http.Get(url) // #nosec G107 — sidecar intentionally downloads from configured URL
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY, 0755) // #nosec G302 G304 — downloaded binaries are managed by the sidecar manager
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}

	return binPath, nil
}

// Execute runs a sidecar binary with the given arguments and returns its output.
func (m *Manager) Execute(name string, args ...string) ([]byte, error) {
	binPath, err := m.EnsureBinary(name)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binPath, args...) // #nosec G204 — sidecar intentionally executes managed binaries
	return cmd.CombinedOutput()
}

func findSystemFFmpeg() (string, error) {
	candidates := []string{
		"/usr/local/bin/ffmpeg",
		"/usr/bin/ffmpeg",
		"/opt/homebrew/bin/ffmpeg",
	}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, `C:\Program Files\ffmpeg\bin\ffmpeg.exe`)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("system ffmpeg not found")
}
