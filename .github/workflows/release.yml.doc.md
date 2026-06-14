What it does: Builds cross-platform `sin-websearch` binaries and creates a GitHub Release with artifacts when a version tag is pushed.

Dependencies: Triggered by Git tags matching `v*` (excluding `v*-ceo-audit-v*` bulk-deploy markers). Uses `actions/setup-go`, `softprops/action-gh-release`, and the built-in `GITHUB_TOKEN`.

Important config:
- Builds for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`.
- Generates `sha256sum` checksums in `dist/checksums.txt`.
- Uses `-ldflags="-s -w"` to strip debug info and reduce binary size.
- `generate_release_notes: true` creates release notes from merged PRs.

Caveats:
- Does not sign artifacts or SBOMs. If supply-chain signing is needed, add a signing step.
- Does not publish to package managers (Homebrew, apt, etc.).
