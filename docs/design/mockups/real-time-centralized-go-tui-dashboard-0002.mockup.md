---
id: real-time-centralized-go-tui-dashboard-0002
story: docs/stories/real-time-centralized-go-tui-dashboard-20260626121705-0002/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — real-time-centralized-go-tui-dashboard-0002

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Remote Hardware Monitor           ● 6/7 online              ⏱ 2026-06-26 14:03:12 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌───────────────┬────────┬───────────┬──────┬──────┬──────┬───────┬──────┬───────┬─────────┐ ║
║   │ HOST          │ ORIGIN │ CONN      │ CPU% │ MEM% │ STO% │ TEMP  │ GPU% │ PWR   │ NET     │ ║
║   ├───────────────┼────────┼───────────┼──────┼──────┼──────┼───────┼──────┼───────┼─────────┤ ║
║   │ ▸ workstation │ local  │ ● ONLINE  │  72% │  48% │  31% │  61°C │  64% │  110W │ ● 8ms   │ ║
║   │   dgx-spark   │  edge  │ ● ONLINE  │  88% │  63% │  54% │  74°C │  91% │  142W │ ● 12ms  │ ║
║   │   strix-halo  │ local  │ ● ONLINE  │  40% │  35% │  22% │  52°C │  n/a │   88W │ ● 9ms   │ ║
║   │   mac-mini    │ local  │ ● ONLINE  │  18% │  29% │  67% │  45°C │  38% │   22W │ ● 15ms  │ ║
║   │   rpi-5       │  edge  │ ● ONLINE  │  35% │  52% │  80% │  58°C │  n/a │   n/a │ ● 24ms  │ ║
║   │   alienware   │ local  │ ● ONLINE  │  61% │  71% │  49% │  69°C │  77% │   96W │ ● 20ms  │ ║
║   └───────────────┴────────┴───────────┴──────┴──────┴──────┴───────┴──────┴───────┴─────────┘ ║
║                                                                                                ║
║   GPU% n/a = unsupported vendor (story 0006) · PWR n/a = needs privileged helper (0005)        ║
║   NET = reachability symbol + latency (0007) · ORIGIN = federation hub (0009, edge = relayed). ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● streaming · gRPC low-bw · poll 2s · 1 host off-screen (nuc-lab ○)                     ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · / filter · L logs · ? help                           ║
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
| CPU/MEM/STO gauge fill | `severity.* by %` | nominal #5fd7af→critical #ff5f5f | — | blocks █▓▒░ |
| ORIGIN (federation) label | `structure.label` | #949494 (246) | faint | origin:edge |
| sparkline trend | `structure.text_secondary` | #c6c6c6 (251) | — | ▁▂▃▄▅▆▇█ |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
