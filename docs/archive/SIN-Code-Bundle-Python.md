# Archive: SIN-Code-Bundle (Python)

This repository is **archived** and has been superseded by the unified `sin-code` Go binary.

## What it was

A Python-based unified CLI and MCP server that bundled the SIN-Code agent-engineering stack. It provided the `sin` command and served as the original integration point for the SIN-Code tools.

## What replaced it

- `sin-code` (Go binary) at `~/.local/bin/sin-code`.
- Source: `github.com/OpenSIN-Code/SIN-Code-Bundle`

## Why it was archived

The Go binary provides a single, self-contained executable with no Python dependency or venv management. It exposes all MCP tools under short, consistent names (`sin_discover`, `sin_execute`, etc.) and does not shadow the new tools with the legacy `sin-code-bundle_` prefix.

## Last known version

v0.2.0

## Migration

| Old (Python) | New (Go) |
| --- | --- |
| `sin status` | `sin-code tui` |
| `sin bootstrap .` | `sin-code adw` |
| `sin sin-code run discover --path .` | `sin-code discover --path .` |
| `sin serve` | `sin-code serve` |
| `sin sin-code agents-md --output AGENTS.md` | AGENTS.md ships with the repo |

## Read-only

No new issues, PRs, or releases are accepted. The Python repo remains available for historical reference.
