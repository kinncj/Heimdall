---
adr: "0004"
title: "Optional privileged helper and privilege tiers"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
  - "Security"
---

# 0004 — Optional privileged helper and privilege tiers

## 1. Context

Some signals need elevated privileges: CPU package power/RAPL, full thermal
sensors, and certain GPU counters. Running the entire daemon as root to obtain them
would make a network-facing, adapter-loading process the largest possible attack
surface. Most metrics need no privileges at all. We want full coverage **and** an
unprivileged-by-default posture (E2, story
`optional-privileged-metrics-helper-0005`).

## 2. Goals / Non-Goals

**Goals:**
- Run the network-facing daemon **unprivileged** by default.
- Obtain privileged counters via a **separate, minimal** privileged component.
- Degrade gracefully (`INSUFFICIENT_PERMISSION` / `UNAVAILABLE`) when the helper is
  absent.

**Non-Goals:**
- Granting the helper any general-purpose capability (it is not a privileged shell).
- Coupling the control plane to the helper (control is unprivileged; ADR-0007).

## 3. Proposal

Two privilege tiers on a host:

- **Tier 0 — `heimdall-daemon` (unprivileged):** owns adapters, sampler,
  transport, control executor, log tailer. This is the only process on the network.
- **Tier 1 — `heimdall-helper` (privileged, separate unit):** exposes **only** a
  fixed, read-only counter set (RAPL/power, full thermal, extra GPU counters) over
  a **local Unix domain socket**. It has **no network listener** and **no execute
  API**.

The daemon connects to the helper's socket; the socket is owner/permission
restricted and **peer-credential checked** (only the daemon uid may connect). The
helper answers a typed counter schema only — it accepts no command, path, or
argument selecting code to run. Helper-backed adapters implement the same `Adapter`
contract (ADR-0003); when the helper is not installed/running, those adapters
report `INSUFFICIENT_PERMISSION` or `UNAVAILABLE` and the daemon keeps running.

Deployment: separate systemd unit / launchd plist; the helper is optional and
independently installable.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Run whole daemon as root | Simple; full access | Huge attack surface on a network process | Fails unprivileged-default |
| setuid/capabilities on the daemon | No second process | Privilege embedded in the network-facing binary | Surface still too large |
| Drop privileged metrics entirely | Smallest surface | Loses power/thermal/GPU coverage | Fails coverage goal |
| Helper with a generic exec API | Flexible | A privileged executor = escalation magnet | Unacceptable risk |

## 5. Trade-offs and Risks

- Two components to package and operate per host. Mitigated by making the helper
  optional and the daemon fully functional without it.
- IPC schema is another contract to version. Mitigated by keeping it a tiny fixed
  counter set.
- A compromised helper is privileged; mitigated by zero execute API, no network,
  peer-cred checks, and minimal code.

## 6. Impact

**FinOps:** Negligible; one extra lightweight local process where installed.

**SRE:** Helper failure degrades affected metrics only; never crashes the daemon.
Runbook: install/enable helper, diagnose `INSUFFICIENT_PERMISSION` on power/thermal.

**Security:** Major reduction in attack surface — the privileged code is tiny, has
no network, and cannot execute arbitrary work (threat-model Boundary 4). Clear
privilege separation is the core security property of the host.

**Team:** Slightly more packaging work (two units). Helper code is small and
security-reviewed; most contributors only touch Tier-0 adapters.

## 7. Decision

Default to an unprivileged daemon and obtain privileged counters from an optional,
minimal `heimdall-helper` over a peer-cred-checked local socket exposing a fixed
read-only counter set with no execute API and no network. Absent the helper,
affected metrics degrade cleanly. This buys full coverage without making the
network process privileged.

Status: **proposed**

## 8. Next Steps

- [ ] Specify the helper IPC counter schema (fixed, read-only) — Architect
- [ ] Implement peer-cred check + restrictive socket perms — helper
- [ ] Package systemd unit + launchd plist for the helper — INFRA
- [ ] Add `INSUFFICIENT_PERMISSION` degradation test when helper absent — QA
