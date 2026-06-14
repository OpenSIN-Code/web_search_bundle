# video_bench_test.go

Benchmarks for video helper functions.

## Related files
- `video.go` — video analysis engine.
- `engines_test.go` — unit tests for the same helpers.

## Important details
- Benchmarks VTT parsing, HTML stripping, line deduplication, and video source detection.

## Caveats
- Uses hand-crafted sample data; not real subtitle or video files.
