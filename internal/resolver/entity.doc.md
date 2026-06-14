# entity.go

Resolves a topic into platform-specific handles, repos, subreddits, and hashtags.

## Related files
- `orchestrator/orchestrator.go` — calls `Resolve` before searching.
- `entity_bench_test.go` — benchmarks.

## Important details
- Built-in knowledge map for a handful of topics (e.g., "openclaw", "peter steinberger", "kanye west").
- Resolves X handles, GitHub users/repos, subreddits, YouTube channels, and TikTok hashtags in parallel.
- Caches results per process.

## Caveats
- YouTube and TikTok resolvers are currently no-ops.
- Knowledge map is hardcoded; cannot be updated without code changes.
