# http.go

HTTP REST API server for `sin-websearch`.

- Exposes endpoints: `/health`, `/api/v1/search`, `/api/v1/search/stream`, `/api/v1/pulse`, `/api/v1/resolve`, `/api/v1/alchemist`, `/api/v1/alchemist/swarm`.
- Uses `internal/orchestrator` for search/pulse/resolve.
- Uses `internal/alchemist` for autonomous research loops and swarms.
- Configurable port via `internal/config`.
- All routes enforce POST where appropriate and return JSON.
- Alchemist endpoints default to `headless` safety and require `run_cmd`.

See `docs/alchemist.md` for HTTP usage examples.
