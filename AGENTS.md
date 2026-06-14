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
- `internal/mcp/` — MCP server for agent integration (consumed by SIN-Code as the `websearch` ecosystem skill)
- `internal/server/` — HTTP REST API
- `internal/secrets/` — Infisical / env secret loading
- `internal/config/` — Viper-based configuration
- `internal/experiment/` — fixed-budget autonomous research loop runner
- `internal/alchemist/` — Karpathy-style autonomous optimization daemon
- `internal/alchemist/swarm.go` — multi-strategy parallel alchemist runs
- `internal/alchemist/literature.go` — sin-websearch hypothesis refresh

## Build & Test

Requires Go 1.25+. CI tests Go 1.25 and 1.26.

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
- `websearch_alchemist` — autonomous research loop / swarm

## HTTP Endpoints

All endpoints support optional bearer-token authentication when `token` is configured, and per-IP token-bucket rate limiting (default 10 rps, burst 20) via `rate_limit_rps` / `rate_limit_burst`:

```bash
curl -X POST http://localhost:8787/api/v1/search \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"query":"OpenSIN"}'
```

- `POST /api/v1/search` — search
- `POST /api/v1/pulse` — pulse
- `POST /api/v1/resolve` — resolve
- `POST /api/v1/watch` — video analysis
- `POST /api/v1/vbrief` — offline HTML video briefing
- `POST /api/v1/vprompt` — Vision-LLM prompt for a video
- `POST /api/v1/alchemist` — autonomous research loop
- `POST /api/v1/alchemist/swarm` — multi-strategy swarm
- `GET /health` — health check

## Quality & Security

- CEO-Audit: **A+ 100.0/100** (see `CEO_AUDIT_REPORT.md`)
- `govulncheck` → 0 vulnerabilities
- `gosec` → 0 findings
- `golangci-lint` → 0 findings
- CI runs on every push/PR: `ci.yml` + `ceo-audit.yml`
- SBOMs: `sbom.spdx.json` (SPDX) and `bom.json` (CycloneDX)
- MIT `SPDX-License-Identifier` header in every `.go` file

## Development

```bash
make build    # build the binary
make test     # run all tests
make cover    # coverage report
make vet      # go vet
make lint     # golangci-lint
make sec      # gosec + govulncheck
make audit    # CEO-Audit
make clean    # remove artifacts
```

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
- Issue #4: Alchemist Daemon (Karpathy-style autoresearch)
- Issue #5: Swarm-Alchemist + Literature-Loader

All code in this repo is built from those issues and comments.
