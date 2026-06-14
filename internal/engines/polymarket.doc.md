# polymarket.go

Polymarket prediction-market search engine.

## Related files
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.

## Important details
- Queries `https://clob.polymarket.com/markets` for active markets.
- Client-side filters by query string against question+description.
- Engagement is mapped from `volume24h`.

## Caveats
- Does not support empty queries (returns all active markets).
- Date parsing ignores malformed values.
