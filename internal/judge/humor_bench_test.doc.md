# humor_bench_test.go

Benchmarks for the virality/humor judge.

## Related files
- `humor.go` — scoring implementation.
- `orchestrator/orchestrator.go` — uses the judge.

## Important details
- Benchmarks single-result scoring and `BestTakes` selection with 10/100 items.

## Caveats
- Uses a small, repetitive set of sample texts.
- Not a measure of real-world humor quality.
