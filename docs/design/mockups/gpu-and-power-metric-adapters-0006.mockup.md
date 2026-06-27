---
id: gpu-and-power-metric-adapters-0006
story: docs/stories/gpu-and-power-metric-adapters-20260626121705-0006/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — gpu-and-power-metric-adapters-0006

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · GPU & Power              ● dgx-spark online                 ⏱ 2026-06-26 14:08:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ GPU & POWER — dgx-spark (NVML, 2 GPUs) ──────────────────────────────────────────────────┐ ║
║   │ GPU0  RTX 6000        GPU1  RTX 6000                                                     │ ║
║   │ util ██████████████▓▒░░ 91%   util ██████████▓▒░░░░░░ 64%                                │ ║
║   │ vram ███████████▓▒░░░░░ 28/40  vram ██████▓▒░░░░░░░░░░ 18/40                             │ ║
║   │ temp 71°C ↑           temp 63°C →                                                        │ ║
║   │ pwr  142W             pwr  104W                                                          │ ║
║   │                                                                                          │ ║
║   │ POWER (read-only)     package 246W   cpu 96W   gpu 246W                                  │ ║
║   │ profile balanced      ▸ no control offered to change profile                             │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   btop/mactop-style gradient gauges █▓▒░ for util and VRAM; two GPUs shown side by side.       ║
║   Power profile is displayed READ-ONLY — the dashboard offers no control to change it.         ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS GPU via NVML · 2 GPUs · power read-only (profile balanced)                              ║
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
| GPU util/VRAM gauge | `severity.* by %` | nominal→critical | — | █▓▒░ per GPU |
| power watts value | `structure.value` | #eeeeee (255) | bold | read-only profile = structure.muted |
| vendor unavailable — | `states.unavailable` | #a8a8a8 (248) | faint | AMD/Pi graceful |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
