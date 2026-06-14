# orchestrator_bench_test.go

Benchmarks for the fan-out search orchestrator.

## Related files
- `orchestrator.go` — implementation.
- `cache/cache.go`, `engines/common.go` — dependencies.

## Important details
- Uses in-memory `benchEngine` and an in-memory cache.
- Suppresses stdout/stderr to keep benchmark output clean.
- Benchmarks search, pulse, streaming, cache hits, and summary printing.

## Caveats
- Synthetic engines avoid network latency; real performance will differ.
