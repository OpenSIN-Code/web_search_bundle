CLI command to show a summary of the last alchemist run.

- Reads experiment history via `alchemist.NewHistory`.
- Prints total experiments, committed/discarded/error counts, success rate, and recent entries.

Related files: `alchemist_cmd.go`, `alchemist_history_cmd.go`, `internal/alchemist/history.go`.
