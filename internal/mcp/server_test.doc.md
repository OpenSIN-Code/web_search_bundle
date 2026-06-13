# server_test.go

Tests for the MCP server tool handlers.

- Imports: `mcp` (internal), `context`, `testing`, `internal/engines`, `internal/orchestrator`
- Uses a `mockOrchestrator` implementing the `mcp.Orchestrator` interface.
- Tests: search, pulse, resolve, missing-argument handling, and alchemist missing `run_cmd`.
- Does not test video analysis paths (they require sidecar binaries).
