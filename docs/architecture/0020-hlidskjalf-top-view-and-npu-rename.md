---
adr: "0020"
title: "Hliðskjálf full-screen top view, NPU rename, and screen-aware TUI"
status: accepted
date: "2026-06-30"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0020 — Hliðskjálf full-screen top view, NPU rename, and screen-aware TUI

## 1. Context

The dashboard had a process table bound to `t` ("top"), but no mactop/btop-style
single-host system view. Users watching one host — often over SSH from Termius on
a phone or tablet — wanted per-core CPU, memory, power, accelerator, network, and
disk at a glance, on any screen size.

Two pre-existing constraints shaped the work. The accelerator power metric was
named `power.ane` (Apple Neural Engine), which is Apple-specific for what is
really a generic NPU other platforms also have. And the wire must stay backward
compatible: a v2.2 dashboard runs against older daemons/hubs in a mixed fleet
(see [ADR 0001](0001-versioned-grpc-shared-schema-in-common.md)).

## 2. Goals / Non-Goals

**Goals:**
- A full-screen single-host "top" view entered with `t`, exited with `esc`/`q`.
- Reuse data already collected (`model.history`, per-core values) — render only.
- Generic accelerator terminology (`power.npu`) without breaking a mixed fleet.
- New cross-platform metrics (load, swap, per-core freq) that degrade gracefully.
- Both the top view and the host detail view reflow to the terminal width.

**Non-Goals:**
- No new binary; the view lives in `heimdall-dashboard`.
- No fleet aggregation in the top view — it shows the single focused host.
- No TSDB; trends keep using the in-memory ring buffers from
  [ADR 0008](0008-in-memory-ring-buffers-vs-tsdb.md).

## 3. Proposal

- **Keys:** `p` opens the (renamed) process table; `t` opens the new top view.
- **`topview` package:** a pure, render-only Bubble Tea sub-model. The dashboard
  constructs it from the focused host's snapshot + history, delegates
  `Update`/`View` while active, refreshes it on each tick (preserving scroll),
  resizes it on `WindowSizeMsg`, and leaves on `esc`/`q`. It never touches
  adapters. `Update` returns an explicit action (`ActBack`/`ActQuit`) so `q`
  quits the app while `esc` returns to the grid.
- **NPU rename (the data-model decision):** the canonical metric key is
  `power.npu`. `power.ane` is kept as a read alias, normalised to `power.npu` in
  one place — `domain.HostRegistry.Observe` — so any older daemon in the fleet
  still renders under NPU. The wire (`MetricSample`) is name+value, so new keys
  are additive and need no proto change.
- **New collectors** (`swap`, `load`, `freq`), each returning `Unavailable`
  (rendered `—`) where the platform can't supply it; `cpu.freq` reads `/sys`
  cpufreq on Linux and falls back to gopsutil elsewhere.
- **Screen-aware layout:** `topview.layout(width)` picks the densest of four
  tiers (wide grid → medium → narrow → iPhone-portrait key-numbers), and the host
  detail view scales its gauges/sparklines and reflows its field rows the same
  way. Every rendered line is clamped to the terminal width.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Extend the modal system for the top view | reuses modal plumbing | couples to the already-large `modal.go`; not full-screen | a dedicated sub-model is cleaner and independently testable |
| Separate `heimdall-top` binary | isolation | another binary to ship/run; not in-dashboard | users want it inside the dashboard |
| Rename `power.ane` with no alias | clean break | breaks mixed fleets reading the old key | BC is a hard constraint; alias is cheap |
| Add an NPU/freq proto field | typed wire | proto regen; wire churn for an additive value | name+value wire already carries new keys |

## 5. Trade-offs and Risks

- The `power.ane` alias is a small permanent normalisation in `Observe`; it costs
  one map lookup per metric and must stay until no v2.1-era daemon remains.
- `npu.util` and `mem.bw` have no collector yet (they render `—` everywhere);
  on Apple Silicon Pro/Max the IOReport channels read ~0 regardless, so this is
  honest, not a regression. Documented as a follow-up.
- The top view refreshes by rebuilding from the latest snapshot each tick; scroll
  is preserved explicitly. A future per-row selection would need more state.

## 6. Impact

**FinOps:** None. No new services or storage; trends remain in bounded RAM.

**SRE:** The view is render-only and cannot fail collection. New collectors are
failure-isolated per [ADR 0003](0003-metric-adapter-contract-and-failure-isolation.md):
an unsupported platform yields `Unavailable`, never an error that drops the host.

**Security:** No new privilege. The top view reads existing snapshots; the new
collectors are unprivileged (load/swap/freq need no root).

**Team:** The `topview` package is small and self-contained; the responsive
pattern mirrors the existing `layout(width)`/`scrollWindow` used by the grid.

## 7. Decision

Ship a dedicated, render-only `topview` sub-model on `t` (process table moves to
`p`), rename the accelerator power metric to the vendor-neutral `power.npu` while
honouring `power.ane` as an ingest alias for mixed fleets, add graceful-degrade
load/swap/freq collectors, and make both the top view and the host detail view
reflow to the terminal width so they stay usable down to a phone-sized terminal.

Status: **accepted**

## 8. Next Steps

- [x] `topview` package + dashboard wiring (`t`/`p`, tick refresh, resize)
- [x] `power.ane` → `power.npu` alias in `HostRegistry.Observe`
- [x] `swap` / `load` / `freq` adapters with graceful degradation
- [x] Responsive host detail view + width-bound tests
- [ ] Real `npu.util` / `mem.bw` collectors where the platform exposes them
- [ ] Optional `ProcessRow.User` (proto field) to restore a USER column
