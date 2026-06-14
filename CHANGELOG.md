# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.4] - 2026-06-14

### Added

- Full CEO-Audit report (`CEO_AUDIT_REPORT.md`) — grade **A (90.8/100)**.
- Backfilled tests for previously untested packages:
  - `internal/engines` — engine helpers, video parsing, transcription
  - `internal/briefing` — HTML briefing generation
  - `internal/cache` — cache set/get/expiration
  - `internal/config` — config loading and key parsing
  - `internal/session` — browser session extraction helpers
  - `internal/sidecar` — sidecar manager and binary lifecycle
  - `internal/secrets` — Infisical client and secret loading
  - `internal/profiles` — profile registry, loading, and defaults
- More benchmarks:
  - `internal/judge/humor_bench_test.go`
  - `internal/mission/orchestrator_bench_test.go`
  - `internal/orchestrator/orchestrator_bench_test.go`
  - `internal/config/config_bench_test.go`

### Fixed

- `internal/secrets/infisical.go`: `getSecret()` now checks environment variables **before** falling back to the Infisical CLI, so the secret loader works when the CLI is not installed.
- `internal/engines/engines_test.go`: replaced `nil` context with `context.TODO()` to satisfy `staticcheck` `SA1012`.

### CI / Tooling

- `golangci-lint` now runs locally and passes with **0 findings**.
- `gosec` and `govulncheck` continue to report **0 findings/vulnerabilities**.

## [0.2.3] - 2026-06-14

### Changed

- Raised minimum Go version to **1.24** (`go.mod` updated).
- Updated CI matrix to Go **1.24** and **1.25**; lint job now runs on Go **1.25**.
- Upgraded local toolchain to Go **1.26.4** via Homebrew.
- Fixed Go 1.26 `vet` finding: non-constant format string in `internal/alchemist/literature.go`.

### Added

- More benchmarks for hot paths:
  - `internal/engines/video_bench_test.go` — VTT parsing, HTML stripping, video source detection
  - `internal/resolver/entity_bench_test.go` — entity resolution and query expansion
  - `internal/verify/verify_bench_test.go` — claim extraction and verification pipeline

### Security

- `govulncheck` now reports **0 vulnerabilities** for our code with Go ≥1.25.

### Documentation

- `README.md` — added HTTP API examples for `/watch`, `/vbrief`, `/vprompt`.
- `AGENTS.md` — updated HTTP endpoint list and added Go 1.24+ requirement.

## [0.2.2] - 2026-06-14

### Security

- Resolved all `gosec` findings (down from 53 to 0):
  - Fixed `G115` integer overflow in `internal/pool/keypool.go` by documenting the modulo-bound conversion.
  - Fixed `G703` path traversal in `internal/session/browser.go` (temp path is hardcoded) and `internal/alchemist/swarm.go` (sanitized strategy names via `sanitizeStrategyName`).
  - Hardened directory permissions from `0755` to `0750` for cache, sidecar, video, imaging, briefing, and alchemist history directories.
  - Hardened internal program file writes from `0644` to `0600` in `internal/alchemist/program_md.go` and `internal/alchemist/swarm.go`.
  - Added `gitCommand` helper in `internal/alchemist/git_ops.go` to centralize intentional git subprocess calls.
  - Added documented `#nosec` annotations for CLI-tool patterns that are inherently safe: reading user-supplied file paths, running configured external binaries (git, ffmpeg, yt-dlp), and HTTP requests to search/video endpoints.
  - Fixed two `G104` unhandled-error findings in `internal/secrets/infisical.go` and `internal/cache/cache.go`.

### Added

- Benchmarks for hot paths:
  - `internal/clustering/cluster_bench_test.go` — `Clusterer.Cluster` and `similarity`.
  - `internal/pool/keypool_bench_test.go` — `KeyPool.Next` and `KeyPool.Count`.

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
