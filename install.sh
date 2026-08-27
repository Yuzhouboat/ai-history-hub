#!/bin/sh
# Installs the latest claude-backup release for this OS/arch.
#
#   curl -fsSL https://raw.githubusercontent.com/Yuzhouboat/ai-history-hub/master/install.sh | sh
set -eu

REPO="Yuzhouboat/ai-history-hub"
BIN="claude-backup"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "error: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac
case "$os" in
  linux|darwin) ;;
  *)
    echo "error: unsupported OS: $os" >&2
    exit 1
    ;;
esac

version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
if [ -z "$version" ]; then
  echo "error: could not determine latest release version" >&2
  exit 1
fi

archive="${BIN}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${version}/${archive}"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

echo "Downloading ${BIN} ${version} for ${os}/${arch}..."
curl -fsSL "$url" -o "${tmpdir}/${archive}"
tar -xzf "${tmpdir}/${archive}" -C "$tmpdir" "$BIN"

install_dir="/usr/local/bin"
installed=0
if [ -w "$install_dir" ]; then
  mv "${tmpdir}/${BIN}" "${install_dir}/${BIN}"
  installed=1
elif command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
  sudo mv "${tmpdir}/${BIN}" "${install_dir}/${BIN}"
  installed=1
fi

if [ "$installed" -eq 0 ]; then
  install_dir="${HOME}/.local/bin"
  mkdir -p "$install_dir"
  mv "${tmpdir}/${BIN}" "${install_dir}/${BIN}"
  echo "note: installed to ${install_dir} — make sure it's on your PATH"
fi

chmod +x "${install_dir}/${BIN}"
echo "Installed ${BIN} to ${install_dir}/${BIN}"
"${install_dir}/${BIN}" --help >/dev/null 2>&1 || true
