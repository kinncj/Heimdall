---
id: opt-in-log-streaming-0011
story: docs/stories/opt-in-log-streaming-20260626121705-0011/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — opt-in-log-streaming-0011

Target: **tui**. Logs are opt-in and travel on a SEPARATE stream from metrics. The logs pane tails
live lines per host (source + level + timestamp) and is rate-limited on the low-bandwidth link.
With no source configured, logs stay off — no lines are streamed.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Logs pane — live, rate-limited

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Logs                 ● live · ⚡ rate-limited               ⏱ 2026-06-26 14:15:06 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ LOGS — dgx-spark  (live · opt-in · separate stream) ─────────────────────────────────────┐ ║
║   │ source /var/log/syslog        ⚡ rate-limited (37 dropped)                               │ ║
║   │ ────────────────────────────────────────────────────────────                             │ ║
║   │ 14:15:02  INFO  systemd      Started hwmon-daemon.service                                │ ║
║   │ 14:15:03  WARN  nvml         GPU1 ECC retired page detected                              │ ║
║   │ 14:15:04  INFO  hwmon        stream resumed (offset 0x1f3a)                              │ ║
║   │ 14:15:05  ERR   ping-prober  target 1.1.1.1 timeout (isolated)                           │ ║
║   │ 14:15:06  INFO  hwmon        poll 2s · 6 hosts live                                      │ ║
║   │ ▸ tailing… (newest at bottom)                                                            │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Live tailed lines per host carry source + level + timestamp. The log stream is SEPARATE      ║
║   from metrics and is rate-limited; ⚡ rate-limited (N dropped) signals drops on low-bw link.  ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS logs live · source=/var/log/syslog · ⚡ rate-limited (37 dropped)                       ║
║ ↑/↓ scroll · f follow · / filter level · esc close · keyboard-only · ? help                    ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Live tailed lines per host with source label, level (INFO/WARN/ERR) and timestamp; newest at bottom.
- The log stream is independent of the metric stream and is rate-limited; ⚡ rate-limited (N dropped) shows drops.
- Keyboard-only: ↑/↓ scroll, f follow, / filter by level, esc close — ▸ marks the tailing position.

## OFF state — opt-in, no source configured

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Logs                    logs: off (opt-in)                  ⏱ 2026-06-26 14:15:30 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║   Logs are OFF until a source is explicitly configured (opt-in):                               ║
║                                                                                                ║
║   ┌───────────────┬──────────────┬───────────────────┬──────────────────────────┐              ║
║   │ HOST          │ LOGS         │ SOURCE            │ NOTE                     │              ║
║   ├───────────────┼──────────────┼───────────────────┼──────────────────────────┤              ║
║   │   dgx-spark   │ ● live       │ /var/log/syslog   │ opt-in · ⚡ rate-limited │              ║
║   │ ▸ rpi-5       │ off (opt-in) │ — none configured │ no lines streamed        │              ║
║   │   mac-mini    │ off (opt-in) │ — none configured │ no lines streamed        │              ║
║   └───────────────┴──────────────┴───────────────────┴──────────────────────────┘              ║
║                                                                                                ║
║   rpi-5 has no log source → logs: off (opt-in); no log lines are streamed for that host.       ║
║   Metrics keep streaming regardless — logs are a separate, opt-in stream.                      ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS logs off until a source is configured · metric stream unaffected                        ║
║ ↑/↓ nav · ⏎ configure-source info · esc close · keyboard-only · ? help                         ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Hosts with no configured source show logs: off (opt-in); no log lines are streamed for them.
- Logging stays off until a source is explicitly configured — opt-in, not default-on.
- Metric streaming is unaffected; logs are a separate opt-in stream (off ≠ host problem).

