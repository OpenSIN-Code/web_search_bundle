# video.go

HTTP handlers for video analysis, HTML briefing, and Vision-LLM prompts.

## Related files
- `server/http.go` — server setup and routing.
- `internal/engines/video.go` — video analysis engine.
- `internal/briefing/video_html.go` — HTML report generation.
- `internal/prompts/video.go` — prompt builder.

## Important details
- `POST /api/v1/watch` returns a full video analysis.
- `POST /api/v1/vbrief` generates an offline HTML brief.
- `POST /api/v1/vprompt` returns a Vision-LLM prompt.
- Uses a 5-minute context timeout.

## Caveats
- `cleanup` is deferred and errors are logged to stderr.
- Defaults to Claude model/preset if not provided.
