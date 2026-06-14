# server_bench_test.go

Benchmarks for HTTP server hot paths without external network calls.

## Related files
- `server/http.go` — server setup.
- `server/handlers.go` — other handlers tested here.
- `server/alchemist.go` — `buildAlchemistConfig` used in a benchmark.
- `internal/orchestrator/orchestrator.go` — orchestrator used for search/pulse benchmarks.

## Important details
- Uses `httptest` and a mock engine for isolated benchmarks.
- Covers auth middleware, health, search/pulse/resolve handlers, and JSON/SSE helpers.

## Caveats
- Benchmarks do not cover the alchemist or video handlers.
