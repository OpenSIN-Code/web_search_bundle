# Archive: SIN-Code-Orchestrate-Tool

This repository is **archived** and has been superseded by the unified `sin-code` binary.

## What it was

A standalone Go tool for task management, planning, dependency tracking, parallel execution, and rollback planning.

## What replaced it

- `sin-code orchestrate` in the unified SIN-Code-Bundle.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

Task and planning logic was merged into the unified `sin-code` binary to centralize state and provide consistent CLI semantics.

## Last known version

v0.1.6-fixes

## Migration

```bash
# Old
orchestrate -action add -title "Implement feature" -tags "urgent" -format json

# New
sin-code orchestrate --action add --title "Implement feature" --tags "urgent" --format json
```

## Read-only

No new issues, PRs, or releases are accepted.
