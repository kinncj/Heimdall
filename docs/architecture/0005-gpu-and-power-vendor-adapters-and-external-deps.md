---
adr: "0005"
title: "GPU and power vendor adapters and external dependencies"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
---

# 0005 — GPU and power vendor adapters and external dependencies

## 1. Context

GPU and power metrics depend on vendor-specific interfaces: NVIDIA NVML, AMD
counters, Apple SMC/IOKit/`powermetrics`, Linux RAPL/`/sys`, Windows WMI/PDH. Some
require CGO or `dlopen` of a vendor library and may not be present at build or run
time. Heimdall must cross-compile cleanly to static binaries across
Windows/macOS/Linux × amd64/arm64 and run on devices from DGX Spark to Raspberry
Pi (E2, story `gpu-and-power-metric-adapters-0006`). Pulling CGO into the core
binaries would break the static cross-compilation guarantee.

## 2. Goals / Non-Goals

**Goals:**
- Keep core binaries (`daemon`, `hub`, `dashboard`) pure-Go and statically
  cross-compilable.
- Support vendor GPU/power signals where available; degrade where not.
- Contain optional/native dependencies behind the `Adapter` contract.

**Non-Goals:**
- Guaranteeing every metric on every device (hardware varies; report
  `UNAVAILABLE`).
- Bundling proprietary vendor libraries we cannot redistribute.

## 3. Proposal

- All vendor adapters implement the domain `Adapter` interface (ADR-0003); the rest
  of the system is vendor-agnostic.
- **Quarantine native code.** Prefer obtaining privileged/native counters through
  the **helper** (ADR-0004) or behind **Go build tags** in dedicated adapter files,
  so the default core build stays `CGO_ENABLED=0`. Where a vendor lib is needed,
  load it at runtime (`dlopen`) and treat absence as `UNAVAILABLE` rather than a
  build/link failure.
- **Runtime detection.** Each vendor adapter probes for its dependency at startup:
  present → `OK` data; missing library → `UNAVAILABLE`; present but permission-gated
  → `INSUFFICIENT_PERMISSION` (defer to helper).
- **Build matrix.** `make build` cross-compiles the core binaries for the full host
  matrix without CGO. Native-dependent variants, if any, are produced as opt-in,
  platform-tagged builds — never required for the default fleet binary.
- Device class (`dgx-spark`, `strix-halo`, `mac-mini`, `raspberry-pi`, etc.) travels
  in `HostContext.labels`, not in code branches.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Link NVML/vendor libs into core binaries | Direct access | Forces CGO; breaks static cross-compile; link fails where lib absent | Fails cross-compile goal |
| Shell out to vendor CLIs (`nvidia-smi`) | No linking | Fragile parsing; exec surface; not always installed | Reliability + security |
| Drop GPU/power metrics | Simplest | Loses a headline capability for DGX/Strix targets | Fails coverage goal |
| `dlopen` + build tags + helper (chosen) | Static core; optional native | More build configurations | Accepted trade-off |

## 5. Trade-offs and Risks

- More build configurations (tags, optional native variants) add CI matrix
  complexity; mitigated by keeping the **default** fleet binary pure-Go.
- `dlopen` paths vary per OS/driver; mitigated by runtime probing and clean
  `UNAVAILABLE` fallback.
- Vendor API churn (NVML versions); mitigated by isolating it in one adapter behind
  the stable domain contract.

## 6. Impact

**FinOps:** None directly. Avoiding a TSDB and managed exporters keeps GPU metrics
free beyond compute.

**SRE:** Vendor adapter failure is per-metric (ADR-0003). Observability:
`UNAVAILABLE`/`INSUFFICIENT_PERMISSION`/`ERROR` distinguish "no GPU" from "driver
missing" from "needs helper". Runbook: enable GPU metrics on NVIDIA/AMD/Apple hosts.

**Security:** No shelling out to vendor CLIs (avoids an exec surface). Native libs
are `dlopen`ed read-only for counters; privileged paths go through the constrained
helper.

**Team:** Contributors need vendor-API knowledge only within a single adapter.
Build-tag discipline documented in the Makefile.

## 7. Decision

Implement GPU/power as vendor adapters behind the domain `Adapter` contract,
quarantining native dependencies behind build tags and the privileged helper, with
runtime probing that degrades to `UNAVAILABLE`/`INSUFFICIENT_PERMISSION`. The
default fleet binaries stay pure-Go and statically cross-compilable across the full
host matrix.

Status: **proposed**

## 8. Next Steps

- [ ] Enumerate per-vendor counter sources and their privilege needs — Architect
- [ ] Implement NVML/RAPL/SMC adapters with runtime probing — adapters
- [ ] Verify `CGO_ENABLED=0` cross-compile of core binaries in CI matrix — INFRA
