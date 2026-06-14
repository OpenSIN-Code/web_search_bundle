# extractor_test.go

Hermetic unit tests for claim extraction from search results.

## Related files
- `extractor.go` — claim extraction implementation.
- `claim.go` — `Claim` and `Citation` types.
- `internal/engines/common.go` — `Result` type.

## Important details
- Tests extractor construction and `Extract` with duplicate merging.
- Covers sentence splitting, factual heuristic, categorization, and claim deduplication.

## Caveats
- `isFactual` is a heuristic; tests exercise the current markers (numbers, "is/are/was/were/has/have/released/announced/launched").
