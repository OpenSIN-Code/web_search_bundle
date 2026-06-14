# youtube.go

YouTube search engine using the `yt-dlp` sidecar for metadata and transcripts.

## Related files
- `video.go` — video analysis engine.
- `sidecar/manager.go` — downloads and executes `yt-dlp`.
- `common.go` — shared types.

## Important details
- Returns `YouTubeResult` with title, URL, views, likes, channel, and transcript URL.
- Uses `ytsearchN:` query syntax.
- `SearchResults` adapts to the common `Result` type.

## Caveats
- Requires `yt-dlp` to be installed or downloadable by the sidecar manager.
- English auto-captions are used; other languages may be ignored.
