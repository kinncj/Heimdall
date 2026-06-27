---
id: unprivileged-terminal-control-plane-0010
story: docs/stories/unprivileged-terminal-control-plane-20260626121705-0010/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — unprivileged-terminal-control-plane-0010

Target: **tui**. A read-only, allow-listed control plane runs queries AS THE UNPRIVILEGED user.
An overlay offers an allow-list picker, a validated args field and a bounded result pane. A
persistent banner states the safety posture; escalation/non-allow-listed commands are refused.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Control overlay — allow-listed query + bounded result

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Control plane             ● dgx-spark · read-only           ⏱ 2026-06-26 14:14:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ CONTROL PLANE — dgx-spark  (read-only) ──────────────────────────────────────────────────┐ ║
║   │ command ▸ process.list      args ▏--sort=cpu --top=5▕   [⏎ run]                          │ ║
║   │         · disk.df           · fs.ls <allowed>  · net.ifaces                              │ ║
║   │                                                                                          │ ║
║   │ result (bounded) ─────────────────────────────────────────────                           │ ║
║   │   PID    USER     %CPU  %MEM  COMMAND                                                    │ ║
║   │   1042   svc-mon  18.2   3.1  hwmon-daemon                                               │ ║
║   │   2210   svc-mon   6.4   1.2  nvml-collector                                             │ ║
║   │   3398   svc-mon   2.1   0.8  ping-prober                                                │ ║
║   │   … truncated (showing 5 of 214 · bounded output)                                        │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   read-only · runs as svc-mon · no sudo · audited (every invocation logged)                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS allow-listed query process.list · result bounded · audit-logged                         ║
║ ↑/↓ pick · tab args · ⏎ run · esc close · keyboard-only · ? help                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Picker lists only allow-listed commands (process.list, disk.df, fs.ls <allowed>, net.ifaces); ▸ marks focus.
- Args is a validated field (▏ cursor = focus); ⏎ runs. Output is bounded with a … truncated indicator.
- Persistent banner: read-only · runs as svc-mon · no sudo · audited (every invocation logged).
- Keyboard-only: ↑/↓ pick, tab to args, ⏎ run, esc close — explicit focus markers, no mouse path.

## Refusal state — escalation / not allow-listed

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Control plane                 ✋ refused                    ⏱ 2026-06-26 14:14:22 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ CONTROL PLANE — refused ─────────────────────────────────────────────────────────────────┐ ║
║   │ command ▏ sudo systemctl restart hwmon ▕      [⏎ run]                                    │ ║
║   │                                                                                          │ ║
║   │ ✋ refused (not allow-listed / sudo)                                                     │ ║
║   │   reason  command is not on the read-only allow-list                                     │ ║
║   │   reason  escalation (sudo) is never permitted                                           │ ║
║   │   audit   refusal recorded · operator=alice · 14:14:22                                   │ ║
║   │                                                                                          │ ║
║   │ nothing was executed · no elevated privilege used                                        │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   read-only · runs as svc-mon · no sudo · audited (every invocation logged)                    ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS ✋ refused (not allow-listed / sudo) · audit entry written · nothing ran                ║
║ esc close · ↑/↓ pick allow-listed · ⏎ run · keyboard-only · ? help                             ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Trying sudo or a non-allow-listed command yields ✋ refused (not allow-listed / sudo) with reasons.
- Nothing is executed and no elevated privilege is used; the refusal itself is audit-logged with operator + time.
- ✋ refused is symbol + text; the safety banner remains visible.

