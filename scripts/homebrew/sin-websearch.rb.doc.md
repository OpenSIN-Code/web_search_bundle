# sin-websearch.rb

Homebrew formula that installs the official `sin-websearch` cross-platform release binary.

- Fetches the GitHub release binary matching the caller's platform: macOS amd64/arm64 or Linux amd64/arm64.
- Verifies the SHA256 checksum embedded in the formula before installation.
- Installs the binary as `sin-websearch`.
- Regenerated automatically by `scripts/homebrew/update_formula.sh` after each release.

Dependency map: `.github/workflows/release.yml` produces the binaries; `.github/workflows/homebrew.yml` triggers the updater; `scripts/homebrew/update_formula.sh` writes this file.
