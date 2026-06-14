// SPDX-License-Identifier: MIT
// Purpose: Entry point for the sin-websearch CLI.
// Docs: cmd/sin-websearch/main.doc.md
package main

import (
	"fmt"
	"os"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
	"github.com/spf13/cobra"
)

var (
	commit = "none"
	date   = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sin-websearch",
		Short: "Unified Intelligence Gateway for OpenSIN",
		Long: fmt.Sprintf(`sin-websearch orchestrates 20+ sources (Reddit, X, YouTube, TikTok, HN,
Polymarket, GitHub, SearxNG, Perplexity, ...) with entity resolution, humor judge,
intelligent caching, video intelligence, and multi-agent research missions.

Designed as the native research backend for sin-code and any app via HTTP API.

Version: %s
Commit:  %s
Built:   %s`, config.Version, commit, date),
		Version: config.Version,
	}

	rootCmd.AddCommand(newSearchCmd())
	rootCmd.AddCommand(newPulseCmd())
	rootCmd.AddCommand(newBriefCmd())
	rootCmd.AddCommand(newResolveCmd())
	rootCmd.AddCommand(newMonitorCmd())
	rootCmd.AddCommand(newWatchCmd())
	rootCmd.AddCommand(newVBriefCmd())
	rootCmd.AddCommand(newVBriefPromptCmd())
	rootCmd.AddCommand(newMissionCmd())
	rootCmd.AddCommand(newVerifyCmd())
	rootCmd.AddCommand(newProfileCmd())
	rootCmd.AddCommand(newSecretsCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newHTTPCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newAlchemistCmd())
	rootCmd.AddCommand(newCompletionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion scripts",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}
