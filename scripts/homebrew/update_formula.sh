#!/usr/bin/env bash
# Purpose: Fetch the latest sin-websearch release and regenerate the Homebrew formula.
# Docs: scripts/homebrew/update_formula.sh.doc.md
set -euo pipefail

REPO="OpenSIN-Code/web_search_bundle"
FORMULA="${1:-scripts/homebrew/sin-websearch.rb}"

# Fetch the latest release metadata from GitHub.
RELEASE_JSON=$(curl -sL --retry 3 --retry-delay 2 "https://api.github.com/repos/${REPO}/releases/latest")
TAG=$(echo "${RELEASE_JSON}" | jq -r '.tag_name')
VERSION="${TAG#v}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

# Compute SHA256 with the tool available on the current OS.
sha256() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${file}" | awk '{print $1}'
  else
    shasum -a 256 "${file}" | awk '{print $1}'
  fi
}

# Download a release asset and return its SHA256 checksum.
download_sha() {
  local asset="$1"
  local url="https://github.com/${REPO}/releases/download/${TAG}/${asset}"
  local dest="${TMPDIR}/${asset}"
  curl -sL --retry 3 --retry-delay 2 -o "${dest}" "${url}"
  sha256 "${dest}"
}

SHA_DARWIN_AMD64=$(download_sha "sin-websearch-darwin-amd64")
SHA_DARWIN_ARM64=$(download_sha "sin-websearch-darwin-arm64")
SHA_LINUX_AMD64=$(download_sha "sin-websearch-linux-amd64")
SHA_LINUX_ARM64=$(download_sha "sin-websearch-linux-arm64")

mkdir -p "$(dirname "${FORMULA}")"

cat > "${FORMULA}" <<EOF
# Purpose: Homebrew formula for sin-websearch release binaries.
# Docs: scripts/homebrew/sin-websearch.rb.doc.md
class SinWebsearch < Formula
  desc "Unified Intelligence Gateway for OpenSIN"
  homepage "https://github.com/${REPO}"
  version "${VERSION}"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/sin-websearch-darwin-amd64"
      sha256 "${SHA_DARWIN_AMD64}"
    end
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/sin-websearch-darwin-arm64"
      sha256 "${SHA_DARWIN_ARM64}"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/sin-websearch-linux-amd64"
      sha256 "${SHA_LINUX_AMD64}"
    end
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/sin-websearch-linux-arm64"
      sha256 "${SHA_LINUX_ARM64}"
    end
  end

  def install
    bin.install Dir["sin-websearch-*"].first => "sin-websearch"
    chmod 0755, bin/"sin-websearch"
  end

  test do
    system "#{bin}/sin-websearch", "--version"
  end
end
EOF

echo "Updated ${FORMULA} to ${TAG}"
