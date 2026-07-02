---
id: top-view-core-topology-0028
story: docs/stories/top-view-core-topology-20260702184928-0028/story.md
wireframe: docs/design/wireframes/top-view-core-topology-0028.wireframe.md
tokens: docs/design/identity/tokens.json
theme: docs/design/identity/terminal-theme.json
target: tui
mode: dark
status: approved
approved_by: null
approved_at: null
created_at: 2026-07-02
---

# Mockup — top-view-core-topology-0028 (Hliðskjálf delta)

Delta mockup: only the **CPU** and **POWER** panels change; everything else is
pinned by the approved `fullscreen-top-view-0025.mockup.md` and its Styles table.
This document adds the two new regions and their lipgloss bindings. No new
colours — every role already exists in `terminal-theme.json`.

## State 1 — WIDE: CPU panel, hybrid host (12P + 4E)

```text
┌────────────────────────────────────────────┐
│ CPU                                        │
├────────────────────────────────────────────┤
│ util 72%    freq 3.20 GHz    load 2.41     │
│ util  ⣀⣤⣶⣷⣿⣿⣿⣷⣶⣤⣀⣤⣶⣷⣿⣷⣶  72%               │
│ E-cores (4):                               │
│ c0 █▌  12  c1 █   09  c2 ██  21  c3 █  07  │
│ P-cores (12):                              │
│ c4 ███▌71  c5 ██▌ 52  c6 ████ 80           │
│ c7 ██  41  c8 ███ 63  c9 █▌  33            │
│ c10███▌70  c11██▌ 55  c12████ 88           │
│ c13██  44  c14███ 61  c15██▌ 58            │
└────────────────────────────────────────────┘
```

Blocks render in **logical core order** — on this host the E-cluster owns
c0–c3, so E renders first. Bars, wrap, and severity colouring are unchanged
from 0025; only the `⟦core-group-header⟧` lines are new.

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

## State 3 — NARROW: one aggregate per type

```text
│ E (4)  █▌  12%    P (12) ███▌ 64%          │
```

## State 4 — fallback (no `cpu.topology`, or uniform): unchanged from 0025

```text
│ per-core (16):                             │
│ c0 ███▌71  c1 ██▌ 52  c2 ████ 80  …        │
```

No layout shift, no placeholder line.

## Styles (lipgloss) — new regions only

All other regions: see `fullscreen-top-view-0025.mockup.md` Styles table.

| Region | Theme role | Foreground | Attrs | Notes |
|---|---|---|---|---|
| ⟦core-group-header⟧ `P-cores (12):` / `E-cores (4):` | `type.label` (`structure.label`) | #949494 (246) | **bold** (not faint) | text carries the type — never colour-only; count in parens stays in the same run |
| ⟦cluster-label⟧ `cpu clusters:` | `type.label` (`structure.label`) | #949494 (246) | faint | same treatment as `util`/`freq` labels |
| ⟦cluster-type⟧ `P` / `E` in the cluster line | `type.label` (`structure.label`) | #949494 (246) | bold | matches the group headers |
| ⟦cluster-value⟧ `15.0` / `3.1` | `type.value` (`structure.value`) | #eeeeee (255) | bold | standard metric value |
| ⟦cluster-unit⟧ `W` | `type.unit` (`structure.unit`) | #949494 (246) | — | standard unit suffix |
| ⟦cluster-sep⟧ `·` | `structure.border` | #767676 (243) | — | same as header-trail separators |
| NARROW per-type aggregate bar | `render.Gauge` → `color.severity.*` | per value (ramp) | — | identical ramp to 0025 per-core bars |

Severity ramp, glyph sets, NO_COLOR behaviour: unchanged from 0025. The P/E
group headers are plain text runs, so the panel parses identically under
`NO_COLOR` and in monochrome terminals.

## A11y notes for the auditor

- P/E is signalled by text (`P-cores (12):`), reinforced — not replaced — by any
  colour the bars already carry from the severity ramp.
- The cluster line drops at TINY by design (key-numbers tier); it is never
  truncated mid-value.
- Reading order = render order = logical core order.
