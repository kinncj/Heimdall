---
id: host-context-locale-uptime-os-0008
story: docs/stories/host-context-locale-uptime-os-20260626121705-0008/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — host-context-locale-uptime-os-0008

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Host context                ● dgx-spark online              ⏱ 2026-06-26 14:10:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ HOST CONTEXT — dgx-spark ────────────────────────────────────────────────────────────────┐ ║
║   │ os            Ubuntu 22.04.4 LTS          arch     x86_64                                │ ║
║   │ kernel        6.5.0-35-generic            hostname dgx-spark.local                       │ ║
║   │ locale        en_US.UTF-8                 timezone America/New_York (UTC-4)              │ ║
║   │ uptime        12d 04h 37m  (derived)      boot at  2026-06-14 09:26:11                   │ ║
║   │ agent         hwmon-daemon v1.4.2         class    server / DGX                          │ ║
║   │                                                                                          │ ║
║   │ shown on enroll · refreshed periodically while connected                                 │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Host detail shows OS+version, arch, kernel, hostname, locale, timezone, uptime (derived      ║
║   from boot time), agent version and a device-class label.                                     ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS context reported on enroll · refreshes while connected                                  ║
║ ⏎/esc back · ↑/↓ field · r refresh · ? help                                                    ║
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
| context field label | `structure.label` | #949494 (246) | faint | OS/arch/locale/tz/uptime |
| context value | `structure.value` | #eeeeee (255) | bold | — |
| device-class tag | `structure.accent` | #00d7ff (45) | — | {class:dgx-spark} |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
