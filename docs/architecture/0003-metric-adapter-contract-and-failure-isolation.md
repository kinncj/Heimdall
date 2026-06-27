---
adr: "0003"
title: "Metric adapter contract and failure isolation"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
---

# 0003 — Metric adapter contract and failure isolation

## 1. Context

Heimdall collects many metric families (CPU, memory, disk, network, temperature,
ping, GPU, power, host context) across heterogeneous OSes and devices, and must be
extended with new metrics over time. Two failure realities dominate: (1) a metric
may be unavailable or permission-gated on a given host, and (2) a collector can
hang, error, or panic (vendor libraries, syscalls). A single bad collector must
never drop the whole host or other metrics. This is the substrate epic (E1,
story `extensible-solid-metric-adapter-architec-0003`).

## 2. Goals / Non-Goals

**Goals:**
- One SOLID `Adapter` contract shared by daemon and hub.
- Add metrics without modifying existing adapters (OCP).
- Isolate each adapter so a timeout/error/panic degrades only that metric.
- Make permission/availability a first-class, reportable status.

**Non-Goals:**
- Defining each concrete collector's internals (per-platform detail in INFRA).
- A plugin/dynamic-loading system (compile-time registration is sufficient in v1).

## 3. Proposal

Define a narrow domain interface (illustrative):

```
type Adapter interface {
    Name() string
    Collect(ctx context.Context) ([]Metric, error)
    Status() MetricStatus
}
```

- Adapters live in the **infrastructure** layer and implement a **domain**
  interface; they are registered into an `AdapterRegistry` at the composition root
  (`app/cmd/*`). New metric = new adapter registered; nothing existing changes
  (OCP, DIP).
- The `Sampler` runs each `Collect` under a **per-adapter `context.WithTimeout`**
  and a **`recover()`** guard. A timeout, error, or panic is converted into a
  `Metric` bearing a non-OK `MetricStatus` (`UNAVAILABLE` /
  `INSUFFICIENT_PERMISSION` / `ERROR`) and an explanatory `detail`.
- The host stays `ONLINE`; sibling metrics are unaffected. Status travels on the
  wire per `MetricSample.status` so the dashboard renders a per-metric error glyph.

This is the failure-isolation guarantee, verified by the "adapter failure is
isolated" scenario and the integration suite.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| One monolithic collector | Simple wiring | One panic kills all metrics; hard to extend | Fails isolation + OCP |
| Adapters share one timeout/goroutine | Less overhead | Head-of-line blocking; one hang stalls all | Fails isolation |
| Return errors, no status enum | Idiomatic Go | Loses "unavailable vs denied vs error" nuance the UI needs | Fails permission-reporting goal |
| Dynamic plugin system | Hot-add metrics | Complexity, ABI/security risk, not needed in v1 | Over-engineering |

## 5. Trade-offs and Risks

- Per-adapter goroutine + timeout has overhead at high adapter counts; bounded by a
  worker pool and tunable cadence.
- `recover()` can mask real bugs; mitigated by emitting a `self.adapter.panics`
  metric and structured log on every recovery so panics are visible, not silent.
- A misbehaving adapter that returns slowly still consumes its timeout each cycle;
  surfaced via `self.adapter.collect_ms`.

## 6. Impact

**FinOps:** Negligible; pure in-process CPU. Tunable cadence caps cost.

**SRE:** Strong blast-radius containment — failures are per-metric. Self-metrics
(`collect_ms`, `timeouts`, `panics`) make a degrading adapter observable before it
matters. Runbook: interpret a metric stuck at `ERROR`/`INSUFFICIENT_PERMISSION`.

**Security:** `INSUFFICIENT_PERMISSION` cleanly signals when a metric needs the
privileged helper (ADR-0004) rather than silently failing or escalating.

**Team:** Contributors add metrics by writing one adapter + unit test against a
stable contract. Low barrier; high consistency.

## 7. Decision

Adopt a narrow domain `Adapter` interface with compile-time registration via an
`AdapterRegistry`, sampled under per-adapter timeout + panic recovery, with health
expressed as a per-metric `MetricStatus`. This delivers OCP extensibility and
hard failure isolation: one metric failing never drops the host or its siblings.

Status: **proposed**

## 8. Next Steps

- [ ] Freeze the `Adapter` / `AdapterRegistry` interfaces in `domain/` — Architect
- [ ] Implement the `Sampler` timeout + recover harness with self-metrics — daemon
- [ ] Add unit + integration tests for isolation and status mapping — QA
