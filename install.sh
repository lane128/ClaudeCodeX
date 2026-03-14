#!/usr/bin/env bash
set -euo pipefail

REPO="lane128/ClaudeCodeX"
BINARY="ccx"

# --- detect OS ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin|linux) ;;
  *) echo "error: unsupported OS: $OS" >&2; exit 1 ;;
esac

# --- detect arch ---
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "error: unsupported arch: $ARCH" >&2; exit 1 ;;
esac

# --- download to temp file ---
echo "Downloading $BINARY ($OS/$ARCH)..."
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
curl -fsSL "https://github.com/$REPO/releases/latest/download/${BINARY}_${OS}_${ARCH}" -o "$TMP"
chmod +x "$TMP"

# --- install: /usr/local/bin with sudo fallback to ~/.local/bin ---
if [ -w "/usr/local/bin" ]; then
  mv "$TMP" "/usr/local/bin/$BINARY"
  INSTALL_PATH="/usr/local/bin/$BINARY"
elif command -v sudo >/dev/null 2>&1 && sudo mv "$TMP" "/usr/local/bin/$BINARY"; then
  INSTALL_PATH="/usr/local/bin/$BINARY"
else
  mkdir -p "$HOME/.local/bin"
  mv "$TMP" "$HOME/.local/bin/$BINARY"
  INSTALL_PATH="$HOME/.local/bin/$BINARY"
  if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo ""
    echo "Note: $HOME/.local/bin is not in your PATH."
    echo "Add the following line to your shell profile (~/.zshrc or ~/.bashrc):"
    printf '  export PATH="$HOME/.local/bin:$PATH"\n'
    echo ""
  fi
fi

echo "Installed: $INSTALL_PATH"

# --- initialize default settings ---
"$INSTALL_PATH" setting > /dev/null && echo "Default config: ~/.ccx/settings.json"

echo ""
echo "Run 'ccx doctor' to check your network connectivity."
