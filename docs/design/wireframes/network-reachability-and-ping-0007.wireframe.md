---
id: network-reachability-and-ping-0007
story: docs/stories/network-reachability-and-ping-20260626121705-0007/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — network-reachability-and-ping-0007

Target: **tui**. Each host probes a configured target + the public internet, reporting latency
and reachability. The dashboard shows reachability as symbol + text (● online / ◐ degraded /
○ no-internet) with a latency sparkline. A failed probe is isolated; the host stays online.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Per-host network section — reachability, latency, throughput

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Network                  ● online · 8ms                     ⏱ 2026-06-26 14:09:03 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ NETWORK — workstation ───────────────────────────────────────────────────────────────────┐ ║
║   │ reachability  ● online        target 1.1.1.1                                             │ ║
║   │ latency       8 ms            internet up                                                │ ║
║   │ trend (90s)   ▕▁▂▂▃▂▁▂▃▄▃▂▁▂▁▂▃▂▁▂▏ ms                                                   │ ║
║   │                                                                                          │ ║
║   │ throughput    ↓ 4.2 MB/s   ↑ 0.8 MB/s                                                    │ ║
║   │ probe         ok 14:09:02 (every 5s)                                                     │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Reachability shows BOTH a symbol and text (● online); latency carries a sparkline trend      ║
║   in ms; throughput shows ↓/↑. The ping target is explicit.                                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● online · 8ms · internet up · target 1.1.1.1                                           ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Reachability renders ● online with BOTH symbol and text; latency shows a value + 90s sparkline (ms).
- Throughput shows ↓ down / ↑ up rates and the explicit ping target; probe cadence is shown.
- Keyboard: ↑/↓ nav, ⏎ detail, r refresh — keyboard-reachable.

## Reachability states + isolated probe failure

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Network             ◐ 1 degraded · ○ 1 · ⚠ 1                ⏱ 2026-06-26 14:09:20 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   Reachability states are symbol + text; a failed probe is isolated:                           ║
║                                                                                                ║
║   ┌───────────────┬────────────────┬──────────┬────────────────┬──────────────────────┐        ║
║   │ HOST          │ REACHABILITY   │ LATENCY  │ TREND (ms)     │ THROUGHPUT / NOTE    │        ║
║   ├───────────────┼────────────────┼──────────┼────────────────┼──────────────────────┤        ║
║   │   workstation │ ● online       │      8ms │ ▁▂▂▃▂▁▂▃       │ ↓4.2 ↑0.8 MB/s       │        ║
║   │   dgx-spark   │ ● online       │     12ms │ ▂▂▃▃▄▃▂▃       │ ↓1.2 ↑0.3 MB/s       │        ║
║   │ ▸ strix-halo  │ ◐ degraded     │    180ms │ ▃▅▆▇█▆▅▆       │ high latency / loss  │        ║
║   │   rpi-5       │ ○ no-internet  │        — │ ────────       │ target unreachable   │        ║
║   │   alienware   │ ⚠ probe failed │        — │ ────────       │ ping ⚠ probe failed  │        ║
║   └───────────────┴────────────────┴──────────┴────────────────┴──────────────────────┘        ║
║                                                                                                ║
║   alienware ping ⚠ probe failed but the host stays ● ONLINE and its CPU/MEM/STO/TEMP keep      ║
║   updating — only the reachability metric is in an error state.                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ● 2 · ◐ degraded 1 · ○ no-internet 1 · ⚠ probe failed 1 (all hosts online)              ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- States are distinct symbol+word pairs: ● online, ◐ degraded (high latency/loss), ○ no-internet.
- A failed reachability probe shows ⚠ probe failed for that metric only; the host stays ● ONLINE.
- Other metrics keep updating normally — reachability failure does not take the host offline.
- ○ no-internet (network) is distinct from a host being ○ OFFLINE (no stream) — different columns.

