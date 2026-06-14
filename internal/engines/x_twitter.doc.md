# x_twitter.go

X/Twitter search engine using local browser session cookies.

## Related files
- `session/browser.go` — extracts browser auth tokens.
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.

## Important details
- Reads the `auth_token`/`ct0` cookie from Chrome/Brave/Firefox/Edge.
- Calls the public adaptive search endpoint.
- Uses a hardcoded public Bearer token.

## Caveats
- Requires a logged-in X/Twitter session in a supported browser.
- Chromium cookie decryption is not implemented; only Firefox cookies are usable.
- Fragile: depends on X's internal API shape.
