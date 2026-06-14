# search_engines_test.go

Hermetic unit tests for the HTTP-based search engines using mock transports.

## Related files
- `reddit.go`, `bluesky.go`, `hackernews.go`, `github.go`, `serpapi.go`, `brave.go`, `perplexity.go`, `polymarket.go`, `searxng.go`, `youtube.go`, `x_twitter.go` — engines under test.
- `common.go` — shared types.

## Important details
- `mockRoundTripper` injects fixed HTTP responses without real network calls.
- Covers missing-auth, rate-limit, and parse behaviors.

## Caveats
- YouTube test uses a fake `yt-dlp` script in a temp HOME.
- X/Twitter parse test is isolated from the live adaptive endpoint.
