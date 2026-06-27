---
adr: "0016"
title: "Mímir: pluggable durable sink and hub state restore"
status: accepted
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0016 — Mímir: pluggable durable sink and hub state restore

> Extends **Mímir** (ADR-0011). See the [glossary](../glossary.md). Story: 0021.

## 1. Context

ADR-0008 left a deliberate seam: durable history as "an optional TSDB sink later
added as just another bus subscriber." ADR-0011 (Mímir) built history and an
OpenMetrics `/metrics` endpoint, but kept history **in-memory** — lost on hub
restart by design, listing both "remote-write to operator TSDB" and a "TSDB sink"
as v2 candidates.

The cost of that gap is now concrete. On a hub restart the dashboard goes blank
until daemons resync via keyframe, and offline / last-seen hosts vanish entirely —
a host that was down before the restart has no live daemon to repaint it, so fleet
identity is lost across the bounce. Operators want continuity across restarts
**without** Heimdall owning a storage tier.

This ADR opens the v2 door ADR-0008 and ADR-0011 held ajar.

## 2. Goals / Non-Goals

**Goals:**
- An **optional, off-by-default, pluggable durable sink** the operator points at
  **their own** Prometheus-compatible TSDB via `--tsdb <url>`.
- **Bidirectional** behavior: steady-state `Write` (remote-write) and startup
  `Restore` (PromQL query) behind one `Store` abstraction.
- **Hub state recovery**: seed the registry from the TSDB *before* live traffic, so
  the dashboard repaints instantly and offline/last-seen hosts survive a restart.
- Default behavior **unchanged** (in-memory only) when `--tsdb` is unset.

**Non-Goals:**
- Embedding, packaging, or operating a TSDB. **Heimdall never ships a TSDB.**
- Full-fidelity restore. Info-string metrics and alert state are not restored
  (see §5); they re-derive live within one sample interval.
- A query language or analytics surface beyond what restore needs.
- Depending on the full upstream `prometheus` Go module.

## 3. Proposal

**`Store` interface (the abstraction).** A single port with two methods, registered
as a MetricBus subscriber per ADR-0008/0011:

- `Write(samples)` — steady state. Prometheus **remote-write**: protobuf-encoded
  `WriteRequest` + **snappy** compression, `POST` to `<url>/api/v1/write`. Runs off
  the ingest hot path; failures are best-effort and logged, never block ingest.
- `Restore(ctx) → snapshot` — hub startup. **PromQL instant query** for last-known
  gauge values and labels, plus a **short range query** for recent trend history,
  via `GET <url>/api/v1/query` and `/api/v1/query_range`.

**Wiring.** `--tsdb <url>` is the only switch. Unset → no sink registered, behavior
identical to ADR-0011. Set → the sink is constructed and registered, and the hub
runs `Restore` during boot.

**Restore sequencing.** On startup, if a TSDB is configured, the hub:
1. Queries the TSDB and **seeds the registry before accepting live daemon traffic**.
2. Restores fleet identity, last-known gauge values, labels
   (`host` / `origin_hub` / `tag_<key>`, per ADR-0010), and `last_seen`.
3. Loads recent range history for trends.
4. Then opens ingest. Live keyframes (ADR-0002) reconcile any drift forward.

Effect: the dashboard repaints from the first frame, and hosts that were **offline
before the restart** stay visible with their last-seen state instead of disappearing.

**Best-effort fidelity (documented).** Not everything round-trips through a numeric
TSDB:
- **Info-string metrics** (e.g. `host.os`) are not numeric series → not restored;
  reappear on the next live sample.
- **Alert state** (ADR-0012/0014) is re-derived live by the engine → not restored.
- **Per-core** values may be approximated from aggregates on restore.
These reconverge within one sample interval and are called out in the runbook.

**Dependencies kept minimal.** No upstream `prometheus` module. Only:
- `snappy` (pure-Go) for remote-write block compression.
- A **minimal generated proto** covering exactly `WriteRequest`, `TimeSeries`,
  `Label`, `Sample` — not the full remote-write/types schema. Additive to the
  existing proto surface per ADR-0001.

**SOLID.**
- **DIP** — the hub depends on `Store`, not on Prometheus. The TSDB is an injected
  detail behind a port.
- **OCP** — Prometheus remote-write is the first `Store` impl; InfluxDB or others
  are new impls, not modifications to the hub.
- **SRP** — write, restore, and wire-serialization are separated; each has one
  reason to change.
- **LSP** — every `Store` impl must honor the same write-best-effort / restore-seed
  contract.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Embed/package a TSDB in Heimdall | Turnkey durability | Stateful dep, storage cost, compaction/retention ops, deploy weight | Violates ADR-0008/0011 non-goal; Heimdall is not a storage product |
| Keep in-memory only (status quo) | Simplest, zero deps | Blank dashboard + lost offline hosts on every restart | Fails the continuity goal |
| Disk append-log / local snapshot file | Survives restart, no external dep | New IO + retention + corruption failure modes; no query story for trends | Cost without reuse of operator's existing stack |
| Full `prometheus` Go module for remote-write | Canonical types | Large dependency surface, transitive bloat | Disproportionate; minimal generated proto suffices |
| Pluggable `Store` → operator TSDB, bidirectional (chosen) | Reuses operator spend; off by default; opens ADR-0008 seam; OCP for other backends | Best-effort restore fidelity; outbound surface to secure | Accepted; matches non-goals |

## 5. Trade-offs and Risks

- **Best-effort restore fidelity.** Info-strings and alert state do not round-trip;
  per-core may be approximated. Documented; reconverges in one sample interval.
- **New outbound surface.** Remote-write adds an egress dependency to the operator's
  TSDB needing auth + TLS; a slow/down TSDB must never block ingest (write is
  fire-and-forget with bounded buffering and drop-on-pressure).
- **Double-write cost.** Every applied sample is also serialized, compressed, and
  shipped — extra CPU and egress when the sink is on. Off by default keeps the
  common case free.
- **Clock / staleness on restore.** Restored samples carry TSDB timestamps; the hub
  must treat them as last-known, not live, and let keyframes supersede them — avoid
  resurrecting a stale host as "online."
- **Restore correctness depends on label discipline** (ADR-0011 cardinality rules);
  malformed/high-cardinality labels degrade or bloat restore.
- **Coupling risk** is bounded by `Store`: the hub never imports a TSDB SDK.

## 6. Impact

**FinOps:** Off by default → zero added cost in the common case. When enabled, cost
is **the operator's existing TSDB** plus marginal hub CPU/egress for double-write —
Heimdall still owns **no storage tier**, consistent with ADR-0008/0011. Egress is the
main variable; scales with `samples/s × series`.

**SRE:** Restart now restores fleet view and offline hosts instead of blanking —
a direct MTTR/operability win. New failure modes: TSDB unreachable on write
(best-effort, ingest unaffected) and TSDB unreachable on restore (hub falls back to
empty registry + keyframe rebuild = today's behavior, no regression). Observability:
sink write success/error and restore duration/series-count as self-metrics. Runbook:
set `--tsdb`, scope TSDB credentials, expect best-effort fields to reconverge in one
interval, scrape/restore at one hub level (ADR-0011 de-dupe by `origin_hub`).

**Security:** New egress to the TSDB → **TSDB credentials** must be supplied out of
band (env/file), never logged, never embedded in series. Remote-write should use TLS;
the URL may carry auth that must be redacted in logs. No secrets in labels or values
(reaffirms ADR-0011). Restore reads operator-controlled data into the registry —
treat as untrusted input, bound cardinality and series count.

**Team:** Operators already running Prometheus need no new skill. The proto/snappy
plumbing is contained in the `Store` impl; the rest of the hub sees only the port.

## 7. Decision

Heimdall **never embeds or packages a TSDB.** It gains an optional, off-by-default,
pluggable durable sink behind a `Store` port that the operator points at their own
Prometheus-compatible TSDB via `--tsdb <url>`. The sink is bidirectional: it writes
via Prometheus remote-write (protobuf + snappy) in steady state, and on startup it
restores last-known values, labels, last-seen, and recent history via PromQL — seeding
the registry before live traffic so the dashboard repaints instantly and offline hosts
survive a restart. Restore is best-effort: info-string metrics and alert state are not
restored and reconverge live. Dependencies stay minimal (snappy + a minimal generated
remote-write proto, not the full prometheus module). Default in-memory behavior is
unchanged. This opens the v2 door ADR-0008 and ADR-0011 deliberately left.

Status: **accepted**

## 8. Next Steps

- [ ] Define the `Store` port (`Write` / `Restore`) and register it as a bus subscriber — Architect
- [ ] Implement Prometheus remote-write `Store`: minimal proto + snappy, `--tsdb` wiring — hub
- [ ] Implement startup `Restore` (instant + range PromQL) seeding the registry before ingest — hub
- [ ] Add sink self-metrics (write ok/err, restore duration/series count) — hub
- [ ] Tests: write best-effort under TSDB down, restore seeds registry, offline host survives restart, stale-timestamp handling — QA
- [ ] Runbook: `--tsdb` setup, credential handling, best-effort fields, one-level scrape/restore — Architect
