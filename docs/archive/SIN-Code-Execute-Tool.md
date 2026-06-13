# Archive: SIN-Code-Execute-Tool

This repository is **archived** and has been superseded by the unified `sin-code` binary.

## What it was

A standalone Go tool for safe command execution with timeout, output capture, safety checks, secret redaction, and error analysis.

## What replaced it

- `sin-code execute` in the unified SIN-Code-Bundle.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

The execution logic was merged into the unified `sin-code` binary to provide a single CLI for all agent-engineering tasks. Behavior and safety rules were preserved.

## Last known version

v0.2.4-fixes

## Migration

```bash
# Old
execute -command "go test ./..." -timeout 60 -format json

# New
sin-code execute --command "go test ./..." --timeout 60 --format json
```

## Read-only

No new issues, PRs, or releases are accepted.
