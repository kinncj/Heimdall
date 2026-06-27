---
id: gjallarhorn-dashboard-badge-0019
story: docs/stories/gjallarhorn-dashboard-badge-20260627203215-0019/Story.md
target: tui
status: approved
created_at: 2026-06-27
---

# Wireframe — gjallarhorn-dashboard-badge-0019

Target: **tui**. Surfaces firing alerts on the fleet grid. When any alert fires, the header shows a
`⚠ n alerts` chip in an alert colour. Each host with a firing alert gets a `🔔` badge next to its
name and its whole row renders in the alert colour. Low-fidelity ASCII; colour is specified in the
mockup.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded   ○ offline   ⏱ stale   ⚠ alert (threshold breach)   ▸ focus
  🔔 = host has a firing alert · header chip `⚠ n alerts` = fleet alert total
Min width ≈ 80 cols: optional columns (TEMP/GPU/PWR) collapse first; HOST/STATE/CPU/MEM never drop.

## State 1 — quiet (no alerts firing)

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

**Annotations**

- No alert chip in the header; every row renders in the normal palette.
- This is the baseline the alert state contrasts against.

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

**Annotations**

- Header gains a `⚠ 2 alerts` chip (alert colour) counting the whole fleet.
- `dgx-spark` (TEMP 94°C) and `strix-halo` (DISK 96%) each get a `🔔` badge and an alert-coloured row.
- The breached metric and threshold are spelled out below the grid; STATE shows `⚠` not just colour.
