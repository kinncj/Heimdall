#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Install Heimdall binaries from GitHub Releases.
#
#   curl -fsSL .../scripts/install.sh | sh -s -- dashboard
#   curl -fsSL .../scripts/install.sh | sh -s -- daemon helper
#
# Environment overrides:
#   HEIMDALL_VERSION   release tag to install (default: latest)
#   HEIMDALL_BIN_DIR   install directory (default: /usr/local/bin, else ~/.local/bin)
#   HEIMDALL_REPO      source owner/repo (default: kinncj/Heimdall)
set -eu

REPO="${HEIMDALL_REPO:-kinncj/Heimdall}"
VERSION="${HEIMDALL_VERSION:-latest}"
COMPONENTS="${*:-dashboard}"

err() { echo "install: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || err "missing required tool: $1"; }

need curl
need uname

# --- detect platform -------------------------------------------------------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux) os=linux ;;
  darwin) os=darwin ;;
  *) err "unsupported OS: $os (Windows users: download the .exe from Releases)" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) err "unsupported architecture: $arch" ;;
esac

# --- resolve version -------------------------------------------------------
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$VERSION" ] || err "could not resolve the latest release for ${REPO}"
fi
echo "install: ${REPO} ${VERSION} (${os}/${arch})"

base="https://github.com/${REPO}/releases/download/${VERSION}"

# --- choose an install dir -------------------------------------------------
if [ -n "${HEIMDALL_BIN_DIR:-}" ]; then
  bindir="$HEIMDALL_BIN_DIR"
elif [ -w /usr/local/bin ] 2>/dev/null; then
  bindir="/usr/local/bin"
else
  bindir="${HOME}/.local/bin"
fi
mkdir -p "$bindir"

# --- fetch checksums (best effort) -----------------------------------------
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
sums=""
if curl -fsSL "${base}/SHA256SUMS" -o "${tmp}/SHA256SUMS" 2>/dev/null; then
  sums="${tmp}/SHA256SUMS"
fi

verify() { # <file> <asset-name>
  [ -n "$sums" ] || { echo "  (no SHA256SUMS published; skipping verification)"; return 0; }
  want=$(grep " $2\$" "$sums" 2>/dev/null | awk '{print $1}' | head -1)
  [ -n "$want" ] || { echo "  (no checksum entry for $2; skipping)"; return 0; }
  if command -v sha256sum >/dev/null 2>&1; then
    got=$(sha256sum "$1" | awk '{print $1}')
  else
    got=$(shasum -a 256 "$1" | awk '{print $1}')
  fi
  [ "$want" = "$got" ] || err "checksum mismatch for $2"
  echo "  verified $2"
}

# --- install each requested component --------------------------------------
for c in $COMPONENTS; do
  asset="heimdall-${c}_${os}_${arch}"
  echo "downloading ${asset}"
  curl -fSL "${base}/${asset}" -o "${tmp}/${asset}" || err "download failed: ${asset}"
  verify "${tmp}/${asset}" "${asset}"
  chmod +x "${tmp}/${asset}"
  dest="${bindir}/heimdall-${c}"
  if mv "${tmp}/${asset}" "$dest" 2>/dev/null; then :; else
    echo "  (elevating to write ${bindir})"
    sudo mv "${tmp}/${asset}" "$dest"
  fi
  echo "  installed ${dest}"
done

echo
echo "Done. Make sure ${bindir} is on your PATH."
echo "Try:  heimdall-${COMPONENTS%% *} --help"
