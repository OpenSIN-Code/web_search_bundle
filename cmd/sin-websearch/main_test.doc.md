Smoke tests for the CLI command builders in `cmd/sin-websearch`.

- Verifies every `new*Cmd` factory returns a valid `cobra.Command` with `Use` and `Short`.
- Covers the shell completion command argument validation.
- No external network calls; only command metadata is exercised.

Related files: `main.go`, `*_cmd.go` files.
