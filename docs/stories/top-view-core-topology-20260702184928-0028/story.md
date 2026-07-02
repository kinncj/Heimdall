---
id: "top-view-core-topology-0028"
title: "Hliðskjálf: per-core grid labelled by core type + per-cluster CPU power"
epic: "power-metric-standardization"
priority: "high"
ui: true
design_target: "tui"
qa_bdd: "behave"
adr_required: false
milestone: "v2.6.0"
phase: discover
labels:
  - "type:feature"
  - "priority:high"
  - "area:adapters"
  - "area:tui"
status: approved
issue_number: 5
issue_url: "https://github.com/kinncj/Heimdall/issues/5"
created_at: "2026-07-02T18:49:28+0000"
---

# Hliðskjálf: per-core grid labelled by core type + per-cluster CPU power

## Story

**As an** operator watching a heterogeneous fleet (Apple Silicon, Intel hybrid,
AMD, ARM),
**I want** the top view's per-core bars to say *which kind* of core each bar is
(P/E on Apple Silicon and Intel hybrid, uniform elsewhere), and — where the
hardware exposes it — the watts each CPU cluster is drawing,
**so that** "the machine is busy" becomes "the P-cores are pinned and pulling
18 W while the E-cores idle", per host, at a glance.

## Context

ADR-0021 standardized `power.cpu`/`power.gpu`/`power.total` and left two
next-steps this story picks up:

- The top view (ADR-0020, Hliðskjálf) already renders per-core utilisation bars
  from `cpu.cores` (`KindPerCore`), but the bars are anonymous — on a 12P+4E
  M3 Max you can't tell whether the pinned cores are P or E.
- On Apple Pro/Max the daemon already reads the raw SMC per-cluster power keys
  and *sums* them into `power.cpu`. The cluster split (P vs E) is computed and
  then thrown away.

Ground truth on the fleet:

- **Apple Silicon** exposes core counts per type via unprivileged sysctl
  (`hw.perflevel0` = Performance, `hw.perflevel1` = Efficiency; verified live:
  M3 Max reports 12/4). Logical core order vs type must be verified empirically
  (E-cluster typically owns the low CPU IDs).
- **Intel hybrid (Linux)** exposes `/sys/devices/cpu_core/cpus` and
  `/sys/devices/cpu_atom/cpus` as index ranges — unprivileged.
- **Intel hybrid (Windows)** exposes `EfficiencyClass` via
  `GetLogicalProcessorInformationEx` (`golang.org/x/sys`, already in the module
  graph — no new dependency).
- **AMD / non-hybrid / ARM** are uniform: every core the same type.
- **Per-cluster power** exists only where per-cluster rails are readable:
  Apple Pro/Max SMC keys (P-clusters `PC02`+`PC42`, E/fabric `PC03`+`PC43` on
  M3 Max). RAPL is package-level; there is no per-core or per-cluster wattage
  on Intel/AMD/ARM/Windows — those hosts simply don't get the cluster rows.

## Metric contract (BC-safe: new names only, old dashboards ignore them)

| Metric | Kind | Content |
|---|---|---|
| `cpu.topology` | PerCore | `PerCore[i]` = core-type id of logical core *i* (`0` = performance, `1` = efficiency); `Gauge` = distinct type count; `Detail` = human summary (`12P + 4E`, `16 cores (uniform)`) |
| `power.cpu.pcluster` | Gauge (W) | P-cluster rail sum — Apple Pro/Max SMC only |
| `power.cpu.ecluster` | Gauge (W) | E-cluster rail sum — Apple Pro/Max SMC only |

`cpu.topology` reads `Unavailable` with a reason on platforms with no probe
(e.g. Windows pre-Win10 fallback); the renderer then draws today's unlabelled
grid. Nothing existing is renamed or re-shaped.

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Per-core grid labelled by core type, with per-cluster CPU power

  Scenario: Apple Silicon reports P/E topology unprivileged
    Given a macOS host whose sysctl reports perflevel0=12 and perflevel1=4
    When the cpu adapter collects
    Then metric "cpu.topology" is OK with 16 per-core entries
    And the per-core entries mark the efficiency cores distinctly from the performance cores
    And the detail reads "12P + 4E"

  Scenario: Intel hybrid Linux reports P/E topology from sysfs
    Given a Linux host with cpu_core cpus "0-15" and cpu_atom cpus "16-23"
    When the cpu adapter collects
    Then metric "cpu.topology" is OK with 24 per-core entries
    And entries 0-15 are performance and entries 16-23 are efficiency

  Scenario: Uniform CPU reports a single-type topology
    Given a host with no hybrid-core interface and 32 logical cores
    When the cpu adapter collects
    Then metric "cpu.topology" is OK with 32 per-core entries all of one type
    And the detail reads "32 cores (uniform)"

  Scenario: Apple Pro/Max splits cluster power from the SMC
    Given an Apple host whose SMC P-cluster keys sum to 15.0 W and E-cluster keys to 3.0 W
    When power is assembled
    Then metric "power.cpu.pcluster" is OK at 15.0 W
    And metric "power.cpu.ecluster" is OK at 3.0 W
    And metric "power.cpu" remains the 18.0 W total

  Scenario: Hosts without per-cluster rails emit no cluster metrics
    Given a Linux host reading RAPL package power
    When power is assembled
    Then no "power.cpu.pcluster" or "power.cpu.ecluster" metric is emitted

  Scenario: Top view labels the core grid by type
    Given a snapshot with cpu.cores and a cpu.topology of "12P + 4E"
    When the top view renders the CPU panel
    Then performance-core bars are grouped under a "P" label and efficiency-core bars under an "E" label

  Scenario: Top view shows cluster watts when present
    Given a snapshot with power.cpu.pcluster 15.0 and power.cpu.ecluster 3.0
    When the top view renders the POWER panel
    Then the panel shows the P-cluster and E-cluster watts alongside the CPU total

  Scenario: Top view degrades gracefully without topology (mixed-version fleet)
    Given a snapshot from an older daemon with cpu.cores but no cpu.topology
    When the top view renders the CPU panel
    Then the per-core grid renders unlabelled exactly as before
```

## Out of scope

- Per-core wattage — no vendor exposes it (Apple = cluster rails, Intel/AMD =
  package/domain). Documented in ADR-0021 §5.
- Mapping SMC cluster keys for additional Apple dies (tracked in ADR-0021 next
  steps; unmapped dies keep the safe fallback).
- Dashboard (fleet) view changes — this is the single-host top view.

## Notes

- Core-index → type mapping on macOS must be verified live before the renderer
  trusts it (E-cores are believed to own the low CPU IDs; confirm under load on
  the M3 Max and pin with a test).
- `golang.org/x/sys` moves from indirect to direct for the Windows probe — no
  new module, no ADR trigger.

## A11y audit (tui checklist)

Terminal checklist, audited 2026-07-02 against the implemented render
(`app/internal/tui/topview/`). **No critical or serious violations — merge
unblocked.**

| Check | Result | Evidence |
|---|---|---|
| color-only signaling | pass | P/E is carried by plain-text group headers (`P-cores (12):`, `E-cores (4):`) and the cluster line's `P`/`E` text tags; render tests strip ANSI before asserting, so the labels exist without colour |
| NO_COLOR / monochrome | pass | all new signal is text runs; severity colour on bars is reinforcement only (unchanged from the 0025 audit) |
| min-width / resize | pass | `TestNoLineExceedsWidthHybrid` asserts no line exceeds the frame at 30–120 cols; the cluster line is dropped at TINY by design, never clipped mid-value |
| reading order | pass | groups render in logical core order (whichever type owns c0 first); visual order == reading order |
| no fabricated values | pass | absent cluster rails render nothing (no fake 0); missing topology falls back to the exact pre-feature grid |
| focus / keyboard | pass (n/a) | view keeps its single scroll focus; no new interactive elements |
