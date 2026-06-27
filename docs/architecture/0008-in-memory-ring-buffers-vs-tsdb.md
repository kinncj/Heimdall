---
adr: "0008"
title: "In-memory ring buffers vs TSDB"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
  - "Platform Lead"
---

# 0008 — In-memory ring buffers vs TSDB

## 1. Context

The dashboard shows live values plus short trends (sparklines) per host and metric.
Storing trends in a time-series database (Prometheus/Influx/Timescale) adds a
stateful dependency, storage cost, retention/compaction operations, and deployment
weight — all for a v1 whose value is the **live** fleet view, not long-term
analytics. We need bounded, predictable trend history without a storage tier
(stated non-goal: "no TSDB in v1").

## 2. Goals / Non-Goals

**Goals:**
- Per-`(host, metric)` trend history sufficient for on-screen sparklines.
- Bounded, predictable memory; no disk, no external store.
- Simple operability and a clean v2 upgrade path if long-term retention is later
  needed.

**Non-Goals:**
- Long-term retention, downsampling, or a metric query language.
- Historical analytics across restarts (history is ephemeral by design).

## 3. Proposal

The hub keeps a **fixed-size ring buffer per `(host, metric)`** in memory. Newest
sample overwrites oldest; depth is configured for the on-screen trend window. Total
memory is bounded and computable:

```
memory ≈ hosts × metrics_per_host × ring_depth × sample_size
```

Properties:
- **No persistence.** A hub restart loses history by design; daemons reconnect and
  resync via keyframe (ADR-0002 / events.md). Live state recovers immediately.
- **Delete-reassignment.** Removing a host or adapter purges its ring buffers so no
  orphan series remain (verified by the delete-reassignment integration test).
- **Self-observability.** `self.ring.occupancy` is published as a metric.
- **v2 door (not built now).** Because trends are derived from the **MetricBus**,
  an optional TSDB sink can later be added as just another bus subscriber — no
  change to ingest or the dashboard. No speculative abstraction is added in v1.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Embedded TSDB (Prometheus/Influx/Timescale) | Long history, queries | Stateful dep, storage cost, compaction ops, heavy deploy | Over-scoped for v1 live view |
| Write metrics to disk (append log) | Survives restart | IO + retention management; still no query story | Cost without v1 benefit |
| No history (latest value only) | Simplest | No sparklines/trends — a core UX requirement | Fails dashboard goal |
| Per-(host,metric) ring buffers (chosen) | Bounded, simple, no storage | History lost on restart | Accepted; matches non-goals |

## 5. Trade-offs and Risks

- **Restart loses history.** Accepted; live state recovers via keyframe within
  seconds. Documented operator expectation.
- **Memory scales with fleet × metrics.** Mitigated by bounded depth, an occupancy
  self-metric, a per-hub host cap, and federation (ADR-0006) instead of one giant
  hub.
- Operators may want long retention later; mitigated by the bus-subscriber sink
  path kept open for v2.

## 6. Impact

**FinOps:** **No storage tier, no managed TSDB** — the largest cost avoidance in the
system. Hub cost is bounded RAM + compute. Sizing is a simple, predictable formula.

**SRE:** Far simpler to operate — no compaction, retention, WAL, or disk-full
failure mode. Failure mode is bounded RAM; mitigate with occupancy alerts and host
caps. Runbook: size ring depth, read occupancy, when to federate vs scale a hub.

**Security:** Less persistent data at rest (less to exfiltrate). No database to
secure/patch. In-memory state is volatile.

**Team:** No TSDB expertise required to operate v1. The ring buffer is a small,
well-understood data structure.

## 7. Decision

Store trend history in fixed-size, per-`(host, metric)` in-memory ring buffers with
no persistence, accepting history loss on restart in exchange for radically simpler
operations and zero storage cost. Keep a clean v2 upgrade path by sourcing any
future TSDB sink from the existing MetricBus, without building that abstraction now.

Status: **proposed**

## 8. Next Steps

- [ ] Set default ring depth + per-hub host cap and document the sizing formula — Architect
- [ ] Implement ring buffers + `self.ring.occupancy` + purge-on-delete — hub
- [ ] Add the delete-reassignment integration test (no orphan buffers) — QA
