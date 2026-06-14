# alchemist.go

HTTP handlers for the alchemist and swarm endpoints.

## Related files
- `server/http.go` — server setup and routing.
- `server/handlers.go` — other HTTP handlers.
- `internal/alchemist/*.go` — daemon and swarm implementations.
- `internal/config/config.go` — server config.

## Important details
- `POST /api/v1/alchemist` runs a single autonomous research loop.
- `POST /api/v1/alchemist/swarm` runs multiple strategies in parallel.
- Default budget/runtime are 30s and 5m; safety mode defaults to `headless`.

## Caveats
- Long-running handlers use a 1-hour max cap for swarm safety.
- The `run_cmd` must be non-empty; requests are validated.
