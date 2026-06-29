#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Cross-compile the Heimdall binaries for all supported platforms into dist/.
# Used by `make release` and the release GitHub Actions workflow.
#
# Release binaries are built CGO-free so they cross-compile cleanly and run with
# no shared-library dependencies. On macOS this means power/GPU fall back to
# `powermetrics` (sudo) rather than IOReport — build locally for the no-sudo path.
set -euo pipefail

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
OUT="${OUT:-dist}"
PLATFORMS="${PLATFORMS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64}"
COMPONENTS="${COMPONENTS:-dashboard daemon hub helper cli}"

rm -rf "$OUT"
mkdir -p "$OUT"

for c in $COMPONENTS; do
  for p in $PLATFORMS; do
    os="${p%/*}"; arch="${p#*/}"
    ext=""; [ "$os" = "windows" ] && ext=".exe"
    bin="heimdall-${c}_${os}_${arch}${ext}"
    echo "building ${bin}"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
      go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
      -o "$OUT/$bin" "./app/cmd/${c}"
  done
done

( cd "$OUT" && { sha256sum heimdall-* 2>/dev/null || shasum -a 256 heimdall-*; } > SHA256SUMS )

echo "release artifacts (version ${VERSION}) in ${OUT}/:"
ls -1 "$OUT"
