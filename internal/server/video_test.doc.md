# video_test.go

Unit tests for the video analysis, briefing, and vision prompt handlers in `video.go`.

## What it tests

- Rejects malformed JSON and non-POST methods for `/api/v1/watch`, `/api/v1/vbrief`, and `/api/v1/vprompt`.
- Returns 500 when `sidecar.NewManager` fails (simulated by pointing `HOME` at `/dev/null`).
- Returns 500 when the video engine fails during analysis (simulated by pre-installing
  fake `yt-dlp` and `ffmpeg` scripts that exit immediately).

## Dependencies

- `internal/sidecar` — sidecar manager used by the video handlers.
- `internal/engines` — video engine called by the handlers.

## Notes

- Tests avoid real network calls and real binary downloads by faking the sidecar environment.
- Missing URL cases are already covered in `http_test.go`.

## Run

```bash
go test ./internal/server -run 'TestHandleWatch|TestHandleVideo'
```
