---
id: "amd-gpu-npu-collection-0026"
title: "AMD GPU + XDNA NPU metrics (amd-smi preferred, amdgpu sysfs fallback)"
epic: "fleet-accelerator-coverage"
priority: "high"
ui: false
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
created_at: "2026-06-30T19:01:58+0000"
---

# AMD GPU + XDNA NPU metrics (amd-smi preferred, amdgpu sysfs fallback)

## Story

**As an** operator running a mixed fleet that includes AMD machines (Strix Halo
laptops like the HP ZBook Ultra G1a "CaptainCanuck", and future Radeon discrete
cards),
**I want** Heimdall to report GPU utilisation, VRAM, temperature, and power on
AMD hardware — and a best-effort XDNA NPU utilisation where the driver exposes
it —
**so that** an AMD host stops showing a blank/sparse accelerator panel and reads
the same `gpu.*` / `power.gpu` keys an NVIDIA host already does.

## Context

GPU collection is `nvidia-smi`-only today. `helper.PrivilegedMetrics`
(`app/internal/helper/collect.go`) runs the darwin IOReport/SMC path, the Linux
RAPL/hwmon path, the Windows path, then `runNvidiaSMI`. There is no AMD branch
anywhere, so on a Strix Halo box the privileged adapter emits **zero** `gpu.*`
keys.

Confirmed live on the fleet (hub `192.168.1.229:9090`):

| Host | Chip | nvidia-smi | Keys reported |
|------|------|------------|---------------|
| `promaxgb10-bdd9` | GB10 Grace-Blackwell | present | `gpu.util`, `gpu.temp`, `power.gpu` |
| `CaptainCanuck` | Strix Halo (Radeon 8060S iGPU) | absent | none — only `mem.*` |

`CaptainCanuck` is CachyOS Linux (kernel 7.1.2, x86_64). The iGPU exposes
metrics two ways:

1. **`amd-smi`** (ROCm) — richer, structured (`amd-smi metric --json` / CSV):
   util, VRAM used/total, edge temp, socket/GPU power, clock. May be absent on a
   stock laptop install.
2. **amdgpu sysfs** — always present with the in-tree driver, no extra package:
   - `/sys/class/drm/card*/device/gpu_busy_percent` → util
   - `/sys/class/drm/card*/device/mem_info_vram_used` + `mem_info_vram_total` → VRAM
   - hwmon (`/sys/class/drm/card*/device/hwmon/hwmon*/`): `power1_average` (µW)
     → power, `temp1_input` (m°C) → temp, `freq1_input` → clock

XDNA NPU (`amdxdna` driver) util has no stable userspace counter yet; treat it
as best-effort and degrade to Unavailable when absent — same as the M3 Max
`npu.util` gap.

Precedence mirrors the existing NVIDIA merge: AMD readings fill names not
already set by an earlier source, so a host that somehow has both vendors does
not double-count, and the unprivileged-but-readable sysfs path works without the
root helper (like `nvidia-smi`).

Unified-memory caveat: on an iGPU, `mem_info_vram_*` reflects the GTT/VRAM
carve-out, not system RAM. We report it as `gpu.vram` (percent of the GPU's own
budget) and keep system memory on the existing `mem.*` keys — no conflation.

## Acceptance Criteria

```gherkin
@story:amd-gpu-npu-collection-0026 @epic:fleet-accelerator-coverage @priority:high
Feature: AMD GPU and XDNA NPU metric collection

  Scenario: amd-smi present is the preferred source
    Given a Linux host with an AMD GPU
    And "amd-smi" is on PATH and returns a metric snapshot
    When the privileged collector runs
    Then it reports "gpu.util" as a percent
    And it reports "gpu.vram" as a percent with a "used / total GB" detail
    And it reports "gpu.temp" in celsius
    And it reports "power.gpu" in watts
    And it does not shell out to sysfs for any name amd-smi already provided

  Scenario: amd-smi absent falls back to amdgpu sysfs
    Given a Linux host with an AMD GPU
    And "amd-smi" is not on PATH
    And the amdgpu sysfs nodes are readable
    When the privileged collector runs
    Then it reads gpu_busy_percent for "gpu.util"
    And it reads mem_info_vram_used and mem_info_vram_total for "gpu.vram"
    And it reads the hwmon power1_average for "power.gpu"
    And it reads the hwmon temp1_input for "gpu.temp"

  Scenario: AMD readings do not shadow an existing source
    Given a host where an earlier source already set "power.gpu"
    When the AMD source runs and also produces "power.gpu"
    Then the earlier value is kept and the AMD value is dropped

  Scenario: best-effort NPU when the driver does not expose it
    Given an AMD host whose amdxdna driver exposes no utilisation counter
    When the privileged collector runs
    Then "npu.util" is reported as Unavailable
    And the GPU metrics are still reported normally

  Scenario: a non-AMD host is unaffected
    Given a host with no AMD GPU and no amd-smi and no amdgpu sysfs nodes
    When the privileged collector runs
    Then no "gpu.*" key originates from the AMD source
    And the NVIDIA and Apple paths behave exactly as before

  Scenario: malformed amd-smi output degrades gracefully
    Given "amd-smi" is on PATH but returns empty or unparseable output
    When the privileged collector runs
    Then the AMD source yields no metrics
    And the collector falls back to amdgpu sysfs
```

## Definition of Done

- [ ] Unit tests green (amd-smi parser + sysfs parser, table-driven)
- [ ] Integration tests green (PrivilegedMetrics merge precedence with a fake AMD source)
- [ ] Cucumber/Behave scenarios green
- [ ] Non-Linux build still compiles (sysfs path behind a build tag / runtime guard)
- [ ] NVIDIA and Apple paths verified unchanged
- [ ] Docs: install guide for amd-smi (ROCm) + sysfs fallback note in metrics docs
- [ ] CHANGELOG entry added (v2.2.1)
- [ ] Validated live on CaptainCanuck: `gpu.util` / `power.gpu` / `gpu.vram` appear
- [ ] PR description references this story

## ADR Links

<!-- adr_required: false — additive collector, no architectural decision; the
     amd-smi-over-sysfs precedence is documented inline in the story Context. -->
