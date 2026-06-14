# reddit.go

Reddit public JSON search engine (no API key required).

## Related files
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.

## Important details
- Uses `https://www.reddit.com/search.json`.
- Sets a custom `User-Agent`.
- Maps self-posts to their permalink.

## Caveats
- Reddit may rate-limit or block requests; this is a best-effort public endpoint.
- No authentication.
