---
id: centralized-dashboard-federation-relay-0009
story: docs/stories/centralized-dashboard-federation-relay-20260626121705-0009/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Mockup — centralized-dashboard-federation-relay-0009

High-fidelity terminal render with the Heimdall brand applied (dark mode; light mirrors via
`docs/design/identity/terminal-theme.json`). Steel wordmark, electric-blue (#00d7ff) signature,
98-col double-border frame. Colour reinforces state that is always also glyph + word — NO_COLOR-safe.

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Federation              ↑ relaying · 3 subs                 ⏱ 2026-06-26 14:13:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ FEDERATION — this hub: hq-hub (role: relay) ─────────────────────────────────────────────┐ ║
║   │ upstream  cloud-hub   ↑ relaying  (authenticated)                                        │ ║
║   │ downstream edge-hub   ● 2 hosts relayed in                                               │ ║
║   │ subscribers 3 dashboards attached (consistent view)                                      │ ║
║   │ role       relay  (collects local + relays upstream)                                     │ ║
║   │                                                                                          │ ║
║   │ loop / duplication prevention is enforced in the backend (note)                          │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Federated hosts are labelled by origin hub:                                                  ║
║                                                                                                ║
║   ┌───────────────┬────────────┬───────────┬───────┬──────────────────────┐                    ║
║   │ HOST          │ ORIGIN     │ CONN      │ CPU%  │ NOTE                 │                    ║
║   ├───────────────┼────────────┼───────────┼───────┼──────────────────────┤                    ║
║   │   workstation │ hq (local) │ ● ONLINE  │   72% │ collected locally    │                    ║
║   │   mac-mini    │ hq (local) │ ● ONLINE  │   18% │ collected locally    │                    ║
║   │ ▸ dgx-spark   │ edge (fed) │ ● ONLINE  │   88% │ relayed via edge-hub │                    ║
║   │   rpi-5       │ edge (fed) │ ● ONLINE  │   35% │ relayed via edge-hub │                    ║
║   └───────────────┴────────────┴───────────┴───────┴──────────────────────┘                    ║
║                                                                                                ║
║   hq (local) hosts are collected here; edge (fed) hosts are relayed in from edge-hub. Several  ║
║   dashboards can subscribe at once and see a consistent view.                                  ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS role=relay · upstream ↑ relaying · downstream edge-hub ● · subs 3                       ║
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
| relaying ↑ (Bifröst) | `states.relaying` | #5fd7ff (81) | — | upstream link |
| reconnecting ↯ | `states.reconnecting` | #ffaf00 (214) | — | link drop |
| origin hub label | `structure.label` | #949494 (246) | faint | local vs edge |
| role badge | `structure.accent` | #00d7ff (45) | — | relay |

## Accessibility

- **Focus**: reverse video + leading `▸` (`structure.focus`, electric-blue) — never colour alone.
- **State signalling**: every status is glyph + word (`●○◐⏱⚠⚿—↑↯⚡✋`); colour only reinforces.
- **Contrast**: all pairs are AA-verified theme pairs (WCAG 2.2 AA; see `docs/design/identity/README.md`).
- **NO_COLOR / degraded**: colours collapse to terminal default; meaning persists via glyph, word, value, gauge density (`░▒▓█`), reverse focus.
