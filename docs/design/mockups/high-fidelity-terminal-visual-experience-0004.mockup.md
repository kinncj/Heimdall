---
id: high-fidelity-terminal-visual-experience-0004
story: docs/stories/high-fidelity-terminal-visual-experience-20260626121705-0004/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — high-fidelity-terminal-visual-experience-0004

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · render modes                   ● 6/7 online                 ⏱ 2026-06-26 14:03:12 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ HIGH-FIDELITY  truecolor + anim ──────────┐  ┌ DEGRADED  NO_COLOR / dumb / slow ─────────┐ ║
║   │ dgx-spark        ● ONLINE  ⠹ live         │  │ dgx-spark        [ON] ONLINE  busy        │ ║
║   │ ╓─ depth panel (shadow) ──────╖           │  │ +- plain panel ---------------+           │ ║
║   │ CPU  █████████▓▒░░ 88% ↑                  │  │ CPU  [######## ] 88% ^                    │ ║
║   │ MEM  ██████▓▒░░░░░ 63% →                  │  │ MEM  [######   ] 63% =                    │ ║
║   │ STO  █████▓▒░░░░░░ 54% →                  │  │ STO  [#####    ] 54% =                    │ ║
║   │ TEMP ████████▓▒░░░ 74°C ↑                 │  │ TEMP [#######  ] 74C ^                    │ ║
║   │ GPU  ██████████▓▒░ 91% ↑                  │  │ GPU  [######## ] 91% ^                    │ ║
║   │ VRAM ███████▓▒░░░░ 28/40GB                │  │ VRAM [######   ] 28/40GB                  │ ║
║   │ PWR  ██████▓▒░░░░░ 142W                   │  │ PWR  [#####    ] 142W                     │ ║
║   │ prof balanced (read-only)                 │  │ prof balanced (read-only)                 │ ║
║   │ ╙─────────────────────────────╜           │  │ +-----------------------------+           │ ║
║   │ gradient █▓▒░ · braille spinner           │  │ ascii # = bars · no animation             │ ║
║   └───────────────────────────────────────────┘  └───────────────────────────────────────────┘ ║
║                                                                                                ║
║   Same host, two modes. Every critical field — host, CONN, CPU/MEM/STO/TEMP and the new        ║
║   GPU/VRAM/PWR gauges + read-only profile — is present and readable in BOTH.                   ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS render=high-fidelity · v toggles mode · auto-detected                                   ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · v fidelity · ? help                                  ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

## Styles (lipgloss)

Each region maps to a role in `terminal-theme.json` (concrete dark-mode value). All colours come from `palette.json` — no drift.

| Region | Theme role | Foreground | Attrs | Border / BG |
|---|---|---|---|---|
| App frame | `borders.app / structure.border` | #585858 (240) | — | DoubleBorder ║═╔╗╚╝ |
| Title `⬢ HEIMDALL` | `structure.title` | #d0d0d0 steel (252) | bold | wordmark; tagline rule in signature blue |
| Brand signature / eye | `structure.accent` | #00d7ff electric-blue (45) | — | tagline dashes, brand mark |
| Clock / status bar | `structure.text_muted` | #949494 (246) | — | — |
| Section heading | `structure.heading` | #eeeeee (255) | bold | — |
| Panel surface | `structure.panel / borders.panel` | #585858 border | — | bg #1c1c1c (234), NormalBorder ┌─┐│└┘ |
| Column label | `structure.label` | #949494 (246) | faint | — |
| Metric value | `structure.value` | #eeeeee (255) | bold | — |
| Unit suffix | `structure.unit` | #949494 (246) | — | — |
| Footer keybinding | `structure.keybinding` | #eeeeee (255) | underline+bold | — |
| Focused row | `structure.focus` | #00d7ff electric-blue (45) | reverse | leading ▸ |
| online ● | `states.online` | #5fd75f (77) | — | — |
| stale ⏱ | `states.stale` | #d7af87 (180) | faint | — |
| offline ○ | `states.offline` | #a8a8a8 (248) | — | — |
| Depth panel (hi-fi) | `borders.depth` | #585858 (240) | — | ThickBorder + shadow #080808 |
| Gradient gauge (hi-fi) | `severity.* ramp` | cool→warm | — | █▓▒░ |
| Spinner (hi-fi) | `structure.accent` | #00d7ff (45) | — | braille ⠹⠸⠼ |
| Degraded frame | `borders.ascii` | terminal default | — | +-|; NO_COLOR; no anim |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
