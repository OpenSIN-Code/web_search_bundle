# video_test.go

Unit tests for video prompt builders.

## Related files
- `video.go` — prompt builder implementation.
- `video_bench_test.go` — benchmarks.

## Important details
- Tests default prompt values, model-specific prefixes, all presets, no-frames, transcript truncation, token estimation, duration formatting, and list helpers.

## Caveats
- Uses a synthetic `engines.VideoAnalysis`; no real frames are read.
