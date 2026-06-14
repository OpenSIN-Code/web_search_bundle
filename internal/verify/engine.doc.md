# engine.go

Runs the full verification pipeline on search results and produces a report.

## Related files
- `extractor.go` — claim extraction.
- `claim.go` — data types and status constants.
- `engines/common.go` — input result type.

## Important details
- Confidence is derived from `len(sources) / MinSourcesPerClaim`.
- Supports contradiction detection (placeholder `opposes` always returns false).
- Produces a formatted text report with emoji summaries.

## Caveats
- `opposes` is not implemented; contradiction detection is currently disabled.
- Text formatting uses emoji; may not render in all terminals.
