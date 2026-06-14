# cluster_test.go

Unit tests for the story clustering package.

## Related files
- `cluster.go` — `Clusterer` implementation under test.
- `cluster_bench_test.go` — benchmarks.

## Important details
- Verifies that similar titles are grouped into the same cluster.
- Uses synthetic items from Reddit and Hacker News.

## Caveats
- Small test set; does not validate threshold edge cases.
