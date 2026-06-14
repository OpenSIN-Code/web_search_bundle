# claim.go

Defines the citation discipline and core data types for claim verification.

## Related files
- `extractor.go` — extracts claims from search results.
- `engine.go` — scores and classifies claims.
- `verify_bench_test.go` — benchmarks the pipeline.

## Important details
- `DefaultDiscipline` requires 2 sources per claim and a confidence threshold of 0.7.
- Status constants: `verified`, `weak`, `contested`, `contradicted`, `unverified`.

## Caveats
- Discipline values are opinionated defaults; callers may override them.
