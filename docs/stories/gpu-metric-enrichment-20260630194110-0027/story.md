---
id: "gpu-metric-enrichment-0027"
title: "Richer GPU metrics — clocks, memory utilisation, fan (NVIDIA + AMD)"
epic: "fleet-accelerator-coverage"
priority: "high"
ui: true
design_target: "tui"
qa_bdd: "behave"
adr_required: false
milestone: "v2.2.1"
phase: discover
labels:
  - "type:feature"
  - "priority:high"
  - "area:adapters"
status: approved
issue_number: null
issue_url: null
created_at: "2026-06-30T19:41:10+0000"
---

# Richer GPU metrics — clocks, memory utilisation, fan (NVIDIA + AMD)

## Story

**As an** operator watching GPU-heavy hosts (DGX/GB10, NVIDIA workstations, AMD
Strix Halo),
**I want** more than just `gpu.util` / `gpu.temp` / `power.gpu` — at least the
**core clock**, **memory-controller utilisation**, and **fan** where the hardware
exposes them,
**so that** the top-view GPU panel shows a real accelerator picture instead of
three numbers.

## Context

Today the GPU panel collects only `gpu.util`, `gpu.vram`, `gpu.temp`,
`power.gpu`. Live fleet gaps:

- **promaxgb10 (GB10/DGX)** reports util/temp/power but **no `gpu.vram`** — its
  `nvidia-smi` memory query returns N/A on the unified-memory part.
- Neither NVIDIA nor AMD reports clocks, memory-controller utilisation, or fan.

New keys (all degrade to Unavailable where the platform can't supply them — never
faked):

| Key | Unit | NVIDIA source | AMD source |
|---|---|---|---|
| `gpu.clock` | MHz | `clocks.current.graphics` | amd-smi clock col / sysfs hwmon `freq1_input` |
| `gpu.mem.util` | percent | `utilization.memory` | amd-smi `umc_activity` |
| `gpu.fan` | percent | `fan.speed` | (amd-smi only, when a `%` fan is reported) |

The NVIDIA query string gains the three fields; the parser reads them by index
and skips any that come back `[N/A]`/`[Not Supported]`. AMD adds `--clock` to the
amd-smi call and token-matches the new columns, plus reads `freq1_input` from the
amdgpu hwmon for the sysfs fallback.

## Acceptance Criteria

```gherkin
@story:gpu-metric-enrichment-0027 @epic:fleet-accelerator-coverage @priority:high
Feature: Richer GPU metric collection

  Scenario: NVIDIA reports clock, memory util, and fan
    Given an NVIDIA host
    When the collector parses an nvidia-smi row that includes graphics clock, memory utilisation, and fan speed
    Then it reports "gpu.clock" in MHz
    And it reports "gpu.mem.util" as a percent
    And it reports "gpu.fan" as a percent

  Scenario: NVIDIA fields that read N/A are skipped, not faked
    Given an NVIDIA host where fan speed reads "[N/A]"
    When the collector parses the row
    Then "gpu.fan" is not reported
    And the other GPU metrics are still reported

  Scenario: AMD reports clock and memory util via amd-smi
    Given an AMD host with amd-smi present
    When the collector parses the amd-smi CSV with clock and umc_activity columns
    Then it reports "gpu.clock" in MHz
    And it reports "gpu.mem.util" as a percent

  Scenario: AMD sysfs fallback reads the gfx clock
    Given an AMD host without amd-smi
    And the amdgpu hwmon exposes freq1_input
    When the collector runs
    Then it reports "gpu.clock" derived from freq1_input

  Scenario: the top-view GPU panel surfaces the new metrics
    Given a host reporting gpu.clock, gpu.mem.util, and gpu.fan
    When the full-screen top view renders the GPU panel at a wide tier
    Then clock, memory utilisation, and fan are visible
    And a host missing them renders a dash, not an error
```

## Definition of Done

- [ ] Unit tests green (extended nvidia + amd parsers)
- [ ] Top-view GPU panel renders the new metrics; missing ones show `—`
- [ ] Existing GPU metrics and the NVIDIA/AMD/Apple paths unchanged
- [ ] `docs/metrics.md` lists the new keys
- [ ] CHANGELOG entry (v2.2.1)
- [ ] PR description references this story

## ADR Links

<!-- adr_required: false — additive metric keys, no architectural change. -->
