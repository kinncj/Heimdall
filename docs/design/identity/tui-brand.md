---
id: heimdall-tui-brand
target: tui
status: approved
created_at: 2026-06-26
supersedes: "the placeholder gold accent in palette.json/tokens.json"
---

# Heimdall — TUI Brand & Header Spec

Canonical brand, derived from the design assets in [`/assets`](../../../assets). These are the
source of truth for the dashboard's chrome (header, status bar, splash). Implemented in epic 3
(`real-time-centralized-go-tui-dashboard-0002`, `high-fidelity-terminal-visual-experience-0004`).

## Brand

| Element | Value |
|---|---|
| Name | **Heimdall** |
| Tagline | **Watch Over All Realms** |
| Brand mark | `⬢` (hex sigil) + the winged-helm watchman (full art: `assets/ICON.png`, `assets/LOGO.png`) |
| Wordmark | **steel/silver**, bold — dark `#d0d0d0` (ansi 252) · light `#1c1c1c` (ansi 234) |
| Signature accent | **electric blue** (Heimdall's eye / Bifröst glow) — dark `#00d7ff` (ansi 45) · light `#005f87` (ansi 24) |
| online / streaming / relay | green `#5fd75f` (ansi 77) |
| rate-limited / transitional | amber `#ffaf00` (ansi 214) |
| Surface | near-black `#121212` (ansi 233) |

Contrast (WCAG 2.2 AA, recomputed): steel-on-bg **12.15:1**, signature-on-bg **10.85:1**,
signature-on-panel **9.87:1**, light signature-on-white **7.03:1** — all pass. The signature is a
reinforcement accent; every state remains glyph + word, so the chrome survives `NO_COLOR`.

## Assets — which form to use

Per the brand owner: **use the PNG where it fits; use the ASCII `.txt` only where images can't render.**

| Context | Use | Files (`/assets`) |
|---|---|---|
| Docs, README, design portal (images render) | **PNG** | `LOGO.png` · `ICON.png` · `*_NO_BG.png` (transparent) · `TUI_HEADER_FAT/SKINNY.png` · `TUI_STATUS_BAR.png` |
| The terminal UI (images cannot render) | **ASCII `.txt`** | `ASCII_ART.txt` (splash) · `ICON_ASCII_ART.txt` · `LOGO_ASCII_ART.txt` |

## Splash (startup / about screen)

The terminal splash embeds the real art **`assets/ASCII_ART.txt`** (80×235, plain `@`-glyph render)
via `go:embed`, shown when the terminal is wide enough; the wordmark below is the narrow (≲80 col)
fallback. The inline header uses the compact `⬢` sigil, never the full art.

```text
        ╦ ╦ ╔═╗ ╦ ╔╦╗ ╔╦╗ ╔═╗ ╦   ╦
        ╠═╣ ║╣  ║ ║║║  ║║ ╠═╣ ║   ║
        ╩ ╩ ╚═╝ ╩ ╩ ╩ ═╩╝ ╩ ╩ ╩═╝ ╩═╝
            -=- watch over all realms -=-
```
- Wordmark `structure.title` (steel, bold). Tagline rule `-=-` in `structure.accent` (electric blue).

## Header — fat (logo + wordmark + tagline)

```text
┌┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┐
┊ ⬢  H E I M D A L L                                                       ● 6/7 ONLINE       ┊
┊    -- watch over all realms --                                           🕐 14:03:12          ┊
└┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┘
```

## Header — skinny (default dashboard chrome)

```text
┌┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┐
┊ ⬢  HEIMDALL          │          ● 6/7 ONLINE          │          🕐 14:03:12               ┊
└┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┘
```

## Status bar (footer chrome)

```text
┌┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┐
┊ ⬢ HEIMDALL │ ● streaming │ poll 2s │ low-bw gRPC │ ↑ edge relay │ ⚡ rate-limited │ 🕐 14:03:12 ┊
└┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┘
```

## Lipgloss mapping (from `terminal-theme.json`)

| Chrome element | Theme role | Style |
|---|---|---|
| Chrome border | `structure.border` | dashed (lipgloss custom Border, runes `┄ ┊`) — distinct from the solid panel frames |
| `⬢` + wordmark | `structure.title` | steel `#d0d0d0`, bold |
| Tagline `-- … --` | `structure.accent` | electric-blue `#00d7ff` |
| `● n/m ONLINE` | `states.online` | green `#5fd75f`, dot + word |
| `↑ edge relay` (Bifröst) | `states.relaying` | `#5fd7ff` |
| `⚡ rate-limited` | `states.rate_limited` | amber `#ffaf00` |
| `🕐 HH:MM:SS` | `structure.text_muted` | `#949494` |
| separators `│` | `structure.border` | — |

Notes: counts (`6/7`) and the clock are always present so status never depends on colour. The skinny
header is the default; the fat header is used on the splash / wide terminals; both collapse gracefully
below ~80 columns (drop the tagline, then abbreviate `ONLINE` → the `●` dot + count).
