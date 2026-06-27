---
id: optional-privileged-metrics-helper-0005
story: docs/stories/optional-privileged-metrics-helper-20260626121705-0005/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — optional-privileged-metrics-helper-0005

Target: **tui**. Power and full thermal metrics need an OPTIONAL privileged helper. When it is
absent, those cells show a ⚿ needs-helper affordance (insufficient permission), NOT an error, and
the unprivileged daemon keeps running. When present, the cells populate over a local socket.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Needs-helper vs populated (both states side by side)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · privileged helper                ● both online              ⏱ 2026-06-26 14:07:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ helper – absent  (rpi-5) ─────────────────┐  ┌ helper ✓ installed  (dgx-spark) ──────────┐ ║
║   │ rpi-5            ● ONLINE                 │  │ dgx-spark        ● ONLINE                 │ ║
║   │ helper – absent                           │  │ helper ✓ installed                        │ ║
║   │                                           │  │                                           │ ║
║   │ cpu        35%   ● ok                     │  │ cpu        88%   ● ok                     │ ║
║   │ mem        52%   ● ok                     │  │ mem        63%   ● ok                     │ ║
║   │ temp(soc)  58°C  ● ok                     │  │ temp(soc)  74°C  ● ok                     │ ║
║   │ power      ⚿ needs helper                 │  │ power      142W  ● ok                     │ ║
║   │ temp(pkg)  ⚿ needs helper                 │  │ temp(pkg)  78°C  ● ok                     │ ║
║   │                                           │  │                                           │ ║
║   │ affordance, not an error                  │  │ privileged values populated               │ ║
║   └───────────────────────────────────────────┘  └───────────────────────────────────────────┘ ║
║                                                                                                ║
║   Power and full-thermal cells render ⚿ needs helper when the optional helper is absent,       ║
║   and populate when it is installed. Both daemons stay ● ONLINE throughout.                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ⚿ needs helper = insufficient permission (affordance, not error)                        ║
║ q quit · ↑/↓ nav · ⏎ detail · i install-helper info · ? help                                   ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Left rpi-5 (helper – absent): power and temp(pkg) render ⚿ needs helper; other metrics ● ok.
- Right dgx-spark (helper ✓ installed): the same privileged cells show real values (142W, 78°C).
- ⚿ needs helper is INSUFFICIENT_PERMISSION shown as a symbol+text affordance — never a red error.
- Both hosts stay ● ONLINE; the daemon does not crash when privileged metrics are unavailable.
- Keyboard: i surfaces install-helper info; ↑/↓ nav, ⏎ detail — keyboard-reachable.

## Per-host helper status (separate privileged unit, no sudo)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · privileged helper            ● 5 hosts online               ⏱ 2026-06-26 14:07:10 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   Per-host helper status. The daemon never invokes sudo or runs as root:                       ║
║                                                                                                ║
║   ┌───────────────┬─────────────┬─────────────────────┬──────────────────────────────┐         ║
║   │ HOST          │ HELPER      │ PRIVILEGED METRICS  │ CONNECTION (never sudo)      │         ║
║   ├───────────────┼─────────────┼─────────────────────┼──────────────────────────────┤         ║
║   │   workstation │ ✓ installed │ power · temp(pkg)   │ ● ONLINE · local socket      │         ║
║   │   dgx-spark   │ ✓ installed │ power · temp(pkg)   │ ● ONLINE · local socket      │         ║
║   │   mac-mini    │ ✓ installed │ gpu · power (Apple) │ ● ONLINE · local socket      │         ║
║   │ ▸ rpi-5       │ – absent    │ ⚿ needs helper      │ ● ONLINE · unprivileged      │         ║
║   │   strix-halo  │ – absent    │ ⚿ needs helper      │ ● ONLINE · unprivileged      │         ║
║   └───────────────┴─────────────┴─────────────────────┴──────────────────────────────┘         ║
║                                                                                                ║
║   The helper is a SEPARATE privileged unit; the unprivileged daemon reads privileged           ║
║   metrics over a LOCAL SOCKET. Absent helper ⇒ ⚿ needs helper, daemon keeps running.           ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS helper optional · daemon unprivileged · reads via local socket                          ║
║ q quit · ↑/↓ nav · ⏎ detail · i install-helper info · ? help                                   ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- A HELPER column shows ✓ installed vs – absent; PRIVILEGED METRICS lists what the helper unlocks.
- The helper runs as its own privileged unit; the daemon reads values over a LOCAL SOCKET and never sudo/root.
- Absent helper degrades to ⚿ needs helper (symbol+text) with the host still ● ONLINE.

