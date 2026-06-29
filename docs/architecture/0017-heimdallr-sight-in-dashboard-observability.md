---
adr: "0017"
title: "Heimdallr's sight — hub-mediated control & log channel for in-dashboard observability"
status: accepted
date: "2026-06-29"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0017 — Heimdallr's sight — hub-mediated control & log channel for in-dashboard observability

## 1. Context

Heimdall ships two host-local capabilities behind the daemon's `--control-listen`
port: `ControlPlaneService` (read-only, allow-listed, audited commands —
`process.list`, `disk.df`, `dir.list`) and `LogStreamService` (tail declared log
sources). Today the dashboard reaches them by connecting **directly to the
daemon**: `heimdall-dashboard --control <host:port> --run process.list`.

That contradicts the deployment model the rest of the fleet follows. Per the
federation design (guide 05, ADR 0006):

> Only the hubs listen (`:9090`). Daemons and dashboards dial out and need no
> inbound ports.

Daemons are **outbound-only clients** — frequently behind NAT, on private LANs,
or on a tailnet where only hubs are addressable. So a dashboard cannot, in
general, open a connection to a daemon, and the hub cannot synthesize a reachable
daemon address (the peer IP is often NAT'd, and the daemon may bind `localhost`).
Direct `--control` works only on a flat LAN where the daemon happens to be
reachable.

To make logs and the control plane work **everywhere the fleet already works**,
they must be **mediated by the hub**, exactly like metrics: the daemon pushes
outbound, the dashboard subscribes from the hub.

## 2. Goals / Non-Goals

**Goals:**
- Logs and control reach the dashboard with **no inbound port on the daemon** and
  no addresses typed by the operator.
- The dashboard discovers which hosts expose logs / control and reads them from
  the host detail view (`l` logs, `t` top).
- Strictly backward compatible and additive on the wire (versioning): old daemons,
  old hubs, and hosts without the capability behave exactly as before.
- Bandwidth-safe: log volume cannot starve metrics.
- Cross-platform process listing (Linux, macOS, Windows), with a seam for the
  privileged helper to enrich it.

**Non-Goals:**
- Any write/mutate command — control stays read-only, allow-listed, audited.
- A new identity system — reuse the hub's existing enrollment-token / TLS auth.
- **The daemon acting as a server or holding a reverse request channel.** Today
  the daemon is a pure outbound producer; it never serves requests. A future
  "active socket" where a hub can ask a daemon on demand is explicitly out of
  scope (see §3.6).
- On-demand / interactive command execution against a daemon (follows from the
  above — there is no request path *to* the daemon today).
- Removing the existing direct `--control-listen` mode (kept for explicit LAN use).

## 3. Proposal

### 3.1 Daemon pushes observability data outbound (no server, no reverse channel)

The daemon stays a **pure outbound producer** — it never listens and never serves
a request. When configured, it adds two payloads to what it already pushes to the
hub on its existing outbound `MetricStreamService.Stream`:

- **Log lines.** It tails its `--log-source alias=path` files locally and pushes
  new lines to the hub (opt-in per source, rate-limited, bounded). This is a
  *push tail*, not a pull — there is no `LogStreamService` server on the daemon.
- **Process table.** It runs the allow-listed, OS-appropriate process command
  locally and pushes the table on a **slower, bounded cadence**
  (`--process-interval`, default off) as an inventory payload — not on every
  metric tick.

Both are additive fields on the snapshot envelope (or a sibling push), so old
hubs ignore them and the wire stays compatible (§3.7). The daemon decides what to
collect; nobody asks it.

### 3.2 Hub buffers per host

The hub extends its per-host store: the **latest** process table and a **bounded
ring** of recent log lines per host. This reuses the aggregation it already does
for snapshots — no new connection state, no request routing, no `hostID → channel`
map.

### 3.3 Dashboard reads from the hub

The dashboard talks **only to the hub** (its existing `Subscribe` family). The
subscription carries the latest process table with the snapshot; a hub RPC streams
the buffered + live log lines for a chosen host/source. The dashboard never
contacts a daemon, never holds a daemon address, and never needs a control token —
the hub is the trust boundary it already authenticates against.

### 3.4 Capability advertisement (additive labels)

The daemon advertises **which log aliases exist** via a reserved label
`_logs=app,sys` on the existing Realms label channel (additive; old peers ignore
it). Control availability is known to the hub from the open `Connect` channel, so
the hub stamps `_control=1` on the host. **No address or port is ever advertised**
— there is nothing for the dashboard to dial. Reserved `_`-prefixed labels are
filtered out of user-facing tags (grouping dimensions, filter scopes, detail
tags).

### 3.5 In-TUI modals (unchanged UX, hub data source)

From the host detail view, shown only when the host advertises the capability:
- **`l` — logs.** Lists the `_logs` aliases; `enter` selects one; the same modal
  streams that source's buffered + live lines **from the hub** into a scrollable
  viewport.
- **`t` — top.** Shows the latest process table the hub holds for that host and
  refreshes it as the daemon pushes new ones.
- **`esc`** unwinds one level (stream → list → detail), the app's universal back.

### 3.6 Cross-platform process collection + helper seam

The daemon collects the process table locally with an OS-appropriate command
(`ps -eo …` on Linux/macOS, `tasklist` on Windows) resolved by a per-OS lookup,
not a call-site branch. Collection goes through a `ProcessSource` abstraction
(DIP): daemon-local by default, and **helper-backed when the privileged helper is
present**, so `top` shows fuller detail on hosts that run `heimdall-helper`. The
helper path is additive. The same read-only allow-list still bounds what is run.

### 3.7 Bandwidth, gRPC, versioning

- **Bandwidth.** Pushing logs and a process table for every host always — even
  when nobody is watching — is the cost of having no reverse channel (§3.8). It is
  bounded by: opt-in per daemon (no `--log-source`/`--process-interval` ⇒ nothing
  pushed), per-source rate limiting on the tail, a slower process cadence than the
  metric tick, and a capped log ring at the hub.
- **gRPC.** No new connection: payloads ride the daemon's existing outbound
  `Stream`. The hub→dashboard log read is one server-stream RPC. No inbound
  listener anywhere but the hub.
- **Versioning.** New fields/messages are **additive** in `monitoring.v1`. Old
  daemon → pushes neither payload → host has no observability capability. Old hub →
  ignores the new fields → dashboard sees no capability and hides the affordance.
  No lockstep upgrade.

### 3.8 Removed: the daemon-as-server surface (the one breaking change)

The push model makes the daemon's listening control plane redundant and
contradictory, so it is **removed** — the single accepted backward-compatibility
break:

- Daemon flags removed: `--control-listen`, `--control-token`,
  `--control-tls-cert`, `--control-tls-key`. `--log-source` is **kept** but now
  configures what the daemon *pushes*, not a server.
- The daemon no longer registers `ControlPlaneService` / `LogStreamService`
  servers and opens no inbound port.
- Dashboard flags removed: `--control`, `--run`, `--tail` (the direct-to-daemon
  client). Logs and the process view live in the TUI, served by the hub.

### 3.9 Future (v2): everything over a persistent socket, hub still the hub

The push-only model is a deliberate v1. The intended v2 makes **every** link a
persistent, bidirectional socket, with the hub remaining the sole mediator — no
daemon ever listens, and no daemon ever talks to a dashboard directly:

```
daemon  ──▶ socket ◀──  HUB  ──▶ socket ◀──  dashboard
        (outbound)              (outbound)
```

- **daemon → socket ← hub.** The daemon holds a live outbound socket to the hub.
  Because it stays open, the hub can push requests *down* it on demand — run an
  allow-listed command now, start/stop tailing a source — instead of the daemon
  blindly pushing everything. This restores interactivity and removes the
  always-on bandwidth cost of v1, with the daemon still never acting as a server
  (it dialed out; the hub speaks over the daemon's own connection).
- **hub → socket ← dashboard.** The dashboard holds a live outbound socket to the
  hub and issues on-demand requests (`run X on host Y`, `tail Z`) that the hub
  routes to the owning daemon and streams back.
- **The hub is still THE HUB.** It remains the only listener, the only trust
  boundary, and the only thing either side connects to. Daemons and dashboards
  never discover or address each other.

v2 is out of scope for this ADR. v1 ships the push-only path; the wire stays
additive so v2 can layer request frames onto the same streams without another
break.

**Roadmap.** Once v1.6.0 ships, v2 begins on a `feature/sockets` branch and lands
as **v2.0.0** — persistent sockets end to end, further bandwidth reduction
(request-driven instead of always-on push), and the interactive features that the
socket unlocks (on-demand `top`, richer log control, and similar). The hub remains
the sole mediator throughout. A dedicated v2 ADR will supersede the relevant
sections here when that work starts.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Dashboard → daemon direct (status quo) | simplest | daemons are outbound-only / NAT'd; needs an inbound port | violates the "only hubs listen" model; flags removed in §3.8 |
| Hub synthesizes `peerIP:port` | no daemon change | peer IP often NAT'd/unreachable; daemon may bind localhost | not reachable in the general case |
| Daemon reverse channel (serve requests in reverse over an outbound bidi stream) | interactive, on-demand, no always-on push | the daemon holds a live request-serving connection — interactivity we are explicitly deferring | this **is** the v2 socket model (§3.9); deferred, not rejected |
| New top-level proto package | clean isolation | another schema + registration surface | additive fields in `monitoring.v1` suffice and stay version-compatible |

## 5. Trade-offs and Risks

- **Always-on push (the v1 cost).** With no reverse channel, a configured daemon
  pushes logs and a process table whether or not anyone is watching. Bounded by
  opt-in, rate limits, a slow process cadence, and a capped hub ring — but it is
  real bandwidth the v2 socket (§3.9) would reclaim.
- **No interactivity.** v1 cannot run a one-off command or an arbitrary
  allow-listed query on demand — the dashboard only sees what the daemon chose to
  push. On-demand execution waits for v2.
- **Freshness.** The process table is only as fresh as `--process-interval`; logs
  are live (push tail). The `t` modal shows "last updated" so a slow cadence is not
  mistaken for a hang.
- **Hub memory.** A bounded log ring + the latest process table per host add
  per-host state at the hub; sized to stay small and dropped when a host departs.
- **Breaking change.** Removing `--control-listen` / `--control` / `--run` /
  `--tail` (§3.8) breaks any script using them. Authorized as the single accepted
  break; called out in the changelog and guides.
- **Reserved-label leakage.** `_`-prefixed labels must be filtered everywhere user
  tags are shown, centrally.

## 6. Impact

**FinOps:** Bounded, opt-in push (logs + a slow process table) on the daemon's
existing stream. No new connections. Hub holds a small bounded buffer per host.

**SRE:** The daemon stops listening entirely — one fewer inbound surface and no
control server to crash. Pushed payloads ride the existing stream; the hub buffers
are bounded and dropped when a host departs. A daemon that goes offline simply
stops refreshing its logs/process table; the dashboard shows them as stale, never
hangs. No request routing, no per-request lifecycle.

**Security:** **Removes** an inbound surface — daemons no longer expose
`--control-listen`. The hub authenticates daemons and dashboards as today; the
process command stays read-only and allow-listed at the daemon, run on a timer
rather than on external request. The dashboard never learns a daemon address or
any daemon token. Log content now transits the hub, so hub trust covers it (same
as metrics).

**Team:** Two localized patterns — additive push payloads (daemon → hub → buffer)
and Bubble Tea modal sub-models reading from the hub. No reverse-routing machinery.

## 7. Decision

Make logs and the process view **push-based and hub-mediated**: the daemon, a pure
outbound producer, pushes tailed log lines and a periodic process table to the hub
on its existing stream; the hub buffers them per host; the dashboard reads them
from the hub and renders the `l` and `t` modals. The daemon never listens — its
control-server flags (`--control-listen` et al.) and the dashboard's direct-client
flags (`--control`/`--run`/`--tail`) are **removed**, the single accepted breaking
change. Capability is advertised as additive reserved labels (`_logs`, hub-stamped
`_control`), never an address. The wire stays additive so the v2 socket model
(§3.9) — persistent daemon↔hub↔dashboard sockets with the hub still the sole
mediator — can add on-demand interactivity later without another break.

Status: **accepted**
