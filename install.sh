#!/usr/bin/env bash
set -e

# Kybernis Audit One-Line Installer
# Usage: curl -sSL https://kybernis.com/install-audit.sh | bash

REPO="kybernis/kybernis-audit"
BINARY="kybernis-audit"
INSTALL_DIR="/usr/local/bin"

echo "🛡️ Installing Kybernis Audit..."

# Determine OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# For now, we simulate the download since this is a local build for the MVP.
# In production, this would fetch from GitHub Releases:
# DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/${BINARY}_${OS}_${ARCH}.tar.gz"

echo "Detected OS: $OS, Arch: $ARCH"
echo "Building from source for local installation..."

# Ensure Go is installed for the source build
if ! command -v go &> /dev/null; then
    echo "❌ Error: 'go' is required to build kybernis-audit locally."
    exit 1
fi

go build -o $BINARY ./cmd/kybernis-audit

echo "Moving binary to $INSTALL_DIR (requires sudo)..."
sudo mv $BINARY $INSTALL_DIR/

echo ""
echo "✅ Kybernis Audit installed successfully!"
echo "Run 'kybernis-audit --help' to get started."
