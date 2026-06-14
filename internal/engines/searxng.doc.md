# searxng.go

SearxNG proxy search engine.

## Related files
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.

## Important details
- Reads `SEARXNG_URL` or defaults to `http://localhost:8080`.
- Requests JSON format with `time_range=month`.
- Maps the first backend engine name to the result.

## Caveats
- Requires a running SearxNG instance.
- Results are limited by `numResults` after the response is received.
