# sidecar_test.go

Hermetic unit tests for the sidecar manager.

## Related files
- `manager.go` — the sidecar manager implementation.
- `engines/video.go` / `engines/youtube.go` — consumers of sidecar binaries.

## Important details
- Tests binary registration, download URLs, caching, and execution.
- Uses fake scripts for `yt-dlp` and `scrapecreators`.

## Caveats
- `ffmpeg` tests tolerate system absence.
- Tests create a temp `binDir` rather than using the real `~/.sin-websearch/bin`.
