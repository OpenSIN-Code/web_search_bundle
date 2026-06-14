# video_bench_test.go

Benchmarks for video prompt construction.

## Related files
- `video.go` — prompt builder implementation.
- `video_test.go` — unit tests.

## Important details
- Builds prompts with synthetic analyses of varying frame/word counts.
- Benchmarks all presets and the preset/model list helpers.

## Caveats
- Synthetic frame paths are not real image files.
