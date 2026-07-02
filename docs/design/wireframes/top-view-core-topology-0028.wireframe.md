---
id: top-view-core-topology-0028
story: docs/stories/top-view-core-topology-20260702184928-0028/story.md
target: tui
status: approved
created_at: 2026-07-02
---

# Wireframe — top-view-core-topology-0028 (Hliðskjálf delta)

Target: **tui**. Delta wireframe: this story changes only the **CPU** and **POWER**
panels of the existing top view (`fullscreen-top-view-0025.wireframe.md`); chrome,
tiers, scroll, and every other panel are unchanged. Colour and lipgloss roles are
deferred to the mockup.

Two additions:

1. **CPU panel** — the per-core bars are grouped by core type when `cpu.topology`
   is OK: a `P` block and an `E` block, each with its own label and core count.
   Core indices keep their real logical IDs (c0…cN), they are just grouped.
2. **POWER panel** — when `power.cpu.pcluster` / `power.cpu.ecluster` exist, the
   CPU row gains a cluster split rendered as sub-values of `cpu`.

## Fallback behaviour (BC — mixed-version fleet)

- `cpu.topology` missing or non-OK → CPU panel renders **exactly today's**
  unlabelled `per-core (N):` grid. No layout shift, no placeholder.
- Cluster power metrics absent → POWER panel renders exactly today's row
  (`total / cpu / gpu / npu`). The cluster line only appears when at least one
  cluster metric is OK.
- Uniform topology (single type) → no P/E labels; render today's grid (a
  "16 cores (uniform)" label adds nothing over `per-core (16):`).

## State 1 — WIDE (≥100): CPU panel with P/E grouping (hybrid host, 12P+4E)

```text
┌────────────────────────────────────────────┐
│ CPU                                        │
├────────────────────────────────────────────┤
│ util 72%    freq 3.20 GHz    load 2.41     │
│ util  ⣀⣤⣶⣷⣿⣿⣿⣷⣶⣤⣀⣤⣶⣷⣿⣷⣶  72%               │
│ P-cores (12):                              │
│ c4 ███▌71  c5 ██▌ 52  c6 ████ 80           │
│ c7 ██  41  c8 ███ 63  c9 █▌  33            │
│ c10███▌70  c11██▌ 55  c12████ 88           │
│ c13██  44  c14███ 61  c15██▌ 58            │
│ E-cores (4):                               │
│ c0 █▌  12  c1 █   09  c2 ██  21  c3 █  07  │
└────────────────────────────────────────────┘
```

- Group header is `<type>-cores (<count>):` — text label, never colour-only.
- Which block renders first follows **logical core order** (whichever type owns
  index 0 comes first), so bars always match the host's real core IDs.
- Within a block, wrapping/columns reuse the existing per-core matrix logic —
  only the grouping and headers are new.

## State 2 — WIDE: POWER panel with cluster split (Apple Pro/Max)

```text
┌────────────────────────────────────────────┐
│ POWER                                      │
├────────────────────────────────────────────┤
│ total 22.4 W                               │
│ cpu 18.1 W   gpu 3.9 W   npu  —            │
│ cpu clusters: P 15.0 W · E 3.1 W           │
│ pwr  ⣀⣀⣤⣶⣷⣿⣷⣶⣤⣀⣤⣶  22 W                    │
└────────────────────────────────────────────┘
```

- The cluster line sits directly under the `cpu` row it decomposes.
- If only one cluster metric is OK, show only that one (`cpu clusters: P 15.0 W`).
- Non-Apple hosts: line absent entirely (see Fallback).

## State 3 — MEDIUM (60–99): same grouping, existing wrap

```text
┌──────────────────────────────────────────────────────────┐
│ CPU                                                      │
├──────────────────────────────────────────────────────────┤
│ util 72%   freq 3.20 GHz   load 2.41                     │
│ util ⣀⣤⣶⣷⣿⣿⣿⣷⣶⣤⣀⣤⣶⣷⣿⣷⣶  72%                              │
│ P-cores (12):                                            │
│ c4 ███▌71   c5 ██▌ 52   c6 ████ 80   c7 ██  41           │
│ c8 ███ 63   c9 █▌  33   c10███▌70   c11██▌ 55            │
│ c12████ 88  c13██  44   c14███ 61   c15██▌ 58            │
│ E-cores (4):                                             │
│ c0 █▌  12   c1 █   09   c2 ██  21   c3 █   07            │
└──────────────────────────────────────────────────────────┘
```

## State 4 — NARROW (40–59): aggregate per type replaces the matrix

Today NARROW collapses per-core to `aggregate bar + N`. With topology it becomes
one aggregate per type; without it, unchanged:

```text
│ P (12) ███▌  64%   E (4) █▌  12%   │
```

## State 5 — TINY (<40): unchanged

TINY drops per-core entirely today; topology adds nothing. The POWER key-number
line stays `pwr 22.4W` — cluster split is not shown at TINY.

## Legend

```
P-cores (12):  group header — Performance cores, count in parens
E-cores (4):   group header — Efficiency cores
c<N>           real logical core id (grouping never renumbers)
cpu clusters:  P <w> W · E <w> W   — SMC cluster rails (Apple Pro/Max only)
—              unavailable, as everywhere else in the view
```

## Accessibility risk flags (for the a11y auditor)

- **color-only-signaling**: P/E distinction is carried by the text group headers
  (`P-cores (12):`), never by bar colour alone. The mockup may *add* colour but
  must keep the text labels.
- **min-width-resize**: the cluster line (`cpu clusters: P 15.0 W · E 3.1 W`,
  ≤36 chars) fits MEDIUM and WIDE; it is dropped at TINY by design, not clipped.
- **no-color-support**: group headers and counts parse under `NO_COLOR`.
- **screen-reader order**: P block, then E block, in logical-core order — reading
  order matches visual order.
