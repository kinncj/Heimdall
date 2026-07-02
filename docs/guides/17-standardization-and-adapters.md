# Standardization & Adapters — how to add a cross-platform metric

Heimdall watches a mixed fleet (Apple/Intel/AMD/ARM, macOS/Linux/Windows). The
whole point is that a metric means the **same thing everywhere**, no matter what
the platform calls it. This guide is the rule for keeping it that way. If you are
adding or changing a metric, read this first.

See [ADR-0022](../architecture/0022-metric-standardization-boundary.md) for the
decision and the failure that motivated it, and
[ADR-0021](../architecture/0021-power-metric-standardization-and-source-layering.md)
for the power example.

## The one rule

> **Adapters translate platform specifics into neutral metrics. The `domain`
> core turns neutral metrics into display-ready structures. Consumers (TUI, CLI,
> dashboard) render those structures and nothing more.**

Knowledge flows one way: platform → adapter → neutral metric → domain → consumer.
No platform names, no vendor branching, and no classification logic ever reach a
renderer.

## The three layers

| Layer | Package | Owns | Must NOT contain |
|---|---|---|---|
| **Adapters** | `app/internal/adapters`, `app/internal/helper` | reading each platform's own scheme and emitting a **neutral** metric | cross-host assumptions; UI concerns |
| **Domain** | `app/internal/domain` | the neutral taxonomy + functions that turn neutral metrics into display-ready structures | any framework / transport / UI import |
| **Consumers** | `app/internal/tui/*`, `app/internal/cli`, dashboard | rendering (colours, bars, boxes, JSON) | bucketing, classification, platform branching |

## Worked example — core types (P/E)

Each platform names core types differently:

- macOS: `hw.perflevel0/1` sysctl counts.
- Intel Linux: `/sys/devices/cpu_core/cpus` and `cpu_atom/cpus` index ranges.
- Windows: `GetLogicalProcessorInformationEx` `EfficiencyClass` values.

**Adapter** (`app/internal/adapters/cputopo_*.go`) translates each into the neutral
taxonomy `domain.CorePerf` / `CoreEff` / `CoreLP` and emits one metric,
`cpu.topology`, whose `PerCore[i]` is the type id of logical core `i`. Platform
code stops here.

**Domain** (`app/internal/domain/coretype.go`) turns that neutral layout into
display-ready groups:

```go
groups, ok := domain.CoreGroups(coreUtil, coreType) // []CoreGroup{Label, Indices, Util}
summary := domain.CoreTypeSummary(coreType)          // "12P + 4E"
```

**Consumers** call those and render:

- TUI top view: paints each group under a `P-cores (N):` / `E-cores (N):` header.
- CLI `host`: serialises them as `core_groups` (type, cores, util_pct).

Neither consumer buckets cores or knows what a "perflevel" or "EfficiencyClass"
is. Add a fourth platform and only its adapter changes.

## Wire discipline (the trap that bit us)

A per-core metric is transmitted as its `PerCore` slice via the protobuf
`per_core` **oneof** — so its scalar `Gauge` is **dropped on the wire** and
arrives as `0` at the dashboard/CLI.

- **Never** key logic off a per-core metric's `Gauge`. Standardization functions
  take `PerCore`.
- Carry any human summary in `Detail` (which does cross the wire), as
  `cpu.topology` does.
- This is exactly why the P/E grid silently showed the plain grid on every host
  in an early cut: the grouping read `Gauge`, which the wire had zeroed. It
  passed locally because `Gauge` is set in-process. Add a transport round-trip
  test for any new per-core metric.

## Checklist for a new cross-platform metric

1. **Adapter per platform** (build-tagged): read the native source, emit a neutral
   metric. No platform strings in the value.
2. **Pure parser**, build-tag-free, unit-tested on any OS (split the shell-out /
   syscall from the parsing).
3. If consumers need a computed *view* (grouping, classification, a summary), put
   it in **`domain`** as a pure function — never in a renderer.
4. Consumers call the domain function and render. If you find yourself writing a
   `switch` on platform or bucketing values in a TUI/CLI file, stop — it belongs
   in an adapter or `domain`.
5. Per-core metric? Data goes in `PerCore`, summary in `Detail`, and add a
   transport round-trip test. Never rely on `Gauge`.
6. Explain gaps: a platform that can't supply the metric returns
   `Unavailable`-with-reason (ADR-0003), never a fabricated value.
