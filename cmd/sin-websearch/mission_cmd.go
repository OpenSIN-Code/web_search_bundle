// SPDX-License-Identifier: MIT
// Purpose: Mission command for multi-agent research.
// Docs: cmd/sin-websearch/mission_cmd.doc.md
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/OpenSIN-Code/web_search_bundle/internal/mission"
	"github.com/OpenSIN-Code/web_search_bundle/internal/profiles"
	"github.com/spf13/cobra"
)

func newMissionCmd() *cobra.Command {
	var (
		profileName string
		jsonOutput  bool
	)
	cmd := &cobra.Command{
		Use:   "mission [topic]",
		Short: "Run a multi-agent research mission",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orch, err := buildOrchestrator()
			if err != nil {
				return err
			}
			registry, err := profiles.NewRegistry("")
			if err != nil {
				return err
			}
			profile, err := registry.Get(profileName)
			if err != nil {
				return err
			}
			missionOrch := mission.NewOrchestrator(orch)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			m, err := missionOrch.Run(ctx, args[0], profile)
			if err != nil {
				return err
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(m)
			}
			fmt.Printf("Mission %s (%s)\nTopic: %s\nResults: %d\nSynthesis: %s\n",
				m.ID, m.Status, m.Topic, len(m.AllResults), m.Synthesis)
			if m.Verification != nil {
				fmt.Printf("\n%s", m.Verification.FormatText())
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "competitive-analysis", "Research profile")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
