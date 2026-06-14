# helpers_test.go

Additional hermetic unit tests for engine helpers and constructors.

## Related files
- All engine files that define the private helpers tested here.
- `common.go` — shared `Engine` interface used for name verification.

## Important details
- Covers edge cases for `dedupeLines`, `atURIToWeb`, `parseTime`, and `parseVideoTime`.
- Expands `Name()` coverage to engines that were not in the original `TestEngineNames` table.

## Caveats
- Helper functions are private and spread across multiple engine files.
- No external network calls; all tests are hermetic.
