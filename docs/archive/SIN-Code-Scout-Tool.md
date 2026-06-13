# Archive: SIN-Code-Scout-Tool

This repository is **archived** and has been superseded by the unified `sin-code` binary.

## What it was

A standalone Go tool for code search, pattern detection, regex and semantic search, usage counts, and dead-code detection.

## What replaced it

- `sin-code scout` in the unified SIN-Code-Bundle.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

Code search was integrated into the unified `sin-code` binary to share indexing and ranking infrastructure with other tools.

## Last known version

v0.1.5-fixes

## Migration

```bash
# Old
scout -query "func main" -path /repo -search_type regex -format json

# New
sin-code scout --query "func main" --path /repo --search_type regex --format json
```

## Read-only

No new issues, PRs, or releases are accepted.
