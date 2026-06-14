# engine_test.go

Hermetic unit tests for the verification engine pipeline.

## Related files
- `engine.go` — verification engine implementation.
- `claim.go` — citation discipline and result types.
- `extractor.go` — claim extraction logic.
- `internal/engines/common.go` — `Result` type.

## Important details
- Tests engine construction with default and custom discipline.
- Covers `Verify` with duplicate results, no results, weak claims, and unverified claims.
- Tests `VerificationReport.FormatText()` including truncation.

## Caveats
- `detectContradictions` is exercised indirectly because the public `opposes` method currently returns `false`.
