# keypool.go

Rotates a pool of API keys with basic rate-limit handling.

## Related files
- `engines/serpapi.go` — uses the key pool.
- `keypool_test.go` / `keypool_bench_test.go` — tests.

## Important details
- Round-robin selection with atomic offset.
- Temporarily bans keys after HTTP 429.
- Cleans expired bans on `Next` and `Count`.

## Caveats
- Uses `sync.Mutex` + `atomic.Uint64`; potential contention under very high concurrency.
- All keys are unbanned in benchmarks.
