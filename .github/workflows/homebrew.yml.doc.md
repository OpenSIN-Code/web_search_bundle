# homebrew.yml

GitHub Actions workflow that keeps the Homebrew formula synchronized with the latest release.

- Triggers on `release: published`.
- Checks out the `main` branch.
- Runs `scripts/homebrew/update_formula.sh` to refresh version, URLs, and SHA256 checksums.
- Commits and pushes the updated formula directly to `main` using the built-in `GITHUB_TOKEN`.

Dependency map: calls `scripts/homebrew/update_formula.sh`, which writes `scripts/homebrew/sin-websearch.rb`.
