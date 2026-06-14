# entity_test.go

Hermetic unit tests for entity resolution and query expansion.

## Related files
- `entity.go` — the resolver implementation under test.
- `entity_bench_test.go` — benchmarks.

## Important details
- Covers known-topic resolution, case-insensitive lookup, caching, unknown topics, and `ExpandQueries`.
- Uses only the built-in knowledge map; no external calls.

## Caveats
- YouTube and TikTok resolvers are currently no-ops and are not exercised.
