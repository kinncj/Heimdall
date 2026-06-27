---
adr: "0013"
title: "Yggdrasil: interactive dashboard grouping + enriched snapshots"
status: accepted
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0013 — Yggdrasil: interactive dashboard grouping + enriched snapshots

> **Yggdrasil** /ˈɪɡ.drə.sɪl/ (*IG-druh-sil*) — the world-tree binding the realms,
> and grouping the fleet along it. **Bifröst** is the federation it grows on. See the
> [glossary](../glossary.md).

## 1. Context

ADR-0010 added the data layer for fleet grouping: host/hub tags (Realms),
hub-to-host tag inheritance, the origin-hub `hub` label, and the `Dimension`
(`HostView -> string`) abstraction. v1.2.0 shipped that layer. The grouping UI was
never built, and a defect blocks it.

The dashboard subscribes to a hub through `FederationService.Subscribe`. The two
update paths disagree:

- **Initial state (on connect):** the hub sends registry-derived snapshots. These
  carry **merged effective labels** — origin hub as `hub`, plus inherited hub tags.
- **Live updates:** the hub forwards the **raw daemon snapshot** via `h.publish(snap)`.
  The raw snapshot carries only the daemon's own tags. The `hub` label and inherited
  hub tags are **missing**.

So a host groups correctly on connect, then loses its group key on the next live
tick. Any grouping built on the current live path is broken by construction. The
enrichment the registry already computes must reach live updates too.

## 2. Goals / Non-Goals

**Goals:**
- **Consistent enrichment.** Live updates carry the same merged effective labels and
  alert state as the initial-state snapshots — one host has one set of labels
  regardless of which path delivered it.
- **Interactive grouping in the TUI.** Group the grid by a `Dimension`
  (`origin-hub`, `os`, `tag:<key>`), toggled with a key; section headers per group.
- **Filter / search.** A `/` filter over host name and tag values.
- **SOLID.** Grouping and filtering depend on the `Dimension` abstraction (ADR-0010),
  not per-axis special-casing.
- **Demo parity.** `--demo` exercises grouping with tagged, multi-hub fake hosts.

**Non-Goals:**
- New tag semantics or a tag query language (ADR-0010 owns the merge rule; this is UI).
- Persisting grouping/filter state across restarts.
- Server-side grouping — grouping is a dashboard concern over the subscribed stream.
- Changing the federation relay envelope or transport topology.

## 3. Proposal

**Part 1 — Publish enriched snapshots to dashboards.** Move the enrichment that
the registry already performs for initial state onto the live path. Instead of
`h.publish(snap)` forwarding the raw daemon snapshot, the hub publishes the
**registry-derived snapshot** for that host: raw daemon metrics + **merged effective
labels** (origin hub as `hub`, inherited hub tags, host tags with host-over-hub
precedence per ADR-0010) + **alert state** (ADR-0014). The registry is the single
source of truth for what a host *is*; both delivery paths read from it. The raw
daemon snapshot stays the ingestion input; it stops being the dashboard output.

Concretely: on snapshot apply, the hub updates the registry entry, then derives the
enriched snapshot from that entry and publishes *that*. Initial-state and live paths
converge on one builder. This removes the divergence rather than patching the live
path to re-merge labels independently (which would re-create two sources of truth).

**Part 2 — Interactive grouping/filter/search (TUI).** Build on the ADR-0010
`Dimension` abstraction:

- **Group toggle.** A key cycles the active `Dimension`: `none` → `origin-hub` →
  `os` → `tag:<key>` (over the keys present in the fleet). The grid renders one
  **section header** per distinct dimension value, hosts bucketed under it; missing
  values bucket as `untagged` / `unknown`.
- **Filter/search.** `/` opens an input; the predicate matches against host name and
  effective tag **values** (substring, case-insensitive). Filtering composes with
  grouping — filter first, group the survivors.
- **SOLID.** The grid view depends on `Dimension` (a function) and a `Filter`
  predicate, not concrete fields (DIP). Adding an axis registers a `Dimension`, not a
  `switch` edit (OCP). Group-key extraction, filtering, and rendering are separate
  (SRP). No fat view interface — the grid takes only the dimension and predicate it
  uses (ISP).

**Part 3 — Demo mode.** The fake fleet (`app/internal/fake`, backing `--demo`)
emits hosts across multiple origin hubs with varied OS and tags so every dimension
and the filter are exercisable without a real federation. Demo enrichment flows
through the same registry-derived builder as production, so the demo can't pass while
the real path is broken.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Re-merge labels on the live path independently | Localized fix | Two label-merge implementations → drift; same class of bug returns | Re-creates the divergence it fixes |
| Enrich at the dashboard from raw snapshot + a separate label feed | Hub stays simple | Dashboard needs hub registry state it doesn't have; extra transport | Pushes hub knowledge to the client |
| Publish registry-derived enriched snapshot on both paths (chosen) | One source of truth; live == initial | Hub does a small derive per publish | Accepted |
| Hardcode origin-hub + OS grouping in the grid | Quick | Tags never fit; every axis edits the grid | Violates SOLID; ignores ADR-0010 |
| Server-side grouping | Thin client | Grouping is a view concern; couples hub to UI layout | Wrong layer |

## 5. Trade-offs and Risks

- **Per-publish derive cost.** The hub now builds an enriched snapshot on each live
  tick instead of forwarding bytes it already has. Cost is a map merge over a small
  label set; bounded and in-memory. Acceptable for the consistency win.
- **Enrichment is now load-bearing for live correctness.** A bug in the builder
  breaks every live update, not just initial state. Mitigated by both paths sharing
  one tested builder and by demo exercising it.
- **Tag cardinality** can fragment the grouped view (ADR-0010). Mitigated by
  `untagged` bucketing and grouping being a toggle, not always-on.
- **Filter/group state is ephemeral.** Lost on restart by design; cheap to re-apply.
- **Relay envelope unchanged.** Enrichment happens at the dashboard publish boundary,
  not in the federation relay, so parent-hub forwarding is untouched.

## 6. Impact

**FinOps:** Negligible. One in-memory snapshot derive per host per live tick; no new
storage or service. Net operator-time saving from a fleet that groups correctly under
load.

**SRE:** Removes a correctness defect (live updates losing group keys), so grouping by
origin-hub/tag is trustworthy during a partial federation outage. New failure mode:
a faulty enriched-snapshot builder degrades *all* live updates — covered by tests and
by demo running the same path. Observability: dimension/group counts derive from the
enriched `HostView`s already on screen.

**Security:** No new surface. Labels are organizational, not authz (ADR-0010); they
carry no secrets. No transport or trust-boundary change — the enrichment runs hub-side
on data the hub already holds.

**Team:** Two concepts: the enriched-snapshot builder as the single output path, and
the grouping/filter interaction over the existing `Dimension`. Both small; the proto
is unchanged in this ADR.

## 7. Decision

Make the hub publish the **registry-derived, enriched** snapshot (merged effective
labels + alert state) to dashboards on the live path, the same data the initial-state
snapshots already carry — so a host's `hub` label and inherited hub tags are present
on every update, not just on connect. Build interactive grouping (by `Dimension`),
`/` filter/search over host name and tag values, and per-group section headers in the
TUI on the ADR-0010 abstraction with no per-axis special-casing. `--demo` exercises
all of it via tagged, multi-hub fake hosts running the same enrichment path.

Status: **accepted**

## 8. Next Steps

- [ ] Replace `h.publish(snap)` with a registry-derived enriched-snapshot publish; share one builder with the initial-state path — hub
- [ ] Add the group toggle key + per-group section headers driven by `Dimension` — dashboard
- [ ] Add `/` filter/search over host name + effective tag values, composed with grouping — dashboard
- [ ] Extend the fake fleet with tagged, multi-hub hosts and route demo through the enriched builder — `app/internal/fake`
- [ ] Tests: live == initial enrichment parity; group-key extraction; filter+group composition — QA
- [ ] Implements story 0018 (Yggdrasil grouping); coordinates with ADR-0014 (alert state on the enriched snapshot) — Architect
