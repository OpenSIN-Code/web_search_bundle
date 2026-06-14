# video_test.go

Hermetic tests for video analysis helpers and orchestration.

## Related files
- `video.go` — production code under test.
- `internal/sidecar/manager.go` — used to inject fake `yt-dlp` and `ffmpeg` binaries.
- `internal/imaging/resize.go` — used for the `WatchWithResize` test.

## Important details
- Creates fake `yt-dlp` and `ffmpeg` shell scripts in a temp `HOME/.sin-websearch/bin` directory.
- Covers `Cleanup`, `getMetadata`, `extractFrames`, `getTranscript`, `Watch`, and `WatchWithResize`.
- Tests metadata fallback, native transcript extraction, focused mode, and missing binary errors.

## Caveats
- Fake binaries produce canned JSON/VTT/frame output instead of touching real video files.
- `WatchWithResize` only exercises the empty-frames short-circuit; real image resizing is tested in `internal/imaging`.
- The `whisper` transcription path is not exercised here because it requires HTTP mocking; see `whisper_test.go`.
