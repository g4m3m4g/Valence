#!/bin/sh
# Valence — one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/g4m3m4g/Valence/main/scripts/install.sh | sh
set -e

REPO="g4m3m4g/Valence"
BINARY="valence"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Resolve latest release tag
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$VERSION" ]; then
  echo "error: could not determine latest version — check your internet connection" >&2
  exit 1
fi

ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing valence ${VERSION} (${OS}/${ARCH})..."

# Download and extract into a temp dir, then move the binary
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP" "$BINARY"
chmod +x "$TMP/$BINARY"

# Install — try without sudo first, fall back with sudo
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "  $INSTALL_DIR is not writable, trying with sudo..."
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "valence ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo "Run: valence -h"
