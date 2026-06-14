# humor_test.go

Hermetic unit tests for the `HumorJudge` scoring logic.

## Related files
- `humor.go` — the judge implementation under test.
- `humor_bench_test.go` — benchmarks.

## Important details
- Verifies virality thresholds, engagement calculation, humor detection, and best-takes selection.
- Uses only synthetic inputs; no external calls.

## Caveats
- Tests rely on the current hardcoded patterns and signals in `humor.go`.
- Does not validate real-world humor quality.
