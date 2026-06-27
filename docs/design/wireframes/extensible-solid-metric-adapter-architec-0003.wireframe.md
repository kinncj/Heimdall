---
id: extensible-solid-metric-adapter-architec-0003
story: docs/stories/extensible-solid-metric-adapter-architec-20260626121705-0003/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — extensible-solid-metric-adapter-architec-0003

Target: **tui**. The dashboard viewed through the adapter architecture: each metric is an
independent adapter (SOLID). The UI isolates a single adapter's outcome to one cell and absorbs
new adapters as new columns. Three NON-OK outcomes are kept visually distinct.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Per-host adapter status — isolated failure

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · dgx-spark adapters            ● ONLINE · ⚠ 1 error          ⏱ 2026-06-26 14:03:09 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   Host dgx-spark · ● ONLINE · adapters: 4 registered, 3 ok, 1 failing                          ║
║                                                                                                ║
║   ┌──────────────────┬────────┬────────────┬────────┬──────────┬────────────────────────┐      ║
║   │ ADAPTER          │ SIGNAL │ STATUS     │ VALUE  │ LAST OK  │ NOTE                   │      ║
║   ├──────────────────┼────────┼────────────┼────────┼──────────┼────────────────────────┤      ║
║   │ cpu.adapter      │  CPU   │ ● ok       │    88% │ 14:03:08 │ nominal                │      ║
║   │ mem.adapter      │  MEM   │ ● ok       │    63% │ 14:03:08 │ nominal                │      ║
║   │ storage.adapter  │  STO   │ ● ok       │    54% │ 14:03:08 │ nominal                │      ║
║   │ temp.adapter     │  TEMP  │ ⚠ error    │     -- │ 14:01:55 │ sensor read timeout    │      ║
║   └──────────────────┴────────┴────────────┴────────┴──────────┴────────────────────────┘      ║
║                                                                                                ║
║   temp.adapter failed in isolation: only TEMP shows ⚠ error. CPU/MEM/STO keep streaming and    ║
║   the host stays ● ONLINE — one failure ≠ host failure.                                        ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● host online · ⚠ temp.adapter error (isolated) · 3/4 ok                                ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Adapter table: one row per adapter with SIGNAL/STATUS/VALUE/LAST OK; temp.adapter shows ⚠ error + reason.
- The other three stay ● ok and keep streaming; the host remains ● ONLINE.
- Status is glyph + word and the ⚠ cell keeps its column width so the table never reflows.

## Three distinct non-OK cell states (— / ⚿ / ⚠)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · adapter cell states           ● 3 hosts online              ⏱ 2026-06-26 14:03:30 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   Three NON-OK states are visually distinct (symbol + word), never colour:                     ║
║                                                                                                ║
║   ┌────────────────┬──────────────────┬───────────────────────────────────────┬──────────────┐ ║
║   │ STATE          │ CELL RENDERS     │ MEANING                               │ HOST STAYS   │ ║
║   ├────────────────┼──────────────────┼───────────────────────────────────────┼──────────────┤ ║
║   │ unavailable    │ — unavailable    │ vendor / platform path absent         │ ● ONLINE     │ ║
║   │ needs helper   │ ⚿ needs helper   │ insufficient permission (opt. helper) │ ● ONLINE     │ ║
║   │ error          │ ⚠ error          │ adapter ran but failed to read        │ ● ONLINE     │ ║
║   └────────────────┴──────────────────┴───────────────────────────────────────┴──────────────┘ ║
║                                                                                                ║
║   In the grid, each maps to its own metric cell:                                               ║
║                                                                                                ║
║   ┌───────────────┬───────────┬────────────────┬────────────────┬────────────┐                 ║
║   │ HOST          │ CONN      │ GPU%           │ PWR            │ TEMP       │                 ║
║   ├───────────────┼───────────┼────────────────┼────────────────┼────────────┤                 ║
║   │   dgx-spark   │ ● ONLINE  │ 91%            │ 142W           │    ⚠ error │                 ║
║   │   strix-halo  │ ● ONLINE  │ — unavailable  │ 88W            │       52°C │                 ║
║   │ ▸ rpi-5       │ ● ONLINE  │ — unavailable  │ ⚿ needs helper │       58°C │                 ║
║   └───────────────┴───────────┴────────────────┴────────────────┴────────────┘                 ║
║                                                                                                ║
║   All three hosts stay ● ONLINE; only the affected cell changes. — vs ⚿ vs ⚠ are different     ║
║   glyph+word pairs so they are told apart on monochrome terminals.                             ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS — unavailable · ⚿ needs helper · ⚠ error  (all isolated, hosts online)                  ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Taxonomy: — unavailable (vendor/path), ⚿ needs helper (insufficient permission, story 0005), ⚠ error (read failed).
- Each is a different symbol + word pair, so they are distinguishable without colour and on NO_COLOR.
- All three are isolated to their cell; every host stays ● ONLINE — a missing capability is an affordance, not a crash.
- Focus ▸ on rpi-5 shows two non-OK cells (GPU — unavailable, PWR ⚿ needs helper) coexisting.

## New adapter → new column (Open/Closed)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · adapters               ● 7/7 · +gpu loaded                  ⏱ 2026-06-26 14:06:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   BEFORE — registry: [cpu, mem, sto, temp]                                                     ║
║   ┌────────────────┬─────────┬─────────┬─────────┬─────────┐                                   ║
║   │ HOST           │ CPU%    │ MEM%    │ STO%    │ TEMP    │                                   ║
║   ├────────────────┼─────────┼─────────┼─────────┼─────────┤                                   ║
║   │   dgx-spark    │     88% │     63% │     54% │    74°C │                                   ║
║   │   alienware    │     61% │     71% │     49% │    69°C │                                   ║
║   └────────────────┴─────────┴─────────┴─────────┴─────────┘                                   ║
║                                                                                                ║
║   AFTER  — registry: [cpu, mem, sto, temp, +gpu]                                               ║
║   ┌────────────────┬─────────┬─────────┬─────────┬─────────┬───────────┐                       ║
║   │ HOST           │ CPU%    │ MEM%    │ STO%    │ TEMP    │ GPU% ◀new │                       ║
║   ├────────────────┼─────────┼─────────┼─────────┼─────────┼───────────┤                       ║
║   │   dgx-spark    │     88% │     63% │     54% │    74°C │       91% │                       ║
║   │   alienware    │     61% │     71% │     49% │    69°C │       77% │                       ║
║   └────────────────┴─────────┴─────────┴─────────┴─────────┴───────────┘                       ║
║                                                                                                ║
║   Adding gpu.adapter appends a GPU% column on the right; existing columns/values are           ║
║   untouched (Open/Closed) — no edits to cpu/mem/sto/temp adapters.                             ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS +gpu.adapter registered · existing adapters unchanged                                   ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Registering gpu.adapter appends a GPU% column flagged ◀new for one render; existing columns are unchanged.
- Extension, not modification — no existing adapter is edited (Open/Closed).
- New/optional columns append at the right and collapse first at narrow widths, protecting core metrics.

