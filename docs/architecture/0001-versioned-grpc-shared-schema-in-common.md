---
adr: "0001"
title: "Versioned gRPC shared schema in common/"
status: proposed
date: "2026-06-26"
supersedes: null
superseded_by: null
deciders:
  - "Architect"
  - "Platform Lead"
---

# 0001 — Versioned gRPC shared schema in common/

## 1. Context

Heimdall ships four binaries from one Go module — `heimdall-daemon`,
`heimdall-helper`, `heimdall-hub`, `heimdall-dashboard` — that must agree exactly
on the wire format for enrollment, metric streaming, federation, control, and
logs. If any binary carries its own copy of the message definitions, the daemon
and hub will drift and produce silent decode mismatches in the field. The
BusinessRepo model forbids a horizontal "shared-protos" repo, and duplicating the
schema under each `app/cmd/*` defeats single-source-of-truth.

A versioning policy is also required: daemons in the field upgrade at a different
cadence than the hub, so the schema must evolve without breaking deployed peers.

## 2. Goals / Non-Goals

**Goals:**
- One canonical, versioned protobuf contract imported by all four binaries.
- A clear compatibility policy that lets daemon and hub upgrade independently.
- Keep the contract language-agnostic and free of generated code in the contract
  directory.

**Non-Goals:**
- Choosing the transport semantics of each RPC (covered by the architecture doc).
- Generating or vendoring Go stubs in `common/` (stubs are a build artifact).

## 3. Proposal

Place the contract at `common/proto/monitoring/v1/monitoring.proto` with
`package monitoring.v1` and `option go_package = "heimdall/common/proto/monitoring/v1;monitoringv1"`.
All binaries import the generated types from this one file.

Compatibility policy:
- **Additive-only within `v1`**: new messages/enums/services, new fields with new
  numbers, new enum values, new rpcs.
- **Breaking changes are forbidden in `v1`** and require a new `monitoring/v2/`
  package; `v1` and `v2` coexist during migration, bridged by the hub.
- Retired fields become `reserved` (number + name); numbers are never reused.
- `MetricStatus` reserves `0 = UNSPECIFIED` so an absent/truncated enum never reads
  as `OK`.

Enforcement: `make lint` runs a proto breaking-change detector against the previous
committed descriptor; `make test-contract` validates daemon and hub against the
schema. The `.proto` is the only thing in the directory besides its README — **no
`go.mod`, no `.go` files**.

## 4. Alternatives Considered

| Option | Pros | Cons | Why Rejected |
|---|---|---|---|
| Schema duplicated per binary | No shared import | Guaranteed drift; silent decode bugs | Violates single source of truth |
| Separate `heimdall-protos` repo | Reusable | Horizontal shared repo; coordinated deploys | Violates BusinessRepo model |
| JSON / ad-hoc framing | Human-readable | No schema enforcement, larger frames, no codegen | Fails low-bandwidth + contract goals |
| Single unversioned proto | Simple now | No path to evolve without breaking field daemons | Fails independent-upgrade goal |

## 5. Trade-offs and Risks

- Codegen adds a build step (`make proto`); mitigated by a Makefile target and CI
  check.
- The additive-only discipline requires reviewer vigilance; mitigated by an
  automated breaking-change gate.
- A `v2` migration is real work; deferred until a breaking change is actually
  needed (no speculative versioning).

## 6. Impact

**FinOps:** Negligible. Compact protobuf framing reduces egress vs JSON on metered
links — a small ongoing saving.

**SRE:** Removes a class of field incidents (schema drift). Breaking-change gate
prevents an incompatible daemon/hub deploy. Runbook: how to roll a `v1→v2`
migration with both registered on the hub.

**Security:** A single audited schema narrows the parsing attack surface; max
message size and field validation live in one place.

**Team:** Engineers learn one proto workflow and the additive-only rule. Low
onboarding cost; protobuf is widely known.

## 7. Decision

Adopt a single versioned protobuf contract at `common/proto/monitoring/v1/`,
imported by all binaries, governed by an additive-only-within-`v1` policy with a
`v2` package for breaking changes. This guarantees daemon/hub agreement and
independent upgrades while honouring the BusinessRepo layout.

Status: **proposed**

## 8. Next Steps

- [ ] Wire `make proto` (protoc + Go plugins) and `make lint` breaking-change check — INFRA
- [ ] Add `make test-contract` validating daemon + hub against the schema — QA
- [ ] Document the `v1→v2` bridge procedure in a runbook — Architect
