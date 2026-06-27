---
id: gjallarhorn-dashboard-badge-0019
story: docs/stories/gjallarhorn-dashboard-badge-20260627203215-0019/Story.md
target: tui
status: approved
created_at: 2026-06-27
---

# Mockup — gjallarhorn-dashboard-badge-0019

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) chrome,
amber/red alert colour, 98-col double-border frame. Colour reinforces alert state that is always
also glyph (`⚠`/`🔔`) + word + value — NO_COLOR-safe.

## State 1 — quiet (no alerts)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL │ ● 6/6 ONLINE │ 🕐 14:31:55                                                        ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌───────────────┬──────────┬──────┬──────┬───────┬───────┬──────┬───────┐                    ║
║   │ HOST          │ STATE    │ CPU% │ MEM% │ DISK% │  TEMP │ GPU% │   PWR │                    ║
║   ├───────────────┼──────────┼──────┼──────┼───────┼───────┼──────┼───────┤                    ║
║   │ ▸ workstation │ ● ONLINE │  72% │  48% │   31% │  61°C │  64% │  110W │                    ║
║   │ dgx-spark     │ ● ONLINE │  88% │  63% │   54% │  74°C │  91% │  142W │                    ║
║   │ strix-halo    │ ● ONLINE │  40% │  35% │   22% │  52°C │  n/a │   88W │                    ║
║   │ mac-mini      │ ● ONLINE │  18% │  29% │   67% │  45°C │  38% │   22W │                    ║
║   │ rpi-5         │ ● ONLINE │  35% │  52% │   80% │  58°C │  n/a │   n/a │                    ║
║   │ alienware     │ ● ONLINE │  61% │  71% │   49% │  69°C │  77% │   96W │                    ║
║   └───────────────┴──────────┴──────┴──────┴───────┴───────┴──────┴───────┘                    ║
║                                                                                                ║
║   no alerts firing · header chip hidden · rows render in the normal palette                    ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   STATUS ● streaming · poll 2s · 0 alerts                                                      ║
║   q quit · ↑/↓ nav · ⏎ detail · a alerts · r refresh · ? help                                  ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

## State 2 — 2 alerts firing

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL │ ● 6/6 ONLINE │ ⚠ 2 alerts │ 🕐 14:32:40                                           ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌───────────────┬──────────┬──────┬──────┬───────┬───────┬──────┬───────┐                    ║
║   │ HOST          │ STATE    │ CPU% │ MEM% │ DISK% │  TEMP │ GPU% │   PWR │                    ║
║   ├───────────────┼──────────┼──────┼──────┼───────┼───────┼──────┼───────┤                    ║
║   │ workstation   │ ● ONLINE │  72% │  48% │   31% │  61°C │  64% │  110W │                    ║
║   │ 🔔 dgx-spark  │ ⚠ ONLINE │  88% │  63% │   54% │  94°C │  91% │  142W │                    ║
║   │ 🔔 strix-halo │ ⚠ ONLINE │  40% │  35% │   96% │  52°C │  n/a │   88W │                    ║
║   │ mac-mini      │ ● ONLINE │  18% │  29% │   67% │  45°C │  38% │   22W │                    ║
║   │ rpi-5         │ ● ONLINE │  35% │  52% │   80% │  58°C │  n/a │   n/a │                    ║
║   │ alienware     │ ● ONLINE │  61% │  71% │   49% │  69°C │  77% │   96W │                    ║
║   └───────────────┴──────────┴──────┴──────┴───────┴───────┴──────┴───────┘                    ║
║                                                                                                ║
║   ⚠ 2 alerts: dgx-spark TEMP 94°C (>90) · strix-halo DISK 96% (>95)                            ║
║   🔔 badge + alert-coloured row mark each firing host · chip counts the fleet total            ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   STATUS ● streaming · poll 2s · ⚠ 2 alerts firing                                             ║
║   q quit · ↑/↓ nav · ⏎ detail · a alerts · r refresh · ? help                                  ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

## Styles (lipgloss)

Each region maps to a role in `terminal-theme.json` (concrete dark-mode value). All colours come from `palette.json` — no drift.

| Region | Theme role | Foreground | Attrs | Border / BG |
|---|---|---|---|---|
| App frame | `borders.app / structure.border` | #585858 (240) | — | DoubleBorder ║═╔╗╚╝ |
| Title `⬢ HEIMDALL` | `structure.title` | #d0d0d0 steel (252) | bold | wordmark |
| Header `⬢` mark | `structure.accent` | #00d7ff electric-blue (45) | — | brand sigil |
| Online `● n/n ONLINE` | `states.online` | #5fd75f (77) | — | dot + count + word |
| Alert chip `⚠ n alerts` | `states.error` | #ff5f5f red (203) | bold | fleet alert total; hidden when 0 |
| Clock `🕐` / status bar | `structure.text_muted` | #949494 (246) | — | — |
| Panel surface / table | `structure.panel / borders.panel` | #585858 border | — | bg #1c1c1c (234), NormalBorder ┌─┐│└┘ |
| Column label row | `structure.label` | #949494 (246) | faint | — |
| Metric value (nominal) | `structure.value` | #eeeeee (255) | bold | — |
| Unit suffix (`%`,`°C`,`W`) | `structure.unit` | #949494 (246) | — | — |
| Alert badge `🔔` | `states.error` | #ff5f5f (203) | — | precedes host name |
| Firing host row | `states.error` | #ff5f5f red (203) | — | whole row recoloured; STATE shows `⚠` |
| Breached metric value | `severity.critical` | #ff5f5f (203) | bold | e.g. `94°C`, `96%` |
| Alert summary line | `states.error` | #ff5f5f (203) | — | metric · threshold breakdown |
| Footer keybinding | `structure.keybinding` | #eeeeee (255) | underline+bold | `a` opens alert list |
| Focused row `▸` | `structure.focus` | #00d7ff (45) | reverse | leading ▸ |

Alert colour ramp (threshold severity): warn `#ffaf00` (amber, 214) → critical `#ff5f5f` (red, 203).
A single firing alert uses critical red; lower-severity breaches use amber. The chip and row pick the
highest active severity.

## Accessibility

- **Alert signalling**: every firing host carries `🔔` + STATE `⚠` + an alert-coloured row + a named
  reason below the grid — never colour alone.
- **Chip**: `⚠ 2 alerts` is glyph + word + count; it disappears (not greys out) when the count is 0.
- **Focus**: reverse video + leading `▸` (`structure.focus`) — distinct from the alert colour.
- **Contrast**: alert red #ff5f5f on bg #1c1c1c is AA for UI/large text; values stay bold (WCAG 2.2 AA).
- **NO_COLOR / degraded**: colours collapse to terminal default; alerts persist via `🔔`, `⚠`, value, and the summary line.
- **Width note**: `🕐` and `🔔` render as 2-cell glyphs; the render budgets for them so borders stay aligned.
