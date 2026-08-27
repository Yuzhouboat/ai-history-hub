#!/bin/sh
# Installs the latest claude-backup release for this OS/arch, plus rclone
# (claude-backup shells out to it for the actual transfer) if it's not
# already on PATH.
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
  linux) rclone_os="linux" ;;
  darwin) rclone_os="osx" ;;
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

# Resolve once where binaries go and whether sudo is usable, so installing
# rclone below doesn't re-prompt or second-guess the choice made for
# claude-backup.
install_dir="/usr/local/bin"
if [ -w "$install_dir" ]; then
  install_mode="direct"
elif command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
  install_mode="sudo"
else
  install_dir="${HOME}/.local/bin"
  mkdir -p "$install_dir"
  install_mode="local"
  echo "note: installing to ${install_dir} — make sure it's on your PATH"
fi

# place_binary <src_path> <dest_name>
place_binary() {
  dest="${install_dir}/$2"
  if [ "$install_mode" = "sudo" ]; then
    sudo mv "$1" "$dest"
  else
    mv "$1" "$dest"
  fi
  chmod +x "$dest"
}

place_binary "${tmpdir}/${BIN}" "${BIN}"
echo "Installed ${BIN} to ${install_dir}/${BIN}"
"${install_dir}/${BIN}" --help >/dev/null 2>&1 || true

if command -v rclone >/dev/null 2>&1; then
  echo "rclone already installed ($(command -v rclone)); leaving it as-is"
elif ! command -v unzip >/dev/null 2>&1; then
  echo "note: rclone is required but 'unzip' is not on PATH to extract it; install rclone yourself from https://rclone.org/install/" >&2
else
  rclone_zip="rclone-current-${rclone_os}-${arch}.zip"
  rclone_url="https://downloads.rclone.org/${rclone_zip}"

  echo "Downloading rclone for ${rclone_os}/${arch}..."
  curl -fsSL "$rclone_url" -o "${tmpdir}/${rclone_zip}"
  unzip -q "${tmpdir}/${rclone_zip}" -d "${tmpdir}/rclone_extract"

  rclone_bin=$(find "${tmpdir}/rclone_extract" -type f -name rclone)
  if [ -z "$rclone_bin" ]; then
    echo "note: downloaded rclone archive but couldn't find the binary inside it; install rclone yourself from https://rclone.org/install/" >&2
  else
    place_binary "$rclone_bin" rclone
    echo "Installed rclone to ${install_dir}/rclone"
  fi
fi
