# http_errors_test.go

Hermetic tests for HTTP engine error branches and edge cases.

## Related files
- All HTTP-based engines: `bluesky.go`, `brave.go`, `github.go`, `hackernews.go`, `perplexity.go`, `polymarket.go`, `reddit.go`, `searxng.go`, `serpapi.go`, `x_twitter.go`.
- `search_engines_test.go` — defines the `mockRoundTripper` reused here.

## Important details
- Uses `mockRoundTripper` and `errorRoundTripper` to avoid real network calls.
- Covers non-200 status codes, invalid JSON responses, transport errors, and adapter error paths.
- Tests Polymarket query filtering and Reddit self-post URL fallback.

## Caveats
- `mockRoundTripperWithRequest` duplicates the standard mock to capture request headers.
- Does not test the live X/Twitter browser session extraction.
