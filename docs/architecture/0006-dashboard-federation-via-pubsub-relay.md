---
adr: "0006"
title: "Dashboard federation via pub/sub relay"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
---

# 0006 — Dashboard federation via pub/sub relay

## 1. Context

Operators want a local hub per site/lab and a roll-up view at a parent (cloud) hub,
with multiple dashboards watching concurrently. A naive design (dashboards dialing
every hub, or hubs re-ingesting each other's daemons) creates N×M connections,
duplicate registration, and loops. We need a clean fan-in/fan-out that reuses the
existing bus abstraction (E4, story `centralized-dashboard-federation-relay-0009`).

## 2. Goals / Non-Goals

**Goals:**
- A child hub relays upstream to a parent hub as a **bus subscriber**.
- Multiple dashboards subscribe to a hub concurrently.
- Cross-hub authentication; loop and duplication prevention.

**Non-Goals:**
- Federating the control plane or logs in v1 (telemetry relay only).
- A global query/aggregation engine (the parent just republishes).

## 3. Proposal

Reuse the **MetricBus** subscriber model for federation:

- A child hub runs a relay that **subscribes to its own bus** and forwards
  `snapshot.applied` events upstream via `FederationService.Relay` as a stream of
  `RelayEnvelope`.
- Each `RelayEnvelope` carries `origin_hub_id` (first ingesting hub) and an ordered
  `path[]` of hub ids traversed.
  - **Loop prevention:** a hub drops any envelope whose `path[]` already contains
    its own id.
  - **Duplication prevention:** upstream de-dupe by `(origin_hub_id, host_id, seq)`.
- The parent hub **republishes** received snapshots onto its own bus. Parent-side
  dashboards subscribe exactly like local ones — fan-out is just multiple bus
  subscribers.
- Cross-hub links use **mTLS + a federation token**; `SubscribeRequest.host_ids`
  scopes what a peer may receive.

This keeps dashboards and the relay symmetric (both are bus subscribers) and adds
no special path to the hot ingest loop.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Dashboards dial every hub directly | No relay component | N×M connections; no roll-up; poor over WAN | Fails roll-up + scale |
| Parent re-ingests child daemons | Reuses ingest | Daemons span trust zones; duplicate registration | Breaks identity model |
| Message broker (Kafka/NATS) between hubs | Mature fan-out | Heavy dependency; ops + cost; overkill for v1 | Violates no-TSDB-style simplicity |
| Relay as bus subscriber (chosen) | Symmetric, minimal, loop-safe | Hub holds relay state | Accepted |

## 5. Trade-offs and Risks

- Multi-hop federation adds latency and a small duplication window; bounded by
  `seq` de-dupe and `path[]`.
- A misconfigured topology could fan traffic widely; mitigated by `path[]` loop
  drop and per-peer rate limits.
- Parent hub memory grows with aggregate fleet; mitigated by host-scoped
  subscriptions and the same bounded ring buffers (ADR-0008); federate rather than
  oversize one hub.

## 6. Impact

**FinOps:** Main cost is **egress on the child→parent link**. Delta/keyframe
shaping and compression (events.md) directly reduce it. No broker/managed-streaming
bill in v1.

**SRE:** Relay outage stops upstream roll-up but local dashboards keep working
(graceful partial failure). Reconnect/backoff applies to the relay too. Runbook:
re-link a child hub, diagnose a duplication/loop alert.

**Security:** New cross-org boundary (threat-model Boundary 2): mTLS + token,
subscription scoping, no control/log federation. Loop/dup prevention also limits
amplification DoS.

**Team:** One reused abstraction (bus subscriber) to learn. Topology config is the
main new operational concept.

## 7. Decision

Federate by having a child hub's relay subscribe to its own MetricBus and forward
telemetry to a parent hub as a `RelayEnvelope` stream, with `origin_hub_id` +
`path[]` for loop/duplication prevention and mTLS + token + host scoping for
cross-hub trust. Dashboards and relays are both bus subscribers, keeping the model
uniform and the ingest path untouched.

Status: **proposed**

## 8. Next Steps

- [ ] Implement relay subscriber + `RelayEnvelope` path/origin handling — hub
- [ ] Add loop-prevention and `(origin,host,seq)` de-dupe tests — QA
- [ ] Document federation topology + token provisioning runbook — Architect
