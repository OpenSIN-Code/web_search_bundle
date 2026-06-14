#!/usr/bin/env bash
# Purpose: install or update the latest sin-websearch binary from GitHub Releases.
# Docs: install.sh.doc.md
set -euo pipefail

REPO="OpenSIN-Code/web_search_bundle"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

# Default install location; override with INSTALL_DIR or --dir <path>.
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

usage() {
    cat <<EOF
Usage: install.sh [OPTIONS]

Install the latest sin-websearch binary from GitHub Releases.

Options:
  -d, --dir DIR     Install directory (default: \$HOME/.local/bin)
  -v, --version TAG Install a specific version tag instead of latest
  -h, --help        Show this help

Environment:
  INSTALL_DIR        Same as --dir
  GITHUB_TOKEN       Optional GitHub token to avoid rate limits
EOF
}

VERSION=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

# Detect OS and architecture.
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        CYGWIN*|MSYS*|MINGW*) echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)            echo "unknown" ;;
    esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [[ "$OS" == "unknown" || "$ARCH" == "unknown" ]]; then
    echo "Unsupported platform: ${OS}/${ARCH}" >&2
    exit 1
fi

EXT=""
if [[ "$OS" == "windows" ]]; then
    EXT=".exe"
fi

BINARY_NAME="sin-websearch-${OS}-${ARCH}${EXT}"

# Fetch the latest release metadata.
if [[ -z "$VERSION" ]]; then
    if command -v curl >/dev/null 2>&1; then
        RELEASE_JSON=$(curl -fsSL ${GITHUB_TOKEN:+-H "Authorization: token ${GITHUB_TOKEN}"} "$API_URL")
    elif command -v wget >/dev/null 2>&1; then
        RELEASE_JSON=$(wget -qO- ${GITHUB_TOKEN:+--header="Authorization: token ${GITHUB_TOKEN}"} "$API_URL")
    else
        echo "curl or wget is required" >&2
        exit 1
    fi
    VERSION=$(echo "$RELEASE_JSON" | grep -oE '"tag_name": *"[^"]+"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ -z "$VERSION" ]]; then
        echo "Could not determine latest release" >&2
        exit 1
    fi
fi

BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
DOWNLOAD_URL="${BASE_URL}/${BINARY_NAME}"
CHECKSUM_URL="${BASE_URL}/checksums.txt"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Installing sin-websearch ${VERSION} for ${OS}/${ARCH}..."

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${BINARY_NAME}"
    curl -fsSL "$CHECKSUM_URL" -o "${TMP_DIR}/checksums.txt"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "${TMP_DIR}/${BINARY_NAME}"
    wget -q "$CHECKSUM_URL" -O "${TMP_DIR}/checksums.txt"
else
    echo "curl or wget is required" >&2
    exit 1
fi

# Verify checksum for the downloaded binary if sha256sum is available.
if command -v sha256sum >/dev/null 2>&1; then
    EXPECTED=$(grep "  ${BINARY_NAME}$" "${TMP_DIR}/checksums.txt" | awk '{print $1}')
    ACTUAL=$(sha256sum "${TMP_DIR}/${BINARY_NAME}" | awk '{print $1}')
    if [[ -z "$EXPECTED" ]]; then
        echo "Could not find checksum for ${BINARY_NAME}" >&2
        exit 1
    fi
    if [[ "$EXPECTED" != "$ACTUAL" ]]; then
        echo "Checksum verification failed for ${BINARY_NAME}" >&2
        exit 1
    fi
else
    echo "sha256sum not available; skipping checksum verification" >&2
fi

# Install.
mkdir -p "$INSTALL_DIR"
TARGET="${INSTALL_DIR}/sin-websearch${EXT}"
cp "${TMP_DIR}/${BINARY_NAME}" "$TARGET"
chmod +x "$TARGET"

echo "sin-websearch ${VERSION} installed to ${TARGET}"

if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* && "$INSTALL_DIR" != "/usr/local/bin" && "$INSTALL_DIR" != "/usr/bin" ]]; then
    echo "Add ${INSTALL_DIR} to your PATH to use 'sin-websearch' directly:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
