# orchestrator_test.go

Unit tests for the fan-out search orchestrator.

## Related files
- `orchestrator.go` — implementation under test.
- `orchestrator_bench_test.go` — benchmarks.
- `cache/cache.go`, `engines/common.go`, `resolver/entity.go` — dependencies.

## Important details
- Uses deterministic `testEngine` to avoid network calls.
- Covers combined results, error recording, timeout, cache hits, streaming, pulse, and summary output.

## Caveats
- Summary tests capture stderr via `os.Pipe`; restore `os.Stderr` on failure to avoid breaking later tests.
