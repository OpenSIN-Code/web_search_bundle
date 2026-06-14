# browser.go

Extracts browser cookies for authenticated sessions (X/Twitter).

## Related files
- `engines/x_twitter.go` — consumes the extracted auth token.
- `session_test.go` — tests.

## Important details
- Supports Chrome, Brave, Edge, and Firefox on macOS, Linux, and Windows.
- Firefox cookies are read from `cookies.sqlite`.
- Chromium cookies are skipped because they require OS-specific decryption.

## Caveats
- Only works when the browser is logged in to X/Twitter.
- Creates a temporary copy of the Firefox cookie DB.
- Hardcoded Chromium decryption is not implemented.
