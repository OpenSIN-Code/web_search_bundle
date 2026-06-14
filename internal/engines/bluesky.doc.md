# bluesky.go

Bluesky AT Protocol search engine.

## Related files
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.

## Important details
- Queries `https://search.bsky.app/search/posts`.
- Maps post URI to a web URL via `atURIToWeb`.
- No authentication required.

## Caveats
- `atURIToWeb` only handles `at://` URIs; otherwise returns the original URI.
- Public search endpoint; availability and rate limits may vary.
