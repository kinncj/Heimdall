---
id: unprivileged-terminal-control-plane-0010
story: docs/stories/unprivileged-terminal-control-plane-20260626121705-0010/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — unprivileged-terminal-control-plane-0010

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Control plane             ● dgx-spark · read-only           ⏱ 2026-06-26 14:14:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ CONTROL PLANE — dgx-spark  (read-only) ──────────────────────────────────────────────────┐ ║
║   │ command ▸ process.list      args ▏--sort=cpu --top=5▕   [⏎ run]                          │ ║
║   │         · disk.df           · fs.ls <allowed>  · net.ifaces                              │ ║
║   │                                                                                          │ ║
║   │ result (bounded) ─────────────────────────────────────────────                           │ ║
║   │   PID    USER     %CPU  %MEM  COMMAND                                                    │ ║
║   │   1042   svc-mon  18.2   3.1  hwmon-daemon                                               │ ║
║   │   2210   svc-mon   6.4   1.2  nvml-collector                                             │ ║
║   │   3398   svc-mon   2.1   0.8  ping-prober                                                │ ║
║   │   … truncated (showing 5 of 214 · bounded output)                                        │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   read-only · runs as svc-mon · no sudo · audited (every invocation logged)                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS allow-listed query process.list · result bounded · audit-logged                         ║
║ ↑/↓ pick · tab args · ⏎ run · esc close · keyboard-only · ? help                               ║
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
| control overlay panel | `borders.panelRounded` | #585858 (240) | — | RoundedBorder; backdrop overlay_backdrop |
| allow-list input | `structure.input_active` | #eeeeee (255) | underline | cursor ▏ |
| banner read-only/no-sudo | `structure.caption` | #949494 (246) | italic | audited |
| refused ✋ | `states.refused` | #ff5f5f (203) | bold | not allow-listed / sudo |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
