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

# --- install: /usr/local/bin (with sudo if TTY available), fallback to ~/.local/bin ---
_install_to() {
  local dest="$1"
  cp "$TMP" "${dest}.tmp" && chmod +x "${dest}.tmp" && mv "${dest}.tmp" "$dest"
}

INSTALL_PATH=""
if [ -w "/usr/local/bin" ]; then
  if _install_to "/usr/local/bin/$BINARY"; then
    INSTALL_PATH="/usr/local/bin/$BINARY"
  fi
elif [ -t 0 ] && command -v sudo >/dev/null 2>&1; then
  if sudo bash -c "cp '$TMP' '/usr/local/bin/${BINARY}.tmp' && chmod +x '/usr/local/bin/${BINARY}.tmp' && mv '/usr/local/bin/${BINARY}.tmp' '/usr/local/bin/$BINARY'"; then
    INSTALL_PATH="/usr/local/bin/$BINARY"
  fi
fi

if [ -z "$INSTALL_PATH" ]; then
  for USER_DIR in "$HOME/.local/bin" "$HOME/.ccx/bin" "$HOME/bin"; do
    mkdir -p "$USER_DIR" 2>/dev/null || continue
    if _install_to "$USER_DIR/$BINARY"; then
      INSTALL_PATH="$USER_DIR/$BINARY"
      if [[ ":$PATH:" != *":$USER_DIR:"* ]]; then
        echo ""
        echo "Note: $USER_DIR is not in your PATH."
        echo "Add the following line to your shell profile (~/.zshrc or ~/.bashrc):"
        printf "  export PATH=\"%s:\$PATH\"\n" "$USER_DIR"
        echo ""
      fi
      break
    fi
  done
fi

if [ -z "$INSTALL_PATH" ]; then
  echo "error: could not find a writable install directory" >&2
  exit 1
fi

echo "Installed: $INSTALL_PATH"

# --- initialize default settings ---
"$INSTALL_PATH" setting > /dev/null && echo "Default config: ~/.ccx/settings.json"

echo ""
echo "Run 'ccx doctor' to check your network connectivity."
