# serpapi.go

SerpAPI search engine with rotating API key pool.

## Related files
- `common.go` — shared types.
- `pool/keypool.go` — key rotation.
- `search_engines_test.go` — mock tests.

## Important details
- Uses `pool.New` for key rotation.
- Bans a key for 5 minutes after HTTP 429.
- 15-second timeout.

## Caveats
- Only returns organic web results; no other SerpAPI verticals.
- `source=sin-websearch` parameter is included.
