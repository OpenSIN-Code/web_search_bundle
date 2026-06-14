# Security Policy

## Supported Versions

Only the latest released version of `sin-websearch` receives security updates.

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| < 0.2.0 | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately:

- Open a draft security advisory at https://github.com/OpenSIN-Code/web_search_bundle/security/advisories/new
- Or email the maintainers at security@opensin-code.org

Please do **not** open public issues for security bugs.

We aim to respond within 5 business days and will release a fix as soon as possible.

## Security Best Practices

### HTTP API Authentication

The HTTP server supports optional bearer-token authentication. When a token is configured, all endpoints require an `Authorization: Bearer <token>` header.

Set the token via:

```yaml
# ~/.config/sin-websearch/sin-websearch.yaml
token: "your-random-secret-token"
```

or via environment variable:

```bash
export SIN_WEBSEARCH_TOKEN="your-random-secret-token"
```

Without a token, the server runs open (default for local development). Always configure a token before exposing the server to any untrusted network.

### Secret Management

- Never commit API keys to the repository.
- Use the Infisical CLI integration or environment variables.
- The `ExportAsJSON()` method redacts secret values before serialization.

### Alchemist Safety

The autonomous research loop can modify files and run arbitrary commands. Always use `--safety headless` in untrusted environments, and never expose the HTTP `/alchemist` or `/alchemist/swarm` endpoints without authentication.

## Security Scanning

The CI pipeline runs:

- `go vet ./...`
- `golangci-lint run ./...`
- `gosec ./...`
- `govulncheck ./...`

Run these locally before pushing:

```bash
go vet ./...
golangci-lint run ./... --timeout=5m
gosec ./...
govulncheck ./...
```
