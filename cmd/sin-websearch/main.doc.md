# main.go

Entry point for the `sin-websearch` CLI. Registers all subcommands and delegates to Cobra.

- The build version is read from `internal/config.Version` so it can be shared by the
  CLI (`--version`) and the HTTP `/health` endpoint.
- `commit` and `date` remain package-level variables and can be injected at build time
  via `-X main.commit=...` and `-X main.date=...`.

Dependencies: every `cmd/sin-websearch/*_cmd.go` file.
