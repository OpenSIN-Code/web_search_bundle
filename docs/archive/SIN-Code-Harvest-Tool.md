# Archive: SIN-Code-Harvest-Tool

This repository is **archived** and has been superseded by the unified `sin-code` binary.

## What it was

A standalone Go tool for fetching URLs and APIs with caching, structure extraction, change detection, and auth management.

## What replaced it

- `sin-code harvest` in the unified SIN-Code-Bundle.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

HTTP fetching and API consumption were merged into the unified `sin-code` binary to share caching, secret handling, and retry logic with other tools.

## Last known version

v0.1.4-fixes

## Migration

```bash
# Old
harvest -url "https://api.example.com/data" -format json

# New
sin-code harvest --url "https://api.example.com/data" --format json
```

## Read-only

No new issues, PRs, or releases are accepted.
