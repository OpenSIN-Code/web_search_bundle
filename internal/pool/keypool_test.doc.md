# keypool_test.go

Unit tests for the API key pool.

## Related files
- `keypool.go` — implementation.
- `keypool_bench_test.go` — benchmarks.

## Important details
- Tests empty pools, round-robin selection, and banning.

## Caveats
- Tests use a small set of keys; not a concurrency test.
