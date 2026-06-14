# cache_test.go

Unit tests for the SQLite-backed `Cache` implementation in `cache.go`.

## Related files
- `cache.go` — the cache implementation under test.
- `cache_test.go` — these tests.

## Important details
- Creates isolated databases in `t.TempDir()` to keep tests hermetic.
- Covers set/get, expiry, video-specific entries, stats, compaction, and close.

## Caveats
- Tests do not exercise concurrent writers; production code should handle that separately.
- No external network calls are made.
