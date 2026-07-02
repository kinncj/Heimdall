---
description: Metric standardization boundary — adapters normalise, domain standardizes, consumers render (ADR-0022)
globs: ["app/**"]
alwaysApply: false
---

# Metric standardization boundary (ADR-0022)

Heimdall watches a mixed fleet. A metric must mean the same thing on every OS,
chip, and vendor. Knowledge flows one way: **platform → adapter → neutral metric
→ `domain` → consumer.**

- **Adapters** (`app/internal/adapters`, `app/internal/helper`): translate each
  platform's own scheme (SMC, RAPL, sysctl, sysfs, EfficiencyClass, Scaphandre)
  into a **neutral** metric. Platform names and vendor branching stop here.
- **`domain`** (framework-free core): the neutral taxonomy plus pure functions
  that turn neutral metrics into display-ready structures (e.g.
  `domain.CoreGroups`, `domain.CoreTypeSummary`). Standardization lives here.
- **Consumers** (`app/internal/tui/*`, `app/internal/cli`, dashboard, hub):
  render only. No bucketing, no classification, no `switch` on platform. If a
  renderer needs a computed view, add a `domain` function and call it.

## Per-core metrics — the wire trap

A per-core metric is transmitted as its `PerCore` slice via the protobuf
`per_core` oneof; the scalar `Gauge` is **dropped on the wire** (arrives as 0).
Never key logic off a per-core metric's `Gauge`. Carry summaries in `Detail`
(which does cross the wire) and add a transport round-trip test.

## Before you edit

If you are adding or changing a metric, read
`docs/guides/17-standardization-and-adapters.md`. If you find yourself writing a
platform `switch` or bucketing values inside a TUI/CLI file, stop — it belongs in
an adapter or `domain`.
