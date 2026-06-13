# sin-websearch

Unified Intelligence Gateway for OpenSIN — a single Go binary that orchestrates 20+ web sources (Reddit, Hacker News, Polymarket, GitHub, Brave, Bluesky, SearxNG, Perplexity, SerpAPI, YouTube) with entity resolution, humor/virality scoring, intelligent caching, multi-agent research missions, and video intelligence.

## Installation

```bash
go install github.com/OpenSIN-Code/web_search_bundle/cmd/sin-websearch@latest
```

## Quick Start

```bash
# Search across all sources
sin-websearch search "OpenClaw"

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
```

## Configuration

Create `~/.config/sin-websearch/sin-websearch.yaml`:

```yaml
serpapi_keys:
  - "your-serpapi-key"
brave_api_key: "your-brave-key"
openrouter_api_key: "your-openrouter-key"
scrapecreators_api_key: "your-sc-key"
groq_api_key: "your-groq-key"
openai_api_key: "your-openai-key"
http_port: 8787
searxng_urls:
  - "http://localhost:8080"
```

## Features

- **Multi-source orchestration**: Reddit, HN, Polymarket, GitHub, Brave, Bluesky, SearxNG, Perplexity, SerpAPI, YouTube
- **Entity resolution**: topic → X handles, GitHub repos, subreddits
- **Humor & virality judge**: score content by engagement and wit
- **Clustering**: merge duplicate stories across sources
- **Video intelligence**: frame extraction, native captions, Whisper transcription
- **Vision prompts**: Claude/GPT-4o/Gemini-ready prompts
- **Offline HTML briefings**: base64-embedded frames
- **Multi-agent missions**: explore + librarian agents with verification
- **MCP server**: integrate with sin-code, Claude, Cursor, etc.
- **HTTP REST API**: any app can call it
- **Infisical secrets**: load keys from Infisical CLI

## Documentation

See `AGENTS.md` for agent-facing conventions and `docs/` for architecture docs.

## License

MIT — OpenSIN-Code
