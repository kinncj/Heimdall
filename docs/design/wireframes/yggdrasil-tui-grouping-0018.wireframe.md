---
id: yggdrasil-tui-grouping-0018
story: docs/stories/yggdrasil-tui-grouping-20260627203215-0018/Story.md
target: tui
status: approved
created_at: 2026-06-27
---

# Wireframe — yggdrasil-tui-grouping-0018

Target: **tui**. Adds in-dashboard grouping, filtering and search to the fleet grid. A status line
echoes the active `group` + `filter`; when grouped, host rows split into sections under dim
`── group (n) ──` headers. The `/` key opens a filter input; `g` cycles the group dimension
(hub → OS → tag key); `esc` clears. Low-fidelity ASCII; colour/styling is specified in the mockup.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded   ○ offline   ⏱ stale   ⚠ alert/error   ▸ focus (also inverse)
  ── group (n) ── = dim section header · / = filter input · g = cycle group dimension
Min width ≈ 80 cols: optional columns (TEMP/GPU/PWR) collapse first; HOST/STATE/CPU/MEM never drop.

## State 1 — grouped by origin hub (two groups, no filter)

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

**Annotations**

- Status line shows the active grouping + filter: `group: hub   filter: —`.
- Host rows are split into sections under dim headers: `── home (3) ──`, `── remote-work-station (1) ──`.
- Footer key hints add `g group · / filter · esc clear`. `g` cycles hub → OS → tag key.

## State 2 — filtering (input line active, term `prod`)

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

**Annotations**

- `/` opens the filter input at the bottom: `/ filter: prod▌` (`▌` = cursor). `⏎` applies, `esc` clears.
- Only hosts whose name or a tag value matches remain; grouping still applies to the survivors.
- Status line reflects the live filter: `group: hub   filter: prod`, count `2/4 hosts`.

## State 3 — empty (filter matches nothing)

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

**Annotations**

- A filter that matches no host shows an empty state inside the grid, not a blank screen.
- Grouping is preserved and re-applies the moment the filter is cleared with `esc`.
- Demo mode behaves identically (story scenario: grouping/filtering work in demo mode).
