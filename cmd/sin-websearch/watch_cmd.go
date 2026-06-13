// Purpose: Watch command for multimodal video analysis.
// Docs: cmd/sin-websearch/watch_cmd.doc.md
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/engines"
	"github.com/OpenSIN-Code/web_search_bundle/internal/sidecar"
	"github.com/spf13/cobra"
)

func newWatchCmd() *cobra.Command {
	var (
		start      string
		end        string
		maxFrames  int
		resolution int
		fps        float64
		whisper    string
		outDir     string
		cleanup    bool
		jsonOutput bool
	)
	cmd := &cobra.Command{
		Use:   "watch [url-or-path]",
		Short: "Watch and analyze a video multimodally",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			sc, err := sidecar.NewManager()
			if err != nil {
				return fmt.Errorf("sidecar init: %w", err)
			}
			engine := engines.NewVideoEngine(sc)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()

			opts := engines.WatchOptions{
				URL:        url,
				Start:      start,
				End:        end,
				MaxFrames:  maxFrames,
				Resolution: resolution,
				FPS:        fps,
				Whisper:    whisper,
				OutDir:     outDir,
			}
			fmt.Fprintf(os.Stderr, "🎬 Analyzing: %s\n", url)
			analysis, err := engine.Watch(ctx, opts)
			if err != nil {
				return fmt.Errorf("watch failed: %w", err)
			}
			if cleanup {
				defer func() {
					if err := engine.Cleanup(analysis); err != nil {
						fmt.Fprintf(os.Stderr, "cleanup: %v\n", err)
					}
				}()
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(analysis)
			}
			fmt.Printf("\n📹 %s\n   Source: %s | Duration: %s | Mode: %s\n   Frames: %d | Transcript: %s\n",
				analysis.Title, analysis.Source, analysis.Duration.Round(time.Second), analysis.Mode,
				analysis.FrameCount, analysis.TranscriptSource)
			fmt.Printf("\nWorking dir: %s\n", analysis.WorkDir)
			return nil
		},
	}
	cmd.Flags().StringVar(&start, "start", "", "Start time")
	cmd.Flags().StringVar(&end, "end", "", "End time")
	cmd.Flags().IntVar(&maxFrames, "max-frames", 0, "Max frames")
	cmd.Flags().IntVar(&resolution, "resolution", 768, "Frame width")
	cmd.Flags().Float64Var(&fps, "fps", 0, "Override FPS")
	cmd.Flags().StringVar(&whisper, "whisper", "", "groq|openai|none")
	cmd.Flags().StringVar(&outDir, "out-dir", "", "Custom output directory")
	cmd.Flags().BoolVar(&cleanup, "cleanup", false, "Delete working dir")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
