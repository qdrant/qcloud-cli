#!/bin/sh
# Install qcloud, the Qdrant Cloud CLI, from GitHub Releases.
#
#   curl -fsSL https://raw.githubusercontent.com/qdrant/qcloud-cli/main/install.sh | sh
#
# For macOS and Linux on x86-64 or ARM64. On Windows, download the zip from
# https://github.com/qdrant/qcloud-cli/releases instead.
#
# Environment variables:
#   QCLOUD_VERSION      release to install, such as v0.32.0 (default: latest)
#   QCLOUD_INSTALL_DIR  directory for the binary (default: $HOME/.local/bin)
#
# The archive is verified against the release's checksums.txt before anything
# is installed. Needs curl or wget, sha256sum or shasum, and tar.
set -eu

repo="qdrant/qcloud-cli"
version="${QCLOUD_VERSION:-latest}"

fail() {
  echo "qcloud install: $*" >&2
  exit 1
}

# Everything runs from main, called on the last line, so a download of this
# script cut off midway runs nothing instead of half an install.
main() {
  case "$(uname -s)" in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    *) fail "unsupported OS $(uname -s). Download a release from https://github.com/$repo/releases" ;;
  esac

  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64 ;;
    arm64 | aarch64) arch=arm64 ;;
    *) fail "unsupported architecture $(uname -m). Download a release from https://github.com/$repo/releases" ;;
  esac
  # A shell under Rosetta reports x86_64 on Apple Silicon; install the native build.
  if [ "$os" = darwin ] && [ "$arch" = amd64 ] && [ "$(sysctl -n hw.optional.arm64 2>/dev/null)" = 1 ]; then
    arch=arm64
  fi

  case "$version" in
    latest) base="https://github.com/$repo/releases/latest/download" ;;
    v*) base="https://github.com/$repo/releases/download/$version" ;;
    *) base="https://github.com/$repo/releases/download/v$version" ;;
  esac

  if command -v curl >/dev/null 2>&1; then
    fetch() { curl -fsSL -o "$2" "$1"; }
  elif command -v wget >/dev/null 2>&1; then
    fetch() { wget -q -O "$2" "$1"; }
  else
    fail "needs curl or wget"
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    sha256() { sha256sum "$1" | cut -d ' ' -f 1; }
  elif command -v shasum >/dev/null 2>&1; then
    sha256() { shasum -a 256 "$1" | cut -d ' ' -f 1; }
  else
    fail "needs sha256sum or shasum to verify the download"
  fi

  archive="qcloud-$os-$arch.tar.gz"
  tmp=$(mktemp -d)
  # dash skips EXIT traps on signals, so turn them into an exit.
  trap 'rm -rf "$tmp"' EXIT
  trap 'exit 1' INT TERM

  echo "Downloading $archive ($version)"
  fetch "$base/$archive" "$tmp/$archive" || fail "could not download $base/$archive"
  fetch "$base/checksums.txt" "$tmp/checksums.txt" || fail "could not download $base/checksums.txt"

  expected=$(awk -v f="$archive" '$2 == f { print $1 }' "$tmp/checksums.txt")
  [ -n "$expected" ] || fail "$archive is not listed in checksums.txt"
  actual=$(sha256 "$tmp/$archive")
  [ "$expected" = "$actual" ] || fail "checksum mismatch for $archive: expected $expected, got $actual"

  tar -xzf "$tmp/$archive" -C "$tmp" qcloud
  install_dir="${QCLOUD_INSTALL_DIR:-$HOME/.local/bin}"
  # A quoted ~ from Dockerfile ENV or YAML env: isn't expanded by the shell.
  case "$install_dir" in "~/"*) install_dir="$HOME/${install_dir#\~/}" ;; esac
  mkdir -p "$install_dir"
  # Absolute, without a trailing slash, so the PATH check below matches.
  install_dir=$(CDPATH= cd -- "$install_dir" && pwd)
  # Copy next to the target, then rename over it: a running qcloud is replaced
  # atomically instead of failing with "text file busy".
  cp "$tmp/qcloud" "$install_dir/.qcloud.new"
  chmod 755 "$install_dir/.qcloud.new"
  mv -f "$install_dir/.qcloud.new" "$install_dir/qcloud"

  echo "Installed $("$install_dir/qcloud" version) to $install_dir/qcloud"

  case ":$PATH:" in
    *":$install_dir:"*) ;;
    *)
      echo "$install_dir is not on your PATH. Add it, for example:"
      echo "  export PATH=\"$install_dir:\$PATH\""
      ;;
  esac
}

main "$@"
