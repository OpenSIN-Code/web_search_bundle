// Purpose: Serve command to start the MCP server.
// Docs: cmd/sin-websearch/serve_cmd.doc.md
package main

import (
	"github.com/OpenSIN-Code/web_search_bundle/internal/mcp"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server on stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			orch, err := buildOrchestrator()
			if err != nil {
				return err
			}
			server := mcp.NewServer(orch)
			return server.Serve()
		},
	}
	return cmd
}
