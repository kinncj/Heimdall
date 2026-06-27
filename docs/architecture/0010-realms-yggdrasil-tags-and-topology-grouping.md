---
adr: "0010"
title: "Realms & Yggdrasil: host/hub tags and topology-aware grouping"
status: accepted
date: "2026-06-27"
supersedes: null
superseded_by: null
deciders:
  - "Kinn Coelho Juliao"
---

# 0010 — Realms & Yggdrasil: host/hub tags and topology-aware grouping

> **Realms** /rɛlmz/ (*relmz*) · **Yggdrasil** /ˈɪɡ.drə.sɪl/ (*IG-druh-sil*) — the
> watched worlds and the world-tree that binds them. See the [glossary](../glossary.md).

## 1. Context

A flat host list does not scale operationally. Operators need to slice the fleet by
**where** a host sits in the federation (which Bifröst hub ingested it) and by
**what** a host is (env, role, region). The federation relay envelope already
carries `OriginHubId` ("first hub that ingested the snapshot") and `Path` (ordered
hub ids traversed); the hub records per-host origin via `recordOrigin` (ADR-0006).
But `HostView`/`Host` do **not** yet carry origin, so the dashboard cannot group on
data that is already on the wire. Tags do not exist at all.

We want both grouping axes without special-casing each one in the dashboard, and
without breaking the wire contract (ADR-0001: add fields only, never renumber).

## 2. Goals / Non-Goals

**Goals:**
- **Topology axis (system-derived):** group hosts by `OriginHubId`, or
  hierarchically by `Path` prefix — by surfacing origin onto `HostView`.
- **Tag axis (user-defined):** hosts carry tags; hubs carry tags that **inherit**
  to their hosts via the relay; define a deterministic merge rule on conflict.
- A single dashboard grouping abstraction: a **dimension** = `HostView -> string`,
  where `origin-hub`, `os`, and `tag:<key>` are all dimensions — no special-casing.
- Backward-compatible proto evolution (add fields only).

**Non-Goals:**
- Tag-based access control or RBAC (tags are organizational, not authz).
- A query language over tags beyond group/filter by dimension.
- Persisting tags server-side beyond what rides the existing snapshot/enroll flow.

## 3. Proposal

**Surface topology onto `HostView`.** Add `origin_hub_id` and `path[]` to the
`Host`/`HostView` domain types, populated from the data the hub already records
(`recordOrigin`, relay envelope). No new transport — the values already traverse
federation; we stop dropping them at the domain boundary.

**Host tags.** A daemon carries tags via `--tags env=prod,role=db` (and the
matching JSON key / env var, per the layered config system). Tags ride
`Host`/`Enroll` so the hub knows them at enrollment and on each snapshot.

**Hub tags with inheritance.** A hub carries its own tags (e.g. `region=apac`). Hub
tags **inherit to every host that hub ingests**, propagated through the relay
envelope so a parent hub sees them too. Tag a hub `region=apac` once and all its
hosts become filterable by it — no per-host fan-out.

**Merge rule (deterministic).** On key conflict, **the host's own tag overrides the
inherited hub tag.** Inheritance flows down the `Path`; a host's explicit tag is
the most specific signal, so it wins. Effective tags =
`merge(inherited_hub_tags, host_tags)` with host keys taking precedence. This rule
is documented and tested.

**Grouping dimension abstraction (SOLID).** Generalize the dashboard to group by a
`Dimension`, a function `HostView -> string`:

- `origin-hub` → `view.OriginHubId`
- `os` → `view.OS`
- `tag:<key>` → effective tag value for `<key>` (or "untagged")
- `path-prefix:<n>` → first `n` hub ids of `Path` (hierarchical topology grouping)

OCP: adding a grouping axis is registering a new `Dimension`, not editing a switch
statement. SRP: each dimension extracts exactly one grouping key. The dashboard
group/filter logic depends on the `Dimension` abstraction, not on concrete fields
(DIP).

**Wire impact (backward-compatible).** Add `origin_hub_id` and `path[]` to
`HostView`; add `tags` (map<string,string>) to `Host`/`Enroll` and to the relay
envelope. All are new field numbers — never renumber existing fields (ADR-0001).
Old peers ignore unknown fields; new peers default-empty.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Hardcode origin-hub + OS grouping | Simple now | Every new axis edits the dashboard; tags never fit | Not extensible; fails SOLID |
| Host tags only (no hub inheritance) | Less wire change | Must tag every host for region/site facts | Operationally heavy; misses the hub-level fact |
| Hub tags only (no host override) | One place to tag | Can't express per-host exceptions | Too coarse |
| Separate code paths per axis | Direct | Combinatorial special-casing | Unmaintainable |
| `Dimension = HostView->string` + hub→host tag inheritance with host override (chosen) | Uniform, extensible, backward-compatible | Slightly more abstract dashboard code | Accepted |

## 5. Trade-offs and Risks

- **Tag cardinality** can fragment the grouped view (one group per unique value).
  Mitigated by grouping being a UI concern and `untagged` bucketing.
- **Inheritance surprise.** Operators may not expect hub tags on hosts. Mitigated by
  showing tag provenance (inherited vs own) and the documented override rule.
- **Wire growth.** Tags add bytes per snapshot/enroll. Bounded by keeping tags
  small key/value maps; delta/keyframe shaping (ADR-0006) still applies.
- **Path-prefix grouping** assumes stable hub ids along `Path`; a re-IDed hub
  re-buckets its hosts. Accepted; hub ids are stable config.

## 6. Impact

**FinOps:** Negligible. A small tags map per host on the wire and in memory; no new
storage tier. Reduces operator time spent eyeballing a flat list.

**SRE:** Grouping by origin-hub/path makes blast radius legible during a partial
federation outage ("all `region=apac` hosts went stale"). No new failure mode beyond
slightly larger payloads. Observability: tag/dimension counts are derivable from
existing `HostView`s.

**Security:** Tags are organizational only — **not** an authz mechanism; documented
explicitly so no one builds access control on them. Tags carry no secrets;
treat values as low-sensitivity labels. No new trust boundary.

**Team:** One new concept (`Dimension`) and one rule to learn (host overrides
inherited hub tag). The proto change is additive and routine.

## 7. Decision

Add two grouping axes behind one abstraction. Surface `OriginHubId`/`Path` (already
on the wire) onto `HostView` for system-derived topology grouping, and add
user-defined tags on both hosts and hubs, where hub tags inherit to their hosts via
the relay and a host's own tag overrides the inherited value on conflict. The
dashboard groups by a `Dimension` (`HostView -> string`) so `origin-hub`, `os`, and
`tag:<key>` are all uniform dimensions with no special-casing. All proto changes are
additive per ADR-0001.

Status: **accepted**

## 8. Next Steps

- [ ] Add `origin_hub_id`/`path[]` to `HostView` and `tags` to `Host`/`Enroll`/relay envelope (additive) — hub
- [ ] Implement the hub→host tag inheritance + host-override merge, with tests — hub
- [ ] Implement the `Dimension` abstraction and register `origin-hub`/`os`/`tag:<key>`/`path-prefix` — dashboard
- [ ] Wire `--tags` flag + JSON key + wizard prompt per the config system — daemon
- [ ] Document the merge rule and that tags are not authz — Architect
- [ ] Note: cross-platform privileged-metrics parity (Linux RAPL/hwmon, Windows) is an extension of ADR-0004/0005 and needs no separate ADR — tracked under E2
