# perplexity.go

Perplexity Sonar search engine via OpenRouter.

## Related files
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.

## Important details
- Reads `OPENROUTER_API_KEY` from the environment.
- 60-second timeout.
- Returns a single answer-style result.
- `SearchResults` adapts the answer to the common `Result` type.

## Caveats
- Only supports the `perplexity/sonar` model.
- No citation extraction currently.
