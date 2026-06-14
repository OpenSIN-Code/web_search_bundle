# brave.go

Brave Search API engine.

## Related files
- `common.go` — shared `Result` and `Engine` interface.
- `search_engines_test.go` — tests using mock clients.
- `orchestrator/orchestrator.go` — consumes search results.

## Important details
- Requires a `BRAVE_API_KEY`.
- 10-second HTTP timeout.
- Returns web results mapped to the common `Result` type.

## Caveats
- Returns an error if the API key is empty.
- Only supports web search; no news/image endpoints.
