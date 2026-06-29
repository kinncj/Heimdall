#!/usr/bin/env bash
# Generate a roff manpage (.1) and a Windows / no-man plain-text help (.txt) for
# every Heimdall binary, straight from each binary's own --help output.
#
# Creative + low-maintenance: the manpage tracks the real `--help`, so there is no
# second source of truth to drift. Unix systems get the `.1`; Windows (which has no
# man) ships the `.txt` next to the binary.
#
# Output goes next to the release binaries so the release workflow's existing
# `dist/heimdall-*` glob attaches and checksums them with no extra wiring.
#
# Usage:
#   bash scripts/gen-manpages.sh           # -> dist/heimdall-*.1 and .txt
#   OUT=/tmp/man bash scripts/gen-manpages.sh
set -euo pipefail

OUT="${OUT:-dist}"
DATE="${MAN_DATE:-$(date +%Y-%m-%d)}"
COMPONENTS="${COMPONENTS:-dashboard daemon hub helper cli}"
mkdir -p "$OUT"

desc() {
  case "$1" in
    dashboard) echo "real-time terminal dashboard for a Heimdall fleet" ;;
    daemon)    echo "unprivileged per-host metric collector that streams to a hub" ;;
    hub)       echo "central gRPC server aggregating daemons and fanning out to dashboards" ;;
    helper)    echo "optional root sidecar exposing privileged power/GPU/thermal metrics" ;;
    cli)       echo "machine- and AI-friendly JSON client for a Heimdall hub" ;;
    *)         echo "Heimdall $1" ;;
  esac
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

for c in $COMPONENTS; do
  bin="heimdall-${c}"

  # Find a runnable binary: one already on PATH, else build a throwaway.
  run="$(command -v "$bin" 2>/dev/null || true)"
  if [ -z "$run" ]; then
    run="${tmp}/${bin}"
    CGO_ENABLED=0 go build -o "$run" "./app/cmd/${c}"
  fi

  help="$("$run" --help 2>&1 || true)"

  # Plain text for Windows / systems without man.
  printf '%s\n' "$help" > "${OUT}/${bin}.txt"

  # roff manpage. Escape backslashes, and lines that begin with a roff control
  # character (. or ') so help text is never interpreted as roff.
  up="$(printf '%s' "$bin" | tr '[:lower:]' '[:upper:]')"
  {
    printf '.TH %s 1 "%s" "Heimdall" "Heimdall Manual"\n' "$up" "$DATE"
    printf '.SH NAME\n%s \\- %s\n' "$bin" "$(desc "$c")"
    printf '.SH SYNOPSIS\n.B %s\n[options]\n' "$bin"
    printf '.SH DESCRIPTION\n.nf\n'
    printf '%s\n' "$help" | sed -e 's/\\/\\\\/g' -e "s/^[.']/\\\\&/"
    printf '.fi\n'
    printf '.SH SEE ALSO\nFull documentation: https://github.com/kinncj/Heimdall/tree/main/docs\n'
  } > "${OUT}/${bin}.1"

  echo "wrote ${OUT}/${bin}.1 and ${bin}.txt"
done
