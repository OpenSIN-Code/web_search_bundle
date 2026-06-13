# AGENTS.md — sin-websearch

## Project Purpose

`sin-websearch` is the Unified Intelligence Gateway for OpenSIN: a single Go binary that bundles web search, social pulse, entity resolution, video analysis, and multi-agent research missions.

## Architecture

- `cmd/sin-websearch/` — CLI commands (Cobra)
- `internal/engines/` — source-specific search engines (Reddit, HN, GitHub, ...)
- `internal/orchestrator/` — parallel fan-out orchestrator
- `internal/resolver/` — entity resolution (topic → handles/repos/subreddits)
- `internal/judge/` — virality and humor scoring
- `internal/clustering/` — same-story merging
- `internal/cache/` — SQLite persistent cache
- `internal/pool/` — API key rotation
- `internal/sidecar/` — external binary management (yt-dlp, ffmpeg, scrapecreators)
- `internal/engines/video.go` — multimodal video analysis
- `internal/prompts/` — Vision-LLM prompt templates
- `internal/briefing/` — offline HTML report generation
- `internal/mission/` — multi-agent research missions
- `internal/verify/` — claim verification with citation discipline
- `internal/profiles/` — research mission profiles
- `internal/mcp/` — MCP server for agent integration
- `internal/server/` — HTTP REST API
- `internal/secrets/` — Infisical / env secret loading
- `internal/config/` — Viper-based configuration

## Build & Test

```bash
go mod tidy
go build ./cmd/sin-websearch
go test ./...
```

## Code Conventions

- Every meaningful file has a `// Purpose:` header and a `.doc.md` companion (CoDocs).
- Public APIs have docstrings.
- Keep logic simple; prefer explicit error handling.
- Use `context.WithTimeout` for all external calls.
- Secrets must never be logged.

## MCP Tools

- `websearch_search` — multi-source search
- `websearch_pulse` — social pulse
- `websearch_resolve` — entity resolution
- `websearch_watch` — video analysis
- `websearch_video_brief` — HTML briefing
- `websearch_video_prompt` — Vision-LLM prompt

## HTTP Endpoints

- `POST /api/v1/search` — search
- `POST /api/v1/pulse` — pulse
- `POST /api/v1/resolve` — resolve
- `GET /health` — health check

## Dependencies

- `github.com/spf13/cobra` — CLI
- `github.com/spf13/viper` — config
- `github.com/mark3labs/mcp-go` — MCP server
- `modernc.org/sqlite` — SQLite (CGO-free)
- `golang.org/x/image` — image processing

## GitHub Issues

The project was bootstrapped from the issues at https://github.com/OpenSIN-Code/web_search_bundle/issues:
- Issue #1: Unified Intelligence Gateway plan
- Issue #2: Video intelligence integration
- Issue #3: Deprecation strategy + multi-agent research patterns

All code in this repo is built from those issues and comments.
