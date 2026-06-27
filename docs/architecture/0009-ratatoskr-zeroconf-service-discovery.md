---
adr: "0009"
title: "Ratatoskr: zeroconf service discovery"
status: proposed
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0009 — Ratatoskr: zeroconf service discovery

> **Ratatoskr** /ˌrɑː.təˈtɒs.kər/ (*rah-tuh-TOSK-er*) — the squirrel that carries
> messages up and down Yggdrasil. See the [glossary](../glossary.md).

## 1. Context

Today every daemon needs an explicit `--hub` address. On a LAN, a homelab, or an
overlay network this is friction: operators hand-configure each host even though
the hub is reachable and announceable. We want daemons to auto-find their hub
without changing the connection model. Heimdall's invariant is that daemons and
dashboards make **outbound** connections and hosts expose **no inbound port** except
the optional control plane (ADR-0007). Discovery must preserve that: it may only
tell a daemon *where* the hub is — the hub must never dial a daemon to ingest.

Discovery also crosses trust boundaries. A daemon that auto-connects to whatever
answers a multicast query is trivially hijacked. Discovery answers *location*, not
*trust*; the existing token (gRPC metadata `x-heimdall-token`) and TLS (ADR-0002,
`secure/tls.go`) must still gate enrollment.

## 2. Goals / Non-Goals

**Goals:**
- A daemon discovers its hub and makes the **same outbound gRPC connection** it
  makes today — no inbound port on the host.
- Pluggable discovery strategies that match the network actually in use (LAN,
  overlay, static seed).
- Optional daemon advertisement so a hub/dashboard can **enumerate**
  discovered-but-not-yet-streaming hosts (visibility only).
- Token + TLS trust is never bypassed by auto-connect.

**Non-Goals:**
- The hub dialing daemons to pull metrics (violates the outbound-only model).
- A global service registry or external discovery dependency (Consul/etcd).
- Replacing explicit `--hub` — discovery is additive and opt-in.

## 3. Proposal

**Hub advertises, daemon discovers.** The hub publishes a service record
(`_heimdall-hub._tcp`) carrying its host/port (and TLS hint). The daemon runs a
`Discoverer`, resolves the hub endpoint, and opens the **identical** outbound stream
it opens for an explicit `--hub`. Discovery only produces a candidate address; the
ingest path is unchanged.

**Daemon flags.** Add `--discover` (alias: `--hub auto`) and a `discoverable` JSON
key, following the layered config system (flags > env > file > defaults,
`HEIMDALL_CONFIG_DIR`), with an optional first-run wizard `Ask` prompt. Explicit
`--hub` always wins over discovery.

**Optional daemon advertisement.** A daemon may advertise `_heimdall-daemon._tcp`.
A hub/dashboard can **enumerate** these to show hosts that exist on the network but
have not yet enrolled/streamed. This is strictly visibility — the hub never dials a
daemon. It closes the "I installed it but don't see it" gap.

**Pluggable `Discoverer` strategy (SOLID seam).** A single interface, multiple
providers, run whichever match the available networks:

- **mDNS provider** — LAN multicast (`_heimdall-hub._tcp`). Pure-Go mDNS library.
- **Overlay provider** — Tailscale/overlay enumeration. mDNS cannot traverse
  overlays (no multicast), so this provider resolves peers via the overlay's own
  API/host list.
- **Static seed-list provider** — a configured set of candidate hubs; the trivial
  fallback and the only option in locked-down networks.

DIP: the daemon depends on the `Discoverer` abstraction, not on mDNS. ISP: the
interface is `Discover(ctx) ([]Endpoint, error)` — providers implement only that.
OCP: a new network type is a new provider, no change to the daemon.

**Trust gate (unchanged).** After discovery yields an endpoint, the daemon still
presents the enrollment token over TLS. Auto-connect **refuses an untokened or
untrusted hub by default**; discovery never downgrades trust. A discovered endpoint
with no configured token is surfaced, not auto-enrolled.

**New dependency.** A pure-Go mDNS library (no CGO — linux/windows builds stay
CGO-free; ADR rationale). Overlay and seed providers add no new runtime dependency.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Keep mandatory `--hub` | Zero new deps; explicit | Manual config per host; poor first-run UX | Doesn't solve the friction |
| Daemon advertises, hub dials it | Hub-driven inventory | Requires inbound port on host; hub→daemon ingest | Breaks outbound-only model (ADR-0007) |
| External registry (Consul/etcd) | Mature, cross-subnet | Heavy stateful dependency; ops + cost | Over-scoped; violates lightweight posture |
| Single mDNS-only impl | Simplest code | Fails on overlays (no multicast) and locked networks | Too narrow for real topologies |
| Pluggable `Discoverer` + token/TLS gate (chosen) | Matches each network; trust preserved | One new pure-Go dep (mDNS) | Accepted |

## 5. Trade-offs and Risks

- **Spoofed hub advertisement.** A rogue host can answer an mDNS query. Mitigated
  because discovery yields location only — enrollment still requires the token over
  TLS, and auto-connect refuses untrusted hubs. Without a configured token, a
  discovered hub is shown, not joined.
- **mDNS noise / multicast on large L2 segments.** Bounded query interval; the
  overlay and seed providers avoid multicast entirely.
- **Daemon advertisement leaks host presence** on the segment. It is opt-in
  (`discoverable`), off by default, and carries no secrets — only a service name.
- **New dependency.** A pure-Go mDNS library must be vetted for CGO-freeness and
  maintenance; this is the reason an ADR is required.

## 6. Impact

**FinOps:** No new infrastructure or managed service. Cost is negligible compute for
periodic announce/resolve. Removes operator labor (manual per-host `--hub`).

**SRE:** Discovery failure degrades to **explicit `--hub`** — no regression for
configured fleets. Failure modes: stale mDNS cache, overlay API outage, multicast
disabled. Observability: a `self.discovery.*` metric (provider, candidates found,
last-resolve age). Runbook: when discovery is silent, fall back to seed list.

**Security:** New surface is the announce/resolve path. Threat model: spoofed
advertisement (mitigated by token+TLS), presence disclosure via daemon
advertisement (opt-in, no secrets), multicast DoS (bounded intervals). Discovery
**never** bypasses the trust gate — this is the load-bearing security property.

**Team:** One new concept (`Discoverer` providers). Operators learn `--discover`
and when to pick a provider. The mDNS dep adds minor supply-chain review burden.

## 7. Decision

Add zeroconf discovery as an additive, opt-in capability: the hub advertises
`_heimdall-hub._tcp`, the daemon resolves it through a pluggable `Discoverer`
(mDNS, overlay, static seed) and opens the same outbound stream it opens for an
explicit `--hub`. The hub never dials daemons; optional `_heimdall-daemon._tcp`
advertisement is visibility-only. Token + TLS enrollment trust is never bypassed by
auto-connect. The one new dependency is a pure-Go mDNS library.

Status: **proposed**

## 8. Next Steps

- [ ] Define the `Discoverer` interface and wire it behind `--discover` / `discoverable` — daemon
- [ ] Implement mDNS, overlay, and seed-list providers; vet the pure-Go mDNS library — daemon
- [ ] Hub advertisement of `_heimdall-hub._tcp` + optional daemon enumeration view — hub/dashboard
- [ ] Add tests proving auto-connect refuses an untokened/untrusted hub — QA
- [ ] Document discovery providers and the trust-gate guarantee — Architect
