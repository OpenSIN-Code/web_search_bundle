# sin-websearch

[![CI](https://github.com/OpenSIN-Code/web_search_bundle/actions/workflows/ci.yml/badge.svg)](https://github.com/OpenSIN-Code/web_search_bundle/actions/workflows/ci.yml)
[![CEO Audit](https://img.shields.io/badge/CEO--Audit-A%2B%20100.0%2F100-brightgreen)](./CEO_AUDIT_REPORT.md)
[![Release](https://img.shields.io/github/v/release/OpenSIN-Code/web_search_bundle)](https://github.com/OpenSIN-Code/web_search_bundle/releases)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Unified Intelligence Gateway for OpenSIN — a single Go binary that orchestrates 20+ web sources (Reddit, Hacker News, Polymarket, GitHub, Brave, Bluesky, SearxNG, Perplexity, SerpAPI, Tavily, DuckDuckGo, YouTube) with entity resolution, humor/virality scoring, intelligent caching, semantic caching, cost-aware provider routing, multi-agent research missions, and video intelligence.

## Installation

```bash
go install github.com/OpenSIN-Code/web_search_bundle/cmd/sin-websearch@latest
```

## Homebrew

Install the latest release binary directly from the formula:

```bash
brew install --formula https://raw.githubusercontent.com/OpenSIN-Code/web_search_bundle/main/scripts/homebrew/sin-websearch.rb
```

Or tap the repository and install from the local formula path:

```bash
brew tap OpenSIN-Code/web_search_bundle https://github.com/OpenSIN-Code/web_search_bundle
brew install --formula "$(brew --repo OpenSIN-Code/web_search_bundle)/scripts/homebrew/sin-websearch.rb"
```

## Quick Start

```bash
# Search across all sources
sin-websearch search "OpenClaw"

# Streaming search (NDJSON incremental results per source)
sin-websearch search "OpenClaw" --stream

# Free keyless search via DuckDuckGo
sin-websearch search "OpenClaw" --engine duckduckgo

# Social pulse (engagement-focused)
sin-websearch pulse "OpenClaw"

# Resolve a topic to platform handles
sin-websearch resolve "Peter Steinberger"

# Multi-agent research mission
sin-websearch mission "OpenClaw vs Hermes" --profile competitive-analysis

# Analyze a video
sin-websearch watch https://youtu.be/dQw4w9YgXcQ

# Generate an offline HTML briefing
sin-websearch vbrief https://youtu.be/dQw4w9YgXcQ --preset tutorial

# Start MCP server
sin-websearch serve

# Start HTTP API
sin-websearch http

# Initialize an alchemist program.md
sin-websearch alchemist init --template go

# Run a single autonomous alchemist loop
sin-websearch alchemist run --cmd "go test -bench=." --target train.py

# Run a multi-strategy swarm
sin-websearch alchemist swarm --cmd "go test -bench=." --runtime 1h

# HTTP API: alchemist loop
curl -X POST http://localhost:8787/api/v1/alchemist \
  -H 'Content-Type: application/json' \
  -d '{"run_cmd":"echo metric: 0.8","target":"train.py","max_experiments":3}'

# HTTP API: alchemist swarm
curl -X POST http://localhost:8787/api/v1/alchemist/swarm \
  -H 'Content-Type: application/json' \
  -d '{"run_cmd":"echo metric: 0.8","target":"train.py","strategies":["minimal"],"runtime":"1m"}'

# HTTP API: video analysis
curl -X POST http://localhost:8787/api/v1/watch \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://youtu.be/dQw4w9YgXcQ","max_frames":80,"resolution":768}'

# HTTP API: video briefing
curl -X POST http://localhost:8787/api/v1/vbrief \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://youtu.be/dQw4w9WgXcQ","preset":"tutorial"}'

# HTTP API: vision prompt
curl -X POST http://localhost:8787/api/v1/vprompt \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://youtu.be/dQw4w9WgXcQ","prompt":"Describe the key visual moments"}'
```

## SIN-Code Integration

`sin-websearch` is the official Go-native websearch skill for [SIN-Code](https://github.com/OpenSIN-Code/SIN-Code). It exposes the same tools as the deprecated Python skill (`websearch_search`, `websearch_pulse`, `websearch_resolve`, `websearch_watch`, `websearch_video_brief`, `websearch_video_prompt`, `websearch_alchemist`) via the MCP server.

Install the skill from SIN-Code:

```bash
sin-code skill install websearch
sin-code mcp list   # should show: websearch  stdio  <SIN_SKILLS_DIR>/web_search_bundle/sin-websearch serve
```

This clones `web_search_bundle`, builds the `sin-websearch` binary, and registers it with SIN-Code. No manual PATH setup or `go install` is required.

## Configuration

Create `~/.config/sin-websearch/sin-websearch.yaml`:

```yaml
serpapi_keys:
  - "your-serpapi-key"
brave_api_key: "your-brave-key"
tavily_api_key: "your-tavily-key"          # Tavily search engine
tavily_default_depth: "basic"              # ultra-fast | fast | basic | advanced
tavily_include_answer: false               # include Tavily-generated answer
duckduckgo_enabled: true                   # free keyless search (no API key)
openrouter_api_key: "your-openrouter-key"
scrapecreators_api_key: "your-sc-key"
groq_api_key: "your-groq-key"
openai_api_key: "your-openai-key"
http_port: 8787
searxng_urls:
  - "http://localhost:8080"
rate_limit_rps: 10.0      # per-IP requests per second
rate_limit_burst: 20      # per-IP burst capacity
cost_aware_routing: false           # route to cheapest capable provider
semantic_cache_enabled: false       # embedding-based cache hits
semantic_cache_threshold: 0.85      # cosine similarity threshold
semantic_cache_ttl: "24h"           # cache entry TTL
nim_api_key: "your-nim-key"         # NVIDIA NIM embeddings (TF-IDF fallback if unset)
mcp_annotations: true               # MCP spec 2025-11-25 tool annotations
mcp_output_schema: true             # structured outputSchema for MCP tools
```

## Features

- **Multi-source orchestration**: Reddit, HN, Polymarket, GitHub, Brave, Bluesky, SearxNG, Perplexity, SerpAPI, Tavily, DuckDuckGo, YouTube
- **Tavily search**: 4-level depth tiering (`ultra-fast`/`fast`/`basic`/`advanced`) with `include_answer` support
- **DuckDuckGo search**: free keyless web search, no API key required
- **Cost-aware routing**: query classification routes to cheapest capable provider
- **Semantic caching**: embedding-based cache hit detection (>0.85 cosine similarity) with NIM or TF-IDF fallback
- **Streaming MCP search**: NDJSON incremental results as each source completes
- **MCP tool annotations & outputSchema**: full MCP spec 2025-11-25 compliance for client-side validation
- **Entity resolution**: topic → X handles, GitHub repos, subreddits
- **Humor & virality judge**: score content by engagement and wit
- **Clustering**: merge duplicate stories across sources
- **Video intelligence**: frame extraction, native captions, Whisper transcription
- **Vision prompts**: Claude/GPT-4o/Gemini-ready prompts
- **Offline HTML briefings**: base64-embedded frames
- **Multi-agent missions**: explore + librarian agents with verification
- **Alchemist autoresearch**: Karpathy-style optimization loops with git safety
- **Swarm-Alchemist**: multi-strategy parallel research with winner selection
- **Literature-Loader**: periodic sin-websearch refresh of hypotheses
- **MCP server**: integrate with [SIN-Code](https://github.com/OpenSIN-Code/SIN-Code), Claude, Cursor, etc.
- **HTTP REST API**: any app can call it
- **Infisical secrets**: load keys from Infisical CLI

## Quality & Security

- **CEO-Audit**: A+ 100.0/100 — see `CEO_AUDIT_REPORT.md`
- **Vulnerabilities**: `govulncheck` reports 0
- **Static analysis**: `gosec` and `golangci-lint` report 0 findings
- **CI/CD**: GitHub Actions run `ci.yml` (build/test/vet), `ceo-audit.yml` (47-gate audit), and `release.yml` (cross-platform binaries) on every push/PR/tag
- **SBOM**: `sbom.spdx.json` (SPDX) and `bom.json` (CycloneDX)
- **Security policy**: see `SECURITY.md` and `.well-known/security.txt`

## Development

```bash
git clone https://github.com/OpenSIN-Code/web_search_bundle.git
cd web_search_bundle
make test     # run tests
make cover    # coverage report
make lint     # golangci-lint
make sec      # gosec + govulncheck
make audit    # CEO-Audit
```

## Deployment

See [docs/deployment.md](docs/deployment.md) for TLS, reverse proxy, systemd, and Docker guidance.

## Documentation

See `AGENTS.md` for agent-facing conventions and `docs/` for architecture docs.

## License

MIT — OpenSIN-Code
