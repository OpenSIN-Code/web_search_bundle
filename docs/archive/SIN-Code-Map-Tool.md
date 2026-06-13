# Archive: SIN-Code-Map-Tool

This repository is **archived** and has been superseded by the unified `sin-code` binary.

## What it was

A standalone Go tool for architecture analysis, module mapping, entry-point detection, hot-path tracing, dependency-graph generation, and orphan detection.

## What replaced it

- `sin-code map` in the unified SIN-Code-Bundle.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

Architecture analysis was merged into the unified `sin-code` binary so agents can use a single tool for exploration, mapping, and refactoring support.

## Last known version

v0.2.5-fixes

## Migration

```bash
# Old
map -path /repo -action map -format json

# New
sin-code map --path /repo --action map --format json
```

## Read-only

No new issues, PRs, or releases are accepted.
