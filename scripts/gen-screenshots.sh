#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Generate documentation screenshots of the Heimdall dashboard TUI.
#
# The dashboard renders a single deterministic frame with `--demo --snapshot`
# (the synthetic fleet), so we capture frames headlessly at a matrix of sizes,
# views, and themes — no terminal recorder required for the static shots.
#
# Outputs, in order of fidelity (each step is skipped if its tool is absent, so
# the script always produces *something*):
#   1. Raw ANSI  (.ansi)  — always; zero dependencies, diff-able, re-render-able.
#   2. Styled HTML (.html) — when `aha` is installed; embeddable in docs/pages.
#   3. Animated GIF (.gif) — when `vhs` is installed; driven by the .tape sources
#                            under docs/screenshots/tapes/ (interactive flows:
#                            logs `l`, top `t`, command `c` modals).
#
# Usage: scripts/gen-screenshots.sh [output-dir]   (default: docs/screenshots)
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="${1:-$root/docs/screenshots}"
bin="$root/bin/heimdall-dashboard"
mkdir -p "$out/html" "$out/ansi" "$out/gif"

if [[ ! -x "$bin" ]]; then
	echo "› building heimdall-dashboard"
	(cd "$root" && go build -o "$bin" ./app/cmd/dashboard)
fi

have() { command -v "$1" >/dev/null 2>&1; }

# Isolate config so passing --mode (a persistable setting) never rewrites the
# operator's real ~/.config/heimdall/dashboard.json.
export HEIMDALL_CONFIG_DIR
HEIMDALL_CONFIG_DIR="$(mktemp -d)"
trap 'rm -rf "$HEIMDALL_CONFIG_DIR"' EXIT

# capture <name> <cols> <rows> <theme> <title> [--detail]
# Renders one demo frame at the given size/theme into .ansi (+ .html via aha).
capture() {
	local name="$1" cols="$2" rows="$3" theme="$4" title="$5"
	shift 5
	local ansi="$out/ansi/$name.ansi"
	echo "› $name (${cols}x${rows}, $theme)"
	COLUMNS="$cols" LINES="$rows" "$bin" --demo --snapshot --mode "$theme" "$@" >"$ansi"
	if have aha; then
		aha --black --title "$title" <"$ansi" >"$out/html/$name.html"
	fi
}

# Static frames. The narrow grid showcases the responsive columns (story 0023);
# the small detail frame shows the scrollable fixed header/footer (story 0024).
capture grid-wide   110 30 dark  "Heimdall — fleet grid (wide)"
capture grid-narrow  64 24 dark  "Heimdall — fleet grid (narrow / portrait)"
capture grid-light  110 30 light "Heimdall — fleet grid (light)"
capture detail-wide 110 34 dark  "Heimdall — host detail" --detail
capture detail-small 88 20 dark  "Heimdall — host detail (small screen)" --detail

# Animated flows for the interactive modals. The .tape files are the committed,
# reproducible source; we only render them when vhs is available.
if have vhs; then
	for tape in "$out"/tapes/*.tape; do
		[[ -e "$tape" ]] || continue
		echo "› vhs $(basename "$tape")"
		(cd "$out" && vhs "$tape")
	done
else
	echo "› vhs not installed — skipping animated GIFs (tapes kept as source)"
fi

echo "done → $out"
echo "  ansi/  raw captures (always)"
if have aha; then echo "  html/  styled, embeddable renders"; fi
if have vhs; then echo "  gif/   animated modal flows"; fi
exit 0
