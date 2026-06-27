---
id: cross-platform-metrics-daemon-enrollment-0001
story: docs/stories/cross-platform-metrics-daemon-enrollment-20260626121705-0001/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — cross-platform-metrics-daemon-enrollment-0001

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Enrollment             ● 5/6 · ◐ 1 enrolling                ⏱ 2026-06-26 14:02:12 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌───────────────┬─────┬──────────┬──────────────┬──────────┬──────────────────────┐          ║
║   │ HOST          │ OS  │ HOST-ID  │ ADDRESS      │ PROTO    │ CONNECTION           │          ║
║   ├───────────────┼─────┼──────────┼──────────────┼──────────┼──────────────────────┤          ║
║   │ ▸ workstation │ win │ h-7a1c   │ 10.0.1.21    │ gRPC+TLS │ ● ONLINE   14:02:10  │          ║
║   │   dgx-spark   │ lin │ h-3f9a   │ 10.0.1.32    │ gRPC+TLS │ ● ONLINE   14:02:08  │          ║
║   │   strix-halo  │ lin │ h-9b2d   │ 10.0.1.33    │ gRPC+TLS │ ● ONLINE   14:02:11  │          ║
║   │   mac-mini    │ mac │ h-1c4e   │ 10.0.1.44    │ gRPC+TLS │ ● ONLINE   14:02:05  │          ║
║   │   rpi-5       │ lin │ h-5d80   │ 10.0.1.55    │ gRPC+TLS │ ● ONLINE   14:02:09  │          ║
║   │   alienware   │ win │ h-2e66   │ 10.0.1.66    │ gRPC+TLS │ ◐ ENROLLING…  (new)  │          ║
║   └───────────────┴─────┴──────────┴──────────────┴──────────┴──────────────────────┘          ║
║                                                                                                ║
║   Mixed OS (Windows / macOS / Linux). A first-time host appears as ONE new row when its        ║
║   daemon completes handshake — alienware just joined: ◐ ENROLLING… → ● ONLINE next tick.       ║
║   Enrollment is mutual-TLS + bearer-token authenticated — no metrics flow until the TLS        ║
║   handshake and token check pass (PROTO shows gRPC+TLS).                                       ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● 5 online · ◐ 1 enrolling · gRPC low-bandwidth · mTLS+token                            ║
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
| enrolling ◐ | `states.enrolling` | #ffaf00 (214) | — | CONNECTION column |
| reconnecting ↯ | `states.reconnecting` | #ffaf00 (214) | — | — |
| TLS/token note | `structure.caption` | #949494 (246) | italic | handshake gate |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
