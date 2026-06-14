# claim_test.go

Hermetic unit tests for citation discipline defaults and claim status constants.

## Related files
- `claim.go` — citation discipline and claim types.
- `engine.go` — verification engine that uses the discipline.
- `extractor.go` — claim extractor that uses the discipline.

## Important details
- Verifies `DefaultDiscipline()` returns sensible defaults.
- Checks that all status constants are non-empty.

## Caveats
- No network or filesystem dependencies; pure value tests.
