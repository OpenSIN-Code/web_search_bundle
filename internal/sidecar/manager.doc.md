# manager.go

Downloads and executes external helper binaries (`yt-dlp`, `scrapecreators`, `ffmpeg`).

## Related files
- `sidecar_test.go` — tests.
- `engines/video.go` and `engines/youtube.go` — use the manager.
- `cmd/sin-websearch/watch_cmd.go` — CLI video command.

## Important details
- Binaries are stored in `~/.sin-websearch/bin`.
- Downloads platform-specific URLs for `yt-dlp` and `scrapecreators`.
- Prefers a system `ffmpeg` via symlink when available.

## Caveats
- Downloading and executing binaries uses `#nosec` annotations; ensure trust in the configured URLs.
- Network download is synchronous and may be slow.
- `ffmpeg` has no download URL on macOS/Linux; relies on system install.
