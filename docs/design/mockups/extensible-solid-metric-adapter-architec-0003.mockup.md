---
id: extensible-solid-metric-adapter-architec-0003
story: docs/stories/extensible-solid-metric-adapter-architec-20260626121705-0003/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — extensible-solid-metric-adapter-architec-0003

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · dgx-spark adapters            ● ONLINE · ⚠ 1 error          ⏱ 2026-06-26 14:03:09 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   Host dgx-spark · ● ONLINE · adapters: 4 registered, 3 ok, 1 failing                          ║
║                                                                                                ║
║   ┌──────────────────┬────────┬────────────┬────────┬──────────┬────────────────────────┐      ║
║   │ ADAPTER          │ SIGNAL │ STATUS     │ VALUE  │ LAST OK  │ NOTE                   │      ║
║   ├──────────────────┼────────┼────────────┼────────┼──────────┼────────────────────────┤      ║
║   │ cpu.adapter      │  CPU   │ ● ok       │    88% │ 14:03:08 │ nominal                │      ║
║   │ mem.adapter      │  MEM   │ ● ok       │    63% │ 14:03:08 │ nominal                │      ║
║   │ storage.adapter  │  STO   │ ● ok       │    54% │ 14:03:08 │ nominal                │      ║
║   │ temp.adapter     │  TEMP  │ ⚠ error    │     -- │ 14:01:55 │ sensor read timeout    │      ║
║   └──────────────────┴────────┴────────────┴────────┴──────────┴────────────────────────┘      ║
║                                                                                                ║
║   temp.adapter failed in isolation: only TEMP shows ⚠ error. CPU/MEM/STO keep streaming and    ║
║   the host stays ● ONLINE — one failure ≠ host failure.                                        ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● host online · ⚠ temp.adapter error (isolated) · 3/4 ok                                ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
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
| needs-helper ⚿ | `states.needs_helper` | #5fd7ff info-cyan (81) | — | NEVER red |
| unavailable — | `states.unavailable` | #a8a8a8 (248) | faint | — |
| error ⚠ | `states.error` | #ff5f5f (203) | bold | fault only |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
