# common.go

Shared types and interfaces for all search engines.

## Related files
- All engine implementations (`reddit.go`, `github.go`, `bluesky.go`, etc.).
- `orchestrator/orchestrator.go` — orchestrates engines.
- `verify/extractor.go` and `verify/engine.go` — use `Result`.

## Important details
- Defines `Result`, `Engine`, and `SourceContext`.
- `Result` is the normalized search result format used across the system.

## Caveats
- `SourceContext.Pool` is typed as `interface{}` for flexibility but requires casts.
- `Engine` interface is intentionally small.
