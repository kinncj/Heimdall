---
id: real-time-centralized-go-tui-dashboard-0002
story: docs/stories/real-time-centralized-go-tui-dashboard-20260626121705-0002/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — real-time-centralized-go-tui-dashboard-0002

Target: **tui**. Global shell = header (title · online count · clock), body (host grid) and
footer (status + keys). Grid now carries GPU%, PWR and NET (reachability) columns plus an
ORIGIN (federation hub) label; full GPU/power/network detail and the logs pane live one level in.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Global shell — live grid (CPU/MEM/STO/TEMP/GPU/PWR/NET)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Remote Hardware Monitor           ● 6/7 online              ⏱ 2026-06-26 14:03:12 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌───────────────┬────────┬───────────┬──────┬──────┬──────┬───────┬──────┬───────┬─────────┐ ║
║   │ HOST          │ ORIGIN │ CONN      │ CPU% │ MEM% │ STO% │ TEMP  │ GPU% │ PWR   │ NET     │ ║
║   ├───────────────┼────────┼───────────┼──────┼──────┼──────┼───────┼──────┼───────┼─────────┤ ║
║   │ ▸ workstation │ local  │ ● ONLINE  │  72% │  48% │  31% │  61°C │  64% │  110W │ ● 8ms   │ ║
║   │   dgx-spark   │  edge  │ ● ONLINE  │  88% │  63% │  54% │  74°C │  91% │  142W │ ● 12ms  │ ║
║   │   strix-halo  │ local  │ ● ONLINE  │  40% │  35% │  22% │  52°C │  n/a │   88W │ ● 9ms   │ ║
║   │   mac-mini    │ local  │ ● ONLINE  │  18% │  29% │  67% │  45°C │  38% │   22W │ ● 15ms  │ ║
║   │   rpi-5       │  edge  │ ● ONLINE  │  35% │  52% │  80% │  58°C │  n/a │   n/a │ ● 24ms  │ ║
║   │   alienware   │ local  │ ● ONLINE  │  61% │  71% │  49% │  69°C │  77% │   96W │ ● 20ms  │ ║
║   └───────────────┴────────┴───────────┴──────┴──────┴──────┴───────┴──────┴───────┴─────────┘ ║
║                                                                                                ║
║   GPU% n/a = unsupported vendor (story 0006) · PWR n/a = needs privileged helper (0005)        ║
║   NET = reachability symbol + latency (0007) · ORIGIN = federation hub (0009, edge = relayed). ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● streaming · gRPC low-bw · poll 2s · 1 host off-screen (nuc-lab ○)                     ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · / filter · L logs · ? help                           ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- One row per host with core metrics plus GPU%, PWR (watts) and NET (● + latency).
- ORIGIN tags federation: local hosts vs edge = relayed from a child hub (story 0009).
- Graceful gaps use text, not blank: GPU% n/a (unsupported vendor) and PWR n/a (needs helper).
- Focus ▸ + inverse on workstation. Keyboard adds L (logs pane) to nav/detail/refresh/filter/help.
- Min width: GPU/PWR/NET collapse before core metrics; nothing is lost, only hidden.

## Host detail — GPU, power, network, logs entry

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Remote Hardware Monitor           ● 6/7 online              ⏱ 2026-06-26 14:03:12 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ HOST DETAIL — dgx-spark (origin: edge) ──────────────────────────────────────────────────┐ ║
║   │ dgx-spark · ● ONLINE · linux x86_64 · 10.0.1.32      last update 14:03:12 (0s ago) · live│ ║
║   │                                                                                          │ ║
║   │ CPU   88%  ▕██████████████████████████████▓▒░░░░▏ 8c/16t   MEM 63% ▕██████▓▒░░░░▏ 41/64GB│ ║
║   │ STO   54%  ▕█████████████████▓▒░░░░░░░░░░░░░░░░░▏ 1.08TB   TEMP 74°C ↑ rising            │ ║
║   │                                                                                          │ ║
║   │ GPU   91%  ▕███████████████████████████████▓▒░░░▏ VRAM 28/40GB · 71°C · 142W  (NVML)     │ ║
║   │ PWR   pkg 142W · cpu 96W · gpu 142W · profile balanced (read-only)                       │ ║
║   │ NET   ● online · 12ms ▕▁▂▂▃▂▁▂▃▄▃▂▁▂▏ · ↓ 1.2MB/s ↑ 0.3MB/s · target 1.1.1.1             │ ║
║   │                                                                                          │ ║
║   │ LOGS  L → open logs pane (opt-in; off unless a source is configured · story 0011)        │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ DETAIL ● live · 90s trends · CPU/MEM/STO/TEMP/GPU/PWR/NET                                      ║
║ ⏎/esc back · ←/→ host · L logs · r refresh · ? help                                            ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- ⏎ opens detail: 90s CPU/MEM/STO/TEMP trends PLUS a GPU gauge (util/VRAM/temp/power via NVML), a POWER breakdown (pkg/cpu/gpu watts + read-only profile) and a NET line (reachability + latency sparkline + throughput ↑/↓ + ping target).
- Power profile is read-only — no control is offered to change it (story 0006).
- L opens the per-host logs pane (opt-in; story 0011). Bars print exact values for monochrome.

## Stale & offline states (unchanged behaviour)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Remote Hardware Monitor      ● 4/7 · ⏱ 1 stale · ○ 2 off    ⏱ 2026-06-26 14:05:31 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌────────────────┬────────────────┬───────┬───────┬───────┬────────┬──────────────────────┐  ║
║   │ HOST           │ CONN           │ CPU%  │ MEM%  │ STO%  │ TEMP   │ LAST UPDATE          │  ║
║   ├────────────────┼────────────────┼───────┼───────┼───────┼────────┼──────────────────────┤  ║
║   │ ▸ dgx-spark    │ ⏱ STALE        │   88% │   63% │   54% │   74°C │ 14:03:12  (2m ago)   │  ║
║   │   nuc-lab      │ ○ OFFLINE      │    -- │    -- │    -- │     -- │ 13:48:55  (17m ago)  │  ║
║   │   workstation  │ ● ONLINE       │   72% │   48% │   31% │   61°C │ 14:05:30  (live)     │  ║
║   └────────────────┴────────────────┴───────┴───────┴───────┴────────┴──────────────────────┘  ║
║                                                                                                ║
║   ⏱ STALE row is dimmed; last-known values are RETAINED with their age. ○ OFFLINE clears       ║
║   values and shows the last update time so stale data is never mistaken for live.              ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ⏱ 1 stale · ○ 2 offline · 4 live                                                        ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · / filter · L logs · ? help                           ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- ⏱ STALE (dimmed, last-known values + age), ○ OFFLINE (values cleared, last-update time), ● ONLINE (live).
- State is symbol + word + timestamp — never colour/dim alone. LAST UPDATE disambiguates staleness.
- Reachability ○ no-internet (story 0007) is distinct from host ○ OFFLINE — see the NET column/detail.

## Help overlay (?)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Remote Hardware Monitor           ● 6/7 online              ⏱ 2026-06-26 14:03:20 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   (grid dimmed behind overlay)                                                                 ║
║                                                                                                ║
║   ┌ ? HELP — keybindings ────────────────────────────────────────────────────────────────────┐ ║
║   │ q / Ctrl-C  quit           ↑ / k  up            r       refresh now                      │ ║
║   │ ⏎ / l       open detail     ↓ / j  down          /       filter hosts                    │ ║
║   │ L           logs pane       v      cycle fidelity g / G   top / bottom                   │ ║
║   │ esc         close overlay   ?      toggle help    : every action keyboard-reachable      │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS help open · grid paused                                                                 ║
║ esc/?  close · q quit                                                                          ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- ? toggles a centred modal listing every binding (now including L logs); esc/? closes it.
- Overlay traps focus and pauses the grid; focus returns to the prior row on close.
- Confirms there is no mouse-only path (keyboard-reachable).

## Filter bar (/) — by origin

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Remote Hardware Monitor           ● 6/7 online              ⏱ 2026-06-26 14:03:25 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   / filter ▏ origin:edge▕                     (2 of 7 match · esc clears)                      ║
║                                                                                                ║
║   ┌───────────────┬────────┬───────────┬──────┬──────┬──────┬───────┬──────┬───────┬─────────┐ ║
║   │ HOST          │ ORIGIN │ CONN      │ CPU% │ MEM% │ STO% │ TEMP  │ GPU% │ PWR   │ NET     │ ║
║   ├───────────────┼────────┼───────────┼──────┼──────┼──────┼───────┼──────┼───────┼─────────┤ ║
║   │   dgx-spark   │  edge  │ ● ONLINE  │  88% │  63% │  54% │  74°C │  91% │  142W │ ● 12ms  │ ║
║   │   rpi-5       │  edge  │ ● ONLINE  │  35% │  52% │  80% │  58°C │  n/a │   n/a │ ● 24ms  │ ║
║   └───────────────┴────────┴───────────┴──────┴──────┴──────┴───────┴──────┴───────┴─────────┘ ║
║                                                                                                ║
║   Filter is a live text field; results update per keystroke. Here origin:edge shows only       ║
║   hosts relayed from the edge hub (federation, story 0009).                                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS filter active: origin:edge · 2 match                                                    ║
║ type to filter · esc clear · ⏎ keep · ↑/↓ nav                                                  ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- / opens a live text field; origin:edge filters to federated hosts and shows a match count.
- The cursor ▏ marks the active input (focus-visible); esc clears, ⏎ keeps, ↑/↓ navigate results.
- Empty results would render '0 of 7 match — adjust filter' rather than a blank pane.

