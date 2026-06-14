# verify_bench_test.go

Benchmarks for the claim extraction and verification pipeline.

## Related files
- `extractor.go` — claim extraction logic.
- `engine.go` — verification engine.
- `claim.go` — citation discipline and result types.

## Important details
- Uses synthetic search results from `engines.Result`.
- Benchmarks claim extraction, verification, sentence splitting, and merge/deduplication.

## Caveats
- Not a functional test; only measures throughput with artificial data.
