---
adr: "0007"
title: "Unprivileged remote control plane"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
  - "Security"
---

# 0007 — Unprivileged remote control plane

## 1. Context

Operators want to run lightweight diagnostics on a remote host from the dashboard
(e.g. list processes, check disk usage, view interface state) without SSHing in.
Any "run a command on a remote host" feature is inherently RCE-shaped and is the
single highest-risk surface in Heimdall. It must be designed so it **cannot** become
arbitrary or privileged execution (E5, story
`unprivileged-terminal-control-plane-0010`). The daemon dials the hub outbound, so
control requests are **brokered hub→daemon over the existing authenticated stream**;
the hub cannot open a connection to the host.

## 2. Goals / Non-Goals

**Goals:**
- Allow a fixed set of **read-only** diagnostic commands from the dashboard.
- Execute **as the unprivileged daemon user** — never sudo, setuid, or shell.
- Validate arguments, bound output, and **audit** every request.

**Non-Goals:**
- Any write/remediation/config-change action (explicit non-goal in v1).
- Arbitrary command execution or interactive shells.
- Routing control through the privileged helper (ADR-0004) — forbidden.

## 3. Proposal

`ControlPlaneService.Execute` is a single authenticated, audit-logged bidi stream.
A `ControlRequest` carries `allowlisted_cmd` (a **logical key**, not a shell
string), typed `args[]`, and an authenticated `actor`.

Enforcement, in order:
1. **AuthN/Z** — the dashboard is authenticated to the hub; `actor` is the
   authenticated principal (not client-asserted). Requests reach the daemon only
   over its authenticated stream.
2. **Allow-list** — `allowlisted_cmd` maps via a **static** table to a fixed
   executable path and an argument validator. Unknown key ⇒
   `INSUFFICIENT_PERMISSION`.
3. **Argument validation** — each arg is checked by type (enum/regex/range); path
   args resolve against an allow-list of roots. No `sh -c`, no glob, no env
   expansion, no shell at any point. Args are passed as `argv`.
4. **Execution** — run as the **unprivileged daemon user**, with a per-command
   timeout, output size cap (`truncated=true` when hit), and bounded concurrency so
   sampling is never starved.
5. **Audit** — append-only log + `control.audit` bus event: `request_id`, `actor`,
   `host_id`, `cmd`, `args`, `exit`, decision, `ts`.

The full set is asserted by the **general-protection** integration test: allow-list
enforcement, `INSUFFICIENT_PERMISSION` handling, no-sudo enforcement, and
enrollment-token rejection.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Generic remote shell / `exec` | Maximal flexibility | Unbounded RCE; catastrophic if abused | Unacceptable risk |
| Allow-list but run via `sh -c` | Easy arg passing | Shell metacharacter injection | Defeats the allow-list |
| Allow privileged (sudo) reads | More data | Escalation path on every host | Violates unprivileged posture |
| Allow-listed, typed-argv, as-user (chosen) | Bounded, auditable | Limited command set | Accepted by design |

## 5. Trade-offs and Risks

- The command set is deliberately small; some diagnostics will be unavailable in
  v1. Acceptable — extend the allow-list deliberately, with review.
- Even read-only commands can disclose data; mitigated by curating the allow-list
  to non-sensitive reads, capping output, and auditing.
- Brokering over the daemon stream couples control to stream health; acceptable
  since the daemon must be connected to be controllable anyway.

## 6. Impact

**FinOps:** Negligible compute; no new infrastructure.

**SRE:** Per-command timeout + concurrency cap protect the daemon. Audit log
supports incident forensics. Runbook: extend the allow-list (review + test),
investigate a denied/suspicious control request.

**Security:** The defining control. Read-only, static allow-list, typed argv, no
shell, no sudo, as-user, audited, helper-unreachable (threat-model Boundary 3).
Designed so the worst case is a bounded, audited, unprivileged read.

**Team:** Adding a command requires an allow-list entry + validator + test +
security review — intentional friction.

## 7. Decision

Implement the control plane as a single authenticated, audited bidi stream that
runs only **static-allow-listed, read-only** commands with **typed argv**, executed
**as the unprivileged daemon user** with no shell, no sudo, and no path to the
privileged helper, brokered hub→daemon over the existing authenticated stream. The
design makes arbitrary or privileged execution structurally impossible, not merely
discouraged.

Status: **proposed**

## 8. Next Steps

- [ ] Define the v1 allow-list (commands, executables, arg validators) — Architect + Security
- [ ] Implement executor (as-user, no-shell, timeout, output cap, audit) — daemon
- [ ] Implement the general-protection integration test — QA
- [ ] Document the allow-list extension + review procedure — Architect
