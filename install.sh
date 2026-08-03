#!/bin/sh
# Sillage installer: detects OS/arch, downloads the latest binary, installs it.
# Usage: curl -fsSL https://raw.githubusercontent.com/Halleck45/sillage/main/install.sh | sh
set -eu

REPO="Halleck45/sillage"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$os" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $os (Linux and macOS only)"; exit 1 ;;
esac
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "Unsupported architecture: $arch"; exit 1 ;;
esac

url="https://github.com/$REPO/releases/latest/download/sillage_${os}_${arch}"

# Pick an install dir: /usr/local/bin if writable, else ~/.local/bin.
dest="/usr/local/bin"
if [ ! -w "$dest" ]; then
  dest="$HOME/.local/bin"
  mkdir -p "$dest"
fi

echo "Downloading sillage (${os}/${arch})..."
curl -fSL --progress-bar "$url" -o "$dest/sillage"
chmod +x "$dest/sillage"

echo ""
echo "Installed: $dest/sillage"
case ":$PATH:" in
  *":$dest:"*) echo "Run it with: sillage" ;;
  *) echo "Note: $dest is not in your PATH. Run it with: $dest/sillage" ;;
esac
