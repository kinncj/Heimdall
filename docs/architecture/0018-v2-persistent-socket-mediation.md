---
adr: "0018"
title: "v2 — persistent socket mediation (daemon ⇄ hub ⇄ dashboard)"
status: proposed
date: "2026-06-29"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0018 — v2: persistent socket mediation (daemon ⇄ hub ⇄ dashboard)

> Status: **proposed** — work happens on `feature/sockets` and ships as **v2.0.0**.
> This ADR seeds the design; it is not implemented yet. It builds on
> [ADR 0017](0017-heimdallr-sight-in-dashboard-observability.md) §3.9.

## 1. Context

v1.6.0 (ADR 0017) delivered host logs and a process view as a **push-only** model:
a daemon, a pure outbound producer, pushes log lines and a periodic process table
to the hub whether or not anyone is watching, and the dashboard reads from the hub.
That is correct and NAT-friendly, but it has two costs:

1. **Always-on bandwidth.** A configured daemon pushes continuously, even when no
   dashboard is looking at it.
2. **No interactivity.** Nothing can ask a daemon to do something *now* — run a
   one-off allow-listed command, start/stop tailing a specific source, raise the
   process cadence while a modal is open. The dashboard only sees what was pushed.

v2 removes both by making every link a **persistent, bidirectional socket**, while
keeping the rule that made v1 correct: **only the hub listens; daemons and
dashboards dial out.**

## 2. Goals / Non-Goals

**Goals:**
- On-demand interactivity: run an allow-listed command, tail a chosen source, or
  request a fresh process table — driven by the dashboard, routed by the hub.
- Bandwidth driven by demand: a daemon pushes a stream only while a dashboard is
  actually subscribed to it.
- Preserve the deployment model: daemons hold an **outbound** socket; they never
  listen. The hub remains the sole listener and trust boundary.
- Strictly additive on the v1.6.0 wire — no third breaking change.

**Non-Goals:**
- Daemon-to-dashboard direct connections (the hub always mediates).
- Re-introducing a listening daemon in any form.

## 3. Proposal (sketch)

```
daemon  ──▶ socket ◀──  HUB  ──▶ socket ◀──  dashboard
        (outbound)              (outbound)
```

- **Daemon control socket.** The daemon holds one long-lived outbound bidi stream
  to the hub. The hub sends request frames *down* it (run command, start/stop
  tail, set cadence); the daemon replies with response frames and streams (log
  lines, process tables) tagged by `request_id`. The daemon dials out and serves
  in reverse over its own connection — still not a server.
- **Hub routing.** The hub maps `hostID → control socket`, authorizes the
  dashboard, forwards requests to the owning daemon, correlates replies, and fans
  streams to subscribers. Per-request cancellation and teardown on disconnect.
- **Dashboard request socket.** The dashboard issues on-demand requests to the hub
  (`run X on host Y`, `tail Z`, `top now`) and consumes the streamed replies.
- **Demand-driven push.** A daemon tails / collects only while the hub has an open
  subscription for it, reclaiming the v1 always-on cost.

The frames layer onto the existing `monitoring.v1` streams as additive messages,
so a v2 hub still speaks v1 to old daemons/dashboards.

## 4. Open questions

- Frame envelope: extend `MetricStreamService.Stream` (already bidi) vs. a new
  `HubControlService` stream. Trade-off: reuse one connection vs. isolation.
- Auth granularity for on-demand command execution beyond the enrollment token.
- Back-pressure and fairness when many dashboards subscribe to one busy daemon.
- Cross-hub (federated) routing of requests along the relay path.

## 5. Decision

Deferred. Recorded here so v2 starts from a written design. Implementation proceeds
on `feature/sockets`; a future revision of this ADR moves it to **accepted** and
supersedes the relevant parts of ADR 0017 when v2.0.0 lands.

Status: **proposed**
