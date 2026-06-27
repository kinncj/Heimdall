---
adr: "0012"
title: "Gjallarhorn: alerting and notifications"
status: accepted
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0012 — Gjallarhorn: alerting and notifications

> **Gjallarhorn** /ˈɡjɑː.lar.hɔːrn/ (*GYAH-lar-horn*) — the horn Heimdall sounds to
> warn of Ragnarök. See the [glossary](../glossary.md).

## 1. Context

Heimdall shows live fleet state but does not tell anyone when something is wrong.
Operators want threshold alerts (e.g. `cpu.util > 90% for 5m`, `mem.used > 85%`)
and an outbound notification when an alert fires and clears. The hub already runs a
per-host **liveness state machine** (enrolling/online/stale/offline) over the same
snapshots that would drive thresholds. Building a parallel evaluation loop would
duplicate state and risk divergence. We extend the existing evaluation instead.

This adds Heimdall's first **outbound notification surface** (webhooks), which
introduces SSRF and secret-handling concerns the rest of the system has not had.

## 2. Goals / Non-Goals

**Goals:**
- **Declarative threshold rules** evaluated in the hub against the same snapshots
  that drive liveness — extend that evaluation, do not build a parallel one.
- An alert **lifecycle** (firing → resolved) with hysteresis / for-duration to
  avoid flapping.
- Surface alert state in the dashboard (fleet alert count + row highlight).
- **Webhook** notification (generic JSON POST, Slack/PagerDuty-compatible) on fire
  and on clear.
- Scope/route alerts by **tag** (ADR-0010).

**Non-Goals:**
- A full alerting platform (silences, schedules, on-call rotations, escalation
  policies) — route to PagerDuty/Slack for that.
- Email/SMS transports in v1 (webhook only; others are future transports).
- Storing alert history durably (alert state is in-memory, like ADR-0008/0011).

## 3. Proposal

**Evaluate alongside liveness.** The liveness state machine already consumes each
applied snapshot per host. Threshold evaluation hooks the **same** snapshot apply,
reading metric values the hub already holds. One evaluation pass, one source of
truth — no second ingestion path.

**Declarative rules.** A rule is `{ metric, comparator, threshold, for_duration,
selector }`, e.g. `cpu.util > 90% for 5m`. Rules live in the layered config (file +
JSON, flags/env for overrides) per the existing config system. The `selector` is a
**tag matcher** (ADR-0010 effective tags, host-over-hub merge), so a rule applies to
`env=prod` or `region=apac` without naming hosts.

**Lifecycle with hysteresis.** Each `(rule, host)` is a small state machine:
`ok → pending` when the condition first holds, `pending → firing` once it holds for
`for_duration`, `firing → resolved` when the condition clears (optionally with a
clear-duration). For-duration and clear-duration give hysteresis so a metric
oscillating around the threshold does not flap notifications. State is in-memory,
consistent with ADR-0008/0011 (lost on restart; re-derives from incoming
snapshots).

**Dashboard surfacing.** A fleet-level **alert count** and per-row **highlight** for
hosts with a firing alert. Alert state is just another read off the hub state the
dashboard already subscribes to — no new transport for the local view.

**Webhook notification.** On `firing` and on `resolved`, POST a generic JSON
payload (Slack/PagerDuty-compatible shape) to a configured URL. Fire **and** clear
are both sent so receivers can auto-resolve. Delivery is best-effort with bounded
retry/backoff; webhook failure must not block evaluation.

**Routing by tag.** Notification targets are selected by the same tag selectors, so
`region=apac` alerts go to the apac webhook. One rule → matched hosts → routed
target.

**SOLID.** SRP: rule evaluation, lifecycle, and notification transport are separate
concerns. OCP: a new transport (email/SMS) implements a `Notifier` interface;
webhook is the first implementation. DIP: evaluation depends on the snapshot
stream and a `Notifier` abstraction, not on HTTP directly. ISP: `Notifier` is
`Notify(ctx, AlertEvent) error` — nothing more.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Parallel alert evaluation loop | Decoupled from liveness | Duplicate snapshot consumption + divergent state | Violates single-source-of-truth |
| Offload to Prometheus Alertmanager via `/metrics` (ADR-0011) | No alerting code in hub | Requires operators to run/configure it; no in-dashboard alerts | Valid path but doesn't serve built-in v1 alerting |
| Bake in Slack/PagerDuty SDKs | Richer integrations | Heavy deps; per-vendor coupling | Generic webhook covers both; stays light |
| Email/SMS transports in v1 | Familiar | SMTP/provider deps + secrets; scope creep | Deferred behind the `Notifier` seam |
| Threshold eval extending liveness + webhook `Notifier` + tag routing (chosen) | One eval pass; light; extensible | First outbound surface (SSRF/secrets) | Accepted |

## 5. Trade-offs and Risks

- **SSRF.** Operator-configured webhook URLs can target internal services. Mitigate:
  validate the URL scheme (https), optionally allowlist/denylist destination
  ranges, and document that webhook config is a trusted operator action.
- **No secrets in payload.** The JSON body carries host id, metric, value,
  threshold, tags, and state — **no tokens, no TLS material**. Auth to the receiver
  (e.g. Slack signing) rides the URL/header config, not the body.
- **Notification storms.** A correlated outage can fire many alerts at once.
  Mitigated by for-duration hysteresis and bounded per-target send rate; full
  grouping/silences are a non-goal (route to a real alerting platform).
- **Restart loses alert state.** Consistent with ADR-0008/0011; state re-derives
  from incoming snapshots. A brief re-fire window after restart is possible and
  documented.
- **Webhook receiver down** must not stall evaluation; delivery is async with
  bounded retry, then dropped (logged + a `self.alert.delivery_failed` metric).

## 6. Impact

**FinOps:** Negligible. In-memory rule state and outbound HTTP POSTs only; no
storage tier, no managed alerting service. Reuses the operator's existing
Slack/PagerDuty spend.

**SRE:** Turns Heimdall from passive to actionable. Failure modes: missed alert
(eval bug — covered by tests against the liveness stream), flapping (hysteresis),
webhook delivery failure (retry + self-metric). Observability:
`self.alert.{active,fired,resolved,delivery_failed}`. Runbook: write a rule, test a
webhook, interpret a re-fire after restart.

**Security:** First **outbound** surface. Threat model: SSRF via webhook URL
(scheme validation + optional egress allowlist), data exposure in payload (no
secrets, low-sensitivity fields only), DoS via storm (hysteresis + rate cap).
Webhook config is a trusted-operator action and documented as such.

**Team:** New concepts: rule syntax, lifecycle states, the `Notifier` seam. All
small. No new runtime dependency (stdlib HTTP client).

## 7. Decision

Add declarative threshold alerting evaluated in the hub against the **same**
snapshots that drive the liveness state machine — extending that evaluation rather
than building a parallel one. Each `(rule, host)` runs a firing→resolved lifecycle
with for-duration/clear-duration hysteresis to prevent flapping. Alert state shows
in the dashboard (count + row highlight) and notifies via a generic
Slack/PagerDuty-compatible webhook on both fire and clear, routed and scoped by tag
(ADR-0010). Webhook is the first `Notifier`; SSRF is mitigated by scheme validation
and optional egress allowlisting, and payloads carry no secrets.

Status: **accepted**

## 8. Next Steps

- [ ] Extend the liveness snapshot-apply path with threshold evaluation — hub
- [ ] Implement the `(rule, host)` lifecycle with for-/clear-duration hysteresis + tests — hub
- [ ] Implement the webhook `Notifier` (generic JSON, fire + clear, bounded retry, SSRF guard) — hub
- [ ] Add rule + target config (file/JSON/flags) with tag selectors (ADR-0010) — hub
- [ ] Surface fleet alert count + row highlight — dashboard
- [ ] Document SSRF posture, no-secrets-in-payload, and restart re-fire behavior — Architect
- [ ] Note: cross-platform privileged-metrics parity (Linux RAPL/hwmon, Windows) is an extension of ADR-0004/0005 and needs no separate ADR — tracked under E2
