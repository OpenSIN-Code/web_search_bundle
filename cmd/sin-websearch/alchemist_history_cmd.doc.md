CLI command to list alchemist experiment history.

- Reads the SQLite-backed history via `alchemist.NewHistory`.
- Prints a formatted table of recent experiments.
- Supports `--limit` to control the number of records.

Related files: `alchemist_cmd.go`, `alchemist_report_cmd.go`, `internal/alchemist/history.go`.
