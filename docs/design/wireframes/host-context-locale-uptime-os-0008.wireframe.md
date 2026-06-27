---
id: host-context-locale-uptime-os-0008
story: docs/stories/host-context-locale-uptime-os-20260626121705-0008/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — host-context-locale-uptime-os-0008

Target: **tui**. Each daemon reports host context (OS + arch, kernel, hostname, locale, timezone,
uptime, boot time, agent version, device class) on enrollment and refreshes it periodically.
The dashboard shows it in host detail and updates in place without re-registering the host.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Host detail — context panel

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Host context                ● dgx-spark online              ⏱ 2026-06-26 14:10:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ HOST CONTEXT — dgx-spark ────────────────────────────────────────────────────────────────┐ ║
║   │ os            Ubuntu 22.04.4 LTS          arch     x86_64                                │ ║
║   │ kernel        6.5.0-35-generic            hostname dgx-spark.local                       │ ║
║   │ locale        en_US.UTF-8                 timezone America/New_York (UTC-4)              │ ║
║   │ uptime        12d 04h 37m  (derived)      boot at  2026-06-14 09:26:11                   │ ║
║   │ agent         hwmon-daemon v1.4.2         class    server / DGX                          │ ║
║   │                                                                                          │ ║
║   │ shown on enroll · refreshed periodically while connected                                 │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Host detail shows OS+version, arch, kernel, hostname, locale, timezone, uptime (derived      ║
║   from boot time), agent version and a device-class label.                                     ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS context reported on enroll · refreshes while connected                                  ║
║ ⏎/esc back · ↑/↓ field · r refresh · ? help                                                    ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Key/value panel: OS+version, arch, kernel, hostname, locale, timezone, uptime (derived) + boot time, agent version and a device-class label.
- Context is captured on enroll and refreshed periodically while the host stays connected.
- Keyboard: ↑/↓ moves field focus, ⏎/esc returns to the grid — keyboard-reachable.

## Context update without re-registration

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Host context                ● dgx-spark online              ⏱ 2026-06-26 14:12:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ CONTEXT UPDATE — same host, no re-register ──────────────────────────────────────────────┐ ║
║   │ before  timezone America/New_York (UTC-4)   uptime 12d 04h 37m                           │ ║
║   │   …     daemon sends a context update (host-id h-3f9a unchanged)                         │ ║
║   │ after   timezone Europe/Lisbon (UTC+1)       uptime 12d 06h 02m                          │ ║
║   │                                                                                          │ ║
║   │ The dashboard reflects the new context in place. The host is NOT re-registered           │ ║
║   │ or duplicated — same row, same host-id; only the context fields change.                  │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS context updated in place · host-id h-3f9a · no duplicate                                ║
║ ⏎/esc back · ↑/↓ field · r refresh · ? help                                                    ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- A before/… /after strip shows timezone + uptime changing after the daemon sends a context update.
- The host-id is unchanged, so the dashboard updates the SAME row — no re-register, no duplicate.
- Only context fields change; the host's metric stream and position are unaffected.

