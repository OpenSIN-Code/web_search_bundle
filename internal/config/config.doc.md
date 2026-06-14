# config.go

Loads and merges application configuration for sin-websearch.

## Related files
- `config_test.go` / `config_bench_test.go` — tests.
- `cmd/sin-websearch/*.go` — CLI commands consume the config.
- `internal/secrets/infisical.go` — alternative secret loading path.

## Important details
- Uses Viper to read `sin-websearch.yaml` from `~/.config/sin-websearch`, `~/.sin-websearch`, or the working directory.
- Merges config with env vars prefixed by `SIN_WEBSEARCH`.
- Backwards-compatible env vars: `SERPAPI_KEY`, `BRAVE_API_KEY`, `OPENROUTER_API_KEY`, etc.
- Defaults: HTTP port 8787, MCP port 8788.

## Caveats
- `MustLoad` panics on error; CLI code should handle that.
- The custom `itoa` avoids `strconv` for the env-key loading hot path.
