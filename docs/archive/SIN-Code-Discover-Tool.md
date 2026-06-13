# Archive: SIN-Code-Discover-Tool

This repository is **archived** and has been superseded by the unified `sin-code` binary.

## What it was

A standalone Go tool that searched files and project structures with relevance scoring, dependency mapping, and related-file grouping.

## What replaced it

- `sin-code discover` in the unified SIN-Code-Bundle.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

To reduce maintenance overhead and provide a single, consistent CLI for all agent-engineering tasks. The functionality was merged into the unified `sin-code` binary without behavior changes.

## Last known version

v0.2.5-fixes

## Migration

```bash
# Old
sin-discover -path . -pattern "**/*.go" -format json

# New
sin-code discover --path . --pattern "**/*.go" --format json
```

## Read-only

No new issues, PRs, or releases are accepted. Forks for historical reference are welcome.
