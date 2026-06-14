# humor.go

Scores content for relevance, virality, and humor signals.

## Related files
- `orchestrator/orchestrator.go` — scores search results.
- `humor_bench_test.go` — benchmarks.

## Important details
- Combines engagement (upvotes + views/100) with regex/keyword humor detection.
- `BestTakes` selects top items by weighted virality + humor.
- Uses hardcoded viral patterns and humor signals.

## Caveats
- Relevance is a placeholder (0.5).
- Sorting in `BestTakes` is O(n²).
- English-centric signal list.
