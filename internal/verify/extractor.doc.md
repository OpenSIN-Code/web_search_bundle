# extractor.go

Extracts factual claims from search results for later verification.

## Related files
- `claim.go` — defines `Claim`, `Citation`, and `CitationDiscipline`.
- `engine.go` — runs the full verification pipeline.
- `engines/common.go` — shared `Result` type used as input.

## Important details
- Splits titles/snippets into sentences and filters for factual markers.
- Deduplicates by merging identical text and aggregating sources.
- `isFactual` uses a simple regex + keyword heuristic; not exhaustive.

## Caveats
- Sentence splitting is naive; short or non-factual sentences are ignored.
- No NLP model; false positives/negatives possible.
