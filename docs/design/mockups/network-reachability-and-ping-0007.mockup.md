---
id: network-reachability-and-ping-0007
story: docs/stories/network-reachability-and-ping-20260626121705-0007/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — network-reachability-and-ping-0007

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Network                  ● online · 8ms                     ⏱ 2026-06-26 14:09:03 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ NETWORK — workstation ───────────────────────────────────────────────────────────────────┐ ║
║   │ reachability  ● online        target 1.1.1.1                                             │ ║
║   │ latency       8 ms            internet up                                                │ ║
║   │ trend (90s)   ▕▁▂▂▃▂▁▂▃▄▃▂▁▂▁▂▃▂▁▂▏ ms                                                   │ ║
║   │                                                                                          │ ║
║   │ throughput    ↓ 4.2 MB/s   ↑ 0.8 MB/s                                                    │ ║
║   │ probe         ok 14:09:02 (every 5s)                                                     │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Reachability shows BOTH a symbol and text (● online); latency carries a sparkline trend      ║
║   in ms; throughput shows ↓/↑. The ping target is explicit.                                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● online · 8ms · internet up · target 1.1.1.1                                           ║
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
| reachability online ● | `states.online` | #5fd75f (77) | — | — |
| no-internet ○ | `states.offline` | #a8a8a8 (248) | — | symbol+text |
| latency sparkline | `structure.text_secondary` | #c6c6c6 (251) | — | ms |
| probe failed ⚠ | `states.error` | #ff5f5f (203) | bold | isolated |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
