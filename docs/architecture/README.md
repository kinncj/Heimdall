# Architecture Decision Records

ADRs document significant architectural decisions made during this project's lifetime. Each decision is captured as a numbered Markdown file in this directory.

## When to Write an ADR

Write an ADR whenever a decision:
- Introduces a new external dependency or service
- Changes the data model or storage layer
- Crosses Clean Architecture boundaries
- Requires a coordinated deploy or migration
- Has a non-obvious trade-off that future engineers need to understand

The orchestrator and architect agents will flag `adr_required: true` on stories that trigger ADRs.

## File Naming

```
docs/architecture/NNNN-short-decision-title.md
```

- `NNNN` is a zero-padded sequential number starting at `0001`
- Use the template at `docs/architecture/adr-template.md`

## Status Values

| Status | Meaning |
|---|---|
| `proposed` | Under discussion — not yet decided |
| `accepted` | Decision made and in effect |
| `deprecated` | Was accepted, no longer applies |
| `superseded` | Replaced by a later ADR (link to it) |

## Index

| # | Title | Status | Date |
|---|---|---|---|
| [0001](0001-versioned-grpc-shared-schema-in-common.md) | Versioned gRPC shared schema in common/ | proposed | 2026-06-26 |
| [0002](0002-daemon-enrollment-identity-and-mtls.md) | Daemon enrollment identity and mTLS | proposed | 2026-06-26 |
| [0003](0003-metric-adapter-contract-and-failure-isolation.md) | Metric adapter contract and failure isolation | proposed | 2026-06-26 |
| [0004](0004-optional-privileged-helper-and-privilege-tiers.md) | Optional privileged helper and privilege tiers | proposed | 2026-06-26 |
| [0005](0005-gpu-and-power-vendor-adapters-and-external-deps.md) | GPU and power vendor adapters and external dependencies | proposed | 2026-06-26 |
| [0006](0006-dashboard-federation-via-pubsub-relay.md) | Dashboard federation via pub/sub relay | proposed | 2026-06-26 |
| [0007](0007-unprivileged-remote-control-plane.md) | Unprivileged remote control plane | proposed | 2026-06-26 |
| [0008](0008-in-memory-ring-buffers-vs-tsdb.md) | In-memory ring buffers vs TSDB | proposed | 2026-06-26 |
| [0009](0009-ratatoskr-zeroconf-service-discovery.md) | Ratatoskr: zeroconf service discovery | accepted | 2026-06-27 |
| [0010](0010-realms-yggdrasil-tags-and-topology-grouping.md) | Realms & Yggdrasil: host/hub tags and topology-aware grouping | accepted | 2026-06-27 |
| [0011](0011-mimir-metrics-history-and-openmetrics-export.md) | Mímir: metrics history and OpenMetrics export | accepted | 2026-06-27 |
| [0012](0012-gjallarhorn-alerting-and-notifications.md) | Gjallarhorn: alerting and notifications | accepted | 2026-06-27 |
| [0013](0013-yggdrasil-interactive-grouping-and-enriched-snapshots.md) | Yggdrasil: interactive dashboard grouping + enriched snapshots | accepted | 2026-06-27 |
| [0014](0014-gjallarhorn-alert-state-on-the-wire-and-dashboard-surfacing.md) | Gjallarhorn: alert state on the wire + dashboard surfacing | accepted | 2026-06-27 |
| [0015](0015-windows-privileged-metrics.md) | Windows privileged metrics | accepted | 2026-06-27 |
| [0016](0016-mimir-durable-sink-and-hub-restore.md) | Mímir: pluggable durable sink and hub state restore | accepted | 2026-06-27 |
| [0017](0017-heimdallr-sight-in-dashboard-observability.md) | Heimdallr's sight: in-dashboard host observability via advertised control plane | accepted | 2026-06-29 |
| [0018](0018-v2-persistent-socket-mediation.md) | v2: persistent socket mediation (daemon ⇄ hub ⇄ dashboard) | accepted | 2026-06-29 |
| [0019](0019-v2-log-search-and-top-sorting.md) | v2: log search and top sorting in the detail modals | accepted | 2026-06-29 |

<!-- Add a row for each ADR as you create them -->
