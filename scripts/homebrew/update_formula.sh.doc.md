# update_formula.sh

CI-friendly script that refreshes `scripts/homebrew/sin-websearch.rb` to match the latest GitHub release.

- Queries the GitHub releases API for the newest tag and version.
- Downloads the release binary for each supported platform (macOS amd64/arm64, Linux amd64/arm64).
- Computes SHA256 checksums locally and writes them into the formula.
- Replaces the formula file in-place with updated `version`, URLs, and `sha256` values.

Dependency map: consumes the release artifacts produced by `.github/workflows/release.yml`; invoked by `.github/workflows/homebrew.yml` after a release is published.

Usage:

```bash
./scripts/homebrew/update_formula.sh              # updates scripts/homebrew/sin-websearch.rb
./scripts/homebrew/update_formula.sh path/to.rb    # updates a custom formula file
```
