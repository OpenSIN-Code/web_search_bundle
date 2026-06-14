# orchestrator_bench_test.go

Benchmarks for the multi-agent mission orchestrator.

## Related files
- `orchestrator.go` — mission orchestrator.
- `orchestrator/orchestrator.go` — base search orchestrator used in benchmarks.
- `profiles/profile.go` — profile types.

## Important details
- Uses in-memory `benchEngine` to avoid network calls.
- Suppresses stdout/stderr to avoid benchmark noise.
- Benchmarks full mission runs, explore/librarian phases, and helpers.

## Caveats
- Synthetic engines return deterministic results; not representative of real latency.
