#!/bin/bash
set -euo pipefail

VERSION=${VERSION:-"latest"}
REPO="connordoman/doman"

# Determine install directory (prefer $HOME/go/bin to avoid sudo)
if [ -z "${INSTALL_DIR:-}" ]; then
  GO_BIN_DIR="$HOME/go/bin"
  if [ -w "$HOME" ] && (mkdir -p "$GO_BIN_DIR" 2>/dev/null || [ -w "$GO_BIN_DIR" ]); then
    INSTALL_DIR="$GO_BIN_DIR"
  else
    INSTALL_DIR="/usr/local/bin"
  fi
fi

# Check required commands
command -v curl >/dev/null 2>&1 || { echo "Error: curl is required but not installed. Aborting." >&2; exit 1; }
command -v tar >/dev/null 2>&1 || { echo "Error: tar is required but not installed. Aborting." >&2; exit 1; }

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
  x86_64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Error: Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case $OS in
  darwin) OS_NAME="Darwin" ;;
  linux) OS_NAME="Linux" ;;
  *) echo "Error: Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Fetch latest version if needed
if [ "$VERSION" = "latest" ]; then
  echo "Fetching latest version..."
  VERSION=$(curl -sf "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || {
    echo "Error: Failed to fetch latest version from GitHub API" >&2
    exit 1
  })
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version" >&2
    exit 1
  fi
fi

VERSION=${VERSION#v}
FILE_NAME="doman_${OS_NAME}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILE_NAME}"

# Check if already installed
if command -v doman >/dev/null 2>&1; then
  CURRENT_VERSION=$(doman version | awk '{print $2}')
  echo "doman is already installed (version: $CURRENT_VERSION)"
  echo "Installing version v${VERSION}..."
else
  echo "Installing doman v${VERSION}..."
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download release archive
echo "Downloading from $DOWNLOAD_URL..."
if ! curl -sfL -o "$TMP_DIR/$FILE_NAME" "$DOWNLOAD_URL"; then
  echo "Error: Failed to download release archive" >&2
  echo "URL: $DOWNLOAD_URL" >&2
  exit 1
fi

# Extract archive
cd "$TMP_DIR"
if ! tar -xzf "$FILE_NAME"; then
  echo "Error: Failed to extract archive" >&2
  exit 1
fi

# Find binary
BINARY=$(find . -name "doman" -type f | head -n 1)
if [ -z "$BINARY" ] || [ ! -f "$BINARY" ]; then
  echo "Error: Could not find doman binary in archive" >&2
  exit 1
fi

chmod +x "$BINARY"

# Ensure install directory exists
if [ ! -d "$INSTALL_DIR" ]; then
  if [ ! -w "$(dirname "$INSTALL_DIR")" ]; then
    echo "Creating directory $INSTALL_DIR (requires sudo)..."
    sudo mkdir -p "$INSTALL_DIR"
  else
    mkdir -p "$INSTALL_DIR"
  fi
fi

# Install binary
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$BINARY" "$INSTALL_DIR/doman"
else
  echo "Installing to $INSTALL_DIR..."
  mv "$BINARY" "$INSTALL_DIR/doman"
fi

echo "✓ Successfully installed doman v${VERSION}"

# Verify installation
if command -v doman >/dev/null 2>&1; then
  doman version
else
  echo "Warning: doman was installed but is not in PATH" >&2
  echo "You may need to add $INSTALL_DIR to your PATH" >&2
  if [ "$INSTALL_DIR" = "$HOME/go/bin" ]; then
    echo "Run: export PATH=\"\$PATH:\$HOME/go/bin\"" >&2
    
    profile_path="$HOME/.profile"
    if [ -n "$ZSH_VERSION" ] || [ "$(basename "$SHELL")" = "zsh" ]; then
      profile_path="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ] || [ "$(basename "$SHELL")" = "bash" ]; then
      profile_path="$HOME/.bashrc"
    elif [ "$(basename "$SHELL")" = "fish" ]; then
      profile_path="$HOME/.config/fish/config.fish"
    else
      echo "Warning: could not determine profile path for shell: $SHELL" >&2
      echo "Please edit your profile manually to add $INSTALL_DIR to your PATH" >&2
      echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >&2
      return 0
    fi

    echo "Edit $profile_path to add $INSTALL_DIR to your PATH:" >&2
    echo "# doman installation path" >&2
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >&2
  fi
  "$INSTALL_DIR/doman" version
fi
