---
id: optional-privileged-metrics-helper-0005
story: docs/stories/optional-privileged-metrics-helper-20260626121705-0005/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — optional-privileged-metrics-helper-0005

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · privileged helper                ● both online              ⏱ 2026-06-26 14:07:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ helper – absent  (rpi-5) ─────────────────┐  ┌ helper ✓ installed  (dgx-spark) ──────────┐ ║
║   │ rpi-5            ● ONLINE                 │  │ dgx-spark        ● ONLINE                 │ ║
║   │ helper – absent                           │  │ helper ✓ installed                        │ ║
║   │                                           │  │                                           │ ║
║   │ cpu        35%   ● ok                     │  │ cpu        88%   ● ok                     │ ║
║   │ mem        52%   ● ok                     │  │ mem        63%   ● ok                     │ ║
║   │ temp(soc)  58°C  ● ok                     │  │ temp(soc)  74°C  ● ok                     │ ║
║   │ power      ⚿ needs helper                 │  │ power      142W  ● ok                     │ ║
║   │ temp(pkg)  ⚿ needs helper                 │  │ temp(pkg)  78°C  ● ok                     │ ║
║   │                                           │  │                                           │ ║
║   │ affordance, not an error                  │  │ privileged values populated               │ ║
║   └───────────────────────────────────────────┘  └───────────────────────────────────────────┘ ║
║                                                                                                ║
║   Power and full-thermal cells render ⚿ needs helper when the optional helper is absent,       ║
║   and populate when it is installed. Both daemons stay ● ONLINE throughout.                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ⚿ needs helper = insufficient permission (affordance, not error)                        ║
║ q quit · ↑/↓ nav · ⏎ detail · i install-helper info · ? help                                   ║
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
| needs-helper ⚿ (power/temp) | `states.needs_helper` | #5fd7ff (81) | — | affordance, not error |
| helper installed ✓ | `states.installed` | #5fd75f (77) | — | — |
| helper absent – | `states.absent` | #a8a8a8 (248) | faint | — |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
