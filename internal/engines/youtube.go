// Purpose: YouTube search via yt-dlp sidecar for metadata and transcripts.
// Docs: internal/engines/youtube.doc.md
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
)

// YouTubeResult represents a YouTube video with transcript metadata.
type YouTubeResult struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Transcript string `json:"transcript"`
	Views      int    `json:"view_count"`
	Likes      int    `json:"like_count"`
	Channel    string `json:"channel"`
	Source     string `json:"source"`
}

// YouTubeEngine searches YouTube using yt-dlp.
type YouTubeEngine struct {
	sidecar *sidecar.Manager
}

// NewYouTubeEngine creates a YouTube engine.
func NewYouTubeEngine(sc *sidecar.Manager) *YouTubeEngine {
	return &YouTubeEngine{sidecar: sc}
}

// Name returns the engine name.
func (e *YouTubeEngine) Name() string { return "youtube" }

// Search queries YouTube via yt-dlp.
func (e *YouTubeEngine) Search(ctx context.Context, query string, numResults int) ([]YouTubeResult, error) {
	searchArg := fmt.Sprintf("ytsearch%d:%s", numResults, query)

	args := []string{
		searchArg,
		"--write-auto-sub", "--sub-lang", "en",
		"--skip-download",
		"--dump-json",
		"--no-warnings",
		"--quiet",
	}

	output, err := e.sidecar.Execute("yt-dlp", args...)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}

	var results []YouTubeResult
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		var video struct {
			Title      string `json:"title"`
			WebpageURL string `json:"webpage_url"`
			ViewCount  int    `json:"view_count"`
			LikeCount  int    `json:"like_count"`
			Channel    string `json:"channel"`
			AutoSubs   map[string][]struct {
				Ext string `json:"ext"`
				URL string `json:"url"`
			} `json:"automatic_captions"`
		}

		if err := json.Unmarshal([]byte(line), &video); err != nil {
			continue
		}

		transcript := ""
		if subs, ok := video.AutoSubs["en"]; ok && len(subs) > 0 {
			transcript = fmt.Sprintf("[Transcript available at %s]", subs[0].URL)
		}

		results = append(results, YouTubeResult{
			Title:      video.Title,
			URL:        video.WebpageURL,
			Transcript: transcript,
			Views:      video.ViewCount,
			Likes:      video.LikeCount,
			Channel:    video.Channel,
			Source:     "youtube",
		})
	}

	return results, nil
}

// SearchResults adapts YouTube results to the common Result interface.
func (e *YouTubeEngine) SearchResults(ctx context.Context, query string, limit int) ([]Result, error) {
	res, err := e.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	var results []Result
	for _, r := range res {
		results = append(results, Result{
			Title:      r.Title,
			URL:        r.URL,
			Snippet:    truncate(r.Transcript, 300),
			Source:     "youtube",
			Engagement: r.Views,
			Author:     r.Channel,
		})
	}
	return results, nil
}
