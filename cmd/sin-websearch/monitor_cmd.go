// SPDX-License-Identifier: MIT
// Purpose: Monitor command for recurring topic tracking.
// Docs: cmd/sin-websearch/monitor_cmd.doc.md
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Track topics over time (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Monitor: add, list, run subcommands (stub)")
			return nil
		},
	}
	return cmd
}
