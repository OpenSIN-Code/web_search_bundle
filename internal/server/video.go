// Purpose: HTTP handlers for video analysis, briefing, and vision prompts.
// Docs: video_handler.doc.md
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/briefing"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/prompts"
	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
)

// WatchRequest configures a video analysis.
type WatchRequest struct {
	URL        string `json:"url"`
	Start      string `json:"start"`
	End        string `json:"end"`
	MaxFrames  int    `json:"max_frames"`
	Resolution int    `json:"resolution"`
	Whisper    string `json:"whisper"`
	OutDir     string `json:"out_dir"`
	Cleanup    bool   `json:"cleanup"`
}

// VideoBriefRequest configures a video HTML briefing.
type VideoBriefRequest struct {
	URL    string `json:"url"`
	Preset string `json:"preset"`
	Embed  bool   `json:"embed"`
}

// VideoPromptRequest configures a vision prompt generation.
type VideoPromptRequest struct {
	URL    string `json:"url"`
	Model  string `json:"model"`
	Preset string `json:"preset"`
}

func (s *HTTPServer) handleWatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req WatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}
	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	sc, err := sidecar.NewManager()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "sidecar: " + err.Error()})
		return
	}
	engine := engines.NewVideoEngine(sc)

	opts := engines.WatchOptions{
		URL:        req.URL,
		Start:      req.Start,
		End:        req.End,
		MaxFrames:  req.MaxFrames,
		Resolution: req.Resolution,
		Whisper:    req.Whisper,
		OutDir:     req.OutDir,
	}
	analysis, err := engine.Watch(ctx, opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "watch: " + err.Error()})
		return
	}

	if req.Cleanup {
		defer func() {
			if err := engine.Cleanup(analysis); err != nil {
				fmt.Fprintf(os.Stderr, "cleanup: %v\n", err)
			}
		}()
	}

	writeJSON(w, http.StatusOK, analysis)
}

func (s *HTTPServer) handleVideoBrief(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req VideoBriefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}
	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	sc, err := sidecar.NewManager()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "sidecar: " + err.Error()})
		return
	}
	engine := engines.NewVideoEngine(sc)

	analysis, err := engine.Watch(ctx, engines.WatchOptions{URL: req.URL})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "watch: " + err.Error()})
		return
	}

	preset := prompts.PresetGeneral
	if req.Preset != "" {
		preset = prompts.Preset(req.Preset)
	}
	built := prompts.BuildVideoPrompt(prompts.VideoPromptRequest{Analysis: analysis, Preset: preset})
	path, err := briefing.GenerateVideoBriefHTML(briefing.VideoBriefOptions{Analysis: analysis, Prompt: built, EmbedFrames: req.Embed})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "html: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"path":        path,
		"frame_count": analysis.FrameCount,
		"title":       analysis.Title,
	})
}

func (s *HTTPServer) handleVideoPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req VideoPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request: " + err.Error()})
		return
	}
	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	sc, err := sidecar.NewManager()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "sidecar: " + err.Error()})
		return
	}
	engine := engines.NewVideoEngine(sc)

	analysis, err := engine.Watch(ctx, engines.WatchOptions{URL: req.URL})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "watch: " + err.Error()})
		return
	}

	model := prompts.ModelClaude
	if req.Model != "" {
		model = prompts.Model(req.Model)
	}
	preset := prompts.PresetGeneral
	if req.Preset != "" {
		preset = prompts.Preset(req.Preset)
	}
	built := prompts.BuildVideoPrompt(prompts.VideoPromptRequest{Analysis: analysis, Model: model, Preset: preset})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"system":      built.System,
		"user":        built.User,
		"image_paths": built.ImagePaths,
	})
}
