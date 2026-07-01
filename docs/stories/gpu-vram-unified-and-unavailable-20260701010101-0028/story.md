---
id: "gpu-vram-unified-and-unavailable-0028"
title: "GPU VRAM on unified memory + explain unavailable GPU metrics"
epic: "fleet-accelerator-coverage"
priority: "high"
ui: false
adr_required: false
milestone: "v2.2.5"
phase: discover
labels:
  - "type:bug"
  - "priority:high"
  - "area:adapters"
status: approved
issue_number: null
issue_url: null
created_at: "2026-07-01T01:01:01+0000"
---

# GPU VRAM on unified memory + explain unavailable GPU metrics

## Story

**As an** operator watching a mixed fleet,
**I want** a `gpu.vram` reading on unified-memory GPUs where one can be derived,
and a *reason* shown whenever a GPU metric can't be read,
**so that** a blank GPU/VRAM cell tells me *why* (unified memory, or a broken
`nvidia-smi`) instead of looking like Heimdall is silently failing.

## Context

Fleet observations (hub `192.168.1.229:9090`) surfaced three separate causes for
missing GPU/VRAM, all confirmed live over SSH:

| Host | Chip / OS | Symptom | Confirmed cause |
|------|-----------|---------|-----------------|
| `promaxgb10-bdd9` | GB10 Grace-Blackwell / Ubuntu ARM | `gpu.util` OK, **no `gpu.vram`** | `nvidia-smi` `memory.used/total` = `[N/A]` (unified LPDDR5X, no discrete VRAM) |
| Apple Silicon (MacBook, Mac mini) | M-series / macOS | **no `gpu.vram`** | no VRAM collection at all; unified memory, no carve-out |
| `homeserver`, `rtx-pro-workstation` | Intel + discrete NVIDIA / Ubuntu | **no `gpu.util`, no `gpu.vram`** | `nvidia-smi` fails: `Driver/library version mismatch` (kernel module `580.95.05` vs userspace `580.159`; reboot pending, 30-week uptime) — a **host** issue, not Heimdall |

The `homeserver`/`rtx-pro` case is a host driver-mismatch that a reboot fixes. It
is in scope here only for the *reporting* gap it exposed: when `nvidia-smi` is
present but exits non-zero, `runNvidiaSMI` (`app/internal/helper/collect.go`)
returns the error and the parser never runs, so every `gpu.*` key silently
disappears. The operator sees a blank and blames Heimdall. Surfacing the reason
would have turned a multi-host investigation into a one-glance "reboot me".

### Evidence — GB10 exposes used GPU memory per-process

Aggregate FB memory is `Not Supported` on GB10, but per-process GPU memory is
queryable:

```
$ nvidia-smi --query-compute-apps=pid,used_memory --format=csv,noheader,nounits
92075, 42625        # ollama llama-server holding 42.6 GB of the unified pool
```

So a meaningful `gpu.vram` can be derived: **used** = Σ compute-apps
`used_memory`; **total** = system RAM total (the physical ceiling of the unified
pool, `124546 MiB` here). This is NVIDIA-only — Apple has no equivalent
per-process GPU-memory query, so Apple stays "unavailable, with a reason".

### Design decisions (locked by operator)

1. **Unified-memory NVIDIA** (GB10 and any host where `nvidia-smi memory.total`
   is `N/A`): derive `gpu.vram` from Σ compute-apps `used_memory` over system RAM
   total, with `Detail` like `42.6 / 124.5 GB (shared)`.
2. **Apple Silicon**: emit `gpu.vram` as **Unavailable** with
   `Detail: "unified memory (no discrete VRAM)"` — no per-process source exists.
3. **`nvidia-smi` present but erroring**: emit `gpu.util` (and `gpu.vram`) as
   **Unavailable** carrying the trimmed `nvidia-smi` error as `Detail` (e.g.
   `nvidia-smi: driver/library version mismatch`) instead of dropping every key.

Out of scope: rebooting the affected hosts (operator action); any change to the
`Helper.Collect` `anyOK` gate (the helper-shadow hypothesis was investigated and
disproven for these hosts).

## Acceptance Criteria

```gherkin
@story:gpu-vram-unified-and-unavailable-0028 @epic:fleet-accelerator-coverage @priority:high
Feature: GPU VRAM on unified memory and explained-unavailable GPU metrics

  Scenario: Unified-memory NVIDIA derives gpu.vram from compute-apps
    Given nvidia-smi reports memory.used and memory.total as "[N/A]"
    And nvidia-smi --query-compute-apps=used_memory reports 42625 MiB in use
    And the host has 124546 MiB of system RAM
    When the privileged collector builds NVIDIA metrics
    Then gpu.vram is reported OK at 34 percent
    And its detail reads "42.6 / 124.5 GB (shared)"

  Scenario: Discrete-VRAM NVIDIA is unchanged
    Given nvidia-smi reports memory.used=2048 and memory.total=8192
    When the privileged collector builds NVIDIA metrics
    Then gpu.vram is reported OK at 25 percent from the aggregate counter
    And the compute-apps fallback is not used

  Scenario: Apple Silicon explains the missing VRAM
    Given the host is Apple Silicon with GPU metrics available
    When the privileged collector builds Apple metrics
    Then gpu.vram is reported Unavailable
    And its detail reads "unified memory (no discrete VRAM)"

  Scenario: A broken nvidia-smi surfaces the reason instead of blanking
    Given nvidia-smi is installed
    But nvidia-smi exits non-zero with "Failed to initialize NVML: Driver/library version mismatch"
    When the privileged collector queries the NVIDIA GPU
    Then gpu.util is reported Unavailable
    And its detail contains "driver/library version mismatch"

  Scenario: No nvidia-smi present contributes nothing
    Given nvidia-smi is not on PATH
    When the privileged collector queries the NVIDIA GPU
    Then no gpu.* metric is emitted by the NVIDIA path
```

## Definition of Done

- [ ] Unit tests green
- [ ] Integration tests green
- [ ] Cucumber/Behave scenarios green
- [ ] Wireframe approved (required when `ui: true`) — N/A
- [ ] Mockup approved (required when `ui: true`) — N/A
- [ ] A11y audit passed (required when `ui: true`) — N/A
- [ ] ADRs linked where required — none
- [ ] CHANGELOG entry added
- [ ] PR description references this story

## ADR Links

<!-- none required -->
