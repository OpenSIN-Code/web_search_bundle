# session_test.go

Hermetic unit tests for browser session extraction.

## Related files
- `browser.go` — implementation.
- `engines/x_twitter.go` — consumer of extracted auth tokens.

## Important details
- Tests cookie DB path discovery per OS.
- Uses an in-memory SQLite cookie DB to test Firefox extraction.
- Verifies Chromium paths are skipped.

## Caveats
- OS-specific tests are skipped on non-target platforms.
- No real browser cookies are accessed.
