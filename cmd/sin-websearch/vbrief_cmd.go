// SPDX-License-Identifier: MIT
// Purpose: Video briefing command that generates offline HTML reports.
// Docs: cmd/sin-websearch/vbrief_cmd.doc.md
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/briefing"
	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/prompts"
	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
	"github.com/spf13/cobra"
)

func newVBriefCmd() *cobra.Command {
	var (
		start       string
		end         string
		maxFrames   int
		resolution  int
		whisper     string
		preset      string
		model       string
		question    string
		synthesis   string
		outputPath  string
		noEmbed     bool
		jpegQuality int
	)
	cmd := &cobra.Command{
		Use:   "vbrief [url-or-path]",
		Short: "Generate an offline HTML video briefing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			sc, err := sidecar.NewManager()
			if err != nil {
				return fmt.Errorf("sidecar init: %w", err)
			}
			engine := engines.NewVideoEngine(sc)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
			defer cancel()

			opts := engines.WatchOptions{
				URL:        url,
				Start:      start,
				End:        end,
				MaxFrames:  maxFrames,
				Resolution: resolution,
				Whisper:    whisper,
			}
			fmt.Fprintf(os.Stderr, "🎬 Analyzing video: %s\n", url)
			analysis, err := engine.Watch(ctx, opts)
			if err != nil {
				return fmt.Errorf("video analysis failed: %w", err)
			}
			builtPrompt := prompts.BuildVideoPrompt(prompts.VideoPromptRequest{
				Model:        prompts.Model(model),
				Preset:       prompts.Preset(preset),
				UserQuestion: question,
				Analysis:     analysis,
			})
			path, err := briefing.GenerateVideoBriefHTML(briefing.VideoBriefOptions{
				Analysis:    analysis,
				Prompt:      builtPrompt,
				Synthesis:   synthesis,
				EmbedFrames: !noEmbed,
				JPEGQuality: jpegQuality,
				MaxWidth:    resolution,
				OutputPath:  outputPath,
			})
			if err != nil {
				return fmt.Errorf("HTML generation failed: %w", err)
			}
			fmt.Fprintf(os.Stderr, "✅ Video briefing saved to: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&start, "start", "", "Start time")
	cmd.Flags().StringVar(&end, "end", "", "End time")
	cmd.Flags().IntVar(&maxFrames, "max-frames", 0, "Max frames")
	cmd.Flags().IntVar(&resolution, "resolution", 1024, "Frame width")
	cmd.Flags().StringVar(&whisper, "whisper", "", "groq|openai|none")
	cmd.Flags().StringVar(&preset, "preset", "general", "Prompt preset")
	cmd.Flags().StringVar(&model, "model", "generic", "Vision model target")
	cmd.Flags().StringVar(&question, "question", "", "User question")
	cmd.Flags().StringVar(&synthesis, "synthesis", "", "AI synthesis text")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output HTML path")
	cmd.Flags().BoolVar(&noEmbed, "no-embed", false, "Don't embed frames")
	cmd.Flags().IntVar(&jpegQuality, "jpeg-quality", 75, "JPEG quality")
	return cmd
}

func newVBriefPromptCmd() *cobra.Command {
	var (
		start    string
		end      string
		preset   string
		model    string
		question string
		formatFl string
	)
	cmd := &cobra.Command{
		Use:   "vbrief-prompt [url-or-path]",
		Short: "Generate a Vision-LLM-ready prompt from a video",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			sc, err := sidecar.NewManager()
			if err != nil {
				return err
			}
			engine := engines.NewVideoEngine(sc)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()
			analysis, err := engine.Watch(ctx, engines.WatchOptions{URL: url, Start: start, End: end, Resolution: 1024})
			if err != nil {
				return err
			}
			built := prompts.BuildVideoPrompt(prompts.VideoPromptRequest{
				Model:        prompts.Model(model),
				Preset:       prompts.Preset(preset),
				UserQuestion: question,
				Analysis:     analysis,
			})
			if formatFl == "json" {
				fmt.Printf(`{"system":%q,"user":%q,"image_paths":%q,"image_count":%d,"token_hint":%q,"model":%q,"preset":%q}`+"\n",
					built.System, built.User, built.ImagePaths, built.ImageCount, built.TokenHint, built.Model, built.Preset)
				return nil
			}
			fmt.Println("=== SYSTEM PROMPT ===")
			fmt.Println(built.System)
			fmt.Println("\n=== USER PROMPT ===")
			fmt.Println(built.User)
			fmt.Println("\n=== IMAGE ATTACHMENTS ===")
			for i, p := range built.ImagePaths {
				fmt.Printf("[%d] %s\n", i+1, p)
			}
			fmt.Printf("\nToken estimate: %s\n", built.TokenHint)
			return nil
		},
	}
	cmd.Flags().StringVar(&start, "start", "", "Start time")
	cmd.Flags().StringVar(&end, "end", "", "End time")
	cmd.Flags().StringVar(&preset, "preset", "general", "Prompt preset")
	cmd.Flags().StringVar(&model, "model", "claude", "Target model")
	cmd.Flags().StringVar(&question, "question", "", "User question")
	cmd.Flags().StringVar(&formatFl, "format", "text", "text or json")
	return cmd
}
