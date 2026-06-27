---
id: yggdrasil-tui-grouping-0018
story: docs/stories/yggdrasil-tui-grouping-20260627203215-0018/Story.md
target: tui
status: approved
created_at: 2026-06-27
---

# Mockup — yggdrasil-tui-grouping-0018

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) chrome,
98-col double-border frame. Section headers and the filter/group status line are new; colour only
reinforces meaning that is always also glyph + word — NO_COLOR-safe.

## State 1 — grouped by origin hub

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL │ ● 4/4 ONLINE │ 🕐 14:22:08                                                        ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   group: hub      filter: —      search: off                                                   ║
║                                                                                                ║
║   ┌───────────────┬──────────┬──────┬──────┬───────┬───────┬──────┬───────┐                    ║
║   │ HOST          │ STATE    │ CPU% │ MEM% │ DISK% │  TEMP │ GPU% │   PWR │                    ║
║   ├───────────────┼──────────┼──────┼──────┼───────┼───────┼──────┼───────┤                    ║
║   │ home (3) ──────────────────────────────────────────────────────────── │                    ║
║   │ workstation   │ ● ONLINE │  72% │  48% │   31% │  61°C │  64% │  110W │                    ║
║   │ mac-mini      │ ● ONLINE │  18% │  29% │   67% │  45°C │  38% │   22W │                    ║
║   │ strix-halo    │ ● ONLINE │  40% │  35% │   22% │  52°C │  n/a │   88W │                    ║
║   │ remote-work-station (1) ───────────────────────────────────────────── │                    ║
║   │ dgx-spark     │ ● ONLINE │  88% │  63% │   54% │  74°C │  91% │  142W │                    ║
║   └───────────────┴──────────┴──────┴──────┴───────┴───────┴──────┴───────┘                    ║
║                                                                                                ║
║   grouped by origin hub · dim ── headers ── show group + count · g cycles dimension            ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   STATUS group=hub · filter=none · 4 hosts · 2 groups                                          ║
║   q quit · ↑/↓ nav · ⏎ detail · g group · / filter · esc clear · ? help                        ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

## State 2 — filtering (input line active)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL │ ● 4/4 ONLINE │ 🕐 14:22:08                                                        ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   group: hub      filter: prod      search: live                                               ║
║                                                                                                ║
║   ┌───────────────┬──────────┬──────┬──────┬───────┬───────┬──────┬───────┐                    ║
║   │ HOST          │ STATE    │ CPU% │ MEM% │ DISK% │  TEMP │ GPU% │   PWR │                    ║
║   ├───────────────┼──────────┼──────┼──────┼───────┼───────┼──────┼───────┤                    ║
║   │ home (1) ──────────────────────────────────────────────────────────── │                    ║
║   │ workstation   │ ● ONLINE │  72% │  48% │   31% │  61°C │  64% │  110W │                    ║
║   │ remote-work-station (1) ───────────────────────────────────────────── │                    ║
║   │ dgx-spark     │ ● ONLINE │  88% │  63% │   54% │  74°C │  91% │  142W │                    ║
║   └───────────────┴──────────┴──────┴──────┴───────┴───────┴──────┴───────┘                    ║
║                                                                                                ║
║   2 of 4 hosts match · matched on host name or tag value (tag: env=prod)                       ║
║                                                                                                ║
║   / filter: prod▌                                          ⏎ apply · esc clear                 ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   STATUS group=hub · filter="prod" · 2/4 hosts · 2 groups                                      ║
║   q quit · ↑/↓ nav · ⏎ detail · g group · / filter · esc clear · ? help                        ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

## State 3 — empty (no match)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL │ ● 4/4 ONLINE │ 🕐 14:22:08                                                        ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   group: hub      filter: ztag      search: live                                               ║
║                                                                                                ║
║   ┌───────────────┬──────────┬──────┬──────┬───────┬───────┬──────┬───────┐                    ║
║   │ HOST          │ STATE    │ CPU% │ MEM% │ DISK% │  TEMP │ GPU% │   PWR │                    ║
║   ├───────────────┼──────────┼──────┼──────┼───────┼───────┼──────┼───────┤                    ║
║   │                                                                       │                    ║
║   │    no hosts match "ztag"                                              │                    ║
║   │    try a different term, or press esc to clear the filter             │                    ║
║   │                                                                       │                    ║
║   └───────────────┴──────────┴──────┴──────┴───────┴───────┴──────┴───────┘                    ║
║                                                                                                ║
║   0 of 4 hosts match · grouping is preserved and re-applies when the filter clears             ║
║                                                                                                ║
║   / filter: ztag▌                                          ⏎ apply · esc clear                 ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   STATUS group=hub · filter="ztag" · 0/4 hosts · empty                                         ║
║   q quit · ↑/↓ nav · ⏎ detail · g group · / filter · esc clear · ? help                        ║
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
| Clock `🕐` / status bar | `structure.text_muted` | #949494 (246) | — | — |
| Group/filter status line | `structure.label` | #949494 (246) | faint | `group:`/`filter:` keys |
| Active filter value (`prod`) | `structure.accent` | #00d7ff (45) | — | reflects live term |
| Panel surface / table | `structure.panel / borders.panel` | #585858 border | — | bg #1c1c1c (234), NormalBorder ┌─┐│└┘ |
| Column label row | `structure.label` | #949494 (246) | faint | — |
| Section header `── group (n) ──` | `structure.subheading` | #b2b2b2 (249) | faint | dim full-width rule |
| Group count `(n)` | `structure.accent` | #00d7ff (45) | — | — |
| Metric value | `structure.value` | #eeeeee (255) | bold | — |
| Unit suffix (`%`,`°C`,`W`) | `structure.unit` | #949494 (246) | — | — |
| `n/a` cell | `structure.text_muted` | #949494 (246) | faint | unsupported metric |
| Filter input `/ filter: …▌` | `structure.input_active` | #eeeeee (255) | — | border_focus #00d7ff; `▌` caret blinks |
| Empty-state message | `structure.caption` | #949494 (246) | faint | centred in panel |
| Footer keybinding | `structure.keybinding` | #eeeeee (255) | underline+bold | `g` `/` `esc` highlighted |
| Focused row `▸` | `structure.focus` | #00d7ff (45) | reverse | leading ▸ |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **Grouping**: section breaks are glyph + word + count (`── home (3) ──`), readable without colour.
- **Filter state**: the active term is echoed as text in the status line and the input row — not colour.
- **Empty state**: explicit message + count (`0/4 hosts`), never a blank panel.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, count, value.
- **Width note**: `🕐` renders as a 2-cell glyph; the render budgets for it so the right border stays aligned.
