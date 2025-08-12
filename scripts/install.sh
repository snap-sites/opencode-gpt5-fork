#!/usr/bin/env bash
set -euo pipefail

# Set REPO=YOURUSER/opencode-gpt5-fork or edit the line below after you fork.
REPO="${REPO:-YOURUSER/opencode-gpt5-fork}"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "Unsupported arch $ARCH" ; exit 1 ;;
esac

TAG="${1:-latest}"
if [ "$TAG" = "latest" ]; then
  URL="https://github.com/$REPO/releases/latest/download/ocx_${OS}_${ARCH}.tar.gz"
else
  URL="https://github.com/$REPO/releases/download/$TAG/ocx_${OS}_${ARCH}.tar.gz"
fi

TMP=$(mktemp -d)
curl -fsSL "$URL" -o "$TMP/ocx.tgz"
tar -xzf "$TMP/ocx.tgz" -C "$TMP"
chmod +x "$TMP/ocx"

# Prefer /usr/local/bin if writable, else ~/.local/bin
DEST="/usr/local/bin"
if [ ! -w "$DEST" ]; then
  DEST="$HOME/.local/bin"
  mkdir -p "$DEST"
fi

mv "$TMP/ocx" "$DEST/ocx"
echo "Installed ocx -> $DEST/ocx"
echo "Ensure $DEST is in your PATH."
