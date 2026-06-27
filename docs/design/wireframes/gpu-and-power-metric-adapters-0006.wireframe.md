---
id: gpu-and-power-metric-adapters-0006
story: docs/stories/gpu-and-power-metric-adapters-20260626121705-0006/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — gpu-and-power-metric-adapters-0006

Target: **tui**. GPU adapters (NVIDIA via NVML, Apple Silicon via the platform path) report
util / VRAM / temperature / power, with btop-style gradient gauges and up to 2 GPUs. Power is
read-only. Unsupported vendors (AMD / Raspberry Pi) degrade to — unavailable without crashing.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## GPU & power detail — 2 GPUs, gradient gauges, read-only profile

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · GPU & Power              ● dgx-spark online                 ⏱ 2026-06-26 14:08:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ GPU & POWER — dgx-spark (NVML, 2 GPUs) ──────────────────────────────────────────────────┐ ║
║   │ GPU0  RTX 6000        GPU1  RTX 6000                                                     │ ║
║   │ util ██████████████▓▒░░ 91%   util ██████████▓▒░░░░░░ 64%                                │ ║
║   │ vram ███████████▓▒░░░░░ 28/40  vram ██████▓▒░░░░░░░░░░ 18/40                             │ ║
║   │ temp 71°C ↑           temp 63°C →                                                        │ ║
║   │ pwr  142W             pwr  104W                                                          │ ║
║   │                                                                                          │ ║
║   │ POWER (read-only)     package 246W   cpu 96W   gpu 246W                                  │ ║
║   │ profile balanced      ▸ no control offered to change profile                             │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   btop/mactop-style gradient gauges █▓▒░ for util and VRAM; two GPUs shown side by side.       ║
║   Power profile is displayed READ-ONLY — the dashboard offers no control to change it.         ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS GPU via NVML · 2 GPUs · power read-only (profile balanced)                              ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Two GPUs shown side by side with gradient util/VRAM gauges, temperature (↑/→) and power draw.
- POWER section reports package/cpu/gpu watts and a profile that is READ-ONLY — no control is offered.
- Apple Silicon hosts collect GPU/power via the platform path; the daemon itself stays unprivileged.
- Gauges print exact numbers (28/40GB, 142W) so data survives monochrome/NO_COLOR.

## Cross-vendor summary — graceful — unavailable

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · GPU & Power             ● 5 hosts · 2 unsupported           ⏱ 2026-06-26 14:08:10 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   GPU/power across vendors — unsupported vendors degrade gracefully:                           ║
║                                                                                                ║
║   ┌───────────────┬───────────────┬───────────┬──────────────┬────────────┬──────────────────┐ ║
║   │ HOST          │ GPU VENDOR    │ UTIL%     │ VRAM         │ GPU °C     │ POWER / NOTE     │ ║
║   ├───────────────┼───────────────┼───────────┼──────────────┼────────────┼──────────────────┤ ║
║   │   dgx-spark   │ NVIDIA (NVML) │       91% │      28/40GB │       71°C │ 142W             │ ║
║   │   alienware   │ NVIDIA (NVML) │       77% │       9/16GB │       68°C │ 96W              │ ║
║   │ ▸ mac-mini    │ Apple Silicon │       38% │       6/16GB │       49°C │ 22W (via helper) │ ║
║   │   strix-halo  │ AMD (unsupp.) │         — │            — │          — │ — unavailable    │ ║
║   │   rpi-5       │ Broadcom (Pi) │         — │            — │          — │ — unavailable    │ ║
║   └───────────────┴───────────────┴───────────┴──────────────┴────────────┴──────────────────┘ ║
║                                                                                                ║
║   NVIDIA via NVML; Apple Silicon via the platform path (needs the helper, no daemon sudo).     ║
║   Unsupported AMD / Raspberry Pi report — unavailable; the daemon keeps collecting the rest.   ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS NVML · Apple path · — unavailable for AMD/Pi (no crash)                                 ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- NVIDIA (NVML) and Apple Silicon report full values; AMD and Raspberry Pi show — unavailable.
- — unavailable is text (symbol+word), distinct from ⚿ needs helper and ⚠ error (story 0003).
- When a GPU is unavailable the daemon keeps collecting CPU/MEM/STO/TEMP and stays ● ONLINE.

