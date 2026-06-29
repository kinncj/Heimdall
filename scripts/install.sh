#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Install Heimdall binaries from GitHub Releases — install only what each machine
# needs.
#
#   curl -fsSL .../scripts/install.sh | sh -s -- dashboard
#   curl -fsSL .../scripts/install.sh | sh -s -- daemon helper
#   curl -fsSL .../scripts/install.sh | sh -s -- --install-location ~/.local/bin daemon
#
# Binaries install to the system bin dir by default (/usr/local/bin), elevating
# with sudo if needed. Override with --install-location <dir> or $HEIMDALL_BIN_DIR.
#
# Environment overrides:
#   HEIMDALL_VERSION   release tag to install (default: latest)
#   HEIMDALL_BIN_DIR   install directory (default: /usr/local/bin)
#   HEIMDALL_REPO      source owner/repo (default: kinncj/Heimdall)
set -eu

REPO="${HEIMDALL_REPO:-kinncj/Heimdall}"
VERSION="${HEIMDALL_VERSION:-latest}"

err() { echo "install: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || err "missing required tool: $1"; }

usage() {
  cat <<EOF
Install Heimdall binaries.

Usage: install.sh [--install-location DIR] [COMPONENT...]

Components: hub  dashboard  daemon  helper   (default: dashboard)

  --install-location DIR   install into DIR (default: /usr/local/bin)
  -h, --help               show this help
EOF
}

# --- parse args: flags + component list ------------------------------------
INSTALL_LOCATION=""
COMPONENTS=""
while [ $# -gt 0 ]; do
  case "$1" in
    --install-location) INSTALL_LOCATION="${2:-}"; shift 2 ;;
    --install-location=*) INSTALL_LOCATION="${1#*=}"; shift ;;
    -h|--help) usage; exit 0 ;;
    --*) err "unknown option: $1" ;;
    *) COMPONENTS="${COMPONENTS} $1"; shift ;;
  esac
done
COMPONENTS="${COMPONENTS# }"
[ -n "$COMPONENTS" ] || COMPONENTS="dashboard"

need curl
need uname

# --- detect platform -------------------------------------------------------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux) os=linux ;;
  darwin) os=darwin ;;
  *) err "unsupported OS: $os (Windows users: use scripts/install.ps1)" ;;
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

# --- choose an install dir (system bin by default) -------------------------
if [ -n "$INSTALL_LOCATION" ]; then
  bindir="$INSTALL_LOCATION"
elif [ -n "${HEIMDALL_BIN_DIR:-}" ]; then
  bindir="$HEIMDALL_BIN_DIR"
else
  bindir="/usr/local/bin"
fi

# sudo is used only when the target dir is not writable by this user.
SUDO=""
if [ ! -d "$bindir" ]; then
  mkdir -p "$bindir" 2>/dev/null || { command -v sudo >/dev/null 2>&1 && SUDO="sudo" && $SUDO mkdir -p "$bindir"; } \
    || err "cannot create ${bindir} (pass --install-location to a writable dir)"
fi
if [ -z "$SUDO" ] && [ ! -w "$bindir" ]; then
  command -v sudo >/dev/null 2>&1 && SUDO="sudo" || err "${bindir} is not writable and sudo is unavailable (use --install-location)"
fi

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
  case "$c" in hub|dashboard|daemon|helper|cli) ;; *) err "unknown component: $c (want hub|dashboard|daemon|helper|cli)" ;; esac
  asset="heimdall-${c}_${os}_${arch}"
  echo "downloading ${asset}"
  curl -fSL "${base}/${asset}" -o "${tmp}/${asset}" || err "download failed: ${asset}"
  verify "${tmp}/${asset}" "${asset}"
  chmod +x "${tmp}/${asset}"
  dest="${bindir}/heimdall-${c}"
  $SUDO mv "${tmp}/${asset}" "$dest" || err "failed to install ${dest}"
  echo "  installed ${dest}"
done

echo
echo "Done. Ensure ${bindir} is on your PATH."
echo "Try:  heimdall-${COMPONENTS%% *} --help"
