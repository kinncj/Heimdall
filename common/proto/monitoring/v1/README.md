# monitoring.v1 — versioned gRPC contract

This directory holds the **single source of truth** for Heimdall's wire protocol.
All four binaries — `heimdall-daemon`, `heimdall-helper`, `heimdall-hub`,
`heimdall-dashboard` — import the generated types from `monitoring.proto`. No
binary defines a private copy of any message or service.

## Why the contract lives in `common/`

Per the BusinessRepo layout, domain-scoped shared code lives in `/common`. The
protocol is shared by every executable in the repo and by no other domain, so it
belongs here — not in a horizontal "shared-protos" repo (which the BusinessRepo
model rejects) and not duplicated under each `app/cmd/*` (which would let the
daemon and hub drift out of sync).

## Versioning policy

The package is `monitoring.v1` and the directory is `monitoring/v1/`. The package
version, directory, and `go_package` move together.

### Stable within v1 — additive only

Once `v1` ships, only **backward-compatible** changes are allowed in this package:

- Add a new `message`, `enum`, or `service`.
- Add a new field with a **new, never-before-used** field number.
- Add a new `enum` value (consumers must tolerate unknown values).
- Add a new `rpc` to an existing service.

These are safe because proto3 readers ignore unknown fields and a new field
defaults cleanly on old readers.

### Forbidden within v1 — breaking changes

The following are **prohibited** in `monitoring.v1`. They require a new package:

- Renumber, reuse, or change the type of an existing field.
- Rename or remove a field, message, enum, enum value, service, or rpc.
- Change a field's cardinality (e.g. singular ⇄ `repeated`, or in/out of a `oneof`).
- Change an rpc's streaming mode or request/response type.
- Tighten semantics in a way old peers would violate (e.g. making an optional
  field load-bearing).

When a field is retired, do **not** delete it — mark it `reserved` (number and
name) so the number is never reused.

### Breaking change ⇒ `v2`

A breaking change creates `common/proto/monitoring/v2/` with `package monitoring.v2`
and `go_package = ".../v2;monitoringv2"`. `v1` and `v2` coexist during migration.
The hub registers both service versions and bridges them until every daemon is
upgraded. A migration ADR documents the cutover and removal of `v1`.

## Enum zero-value guard

`MetricStatus` reserves `METRIC_STATUS_UNSPECIFIED = 0`. proto3 decodes an absent
or truncated enum as `0`; for a monitoring system, "missing" must never read as
"healthy". `OK` is `1`. Consumers treat `0` as do-not-trust.

## Compatibility checking

`make lint` runs a breaking-change detector (e.g. `buf breaking`) against the
previous committed descriptor. CI fails on any backward-incompatible edit inside
`v1`. The wire contract is also exercised by the contract test layer
(`make test-contract`) so daemon and hub are validated against the same schema.

## Regeneration

Generated Go code is produced by a Makefile target (it does **not** live in this
directory and is never hand-edited). This directory contains only the `.proto`
contract and this policy. There is intentionally **no `go.mod` and no `.go` file
here** — the contract is language-agnostic.

```
make proto       # generate Go stubs from monitoring.proto into app/internal/transport/genpb
make lint        # includes proto lint + breaking-change check
make test-contract  # daemon/hub validated against this schema
```
