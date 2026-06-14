// SPDX-License-Identifier: MIT
// Purpose: CLI command registration for alchemist subcommands.
// Docs: alchemist_cmd.doc.md
package main

import (
	"github.com/spf13/cobra"
)

func newAlchemistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alchemist",
		Short: "Autonomous Karpathy-style research loops",
		Long: `Run autonomous experiments with fixed-budget loops.
The agent modifies a target file, runs a verification command (M3 gate),
measures a metric, commits if improved, resets if not. Repeats overnight.

Safety invariants (from SIN-Code v3.5.0):
  • no gate → no daemon (M3: verification cmd required)
  • headless → no git changes (M4: ask=deny enforced)
  • budget exhausted → hook summons the human

Examples:
  # Overnight run (8 hours, auto-commit locally, never push)
  sin-websearch alchemist run --budget 5m --runtime 8h --cmd "go test -bench=."

  # Headless mode (log only, no git changes — safest)
  sin-websearch alchemist run --safety headless --budget 5m

  # View morning report from last run
  sin-websearch alchemist report

  # List experiment history
  sin-websearch alchemist history --limit 20

  # Initialize program.md template
  sin-websearch alchemist init`,
	}

	cmd.AddCommand(newAlchemistRunCmd())
	cmd.AddCommand(newAlchemistSwarmCmd())
	cmd.AddCommand(newAlchemistReportCmd())
	cmd.AddCommand(newAlchemistHistoryCmd())
	cmd.AddCommand(newAlchemistInitCmd())

	return cmd
}
