#!/bin/bash
set -euo pipefail

REPO="willvelida/code-minions"
BINARY="code-minions"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)"
if [ -z "$VERSION" ]; then
  echo "Error: Could not determine latest version"
  exit 1
fi

echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."

# Download and extract
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"

# Verify checksum
verify_checksum() {
  local expected="$1" actual="$2"
  if [ "$actual" != "$expected" ]; then
    echo "Checksum verification failed."
    echo "  Expected: $expected"
    echo "  Got:      $actual"
    exit 1
  fi
  echo "Checksum verified."
}

CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
if curl -fsSL "$CHECKSUM_URL" -o "${TMPDIR}/checksums.txt" 2>/dev/null; then
  EXPECTED_HASH="$(awk -v file="${ARCHIVE}" '$2==file {print $1}' "${TMPDIR}/checksums.txt")"
  if [ -z "$EXPECTED_HASH" ]; then
    echo "Warning: No matching checksum entry found for ${ARCHIVE} - skipping verification"
  elif command -v sha256sum >/dev/null 2>&1; then
    verify_checksum "$EXPECTED_HASH" "$(sha256sum "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    verify_checksum "$EXPECTED_HASH" "$(shasum -a 256 "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
  else
    echo "Warning: Neither sha256sum nor shasum found - skipping checksum verification"
  fi
else
  echo "Warning: Could not download checksums - skipping verification"
fi

tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

# Install
install -d "$INSTALL_DIR"
install "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
echo "Run 'code-minions --help' to get started"