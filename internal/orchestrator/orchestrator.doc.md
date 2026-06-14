# orchestrator.go

Parallel fan-out orchestrator across all configured search engines.

## Related files
- `engines/*.go` — individual search engines.
- `resolver/entity.go` — resolves entity before search.
- `judge/humor.go` — scores results.
- `clustering/cluster.go` — clusters results.
- `cache/cache.go` — optional cache.

## Important details
- `Search` resolves the entity, queries every engine in parallel with a 10s timeout, clusters results, and picks best takes.
- `Pulse` is a search with engagement-boosted scoring.
- `SearchStream` returns results per engine over a channel.

## Caveats
- Only uses the first expanded query (`queries[0]`) from entity resolution.
- Cache TTL is 24 hours.
- Prints summary to stderr.
