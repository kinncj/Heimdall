#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Cross-compile the Heimdall binaries for all supported platforms into dist/.
# Used by `make release` and the release GitHub Actions workflow.
#
# Linux and Windows are built CGO-free so they cross-compile cleanly and run with
# no shared-library dependencies. macOS is built with CGO=1 so the IOReport/SMC
# no-sudo power+GPU path is compiled in — that only cross-compiles from a Mac, so
# a darwin target is refused (loudly) when this script runs off a Mac. In CI the
# darwin binaries come from the dedicated macOS runner (see release.yml), and this
# script is invoked there with PLATFORMS limited to linux + windows.
set -euo pipefail

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
OUT="${OUT:-dist}"
PLATFORMS="${PLATFORMS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64}"
COMPONENTS="${COMPONENTS:-dashboard daemon hub helper cli}"

rm -rf "$OUT"
mkdir -p "$OUT"

HOST_OS="$(uname -s)"

for c in $COMPONENTS; do
  for p in $PLATFORMS; do
    os="${p%/*}"; arch="${p#*/}"
    ext=""; [ "$os" = "windows" ] && ext=".exe"
    bin="heimdall-${c}_${os}_${arch}${ext}"
    if [ "$os" = "darwin" ]; then
      # macOS needs CGO for the IOReport/SMC no-sudo power+GPU path. A CGO-free
      # darwin binary silently loses that, so refuse to build one off a Mac.
      if [ "$HOST_OS" != "Darwin" ]; then
        echo "release: SKIP ${bin} — darwin needs CGO; build on macOS (CI uses a mac runner)" >&2
        continue
      fi
      echo "building ${bin} (CGO)"
      CGO_ENABLED=1 GOOS="$os" GOARCH="$arch" CC="clang -arch ${arch/amd64/x86_64}" \
        go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
        -o "$OUT/$bin" "./app/cmd/${c}"
    else
      echo "building ${bin}"
      CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
        go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
        -o "$OUT/$bin" "./app/cmd/${c}"
    fi
  done
done

# Manpages (.1) + Windows/no-man plain-text help (.txt), generated from --help and
# dropped next to the binaries so they are attached and checksummed below.
OUT="$OUT" bash scripts/gen-manpages.sh || echo "release: manpage generation skipped"

( cd "$OUT" && { sha256sum heimdall-* 2>/dev/null || shasum -a 256 heimdall-*; } > SHA256SUMS )

echo "release artifacts (version ${VERSION}) in ${OUT}/:"
ls -1 "$OUT"
