---
adr: "0002"
title: "Daemon enrollment identity and mTLS"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
  - "Security"
---

# 0002 — Daemon enrollment identity and mTLS

## 1. Context

Daemons run on machines outside the hub's trust zone, dial **outbound** over
untrusted networks, and reconnect frequently on flaky links. The hub must (a)
authenticate which daemon is connecting, (b) give each host a **stable identity**
that survives restarts and reconnects so it never double-registers, and (c) keep
the bootstrap simple enough to deploy across a heterogeneous fleet
(Windows/macOS/Linux, workstation → Raspberry Pi).

## 2. Goals / Non-Goals

**Goals:**
- Authenticate daemons to the hub and the hub to daemons.
- A stable `host_id` that is identical across reconnects (no duplicate hosts).
- A low-friction bootstrap that upgrades to strong per-host credentials.

**Non-Goals:**
- A full PKI/secrets-management product (integrate with one; don't build one).
- Authenticating human operators (that is the dashboard/control-plane concern,
  ADR-0007).

## 3. Proposal

Two-phase trust:

1. **Enroll** — the daemon calls `EnrollmentService.Enroll` over TLS (validating
   the hub's server cert against a pinned CA), presenting a **scoped, one-time
   enrollment token** and a self-declared `Host` (stable `host_id` + `HostContext`).
   Optionally it includes a CSR; the hub returns an issued **client certificate**
   plus stream policy (`sample_interval`, `stale_after`, `offline_after`).
2. **Stream** — subsequent `MetricStreamService.Stream` connections use **mTLS**
   with the issued client cert. The cert subject is bound to `host_id`.

`host_id` is **content-derived and stable**: a hash of a durable machine identifier
plus the daemon's public key. A returning daemon presents the same `host_id` and
its cert; the hub reactivates the existing registry entry (`OFFLINE → ONLINE`)
rather than creating a new one. A second party claiming an enrolled `host_id`
without the matching cert is rejected.

Enrollment tokens are scoped (time/host-count limited), one-time, and never logged.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Shared static API key for all daemons | Trivial | One leak compromises the fleet; no per-host identity | Fails auth + revocation |
| mTLS only, manual cert per host | Strong | Heavy ops to provision certs across a mixed fleet | Too much bootstrap friction |
| Token-only (no mTLS) | Simple | Bearer token repl/leak risk on every connect | Weaker than issued certs |
| Hostname/IP as identity | No bootstrap | Hostnames/IPs change; spoofable; duplicates on reconnect | Fails stable-identity goal |

## 5. Trade-offs and Risks

- The hub becomes a certificate issuer (or must integrate one) — added component.
  Mitigated by keeping issuance minimal and pluggable.
- Enrollment-token distribution is the weakest link; mitigated by scoping,
  one-time use, and rejection logging.
- Clock skew can affect cert/token validity; mitigated by reasonable skew windows.

## 6. Impact

**FinOps:** Minimal. No new managed service required if issuance is embedded;
optional integration with an existing CA/secrets manager.

**SRE:** New failure mode — cert expiry/rotation. Runbooks: rotate hub CA, reissue
client certs, revoke a compromised host. Reconnect/resume keeps a cert blip from
losing history beyond the ring window.

**Security:** Strong per-host identity, revocable; bounded blast radius on a single
host compromise. Enrollment token is the critical secret (see threat-model
Boundary 1). mTLS protects telemetry confidentiality + integrity.

**Team:** Engineers need basic TLS/PKI literacy. Bootstrap is documented; daemon
install ships a token, not a private key.

## 7. Decision

Bootstrap daemons with a scoped one-time enrollment token over server-validated
TLS, then issue per-host client certs for mTLS on the metric stream. Identify
hosts by a stable, content-derived `host_id` bound to the cert so reconnects never
double-register. This balances fleet-scale bootstrap friction against strong,
revocable per-host identity.

Status: **proposed**

## 8. Next Steps

- [ ] Define `host_id` derivation per OS (durable machine id sources) — Architect
- [ ] Implement enroll → CSR → issued-cert flow and rejection paths — daemon/hub
- [ ] Add enrollment-token rejection + reconnect-no-duplicate integration tests — QA
