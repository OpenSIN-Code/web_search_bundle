# hackernews.go

Hacker News Algolia search engine.

## Related files
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.
- `orchestrator/orchestrator.go` — consumes results.

## Important details
- Queries `https://hn.algolia.com/api/v1/search`.
- 10-second timeout.
- Falls back to an HN item page if the story has no external URL.

## Caveats
- Public API; rate limits are unspecified.
- Engagement is mapped from story points.
