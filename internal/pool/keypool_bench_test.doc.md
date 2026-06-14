# keypool_bench_test.go

Benchmarks for the API key pool.

## Related files
- `keypool.go` — implementation.
- `keypool_test.go` — unit tests.

## Important details
- Benchmarks `Next` and `Count` with 10 keys.

## Caveats
- No banned keys during benchmarks; pure rotation/count path.
