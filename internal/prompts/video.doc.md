# video.go

Vision-LLM prompt templates for video analysis.

## Related files
- `engines/video.go` — produces the analysis consumed by prompts.
- `briefing/video_html.go` — embeds generated prompts in HTML reports.
- `server/video.go` — exposes prompt generation via HTTP.

## Important details
- Supports multiple models (Claude, GPT-4o, Gemini, Generic) and presets (bug, tutorial, hook, etc.).
- Builds system + user prompts with frame timestamps and transcript.
- Estimates token usage from frame count and transcript length.

## Caveats
- Transcript is truncated at 15,000 characters.
- Token estimation is a rough heuristic (1500 per frame).
