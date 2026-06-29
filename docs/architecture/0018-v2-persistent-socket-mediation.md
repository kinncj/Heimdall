---
adr: "0018"
title: "v2 — persistent socket mediation (daemon ⇄ hub ⇄ dashboard)"
status: accepted
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

## 3a. Decided approach — **Kinn decided (2026-06-29)**

> Flagged as decided by Kinn: this is the approach to build on `feature/sockets`.

**Reuse the daemon's existing outbound bidirectional metric stream as the control
channel.** No new service, no new connection, and — critically — **no listener and
no open port on any host**.

`MetricStreamService.Stream` is already `rpc Stream(stream Snapshot) returns
(stream StreamControl)` — bidirectional. The daemon already dials it outbound and
sends `Snapshot`s; the hub can already send `StreamControl` back down it. v2:

1. **Hub → daemon directives** ride `StreamControl` (extended additively). The
   daemon adds a receive loop on the stream it already holds. It still **dials out
   and never listens** — the hub drives it over the daemon's own connection.
2. **Daemon → hub replies** ride the `Snapshot` stream (additive fields) or a new
   additive `result` message, correlated by `request_id`.
3. **Hub is the sole interface and trust boundary.** It maps `hostID → stream`,
   authorizes the dashboard, routes directives down, and fans replies up. Nothing
   ever connects *to* a daemon.

**Security model (the constraint driving this):**
- **Daemons run with the least privilege possible** and open **no inbound port** —
  the v1 "only hubs listen" rule is preserved exactly. The reverse channel is the
  daemon's own outbound socket, so no new attack surface is exposed on a host.
- **Privileged work is delegated to the helper**, which may run as root, over the
  **existing local unix socket** (the same channel used for privileged metrics).
  The unprivileged daemon never gains privilege; it asks the helper, which applies
  its own policy. A host with no helper simply cannot satisfy a privileged request
  (returns `insufficient_permission`).
- **The hub is the single door.** Operators secure one listener, not a fleet of
  them. No daemon port to firewall, MITM, or scan.

**Why not a separate `HubControlService` stream?** It would mean a second
long-lived connection per daemon for no security or routing benefit; the existing
bidi stream already carries the exact directions we need. Reuse wins.

### Phasing (each ships green on `feature/sockets`)

- **Phase 1 — demand-driven push.** A new `StreamControl` arm
  `ObservabilityWindow{logs, processes}` tells the daemon to push logs / a process
  table **only while the hub has a dashboard watching that host**. Reclaims the v1
  always-on bandwidth cost. Backward compatible: an old daemon ignores the arm and
  keeps pushing per config; an old hub sends no arm and the daemon defaults to its
  v1 behaviour.
- **Phase 2 — on-demand commands.** A `RunCommand{request_id, allowlisted_cmd,
  args, privileged}` arm. The daemon runs read-only allow-listed commands as itself
  (unprivileged); when `privileged`, it asks the **helper** over the unix socket.
  Results return correlated by `request_id`. The dashboard issues these to the hub,
  never to a daemon.

## 4. Open questions (non-blocking)

- Back-pressure and fairness when many dashboards subscribe to one busy daemon.
- Cross-hub (federated) routing of directives along the relay path (Phase 3).
- Helper command-policy surface (which privileged commands it will satisfy).

## 5. Decision

**Accepted (Kinn decided, 2026-06-29).** Build the v2 socket transport by reusing
the daemon's existing outbound bidi `MetricStreamService.Stream`: hub directives
ride `StreamControl` (additive), daemon replies ride the snapshot stream, the hub
mediates everything, and privileged work delegates to the helper over the local
unix socket. Daemons keep needing no inbound port. Ship in phases — demand-driven
push first, on-demand commands second — each additive and backward compatible on
the v1.6.0 wire.

Status: **accepted**
