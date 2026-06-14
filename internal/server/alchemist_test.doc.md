# alchemist_test.go

Unit tests for the alchemist and swarm HTTP handlers in `alchemist.go`.

## What it tests

- Rejects non-POST methods for the swarm endpoint.
- Rejects malformed JSON for both `/api/v1/alchemist` and `/api/v1/alchemist/swarm`.
- Returns 400 when required configuration is missing or invalid:
  missing `run_cmd`, invalid `budget`, invalid `runtime`, and invalid `safety`.
- `buildAlchemistConfig` default values: repo path, safety mode, budget, and runtime.
- `buildAlchemistConfig` validation errors for invalid runtime.

## Dependencies

- Reuses the `setupGitRepo` helper from `http_test.go` for valid repo paths.
- Does not run the real daemon for invalid-input tests.

## Run

```bash
go test ./internal/server -run 'TestHandleAlchemist|TestBuildAlchemist'
```
