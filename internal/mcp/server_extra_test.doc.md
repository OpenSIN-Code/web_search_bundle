# server_extra_test.go

Additional unit tests for MCP server helpers and error paths.

## Related files
- `server.go` — MCP server implementation under test.
- `server_test.go` — main MCP server tests.

## Important details
- Tests argument helpers (`argString`, `argBool`, `argInt`).
- Exercises error handling for search, pulse, resolve, and alchemist tools.
- Uses `errorOrchestrator` to force failures without network calls.

## Caveats
- Some tests rely on the current working directory not being a git repo to trigger init errors.
- `TestHandleAlchemistSwarmInit` checks protocol-level parsing, not end-to-end swarm behavior.
