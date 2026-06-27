---
adr: "0011"
title: "Mímir: metrics history and OpenMetrics export"
status: proposed
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0011 — Mímir: metrics history and OpenMetrics export

> **Mímir** /ˈmiː.mɪr/ (*MEE-meer*) — keeper of the well of memory beneath
> Yggdrasil. See the [glossary](../glossary.md).

## 1. Context

Two adjacent needs share one root. (1) Operators already run Prometheus-style
observability stacks and want to scrape the **whole Heimdall fleet** through one
endpoint instead of instrumenting each host. (2) The dashboard wants short-range
trends beyond a single live value. Both are read models over the same metric
stream.

ADR-0008 deliberately left the seam: trends come from the **MetricBus**, and "an
optional TSDB sink can later be added as just another bus subscriber." This ADR uses
that seam for both features — **without** introducing a stateful TSDB dependency.
Long-term durable storage remains a non-goal; the v2 TSDB-sink path stays open.

## 2. Goals / Non-Goals

**Goals:**
- A hub `/metrics` endpoint in **OpenMetrics (Prometheus)** format exposing the
  whole fleet, with labels for host id, origin hub, and tags.
- **Bounded in-memory history** (ring buffer per host/metric) for short-range
  trends, sized and retention-bounded.
- Both implemented as **MetricBus subscribers** per ADR-0008's extension point — no
  change to the ingest hot path.

**Non-Goals:**
- Long-term durable storage, downsampling, or a query language (leave the v2
  TSDB-sink path open).
- Surviving restarts: in-memory history is lost on restart, by design.
- A push gateway or remote-write client (scrape model only in v1).

## 3. Proposal

**Both features are bus subscribers.** Per ADR-0008, the hub already publishes
applied snapshots on the MetricBus. Add two subscribers; neither touches ingest:

1. **History subscriber** — maintains a fixed-size ring buffer per `(host, metric)`
   for short-range trends. This is the same bounded, computable structure ADR-0008
   defined (`memory ≈ hosts × metrics_per_host × ring_depth × sample_size`),
   retention bounded by ring depth. Purge-on-delete and a `self.ring.occupancy`
   self-metric carry over.

2. **Export subscriber** — keeps the latest value (and ring-derived series where a
   stack scrapes ranges) and renders **OpenMetrics** on `GET /metrics`. Each series
   is labeled with `host`, `origin_hub` (ADR-0010), and `tag_<key>` for effective
   tags (host-over-hub merge, ADR-0010). A Prometheus server scrapes the hub and
   sees the entire federated fleet through one target.

**Cardinality discipline.** Labels are bounded: host id, origin hub, and a small
tag set. The hub documents that high-cardinality tags inflate scrape cost; tag keys
exported as labels are configurable.

**Restart behavior (documented).** History and export state are in-memory.
A restart loses history; live values and labels rebuild within seconds as daemons
resync via keyframe (ADR-0002/0008). This is acceptable and explicitly documented —
durable retention is a v2 concern served by adding a TSDB **sink** as a third bus
subscriber, with no change to ingest, dashboard, or this endpoint.

**SOLID.** SRP: history and export are separate subscribers with one job each. OCP:
a future TSDB sink is another subscriber, not a modification. DIP: both depend on
the MetricBus abstraction, not on ingest internals.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Embed a TSDB now | Durable history, queries | Stateful dep, storage cost, compaction ops | Contradicts ADR-0008; out of scope |
| Per-host Prometheus exporters | Standard | Inbound port per host; defeats outbound-only model | Breaks ADR-0007 |
| Remote-write to operator TSDB | No scrape endpoint | Push client + retries + auth; couples to their TSDB | Heavier than scrape; v2 candidate |
| `/metrics` + ring history as bus subscribers (chosen) | Reuses ADR-0008 seam; no new stateful dep | History lost on restart | Accepted; matches non-goals |

## 5. Trade-offs and Risks

- **History lost on restart.** Accepted and documented (consistent with ADR-0008);
  live state recovers via keyframe.
- **Scrape cardinality blow-up** from many tags/hosts. Mitigated by bounded labels,
  configurable exported tag keys, and federating hubs instead of one giant target.
- **`/metrics` is a new read surface.** A large fleet makes scrape responses big;
  mitigated by scrape interval guidance and per-hub host caps (ADR-0008).
- **Double counting across federation.** The parent hub's `/metrics` includes
  relayed hosts; operators must scrape at one level or de-dupe by `origin_hub`.
  Documented.

## 6. Impact

**FinOps:** No storage tier and no managed TSDB — consistent with ADR-0008's cost
avoidance. Marginal cost is RAM for ring buffers and CPU to render scrapes. Reuses
the operator's existing Prometheus spend rather than adding ours.

**SRE:** `/metrics` plugs Heimdall into existing alerting/dashboards immediately.
Failure modes: scrape timeout on a huge fleet (tune interval/caps), RAM growth
(occupancy self-metric, host cap). Recovery: restart rebuilds live state in seconds.
Runbook: size ring depth, choose exported tag labels, scrape at the right hub level.

**Security:** `/metrics` exposes fleet telemetry — treat as sensitive. It must sit
behind the same TLS/network controls as the hub; no secrets in metric labels or
values. New read endpoint = new DoS surface; rate-limit/scope scrapers.

**Team:** OpenMetrics is a familiar format; no TSDB expertise needed. The ring
buffer is the same structure already specified in ADR-0008.

## 7. Decision

Serve both short-range history and fleet-wide observability from the existing
MetricBus seam. Add a history subscriber (bounded per-`(host, metric)` ring buffers)
and an export subscriber that renders OpenMetrics on `GET /metrics` labeled by host,
origin hub, and effective tags. No TSDB dependency is introduced; in-memory history
is lost on restart by design and recovers via keyframe. Durable long-term storage
stays a v2 concern served by adding a TSDB sink as another bus subscriber.

Status: **proposed**

## 8. Next Steps

- [ ] Implement the history subscriber (reuse ADR-0008 ring buffers + occupancy + purge-on-delete) — hub
- [ ] Implement the OpenMetrics export subscriber with host/origin-hub/tag labels — hub
- [ ] Make exported tag-label keys configurable; document cardinality guidance — Architect
- [ ] Add tests: label correctness, restart loses history, no orphan series — QA
- [ ] Document restart behavior, scrape-at-one-level, and the open v2 TSDB-sink path — Architect
- [ ] Note: cross-platform privileged-metrics parity (Linux RAPL/hwmon, Windows) is an extension of ADR-0004/0005 and needs no separate ADR — tracked under E2
