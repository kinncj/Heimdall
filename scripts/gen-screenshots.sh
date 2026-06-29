#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Generate documentation screenshots of the Heimdall dashboard TUI.
#
# The dashboard renders a single deterministic frame with `--demo --snapshot`
# (the synthetic fleet), so we capture frames headlessly at a matrix of sizes,
# views, and themes — no terminal recorder required for the static shots. Color is
# forced (CLICOLOR_FORCE) so the captures keep the theme's ONLINE/alert palette
# even though stdout is a pipe.
#
# Outputs, in order of fidelity (each step is skipped if its tool is absent, so
# the script always produces *something*):
#   1. Raw ANSI  (.ansi)  — always; zero dependencies, diff-able, re-render-able.
#   2. Styled HTML (.html) — when `aha` is installed; embeddable in docs/pages.
#   3. PNG image  (img/*.png) — when `aha` + headless `chromium` + ImageMagick are
#                            present; trimmed, retina (2x). These are committed and
#                            embedded in the README and guides (GitHub renders PNG).
#   4. Animated GIF (.gif) — when `vhs` is installed; driven by the .tape sources
#                            under docs/screenshots/tapes/ (interactive flows:
#                            logs `l`, top `t`, command `c` modals).
#
# Usage: scripts/gen-screenshots.sh [output-dir]   (default: docs/screenshots)
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="${1:-$root/docs/screenshots}"
bin="$root/bin/heimdall-dashboard"
mkdir -p "$out/html" "$out/ansi" "$out/img" "$out/gif"

if [[ ! -x "$bin" ]]; then
	echo "› building heimdall-dashboard"
	(cd "$root" && go build -o "$bin" ./app/cmd/dashboard)
fi

have() { command -v "$1" >/dev/null 2>&1; }

# Resolve a headless-Chromium binary and an ImageMagick command, if present.
chrome=""
for c in chromium chromium-browser google-chrome google-chrome-stable; do
	if have "$c"; then chrome="$c"; break; fi
done
im=""
for c in magick convert; do
	if have "$c"; then im="$c"; break; fi
done

# Isolate config so passing --mode (a persistable setting) never rewrites the
# operator's real ~/.config/heimdall/dashboard.json.
export HEIMDALL_CONFIG_DIR
HEIMDALL_CONFIG_DIR="$(mktemp -d)"
trap 'rm -rf "$HEIMDALL_CONFIG_DIR"' EXIT

# capture <name> <cols> <rows> <theme> <title> [--detail]
# Renders one demo frame at the given size/theme into .ansi, .html (aha), and a
# trimmed PNG (chromium + ImageMagick).
capture() {
	local name="$1" cols="$2" rows="$3" theme="$4" title="$5"
	shift 5
	local ansi="$out/ansi/$name.ansi" html="$out/html/$name.html" png="$out/img/$name.png"
	echo "› $name (${cols}x${rows}, $theme)"
	COLUMNS="$cols" LINES="$rows" CLICOLOR_FORCE=1 \
		"$bin" --demo --snapshot --mode "$theme" "$@" >"$ansi"
	have aha || return 0
	aha --black --title "$title" <"$ansi" >"$html"
	[[ -n "$chrome" && -n "$im" ]] || return 0
	# Generous window (content is left/top-anchored); ImageMagick trims the black
	# margin afterwards so the PNG is exactly the frame.
	"$chrome" --headless --no-sandbox --hide-scrollbars --disable-gpu \
		--force-device-scale-factor=2 --window-size="$((cols * 10 + 40)),$((rows * 26 + 40))" \
		--screenshot="$png" "file://$html" >/dev/null 2>&1 || return 0
	"$im" "$png" -bordercolor black -border 20 -trim +repage "$png" >/dev/null 2>&1 || true
}

# Static frames. The narrow grid showcases the responsive columns (story 0023);
# the small detail frame shows the scrollable fixed header/footer (story 0024).
# Heights are generous so the whole fleet (7 demo hosts) fits before the trim.
capture grid-wide    112 22 dark  "Heimdall — fleet grid (wide)"
capture grid-narrow   64 22 dark  "Heimdall — fleet grid (narrow / portrait)"
capture grid-light   112 22 light "Heimdall — fleet grid (light)"
capture detail-wide  110 40 dark  "Heimdall — host detail" --detail
capture detail-small  88 22 dark  "Heimdall — host detail (small screen)" --detail

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
if [[ -n "$chrome" && -n "$im" ]]; then echo "  img/   trimmed PNG screenshots (committed, used in docs)"; fi
if have vhs; then echo "  gif/   animated modal flows"; fi
exit 0
