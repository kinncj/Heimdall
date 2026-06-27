---
adr: "0015"
title: "Windows privileged metrics"
status: accepted
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0015 — Windows privileged metrics

## 1. Context

The privileged helper (`app/internal/helper`) gives Heimdall the metrics the
unprivileged daemon cannot read. It follows a **shell-out pattern** (ADR-0004/0005):
on macOS it runs `powermetrics`; for GPU it runs `nvidia-smi`; on Linux it reads
`/sys` for RAPL/hwmon. Every adapter returns `Unavailable` for metrics it cannot
produce rather than failing (ADR-0003) — the daemon stays up and reports partial
data.

Windows has no path today. A Windows host runs the daemon with CPU temperature and
package power blank, even when the hardware exposes them. ADR-0010's E2 note flagged
cross-platform privileged-metrics parity as an extension of ADR-0004/0005 needing no
new architecture — this ADR records the concrete Windows decision and its one real
limitation.

GPU is already covered: `nvidia-smi` is cross-platform, so GPU metrics work on Windows
through the existing adapter with no change.

## 2. Goals / Non-Goals

**Goals:**
- A **Windows helper path** consistent with the existing shell-out pattern.
- **CPU package temperature** (`temp.pkg`) via PowerShell WMI:
  `Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature`,
  parsing `CurrentTemperature` (tenths of a Kelvin) to Celsius.
- **Build-tagged** (`windows`) so non-Windows builds are unaffected.
- A **pure parser**, unit-tested **cross-platform** (parsing logic separate from the
  shell-out).
- **No new dependency** — PowerShell `Get-CimInstance` ships with Windows; `nvidia-smi`
  GPU already works.

**Non-Goals:**
- **CPU package power (`power.pkg`) on Windows** — documented non-goal (see below).
- A kernel driver, service, or LibreHardwareMonitor bundle.
- Per-core temperatures or non-ACPI thermal sources.
- Changing the helper contract, privilege model, or transport (ADR-0003/0004 stand).

## 3. Proposal

**Windows temperature adapter.** Add a `windows`-build-tagged helper adapter that
shells out to PowerShell:

```
Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature
```

Read `CurrentTemperature`, which the ACPI thermal-zone interface reports in **tenths
of a Kelvin**. Convert: `celsius = currentTemperature/10 - 273.15`. Emit as
`temp.pkg`. If multiple thermal zones return, select deterministically (documented:
e.g. the first/primary zone) and treat absent/zero readings as `Unavailable`.

**Parser is pure and isolated (SOLID).** Split the adapter into:
- a thin **shell-out** layer (build-tagged `windows`) that invokes PowerShell and
  captures stdout, and
- a **pure parser** `func(raw string) (celsius float64, ok bool)` with **no OS
  dependency**, unit-tested on any platform.

SRP: the parser does tenths-of-Kelvin→Celsius and nothing else. DIP: the adapter
depends on the parser, not vice versa. OCP: a future Windows power source (driver,
LibreHardwareMonitor) adds an adapter behind the same metric-adapter contract without
touching the temperature path.

**Power stays Unavailable (ADR-0003).** On Windows, `power.pkg` returns
`Unavailable`. Intel RAPL MSRs are **not** reachable from user space without a kernel
driver; the common third-party path (LibreHardwareMonitor) is an extra dependency and
a driver surface we decline. Per ADR-0003, returning `Unavailable` keeps the daemon
healthy and the metric honestly blank — not a failure.

**GPU unchanged.** `nvidia-smi` is invoked the same way on Windows as elsewhere; the
existing GPU adapter covers it. No work beyond confirming the binary is on `PATH`.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| PowerShell WMI `MSAcpi_ThermalZoneTemperature` → `temp.pkg`; power Unavailable (chosen) | No new dep; matches shell-out pattern; honest gaps | ACPI zone temp is coarse; no power | Accepted — best parity within the no-dep constraint |
| Bundle LibreHardwareMonitor for temp + power | Per-core temp + package power | Extra dependency; ships a kernel driver; larger attack surface | Violates no-new-dependency; driver surface declined |
| Ship a custom kernel driver to read RAPL MSRs | Real package power | Driver signing, maintenance, privilege/security burden | Disproportionate; out of scope |
| Direct WMI from Go via a binding lib | No PowerShell process | New dependency; CGo/Win32 surface | Breaks no-dep + shell-out consistency |
| Leave Windows with no privileged metrics | Zero work | Temp available on hardware but unread | Needless gap; ACPI temp is free |

## 5. Trade-offs and Risks

- **Coarse temperature.** ACPI thermal zones are motherboard/zone-level, not per-core
  die temps; `temp.pkg` on Windows may read lower or differently than RAPL-derived
  package temp on Linux/macOS. Documented; it is a reasonable package proxy.
- **Power gap is permanent within constraints.** `power.pkg` Unavailable on Windows is
  a deliberate, documented non-goal, not a TODO. Revisiting it means accepting a
  driver/dependency and a new ADR.
- **PowerShell startup cost.** Spawning `Get-CimInstance` per poll has process
  overhead. Mitigated by the existing helper poll cadence; bounded like the other
  shell-outs. Tighten cadence later if needed.
- **WMI variability.** Some systems/VMs return no thermal zone or a zero reading;
  the adapter yields `Unavailable` rather than a bogus value (ADR-0003).
- **Build-tag discipline.** The shell-out must be `//go:build windows`; the parser must
  not be, so it tests cross-platform. A misplaced tag breaks the build or the test
  matrix.

## 6. Impact

**FinOps:** Zero. No cloud cost, no new dependency, no managed service — a build-tagged
adapter and a PowerShell invocation on the monitored host.

**SRE:** Brings Windows hosts to near-parity: CPU temperature and (existing) GPU
metrics, with package power honestly `Unavailable`. Failure modes are local to the
host and isolated by ADR-0003 — a missing thermal zone or PowerShell error degrades to
`Unavailable`, never a daemon crash. Observability: the same `Unavailable` signaling
operators already read on macOS/Linux for absent metrics.

**Security:** Small, contained surface. Shells out to the platform's own
`Get-CimInstance` and `nvidia-smi` — no third-party binary, **no kernel driver**, no
new dependency. Declining LibreHardwareMonitor/RAPL-driver paths specifically avoids
adding a signed-driver attack surface. No new privilege beyond the helper's existing
tier (ADR-0004).

**Team:** One platform adapter to learn, following the established shell-out + pure-
parser shape. Cross-platform parser tests mean Windows behavior is verifiable on CI
without Windows runners for the parsing logic.

## 7. Decision

Give the helper a Windows privileged-metrics path consistent with the existing
shell-out pattern: a `windows`-build-tagged adapter runs PowerShell
`Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature`,
and a pure, cross-platform-tested parser converts `CurrentTemperature` (tenths of a
Kelvin) to Celsius as `temp.pkg`. CPU package power (`power.pkg`) is a **documented
non-goal** on Windows — RAPL is not reachable without a kernel driver or
LibreHardwareMonitor, so it stays `Unavailable` per ADR-0003. No new dependency;
`nvidia-smi` GPU already works cross-platform.

Status: **accepted**

## 8. Next Steps

- [ ] Add the `windows`-tagged temperature adapter (shell-out to `Get-CimInstance`) — `app/internal/helper`
- [ ] Add the pure tenths-of-Kelvin→Celsius parser with cross-platform unit tests — `app/internal/helper`
- [ ] Confirm `power.pkg` returns `Unavailable` on Windows and document why — `app/internal/helper`
- [ ] Verify `nvidia-smi` GPU adapter on Windows (PATH check only) — helper
- [ ] Document Windows parity matrix (temp ✓, GPU ✓, power Unavailable) — Architect
- [ ] Implements story 0020 (Windows metrics); extends ADR-0004/0005 per the ADR-0010 E2 note — Architect
