What it does: Downloads and installs the latest `sin-websearch` release binary from GitHub Releases for the current OS/architecture.

Dependencies: Requires `curl` or `wget`, and optionally `sha256sum` for checksum verification. Supports Linux, macOS, and Windows.

Important config values & limits:
- Default install directory: `~/.local/bin` (override with `INSTALL_DIR` or `--dir`).
- Downloads `sin-websearch-<os>-<arch>` from the latest GitHub release.
- Verifies SHA256 checksums from `checksums.txt` when `sha256sum` is available.

Usage examples:
```bash
bash install.sh                  # latest to ~/.local/bin
bash install.sh --dir /usr/local/bin --version v0.2.9
```

Caveats:
- Does not build from source; requires a published GitHub release.
- GitHub API rate limits apply without `GITHUB_TOKEN`.
- Does not add the install directory to `PATH` automatically.
