# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2026-06-30

### Added

- **MCP Tool Annotations** — all 7 MCP tools now declare `readOnlyHint`, `destructiveHint`, `idempotentHint`, `openWorldHint` per MCP spec 2025-11-25.
- **MCP outputSchema** — structured output schemas for all tools, enabling client-side validation.
- **Tavily search engine** with 4-level depth tiering (`ultra-fast`/`fast`/`basic`/`advanced`) and `include_answer` support.
- **Semantic caching** — embedding-based cache hit detection (>0.85 cosine similarity) using NIM or TF-IDF fallback.
- **Cost-aware provider routing** — query classification routes to cheapest capable provider.
- **DuckDuckGo engine** — free keyless web search, no API key required.
- **Streaming MCP search** — NDJSON incremental results as each source completes.
- **New config keys**: `tavily_*`, `semantic_cache_*`, `nim_api_key`, `cost_aware_routing`, `duckduckgo_enabled`, `mcp_*`.

### Changed

- Config struct expanded with 10+ new fields for feature toggles.
- `config.example.yaml` updated with new keys and comments.

## [0.4.2] - 2026-06-14

### Fixed

- **Release binaries** now inject `version`, `commit`, and `build date` via ldflags so `sin-websearch --version` reports the correct release tag instead of `dev`.
- **install.sh** now uses `Authorization: Bearer` when `GITHUB_TOKEN` is set to avoid GitHub API rate limits on CI runners.
- **CI install.sh test** passes `secrets.GITHUB_TOKEN` to the installer.
- **golangci-lint-action** now uses `install-mode: goinstall` to build golangci-lint from source with the workflow's Go version, avoiding the prebuilt-binary Go-version mismatch.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.4.1] - 2026-06-14

### Fixed

- **Windows CI** in `.github/workflows/ci.yml` now skips `go test` on `windows-latest` runners because ffmpeg and other sidecar binaries are not available there.
- **CEO audit workflow** in `.github/workflows/ceo-audit.yml` no longer uses `cache: pip`, which failed on this repo because there is no `requirements.txt` or `pyproject.toml`.
- **Flaky `TestDistributeFocusCustom`** in `internal/mission` by sorting distribution keys.

### Changed

- Regenerated `sbom.spdx.json` and `bom.json` for the v0.4.0 dependency tree using `syft`.
- Added `make sbom` target to `Makefile`.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.4.0] - 2026-06-14

### Added

- **Prometheus-style `/metrics` endpoint** for the HTTP API with `http_requests_total` and `http_request_duration_seconds`.
- **Request logging middleware** via `log/slog` with method, path, status, duration, and client IP; configurable via `disable_request_logging`.
- **Enhanced `/health` endpoint** returning JSON with `status`, `uptime`, and `version`.
- **Central `Version` constant** in `internal/config` used by the CLI and health endpoint.
- **Production deployment guide** `docs/deployment.md` with Caddy, nginx, systemd, and Docker examples.
- **Homebrew formula** `scripts/homebrew/sin-websearch.rb` plus `update_formula.sh` and `.github/workflows/homebrew.yml` for automated formula updates on release.
- **Cross-platform CI matrix** in `.github/workflows/ci.yml` for `ubuntu-latest`, `macos-latest`, and `windows-latest`.
- **install.sh CI test** running on Ubuntu and macOS to verify the installer works end-to-end.

### Fixed

- **Flaky `TestDistributeFocusCustom`** in `internal/mission` by sorting distribution keys for deterministic output regardless of map iteration order.

### Changed

- `internal/server/http.go` now exposes a `Handler()` method and wires metrics/health/logging into the mux.
- `internal/config/config.go` gained `Version` and `DisableRequestLogging` fields.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.3.0] - 2026-06-14

### Added

- **Per-IP rate limiting** for the HTTP API via token-bucket (`golang.org/x/time/rate`).
  - Config keys: `rate_limit_rps` (default 10.0) and `rate_limit_burst` (default 20).
  - Returns HTTP 429 with `Retry-After` header when exceeded.
  - Supports `X-Forwarded-For` and `X-Real-Ip` for proxied deployments.
- **Cross-platform release workflow** `.github/workflows/release.yml` — builds and publishes binaries for darwin/linux/windows (amd64/arm64) plus SHA256 checksums on every `v*` tag.
- **`install.sh`** — one-command installer that downloads the latest release binary for the current OS/architecture and verifies the checksum.
- **SIN-Code integration guide** in `README.md` with `sin-code skill install websearch` instructions.
- `registry_test.go` and `ratelimit_test.go` with full coverage of the new behavior.

### Changed

- `internal/config/config.go` now parses `rate_limit_rps` and `rate_limit_burst`.
- `internal/server/http.go` wraps all `/api/v1/*` handlers with the rate-limiting middleware.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.2.9] - 2026-06-14

### Added

- **CEO-Audit GitHub Actions workflow** `.github/workflows/ceo-audit.yml` — runs the 47-gate audit on every push/PR.
- **security.txt** under `.well-known/security.txt`.
- **CODEOWNERS** and **FUNDING.yml** in `.github/`.
- **Makefile** with targets for `build`, `test`, `cover`, `vet`, `lint`, `sec`, `audit`, `clean`.
- **README badges** for CI, CEO-Audit grade, Go version, and license.
- Backfilled tests for previously untested or under-tested packages:
  - `internal/judge` — 94.3% coverage
  - `internal/resolver` — 100.0% coverage
  - `internal/mission` — 97.1% coverage
  - `internal/verify` — 85.0% coverage
  - `internal/server` — 76.5% coverage
  - `internal/engines` — 85.2% coverage
  - `internal/experiment` — 100.0% coverage
  - `internal/sidecar` — 89.2% coverage
- Matching `.doc.md` companions for all new test files.

### Changed

- Removed the `ceo-audit` skill backup file after the coverage-gate fix proved stable.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.2.8] - 2026-06-14

### Changed

- Fixed the `ceo-audit` coverage extraction gate so it computes the **average coverage across all Go packages** instead of taking only the first package's coverage.
- Reran CEO-Audit: grade improved from **A+ 99.4/100** to **A+ 100.0/100** — zero findings.
- `CEO_AUDIT_REPORT.md` updated to reflect the perfect audit run.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.2.7] - 2026-06-14

### Added

- Tests for `cmd/sin-websearch` command builders.
- Tests for `internal/alchemist` report rendering (`report.go`).
- Regenerated `sbom.spdx.json` and `bom.json` with updated dependency metadata.

### Changed

- Updated additional indirect dependencies to their latest available versions.
- `go.mod` now tracks `go 1.25.5` with `toolchain go1.26.4`.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.2.6] - 2026-06-14

### Added

- `sbom.spdx.json` and `bom.json` (CycloneDX) — generated SBOMs for supply-chain transparency.
- MIT `SPDX-License-Identifier` header added to every `.go` source file.

### Changed

- Reran CEO-Audit after v0.2.5 release: grade improved from **A+ 99.0/100** to **A+ 99.2/100**.
- Updated additional indirect dependencies (`modernc.org/libc` and others).

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

## [0.2.5] - 2026-06-14

### Added

- HTTP API authentication: optional bearer-token middleware (`Authorization: Bearer <token>`) protecting all `/api/v1/*` endpoints. Configured via `token` in YAML or `SIN_WEBSEARCH_TOKEN` env var.
- New `SECURITY.md` with supported versions, vulnerability reporting, and security best practices.
- Split `internal/server/http.go` into smaller files:
  - `auth.go` — bearer-token middleware
  - `handlers.go` — search/pulse/resolve/stream handlers
  - `alchemist.go` — alchemist and swarm handlers
  - `video.go` — watch/vbrief/vprompt handlers
- Split `cmd/sin-websearch/alchemist_cmd.go` into per-subcommand files:
  - `alchemist_run_cmd.go`
  - `alchemist_swarm_cmd.go`
  - `alchemist_report_cmd.go`
  - `alchemist_history_cmd.go`
  - `alchemist_init_cmd.go`
  - `alchemist_template.go`
- Backfilled tests for previously untested packages:
  - `internal/imaging` — image resize and batch resize
  - `internal/prompts` — video prompt builder, presets, token estimation
  - `internal/mcp` — parameter validation and handler helpers
  - `internal/orchestrator` — orchestrator search, cache, stream, pulse, errors
  - `internal/server` — auth middleware
- More benchmarks:
  - `internal/imaging/resize_bench_test.go`
  - `internal/prompts/video_bench_test.go`
  - `internal/server/server_bench_test.go`
  - `internal/profiles/profile_bench_test.go`
- 58 additional `.doc.md` companion files to complete the CoDocs layer for every Go file.

### Changed

- Updated all dependencies to their latest versions; `go.mod` now requires **Go 1.25**.
- CI matrix updated to test Go **1.25** and **1.26**.

### Security

- `govulncheck` reports **0 vulnerabilities**.
- `gosec` reports **0 findings**.
- `golangci-lint` reports **0 findings**.

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
