# Threat Model — Heimdall

STRIDE analysis across Heimdall's four trust boundaries. The wire contract is
`common/proto/monitoring/v1/monitoring.proto`; the architecture is in
`architecture.md`. The highest-risk boundary is the **unprivileged control plane**
(dashboard → daemon); it gets the deepest treatment.

## Assets

| Asset | Sensitivity | Owner |
|---|---|---|
| Enrollment token (bootstrap secret) | Critical | Platform |
| mTLS private keys / issued client certs | Critical | Platform |
| Cross-hub federation credentials | Critical | Platform |
| Host telemetry (metrics, host context) | Medium | Fleet owner |
| Opt-in log lines (may contain secrets) | High | Host owner |
| Control-plane command surface | Critical | Platform |
| Audit log | High | Security |
| Hub in-memory state (registry, ring buffers) | Medium | Platform |

## Trust boundaries

1. **daemon ↔ hub** — host to central/cloud over an untrusted network.
2. **hub ↔ relay (federation)** — child hub to parent hub across networks/orgs.
3. **dashboard ↔ daemon (control plane)** — operator action reaching a host.
4. **helper ↔ daemon (privilege)** — privileged process to unprivileged process
   on the same host.

```
[daemon]==TLS/mTLS==>[hub]==mTLS==>[parent hub]
   ^                    ^
   | local socket       | bus subscribe
[helper]            [dashboard]==control/log (brokered hub->daemon)==>[daemon]
```

---

## Boundary 1 — daemon ↔ hub

**S — Spoofing.** A rogue daemon impersonates a host, or a rogue server
impersonates the hub.
- Likelihood: Medium · Impact: High
- Mitigation: TLS with hub server-cert validation (daemon pins/validates the hub
  CA). Daemon authenticates with a scoped, one-time **enrollment token**, then an
  issued **mTLS** client cert. Stable `host_id` is confirmed by the hub; a second
  daemon claiming an enrolled `host_id` without the matching cert is rejected.

**T — Tampering.** Snapshot frames altered in transit.
- Likelihood: Low · Impact: Medium
- Mitigation: TLS integrity. `seq` monotonicity + keyframe resync detect
  injection/replay gaps; `(host_id, seq)` de-dupe makes replays idempotent.

**R — Repudiation.** A host denies sending data, or the hub denies receipt.
- Likelihood: Low · Impact: Low
- Mitigation: mTLS client identity bound to `host_id`; ingest logs `(host_id, seq,
  ts)` with the authenticated cert subject.

**I — Information Disclosure.** Telemetry leaks to a network observer.
- Likelihood: Medium · Impact: Medium
- Mitigation: TLS 1.3. No secrets in metric payloads by design. Enrollment tokens
  are never logged.

**D — Denial of Service.** A flood of enroll attempts or oversized frames exhausts
the hub.
- Likelihood: Medium · Impact: High
- Mitigation: Per-IP and per-token enroll rate limits; max message size; bounded
  ring buffers; per-host cadence caps via `CadenceUpdate`; back-pressure drops
  oldest rather than blocking ingest.

**E — Elevation of Privilege.** A daemon connection yields hub-side control.
- Likelihood: Low · Impact: High
- Mitigation: The ingest path only writes registry/ring-buffer state; it exposes no
  administrative RPC. Daemons cannot subscribe to other hosts or issue control.

---

## Boundary 2 — hub ↔ relay (federation, E4)

**S — Spoofing.** A rogue hub joins as a child/parent.
- Likelihood: Medium · Impact: High
- Mitigation: Cross-hub **mTLS** + federation token; both ends validate certs.
  `origin_hub_id` is bound to the authenticated peer.

**T — Tampering / loop injection.** A malicious relay forges `path[]` to cause
loops or duplicate storms.
- Likelihood: Medium · Impact: Medium
- Mitigation: A hub drops any envelope whose `path[]` already contains its own id
  (loop prevention); `(origin_hub_id, host_id, seq)` de-dupe (duplication
  prevention); per-peer ingest rate limit.

**R — Repudiation.** A hub denies relaying.
- Likelihood: Low · Impact: Low
- Mitigation: Relay logs authenticated peer id + `origin_hub_id` + `seq` range.

**I — Information Disclosure.** Upstream relay exposes a whole fleet to a parent
org.
- Likelihood: Medium · Impact: High
- Mitigation: mTLS; `SubscribeRequest.host_ids` scoping limits what a peer may
  receive; relay is explicit opt-in config, not automatic.

**D — Denial of Service.** A child floods the parent (or vice versa).
- Likelihood: Medium · Impact: Medium
- Mitigation: Per-peer rate limits and max in-flight; bounded subscriber queues
  with oldest-drop; relay backoff.

**E — Elevation of Privilege.** A federated peer issues control commands across the
boundary.
- Likelihood: Low · Impact: Critical
- Mitigation: `FederationService` carries **telemetry only** — no control or log
  RPC is reachable over federation. Control plane is not federated in v1.

---

## Boundary 3 — dashboard ↔ daemon (control plane, E5) — highest risk

This boundary is RCE-shaped: it ends in command execution on a host. It is
designed to be **read-only, allow-listed, unprivileged, shell-free, audited**.

**S — Spoofing.** An attacker impersonates an authorized operator/dashboard.
- Likelihood: Medium · Impact: Critical
- Mitigation: Dashboard authenticates to the hub (mTLS/token); `ControlRequest.actor`
  is the authenticated principal, not client-asserted. Requests are brokered over
  the authenticated daemon stream; an unauthenticated party cannot reach the daemon.

**T — Tampering.** Argument injection — smuggling a shell metacharacter or path to
widen a command.
- Likelihood: High · Impact: Critical
- Mitigation: `allowlisted_cmd` is a **logical key** mapped to a fixed executable
  and a typed argument validator — never a shell string. `args[]` are validated
  tokens (enum/regex/range), passed as `argv`, never concatenated into a shell.
  No `sh -c`, no glob, no env expansion. Path arguments resolved against an
  allow-list of roots.

**R — Repudiation.** An operator denies running a command.
- Likelihood: Medium · Impact: Medium
- Mitigation: Append-only **audit log**: `request_id`, `actor`, `host_id`, `cmd`,
  `args`, `exit`, decision (allowed/denied), `ts`. Emitted as a `control.audit` bus
  event.

**I — Information Disclosure.** A read-only command still leaks sensitive host data
(e.g. dumping a config with secrets).
- Likelihood: Medium · Impact: High
- Mitigation: Allow-list curated to diagnostic, non-sensitive reads (process list,
  `df`, interface state, uptime). Output is size-capped (`truncated=true`). No
  arbitrary file read; no command that prints credentials. Output is audited.

**D — Denial of Service.** A heavy or wedged command starves the host.
- Likelihood: Medium · Impact: Medium
- Mitigation: Per-command timeout + output cap; per-actor and per-host rate limits;
  the executor runs commands with bounded concurrency so the daemon's sampling is
  never starved.

**E — Elevation of Privilege.** The control path is abused to run privileged or
arbitrary code.
- Likelihood: High (if mishandled) · Impact: Critical
- Mitigation: Commands run **as the unprivileged daemon user**. **No sudo. No setuid.
  No shell. No arbitrary execution.** The daemon never forwards control requests to
  the privileged helper. Denials return `INSUFFICIENT_PERMISSION`. This entire set
  is asserted by the **general-protection** integration test (allow-list
  enforcement, `INSUFFICIENT_PERMISSION` handling, no-sudo enforcement, enrollment
  token rejection).

---

## Boundary 4 — helper ↔ daemon (privilege, E2)

The helper is privileged; the daemon is not. The daemon must not be able to turn
the helper into a general-purpose privileged executor.

**S — Spoofing.** A non-helper process answers on the local socket, or a non-daemon
process queries the helper.
- Likelihood: Low · Impact: High
- Mitigation: Unix domain socket with restrictive file permissions/owner; peer-cred
  (SO_PEERCRED / equivalent) check so only the daemon uid may connect and only the
  expected helper serves.

**T — Tampering.** Counter values manipulated between helper and daemon.
- Likelihood: Low · Impact: Low
- Mitigation: Local socket within the host trust zone; helper exposes a fixed,
  typed counter schema; daemon validates ranges.

**R — Repudiation.** N/A across a same-host local socket (covered by host audit).

**I — Information Disclosure.** Helper leaks privileged data beyond the agreed
counter set.
- Likelihood: Low · Impact: Medium
- Mitigation: Helper exposes **only** a fixed read-only counter set (RAPL/power,
  full thermal, extra GPU counters). No general file/registry read API.

**D — Denial of Service.** Daemon spams the helper.
- Likelihood: Low · Impact: Low
- Mitigation: Helper rate-limits requests; helper failure degrades affected metrics
  to `INSUFFICIENT_PERMISSION`/`UNAVAILABLE`, never crashing the daemon.

**E — Elevation of Privilege.** Daemon coerces the helper into arbitrary privileged
work.
- Likelihood: Medium · Impact: Critical
- Mitigation: The helper has **no execute API** — it answers a fixed set of counter
  requests only. It accepts no command, path, or argument that selects code to run.
  It never touches the network. Minimal attack surface by construction.

---

## Risk register

| # | Threat | Boundary | Likelihood | Impact | Risk | Mitigation | Status |
|---|---|---|---|---|---|---|---|
| 1 | Shell/arg injection via control plane | 3 | High | Critical | **Critical** | Logical-key allow-list, typed argv, no shell | Designed |
| 2 | Privilege escalation via control plane | 3 | High | Critical | **Critical** | As-user, no sudo/setuid, helper unreachable | Designed |
| 3 | Helper abused as privileged executor | 4 | Medium | Critical | **High** | Fixed counter API, no execute, peer-cred | Designed |
| 4 | Rogue hub joins federation | 2 | Medium | High | **High** | Cross-hub mTLS + token | Designed |
| 5 | Fleet telemetry exposed upstream | 2 | Medium | High | **High** | mTLS + host_id subscription scoping, opt-in | Designed |
| 6 | Daemon/host spoofing at enroll | 1 | Medium | High | **High** | Enrollment token + issued mTLS + stable host_id | Designed |
| 7 | Operator spoofing on control plane | 3 | Medium | Critical | **High** | Authenticated actor, brokered over auth stream | Designed |
| 8 | Sensitive read via allow-listed cmd | 3 | Medium | High | **High** | Curated read-only allow-list, output cap, audit | Designed |
| 9 | Enroll/ingest flood DoS on hub | 1 | Medium | High | **High** | Rate limits, max msg size, bounded buffers | Designed |
| 10 | Federation loop / duplicate storm | 2 | Medium | Medium | **Medium** | path[] loop drop, (origin,host,seq) de-dupe | Designed |
| 11 | Telemetry sniffing in transit | 1 | Medium | Medium | **Medium** | TLS 1.3 | Designed |
| 12 | Opt-in logs leak secrets | 3/6 | Medium | High | **High** | Opt-in only, source allow-list, rate cap, redaction guidance | Designed |
| 13 | Snapshot replay/tamper | 1 | Low | Medium | **Low** | TLS + seq monotonicity + de-dupe | Designed |
| 14 | Audit-log repudiation gap | 3 | Medium | Medium | **Medium** | Append-only audit with authenticated actor | Designed |

## Mitigations required before launch

- [ ] Control-plane allow-list is a static logical-key map; **no** `sh -c`, glob, or
      env expansion anywhere in the executor path.
- [ ] Control commands run as the unprivileged user; **no sudo/setuid**; helper is
      unreachable from the control path.
- [ ] `general-protection` integration test passes: allow-list, `INSUFFICIENT_PERMISSION`,
      no-sudo, enrollment-token rejection.
- [ ] mTLS enforced on daemon↔hub (post-enroll) and hub↔relay; server-cert
      validation on the daemon.
- [ ] Enrollment tokens scoped + one-time; never logged; rejection path tested.
- [ ] Helper exposes a fixed read-only counter set over a peer-cred-checked local
      socket; no network listener.
- [ ] Federation enforces `path[]` loop prevention and `(origin_hub_id, host_id,
      seq)` de-dupe; subscription scoping by `host_ids`.
- [ ] Opt-in logs: source allow-list + rate cap; redaction guidance documented;
      off by default.
- [ ] Append-only audit log for every control request with authenticated `actor`.
- [ ] Enroll/ingest rate limits and max message size configured on the hub.
