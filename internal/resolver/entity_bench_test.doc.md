# entity_bench_test.go

Benchmarks for entity resolution and query expansion.

## Related files
- `entity.go` — implementation.
- `orchestrator/orchestrator.go` — uses the resolver.

## Important details
- Benchmarks known and unknown topics, plus `ExpandQueries`.

## Caveats
- Unknown topics still exercise goroutines; benchmarks measure the small built-in map.
