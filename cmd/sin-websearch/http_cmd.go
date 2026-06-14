// SPDX-License-Identifier: MIT
// Purpose: HTTP command to start the REST API server.
// Docs: cmd/sin-websearch/http_cmd.doc.md
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/OpenSIN-Code/web_search_bundle/internal/config"
	"github.com/OpenSIN-Code/web_search_bundle/internal/server"
	"github.com/spf13/cobra"
)

func newHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Start the HTTP REST API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{HTTPPort: 8787}
			}
			orch, err := buildOrchestrator()
			if err != nil {
				return err
			}
			s := server.NewHTTPServer(cfg, orch)
			fmt.Printf("HTTP server starting on port %d\n", cfg.HTTPPort)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5)
				defer cancel()
				_ = s.Shutdown(shutdownCtx)
			}()
			return s.Start()
		},
	}
	return cmd
}
