# handlers_test.go

Unit tests for the core HTTP handlers in `handlers.go`.

## What it tests

- `handleHealth` returns a 200 status payload.
- `handleSearch` accepts valid queries, rejects bad methods / JSON / empty queries,
  and surfaces per-engine errors without failing the whole request.
- `handlePulse` mirrors search for topic-based fan-out.
- `handleResolve` resolves a name via the server's resolver.
- `handleSearchStream` emits SSE `start`, `result`, and `done` events, and
  handles malformed input gracefully.

## Dependencies

- `internal/orchestrator` — used with a mock engine to avoid real network calls.
- `internal/engines` — defines the mock search engine used by the tests.

## Run

```bash
go test ./internal/server -run 'TestHandle'
```
