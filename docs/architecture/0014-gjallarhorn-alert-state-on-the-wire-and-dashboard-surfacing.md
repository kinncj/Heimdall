---
adr: "0014"
title: "Gjallarhorn: alert state on the wire + dashboard surfacing"
status: accepted
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0014 — Gjallarhorn: alert state on the wire + dashboard surfacing

> **Gjallarhorn** /ˈɡjɑː.lar.hɔːrn/ (*GYAH-lar-horn*) — the horn Heimdall sounds to
> warn of Ragnarök. See the [glossary](../glossary.md).

## 1. Context

ADR-0012 added Gjallarhorn alerting. v1.2.0 shipped the **hub side**: the rules
engine evaluating thresholds alongside the liveness state machine, the
firing→resolved lifecycle with hysteresis, and webhook + log notification. What it
did **not** ship is dashboard surfacing — ADR-0012 listed "fleet alert count + row
highlight" as a goal, but the firing state never leaves the hub.

The hub alert engine already knows, each evaluation tick, which `(rule, host)` pairs
are firing. That knowledge stays internal: it drives webhooks and logs, but the
dashboard's `HostView` has no notion of an alert. An operator watching the TUI cannot
see what the hub already knows.

ADR-0013 fixes the live-update path to publish **registry-derived enriched
snapshots** to dashboards. That mechanism is the natural carrier for alert state:
if the registry holds per-host firing alerts, the enriched snapshot already on its
way to the dashboard can carry them too.

## 2. Goals / Non-Goals

**Goals:**
- **Alert state in the registry.** Each evaluation tick writes the set of firing
  rule names per host into the hub registry.
- **Alert state on the wire.** An additive `repeated string alerts` field on
  `Snapshot`, carried through transport onto `HostView`.
- **Dashboard surfacing.** A **fleet alert count** in the header and a per-row
  **badge/highlight** for hosts with one or more firing alerts.
- **Live delivery via ADR-0013.** Alerts ride the enriched snapshot, so they reach
  the dashboard on the same path as labels — no separate alert transport.
- **Demo parity.** `--demo` simulates firing alerts so the badge is visible without a
  hub.

**Non-Goals:**
- Changing the rules engine, lifecycle, or hysteresis (ADR-0012 owns those).
- Alert history or durable alert storage (in-memory, per ADR-0008/0011/0012).
- Changing the webhook payload or adding transports (ADR-0012 owns notification).
- Interactive alert acknowledgement/silencing from the dashboard.

## 3. Proposal

**Write firing state into the registry.** The alert engine already computes the
firing `(rule, host)` set each tick. Extend that tick to write, per host, the list of
**firing rule names** into the registry entry. The registry becomes the single source
of truth for "what is firing on this host," consistent with ADR-0013 making the
registry the source for enriched snapshots.

**Additive proto field.** Add `repeated string alerts = N;` to `Snapshot` in
`common/proto/monitoring/v1`, using a **new field number** after the existing
`host_id..labels` (field 7) — never renumber (ADR-0001). Old peers ignore it; new
peers default to empty. The field carries firing rule names, not rule bodies or
thresholds.

**Carry through transport onto `HostView`.** The enriched-snapshot builder from
ADR-0013 reads the registry's firing rule names and populates `alerts` on the
published snapshot. `HostView` gains an `Alerts []string` field. Because alerts ride
the **same** enriched snapshot as labels, they arrive live and on initial state with
identical guarantees — no second subscription, no separate alert feed.

**Dashboard surfacing.**
- **Header:** a fleet-level alert count = number of hosts with a non-empty `Alerts`
  (or total firing rules; count semantics documented). Zero alerts → quiet header.
- **Row:** hosts with firing alerts get a badge/highlight; the firing rule names are
  available for display (e.g. on a detail/expanded view). Rendering is a read off
  `HostView.Alerts` already on screen — no new transport for the local view, matching
  the ADR-0012 dashboard intent.

**Demo mode.** The fake fleet (`app/internal/fake`, `--demo`) marks a subset of fake
hosts as firing (populating `Alerts` through the same enriched builder) so the count
and badge render without a live hub or real rules.

**SOLID.** SRP: registry write, proto carriage, and rendering are separate. OCP: a
new alert presentation (e.g. severity coloring) reads the same `Alerts` field, no
engine change. DIP: the dashboard depends on `HostView.Alerts`, not on the engine.
ISP: `alerts` is a flat list of names — the minimum the view needs, nothing more.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Separate alert-state subscription/stream | Decoupled from snapshots | Second transport + sync between two streams; alerts can lag labels | Re-introduces multi-source-of-truth ADR-0013 removed |
| Embed full rule bodies/thresholds in the field | Rich client display | Larger payload; duplicates config the hub owns | Names suffice for badge/count; keep it light |
| Dashboard polls a hub alerts endpoint | No proto change | New RPC + polling cadence; not live | Enriched snapshot already flows live |
| Reuse `labels` to encode alerts (e.g. `alert:foo`) | No new field | Conflates organizational labels with transient state; pollutes grouping | Wrong semantics; would leak into ADR-0013 dimensions |
| Additive `repeated string alerts` on the enriched snapshot (chosen) | Live, one source, additive, light | One new proto field to maintain | Accepted |

## 5. Trade-offs and Risks

- **Names only, not bodies.** The wire carries firing rule *names*. A client wanting
  the threshold/metric must look at config. Accepted: the badge needs identity, not
  the rule body.
- **Transient state on a snapshot field.** `alerts` is volatile (changes each tick)
  unlike organizational labels. Documented so no one groups on it as if it were a tag.
- **Restart re-fire window.** Alert state is in-memory (ADR-0012); after a hub restart
  the registry's firing set re-derives from incoming snapshots, so the badge may
  briefly clear then re-light. Documented, consistent with ADR-0008/0011/0012.
- **Count semantics.** "Hosts firing" vs "total firing rules" must be one defined
  number to avoid header/row mismatch. Chosen: header counts hosts with ≥1 firing
  alert; documented.
- **Coupling to ADR-0013.** Surfacing depends on the enriched-snapshot path. If
  ADR-0013 enrichment is broken, alerts also don't reach live — same shared builder,
  same tests guard both.

## 6. Impact

**FinOps:** Negligible. A small `[]string` per host in the registry, on the wire, and
in memory; no new storage tier or service. Reuses ADR-0012 evaluation and the
ADR-0013 publish path.

**SRE:** Closes the loop from "hub knows" to "operator sees" — the dashboard now shows
what the alert engine already decided. Failure modes: stale badge if the enriched
publish stalls (shared with ADR-0013), brief re-fire after restart (ADR-0012).
Observability: existing `self.alert.{active,fired,resolved}` plus a derivable
on-screen count. No new alert-evaluation failure mode (engine unchanged).

**Security:** **Alert names are not secrets.** The field carries low-sensitivity rule
identifiers, consistent with the ADR-0012 "no secrets in payload" posture. **No change
to webhook payloads** — this ADR adds an internal/dashboard surface only, not a new
outbound surface. No new trust boundary; alerts ride the existing authenticated
subscription.

**Team:** One additive proto field and one `HostView` field to learn, plus header/row
rendering. The engine and notification paths are untouched.

## 7. Decision

Surface Gjallarhorn alert state in the dashboard by writing per-host firing rule
names into the hub registry each evaluation tick, adding an additive
`repeated string alerts` field to `Snapshot` (new field number, never renumber per
ADR-0001), and carrying it through transport onto `HostView`. Alerts ride the
**registry-derived enriched snapshot** from ADR-0013, so they reach the dashboard live
on the same path as labels — no separate alert transport. The dashboard shows a fleet
alert count in the header and a per-row badge/highlight. `--demo` simulates firing
alerts so the badge renders without a hub. Alert names are not secrets and webhook
payloads are unchanged.

Status: **accepted**

## 8. Next Steps

- [ ] Write firing rule names per host into the registry each evaluation tick — hub
- [ ] Add `repeated string alerts` to `Snapshot` (new field number, additive) — `common/proto/monitoring/v1`
- [ ] Populate `alerts` in the ADR-0013 enriched-snapshot builder; add `Alerts` to `HostView` — hub
- [ ] Render fleet alert count (header) + per-row badge/highlight — dashboard
- [ ] Simulate firing alerts in the fake fleet for `--demo` — `app/internal/fake`
- [ ] Document: alert names not secrets, no webhook payload change, count semantics, restart re-fire — Architect
- [ ] Implements story 0019 (Gjallarhorn badge); depends on ADR-0013 and ADR-0012 — Architect
