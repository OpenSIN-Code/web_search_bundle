# briefing_test.go

Hermetic unit tests for video briefing HTML generation.

## Related files
- `video_html.go` — the HTML briefing implementation.
- `internal/prompts/video.go` — prompt builder used by briefs.
- `internal/engines/video.go` — video analysis type.

## Important details
- Uses a minimal valid JPEG header for test images.
- Covers `GenerateVideoBriefHTML`, image-to-data-URL, slugify, and duration formatting.
- Tests both embedded and non-embedded frame modes.

## Caveats
- The minimal JPEG is not a fully valid image; only dimension parsing is exercised.
- Missing-frame behavior differs between embed and non-embed modes.
