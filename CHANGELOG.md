# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1] - 2026-06-14

### Fixed

- CI build failed because the `cmd/sin-websearch/` directory was ignored by `.gitignore` (pattern `sin-websearch` matched any directory with that name). Changed to `/sin-websearch` to only ignore the root binary.
- `internal/secrets/` was ignored by the global `.gitignore_global` `secrets/` pattern. Force-added `infisical.go` and its CoDocs companion.
- Resolved all `golangci-lint` `errcheck` findings across the codebase:
  - `internal/engines/video.go` — `os.MkdirAll` errors
  - `internal/engines/whisper.go` — `writer.WriteField` errors
  - `internal/imaging/resize.go` — `os.MkdirAll` error
  - `internal/briefing/video_html.go` — `os.MkdirAll` error
  - `internal/alchemist/daemon.go` — `git.ReturnToMainBranch` and `git.Restore` errors
  - `internal/server/http.go` — `json.Encoder.Encode` error
  - `cmd/sin-websearch/watch_cmd.go` — `engine.Cleanup` error
- Removed unused `sync.Mutex` from `internal/alchemist/daemon.go`.
- Removed unused `redact()` helper from `internal/server/http.go`.

### Changed

- `internal/experiment/loop_test.go` — removed ineffectual `reason` assignments.

## [0.2.0] - 2026-06-14

### Added

- **HTTP API for alchemist and swarm**
  - `POST /api/v1/alchemist` — run a single autonomous research loop.
  - `POST /api/v1/alchemist/swarm` — run a multi-strategy parallel swarm.
  - Request validation, defaults, and safety-mode enforcement.
- **MCP alchemist tool**
  - `websearch_alchemist` supports both single-daemon and swarm modes.
- **Alchemist tests and bug fixes**
  - Smoke tests for strategies, `ProgramMD`, history, literature loader, and swarm.
  - Fixed `ProgramMD` parsing bug where the word "learning" in a hypothesis line switched sections.
  - Fixed `ProgramMD` flush to persist added hypotheses and update `rawContent`.
  - Fixed daemon panic when generating the morning report (nil context passed to `git diff`).
  - Moved SQLite history creation into `daemon.Run()` after the work-branch checkout.
  - Swarm now commits per-strategy `program.<strategy>.md` so the work tree is clean.
- **Documentation**
  - `docs/alchemist.md` with full usage guide for alchemist and swarm.
  - HTTP API examples added to `README.md` and `docs/alchemist.md`.
  - `AGENTS.md` updated with new MCP tools and HTTP endpoints.
- **CI**
  - `.github/workflows/ci.yml` with `go test`, `go build`, `go vet`, `go mod verify` for Go 1.23/1.24, and `golangci-lint`.

### Changed

- `MorningReport.RenderMarkdown()` now falls back to simple rendering when template parsing fails.

### Deprecated

- `OpenSIN-Code/SIN-Code-Websearch-Skill` — archived, superseded by `web_search_bundle`.
- `OpenSIN-Code/SIN-Websearch-SerpAPI-Bundle` — archived, superseded by `web_search_bundle`.
- `OpenSIN-Code/SIN-Code-Bundle` (Python legacy) — no longer exists; active stack is `OpenSIN-Code/SIN-Code` (Go).

### Fixed

- `TestHandleAlchemist` no longer receives empty markdown because `RenderMarkdown()` falls back to `renderSimple()`.

## [0.1.0] - 2026-06-12

### Added

- Initial `sin-websearch` unified intelligence gateway.
- Multi-source search orchestration: Reddit, Hacker News, Polymarket, GitHub, Brave, Bluesky, SearxNG, Perplexity, SerpAPI, YouTube.
- Entity resolution (`resolve`) to map topics to platform handles.
- Social pulse (`pulse`) focused on engagement.
- Humor and virality scoring.
- Story clustering across sources.
- Video intelligence: frame extraction, captions, Whisper transcription.
- Vision-LLM prompt generation for Claude, GPT-4o, Gemini.
- Offline HTML video briefings with base64-embedded frames.
- Multi-agent research missions with verification.
- MCP server exposing search, pulse, resolve, watch, video brief, and video prompt tools.
- HTTP REST API with `/health`, `/api/v1/search`, `/api/v1/search/stream`, `/api/v1/pulse`, `/api/v1/resolve`.
- SQLite-based caching and key rotation.
- Infisical / environment secret loading.
- Cobra CLI and Viper configuration.

[0.2.1]: https://github.com/OpenSIN-Code/web_search_bundle/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/OpenSIN-Code/web_search_bundle/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/OpenSIN-Code/web_search_bundle/releases/tag/v0.1.0
