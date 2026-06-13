# infisical.go

Load secrets from Infisical CLI or environment variables.

- Imports: `encoding/json`, `fmt`, `os`, `os/exec`, `strings`
- Used by: `cmd/sin-websearch/secrets_cmd.go`, `internal/config`
- Loads API keys (SerpAPI, Brave, OpenRouter, etc.) and optionally sets them in the process environment.
- Redacts values in JSON export; never logs raw secrets.
