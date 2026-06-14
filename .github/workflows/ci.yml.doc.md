# CI workflow

Purpose: Define the continuous integration pipeline for the `sin-websearch` project.

Docs: .github/workflows/ci.yml

## What it does

Runs build, test, and lint checks on every push or pull request to `main`.

## Jobs

- `test` — Builds and tests the project across multiple Go versions and operating systems.
  - OS matrix: `ubuntu-latest`, `macos-latest`, `windows-latest`
  - Go matrix: `1.25`, `1.26`
  - Steps per job: `go mod verify`, `go build ./cmd/sin-websearch`, `go test ./...`, `go vet ./...`
  - `fail-fast: false` ensures every OS/Go combination finishes even if another fails.

- `install` — Validates the `install.sh` release installer on `ubuntu-latest` and `macos-latest`.
  - Installs the latest binary into a temporary directory.
  - Verifies the installed binary responds to `sin-websearch --help`.

- `lint` — Runs `golangci-lint` on `ubuntu-latest` with Go 1.26.
  - Preserved from the original workflow.

## Triggers

- `push` to `main`
- `pull_request` targeting `main`

## Important values

- Lint timeout: `5m`
- Default `fail-fast` is disabled for matrix jobs to surface all platform-specific failures.
- `install.sh` requires `bash` and is therefore not tested on Windows runners.

## Caveats

- Windows tests run in the matrix; if they become flaky or require platform-specific fixes, consider a conditional skip rather than removing Windows from the matrix.
- The `install` job downloads the latest GitHub release, so it may fail if no release exists for the current platform or the release API is rate-limited.
