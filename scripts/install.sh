#!/bin/sh
set -e

# Bunkr installer
# Usage: curl -fsSL https://raw.githubusercontent.com/pankajbeniwal/bunkr/main/scripts/install.sh | sh

REPO="pankajbeniwal/bunkr"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/bunkr"

main() {
    # Check for Linux
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    if [ "$OS" != "linux" ]; then
        echo "Error: bunkr only supports Linux (got: $OS)"
        exit 1
    fi

    # Detect architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo "Error: unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    echo "Detected: ${OS}/${ARCH}"

    # Get latest version
    echo "Fetching latest release..."
    LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST" ]; then
        echo "Error: could not determine latest version"
        exit 1
    fi
    echo "Latest version: ${LATEST}"

    # Download binary
    ASSET_NAME="bunkr_${OS}_${ARCH}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET_NAME}"

    echo "Downloading ${DOWNLOAD_URL}..."
    TMP=$(mktemp)
    curl -fsSL -o "$TMP" "$DOWNLOAD_URL"

    # Install
    chmod +x "$TMP"
    mv "$TMP" "${INSTALL_DIR}/bunkr"

    # Create config directory
    mkdir -p "$CONFIG_DIR"

    echo ""
    echo "bunkr ${LATEST} installed to ${INSTALL_DIR}/bunkr"
    echo ""
    echo "Get started:"
    echo "  bunkr init                    # Harden a server"
    echo "  bunkr install uptime-kuma     # Install an app"
    echo "  bunkr list                    # See available recipes"
    echo ""
}

main
