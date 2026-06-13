# Archive: SIN-Code-Grasp-Tool

This repository is **archived** and has been superseded by the unified `sin-code` binary.

## What it was

A standalone Go tool for deep single-file analysis, including structure, dependencies, usage context, and related-file suggestions.

## What replaced it

- `sin-code grasp` in the unified SIN-Code-Bundle.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

Single-file analysis was folded into the unified `sin-code` binary to provide consistent output formats and shared context with other tools.

## Last known version

v0.2.4-fixes

## Migration

```bash
# Old
grasp -path /repo/main.go -format json

# New
sin-code grasp --path /repo/main.go --format json
```

## Read-only

No new issues, PRs, or releases are accepted.
