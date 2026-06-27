---
id: opt-in-log-streaming-0011
story: docs/stories/opt-in-log-streaming-20260626121705-0011/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — opt-in-log-streaming-0011

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Logs                 ● live · ⚡ rate-limited               ⏱ 2026-06-26 14:15:06 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ LOGS — dgx-spark  (live · opt-in · separate stream) ─────────────────────────────────────┐ ║
║   │ source /var/log/syslog        ⚡ rate-limited (37 dropped)                               │ ║
║   │ ────────────────────────────────────────────────────────────                             │ ║
║   │ 14:15:02  INFO  systemd      Started hwmon-daemon.service                                │ ║
║   │ 14:15:03  WARN  nvml         GPU1 ECC retired page detected                              │ ║
║   │ 14:15:04  INFO  hwmon        stream resumed (offset 0x1f3a)                              │ ║
║   │ 14:15:05  ERR   ping-prober  target 1.1.1.1 timeout (isolated)                           │ ║
║   │ 14:15:06  INFO  hwmon        poll 2s · 6 hosts live                                      │ ║
║   │ ▸ tailing… (newest at bottom)                                                            │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Live tailed lines per host carry source + level + timestamp. The log stream is SEPARATE      ║
║   from metrics and is rate-limited; ⚡ rate-limited (N dropped) signals drops on low-bw link.  ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS logs live · source=/var/log/syslog · ⚡ rate-limited (37 dropped)                       ║
║ ↑/↓ scroll · f follow · / filter level · esc close · keyboard-only · ? help                    ║
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
| log line body | `structure.body` | #eeeeee (255) | — | newest at bottom |
| source/level label | `structure.label` | #949494 (246) | faint | — |
| rate-limited ⚡ | `states.rate_limited` | #ffaf00 (214) | — | N dropped |
| off (opt-in) state | `structure.muted` | #949494 (246) | — | no source configured |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
