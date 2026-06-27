---
id: centralized-dashboard-federation-relay-0009
story: docs/stories/centralized-dashboard-federation-relay-20260626121705-0009/Story.md
target: tui
status: approved
created_at: 2026-06-26
---

# Wireframe — centralized-dashboard-federation-relay-0009

Target: **tui**. A hub can act as a relay: collecting local hosts and relaying upstream to a
cloud parent, while multiple dashboards subscribe concurrently. A topology strip shows this hub's
role, upstream link health and subscriber count; federated hosts are labelled by origin hub.

Legend (every state = symbol + text, never colour alone):
  ● online/ok   ◐ degraded/enrolling   ○ offline/no-internet   ⏱ stale   ⚠ error
  ⚿ needs helper (insufficient permission)   — unavailable (vendor/path)   ✓ installed   – absent
  ↑ relaying   ↯ reconnecting   ⚡ rate-limited   ✋ refused   ▸ focus (also inverse)
Min width ≈ 80 cols: optional columns (TREND/GPU/PWR/NET) collapse first; core metrics never drop.

## Topology strip + federated host labels

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Federation              ↑ relaying · 3 subs                 ⏱ 2026-06-26 14:13:00 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ FEDERATION — this hub: hq-hub (role: relay) ─────────────────────────────────────────────┐ ║
║   │ upstream  cloud-hub   ↑ relaying  (authenticated)                                        │ ║
║   │ downstream edge-hub   ● 2 hosts relayed in                                               │ ║
║   │ subscribers 3 dashboards attached (consistent view)                                      │ ║
║   │ role       relay  (collects local + relays upstream)                                     │ ║
║   │                                                                                          │ ║
║   │ loop / duplication prevention is enforced in the backend (note)                          │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Federated hosts are labelled by origin hub:                                                  ║
║                                                                                                ║
║   ┌───────────────┬────────────┬───────────┬───────┬──────────────────────┐                    ║
║   │ HOST          │ ORIGIN     │ CONN      │ CPU%  │ NOTE                 │                    ║
║   ├───────────────┼────────────┼───────────┼───────┼──────────────────────┤                    ║
║   │   workstation │ hq (local) │ ● ONLINE  │   72% │ collected locally    │                    ║
║   │   mac-mini    │ hq (local) │ ● ONLINE  │   18% │ collected locally    │                    ║
║   │ ▸ dgx-spark   │ edge (fed) │ ● ONLINE  │   88% │ relayed via edge-hub │                    ║
║   │   rpi-5       │ edge (fed) │ ● ONLINE  │   35% │ relayed via edge-hub │                    ║
║   └───────────────┴────────────┴───────────┴───────┴──────────────────────┘                    ║
║                                                                                                ║
║   hq (local) hosts are collected here; edge (fed) hosts are relayed in from edge-hub. Several  ║
║   dashboards can subscribe at once and see a consistent view.                                  ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS role=relay · upstream ↑ relaying · downstream edge-hub ● · subs 3                       ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- Strip shows role (relay), upstream link ↑ relaying, downstream edge-hub, and subscriber count.
- Hosts carry an ORIGIN label: hq (local) vs edge (fed = relayed in) so provenance is explicit.
- Multiple dashboards subscribe at once and see a consistent view; link health is symbol + text.

## Upstream reconnecting + re-auth (loop/dup prevention)

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ Heimdall · Federation            ↯ upstream reconnecting               ⏱ 2026-06-26 14:13:40 ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║                                                                                                ║
║   ┌ FEDERATION — this hub: hq-hub (role: relay) ─────────────────────────────────────────────┐ ║
║   │ upstream  cloud-hub   ↯ reconnecting (re-auth pending)                                   │ ║
║   │ downstream edge-hub   ● 2 hosts relayed in                                               │ ║
║   │ subscribers 3 dashboards attached (consistent view)                                      │ ║
║   │ role       relay  (collects local + relays upstream)                                     │ ║
║   │                                                                                          │ ║
║   │ loop / duplication prevention is enforced in the backend (note)                          │ ║
║   └──────────────────────────────────────────────────────────────────────────────────────────┘ ║
║                                                                                                ║
║   Cross-hub link drops: upstream ↯ reconnecting. On reconnect the parent re-authenticates      ║
║   this hub BEFORE accepting data; host data resumes without duplication or corruption.         ║
║   The relay does not create a loop between hubs (backend-enforced).                            ║
╠════════════════════════════════════════════════════════════════════════════════════════════════╣
║ STATUS upstream ↯ reconnecting · re-auth required before data resumes · no loop                ║
║ q quit · ↑/↓ nav · ⏎ detail · r refresh · ? help                                               ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

**Annotations**

- When the cross-hub link drops, upstream shows ↯ reconnecting (re-auth pending) — symbol + text.
- On reconnect the parent re-authenticates this hub before accepting data; data resumes without dup/corruption.
- Loop and duplication prevention is enforced in the backend — noted here, not a UI control.

