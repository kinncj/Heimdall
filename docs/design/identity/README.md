# Heimdall — Visual Identity (TUI)

> "Watch Over All Realms"

> **Brand chrome** — wordmark + accent — is specified canonically in [`tui-brand.md`](./tui-brand.md):
> a **steel `#d0d0d0` wordmark** plus a single **electric-blue signature `#00d7ff`** accent (Heimdall's
> eye / the Bifröst glow), derived from the assets in `/assets`. This README is the colour & contrast
> source of record and is aligned to it — steel-on-bg **12.15:1**, signature-on-bg **10.85:1**; the
> full table is in §6.

**Target:** `tui` (terminal). **Default theme:** dark. **Status:** token files marked `approved` by the maintainer (see §7); this README is the matching AA contrast record.

This directory holds the colour and type source of record for Heimdall, a terminal
hardware-monitoring dashboard. Files:

| File | What it owns |
|---|---|
| `palette.json` | Colours: ramps, named roles, semantic states, severity ramp. Hex **+ nearest ANSI-256 index** + glyph + intended role. The colour source of record. |
| `typography.json` | Terminal "type": SGR attribute roles (bold / faint / italic / underline / reverse). No fonts or sizes. |
| `tokens.json` | **Canonical W3C DTCG** token file — the single edit surface. Composes colour references + spacing (cells) + borders (box-drawing) + type/status roles. The design-system-author emits `terminal-theme.json` from this. |
| `README.md` | This file: intent, dark/light stance, semantic map, severity ramp, **contrast table**. |

---

## 1. Intent

Heimdall guards the Bifröst and misses nothing. Two ideas drive the palette, and **readability of
monitoring data always wins over decoration**:

- **One brand accent — the electric-blue signature** (`#00d7ff`, ANSI 45): Heimdall's all-seeing eye
  and the Bifröst glow. With the **steel/silver wordmark** (`#d0d0d0`, ANSI 252) it marks the title
  (`⬢ Heimdall`), the focused row, and the focused panel border — and nothing else. `secondary` and
  `accent` are intentionally `null`: a monitoring tool earns its colour budget on *state*, not
  branding. More brand colour would be noise.
- **The rainbow bridge is functional, not decorative.** The Bifröst spectrum is spent where it earns
  its keep — the **severity ramp** (cool → warm) on utilisation/temperature gauges, so a full gauge
  literally reads "hot/busy". No neon-on-neon; backgrounds sit near-black so the data carries the
  contrast.

Colour is a *reinforcement* layer only. Every state is a **glyph + word** (and a number/timestamp
where relevant), so the UI is fully legible with colour switched off.

## 2. Dark / light stance

- **Dark is primary.** Assume a near-black terminal (`#121212`). All design decisions are made here
  first; the high-fidelity render and depth panels are tuned for dark.
- **Light is supported.** A parallel set (`themes.light` / `color.light.*`) darkens every hue so it
  clears AA on a white/near-white background. It exists so Heimdall stays legible on light terminals,
  not as a first-class brand surface.
- **Both themes pass WCAG 2.2 AA on `background` *and* `panel` surfaces** (see §6).

### ANSI-256 fallback

Every colour is authored as truecolor (24-bit) hex **and** carries its nearest ANSI-256 index
(`ansi256`) for 256-colour terminals. Neutrals — including the unfocused panel **border** (`#767676` ·
ANSI 243 dark, `#626262` · ANSI 241 light) — are aligned to the ANSI grey ramp for clean degradation.
Where two warm severity stops round to neighbouring ANSI cells they always sit in different regions
(gauge vs title) and each carries a glyph/word/number, so 256-colour rounding never loses meaning.

## 3. Monochrome / `NO_COLOR` behaviour

Under `NO_COLOR`, `TERM=dumb`, or a monochrome terminal, colour collapses to the terminal's default
fg/bg. Meaning survives through **five non-colour carriers**:

1. **Status glyph** — `●  ◐  ○  ⏱  ⚠  ⚿  —  ✓  –  ↑  ↯  ⚡  ✋  ▸`
2. **Status word** — `ONLINE`, `DEGRADED`, `STALE`, `refused`, …
3. **Numeric value / timestamp** — `88%`, `74°C`, `142W`, `14:03:12 (2m ago)`
4. **Gauge block density** — `░▒▓█` fill encodes magnitude independent of colour
5. **Reverse video** for focus (works with zero colours)

This is the accessible baseline render (wireframe `0004`, "degraded" column). It satisfies the
terminal a11y checks `color-only-signaling`, `no-color-support`, and `focus-visible`.

## 4. Semantic state → colour map

14 legend states (stories 0001–0011) map onto **6 hue roles**; the electric-blue signature is brand
chrome (title / focus), not a state colour. Distinctness is guaranteed by
glyph + word; colour reinforces by family. (Dark hex shown; see `palette.json` for light.)

| State | Glyph | Hue role | Dark hex / ANSI | Emphasis | Meaning |
|---|---|---|---|---|---|
| online / ok | `●` | success | `#5fd75f` · 77 | — | live & healthy |
| installed | `✓` | success | `#5fd75f` · 77 | — | helper/adapter present |
| degraded / enrolling | `◐` | caution | `#ffaf00` · 214 | — | transitional / partial |
| reconnecting | `↯` (`◑`) | caution | `#ffaf00` · 214 | — | link re-establishing (re-auth) |
| rate-limited | `⚡` | caution | `#ffaf00` · 214 | — | stream throttled, drops |
| stale | `⏱` | caution (faded) | `#d7af87` · 180 | faint | last-known values + age |
| offline / no-internet | `○` | absence | `#a8a8a8` · 248 | — | no link / probe target down |
| unavailable | `—` | absence | `#a8a8a8` · 248 | faint | vendor/path n/a |
| absent | `–` | absence | `#a8a8a8` · 248 | faint | helper not installed |
| error / probe-failed | `⚠` | critical | `#ff5f5f` · 203 | bold | fault on one metric |
| refused | `✋` | critical | `#ff5f5f` · 203 | bold | not allow-listed / sudo denied |
| needs-helper | `⚿` | info | `#5fd7ff` · 81 | — | insufficient permission — an **affordance, never a red error** |
| relaying | `↑` | info | `#5fd7ff` · 81 | — | hub relaying upstream |
| focus / selected | `▸` | brand | signature `#00d7ff` · 45 | reverse | active row (reverse video + `▸`) |

Two deliberate consolidations: **offline/unavailable/absent** share the neutral "absence" colour
(absence of data is not a fault), and **error/refused** share critical red (both are hard stops). The
glyph and word disambiguate within each family.

## 5. Severity ramp (Bifröst gauges)

A sequential **cool → warm** ramp for `█▓▒░` utilisation/temperature gauges and sparklines. Higher =
hotter/busier. Distinct in ANSI-256; the **monochrome fallback is block density + the printed value**,
so the ramp never depends on colour.

| Stop | Band | Dark hex / ANSI | Light hex / ANSI |
|---|---|---|---|
| 1 · nominal | 0–39 % | `#5fd7af` · 79 | `#006d00` · 22 |
| 2 · moderate | 40–59 % | `#87d75f` · 113 | `#5a7400` · 64 |
| 3 · elevated | 60–74 % | `#ffd75f` · 221 | `#806600` · 94 |
| 4 · high | 75–89 % | `#ff875f` · 209 | `#9c5400` · 130 |
| 5 · critical | 90–100 % | `#ff5f5f` · 203 | `#c81414` · 160 |

`severity.critical` reuses the `error` red on purpose: a maxed gauge and an error are the same alarm.

## 6. Contrast table (WCAG 2.2 AA)

Recomputed from the committed hexes by `verify_identity.py`. **Min** = `4.5:1` for text, `3:1` for
large/UI (gauge blocks, focus, **borders**). Every pair is checked on both the `background` and `panel`
surface. **All 65 pairs PASS** — there is no failing pair in either theme.

### Dark theme

| role | fg | bg | min | ratio | AA |
|---|---|---|---|---|---|
| structure.border | `#767676` | background #121212 | 3:1 | **4.12:1** | PASS |
| structure.border | `#767676` | panel #1c1c1c | 3:1 | **3.75:1** | PASS |
| structure.border_focus | `#00d7ff` | background #121212 | 3:1 | **10.85:1** | PASS |
| structure.border_focus | `#00d7ff` | panel #1c1c1c | 3:1 | **9.87:1** | PASS |
| text.primary | `#eeeeee` | background #121212 | 4.5:1 | **16.15:1** | PASS |
| text.primary | `#eeeeee` | panel #1c1c1c | 4.5:1 | **14.69:1** | PASS |
| text.primary | `#eeeeee` | selection #303030 | 4.5:1 | **11.38:1** | PASS |
| text.secondary | `#c6c6c6` | background #121212 | 4.5:1 | **10.97:1** | PASS |
| text.secondary | `#c6c6c6` | panel #1c1c1c | 4.5:1 | **9.98:1** | PASS |
| text.muted | `#949494` | background #121212 | 4.5:1 | **6.18:1** | PASS |
| text.muted | `#949494` | panel #1c1c1c | 4.5:1 | **5.62:1** | PASS |
| brand.signature | `#00d7ff` | background #121212 | 4.5:1 | **10.85:1** | PASS |
| brand.signature | `#00d7ff` | panel #1c1c1c | 4.5:1 | **9.87:1** | PASS |
| brand.steel | `#d0d0d0` | background #121212 | 4.5:1 | **12.15:1** | PASS |
| brand.steel | `#d0d0d0` | panel #1c1c1c | 4.5:1 | **11.05:1** | PASS |
| semantic.online | `#5fd75f` | background #121212 | 4.5:1 | **10.14:1** | PASS |
| semantic.online | `#5fd75f` | panel #1c1c1c | 4.5:1 | **9.23:1** | PASS |
| semantic.degraded | `#ffaf00` | background #121212 | 4.5:1 | **10.16:1** | PASS |
| semantic.degraded | `#ffaf00` | panel #1c1c1c | 4.5:1 | **9.24:1** | PASS |
| semantic.stale | `#d7af87` | background #121212 | 4.5:1 | **9.25:1** | PASS |
| semantic.stale | `#d7af87` | panel #1c1c1c | 4.5:1 | **8.42:1** | PASS |
| semantic.offline | `#a8a8a8` | background #121212 | 4.5:1 | **7.88:1** | PASS |
| semantic.offline | `#a8a8a8` | panel #1c1c1c | 4.5:1 | **7.17:1** | PASS |
| semantic.error | `#ff5f5f` | background #121212 | 4.5:1 | **6.29:1** | PASS |
| semantic.error | `#ff5f5f` | panel #1c1c1c | 4.5:1 | **5.72:1** | PASS |
| semantic.info | `#5fd7ff` | background #121212 | 4.5:1 | **11.29:1** | PASS |
| semantic.info | `#5fd7ff` | panel #1c1c1c | 4.5:1 | **10.27:1** | PASS |
| severity.nominal | `#5fd7af` | background #121212 | 3:1 | **10.55:1** | PASS |
| severity.moderate | `#87d75f` | background #121212 | 3:1 | **10.63:1** | PASS |
| severity.elevated | `#ffd75f` | background #121212 | 3:1 | **13.50:1** | PASS |
| severity.high | `#ff875f` | background #121212 | 3:1 | **7.92:1** | PASS |
| severity.critical | `#ff5f5f` | background #121212 | 3:1 | **6.29:1** | PASS |
| focus (reverse on signature) | `#121212` | signature-fill #00d7ff | 4.5:1 | **10.85:1** | PASS |

### Light theme

| role | fg | bg | min | ratio | AA |
|---|---|---|---|---|---|
| structure.border | `#626262` | background #ffffff | 3:1 | **6.10:1** | PASS |
| structure.border | `#626262` | panel #eeeeee | 3:1 | **5.26:1** | PASS |
| structure.border_focus | `#005f87` | background #ffffff | 3:1 | **7.03:1** | PASS |
| structure.border_focus | `#005f87` | panel #eeeeee | 3:1 | **6.06:1** | PASS |
| text.primary | `#1c1c1c` | background #ffffff | 4.5:1 | **17.04:1** | PASS |
| text.primary | `#1c1c1c` | panel #eeeeee | 4.5:1 | **14.69:1** | PASS |
| text.secondary | `#3a3a3a` | background #ffffff | 4.5:1 | **11.37:1** | PASS |
| text.secondary | `#3a3a3a` | panel #eeeeee | 4.5:1 | **9.80:1** | PASS |
| text.muted | `#626262` | background #ffffff | 4.5:1 | **6.10:1** | PASS |
| text.muted | `#626262` | panel #eeeeee | 4.5:1 | **5.26:1** | PASS |
| brand.signature | `#005f87` | background #ffffff | 4.5:1 | **7.03:1** | PASS |
| brand.signature | `#005f87` | panel #eeeeee | 4.5:1 | **6.06:1** | PASS |
| brand.steel | `#1c1c1c` | background #ffffff | 4.5:1 | **17.04:1** | PASS |
| brand.steel | `#1c1c1c` | panel #eeeeee | 4.5:1 | **14.69:1** | PASS |
| semantic.online | `#006d00` | background #ffffff | 4.5:1 | **6.59:1** | PASS |
| semantic.online | `#006d00` | panel #eeeeee | 4.5:1 | **5.68:1** | PASS |
| semantic.degraded | `#9c5400` | background #ffffff | 4.5:1 | **5.70:1** | PASS |
| semantic.degraded | `#9c5400` | panel #eeeeee | 4.5:1 | **4.92:1** | PASS |
| semantic.stale | `#6b5526` | background #ffffff | 4.5:1 | **7.11:1** | PASS |
| semantic.stale | `#6b5526` | panel #eeeeee | 4.5:1 | **6.13:1** | PASS |
| semantic.offline | `#626262` | background #ffffff | 4.5:1 | **6.10:1** | PASS |
| semantic.offline | `#626262` | panel #eeeeee | 4.5:1 | **5.26:1** | PASS |
| semantic.error | `#c81414` | background #ffffff | 4.5:1 | **5.89:1** | PASS |
| semantic.error | `#c81414` | panel #eeeeee | 4.5:1 | **5.08:1** | PASS |
| semantic.info | `#005f87` | background #ffffff | 4.5:1 | **7.03:1** | PASS |
| semantic.info | `#005f87` | panel #eeeeee | 4.5:1 | **6.06:1** | PASS |
| severity.nominal | `#006d00` | background #ffffff | 3:1 | **6.59:1** | PASS |
| severity.moderate | `#5a7400` | background #ffffff | 3:1 | **5.34:1** | PASS |
| severity.elevated | `#806600` | background #ffffff | 3:1 | **5.50:1** | PASS |
| severity.high | `#9c5400` | background #ffffff | 3:1 | **5.70:1** | PASS |
| severity.critical | `#c81414` | background #ffffff | 3:1 | **5.89:1** | PASS |
| focus (reverse on signature) | `#ffffff` | signature-fill #005f87 | 4.5:1 | **7.03:1** | PASS |

> The smallest margin is light `semantic.degraded` on `panel` at **4.92:1** (min 4.5:1) — still clears
> the normal-text bar. The unfocused chrome **border** is the tightest UI pair at **3.75:1** on `panel`
> (dark) — above the 3:1 large/UI bar (`#585858`/240 would have been 2.63:1 and is rejected).

### Reproduce

```bash
for f in docs/design/identity/*.json; do python3 -c "import json;json.load(open('$f'))" && echo "OK $f"; done
# Full audit (contrast recompute, ANSI presence, DTCG ref resolution, palette<->tokens drift):
python3 docs/design/identity/verify_identity.py
```

## 7. Approval

`palette.json`, `typography.json`, and `tokens.json` carry `status: approved` in their `_meta`, set by
the maintainer — this agent does not self-approve. This README has been brought in line with that
approved palette (electric-blue signature + steel wordmark) and with the AA-verified `color.border`
token. The design-system-author emits `terminal-theme.json` (ANSI-256/truecolor + lipgloss styles)
from `tokens.json`.
