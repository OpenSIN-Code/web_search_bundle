# cluster_bench_test.go

Benchmarks for story clustering.

## Related files
- `cluster.go` — implementation.
- `cluster_test.go` — unit tests.

## Important details
- Measures `Cluster` with 100 and 1000 synthetic items.
- Also benchmarks `similarity` on a near-duplicate pair.

## Caveats
- Synthetic data is highly repetitive; not representative of real content.
