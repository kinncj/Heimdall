---
id: "fullscreen-top-view-0025"
title: "Full-screen top view (mactop/btop-style single-host system monitor)"
epic: "hlidskjalf-top-view"
priority: "high"
ui: true
design_target: "tui"
qa_bdd: "behave"
adr_required: false
milestone: "v2.2.0"
phase: discover
labels:
  - "type:feature"
  - "priority:high"
status: approved
issue_number: null
issue_url: null
created_at: "2026-06-30T15:28:49+0000"
---

# Full-screen top view (mactop/btop-style single-host system monitor)

## Story

**As an** operator watching a single host (often over SSH from Termius on an
iPhone or iPad),
**I want** a `t` key that opens a full-screen mactop/btop-style system view —
per-core CPU, memory, power, accelerators, network, disk, and processes — while
`p` opens the renamed process table,
**so that** I can read one host's live state at a glance on any screen size,
including very narrow terminals, without the view ever erroring or faking data
it cannot collect.

## Context

Today `t` opens `modalTop`, which is really a process table, not a system "top".
v2.2.0 reassigns the keys and adds a real full-screen single-host dashboard:

- `p` → **processes** (the existing process table, renamed).
- `t` → **top** (new full-screen system view); `esc`/`q` exits.

The new view reads data Heimdall already collects: `model.history`
(`map[HostID]map[string][]float64`) for time-series graphs and
`domain.Metric.PerCore []float64` for per-core values. It renders only; it does
not collect. New metrics (load average, swap, per-core frequency, memory
bandwidth, NPU utilisation) ride the existing additive pipeline and degrade
gracefully where a platform cannot supply them.

Design target is **tui**. BDD framework is **behave**; this story ships
behave `*_steps.py` stubs under `cucumber/` per repo convention.

The responsive behaviour reuses the pattern already established by
`dashboard-small-screen-viewport-and-filter-terms-0022` (height-clamped
`window()`/`scrollWindow` with fixed header/footer, driven by
`tea.WindowSizeMsg`) and `dashboard-responsive-columns-0023`
(`layout(width)` picking the densest plan that fits). This story does not invent
a new responsive mechanism; it applies the existing one to the new view.

Approved design: `docs/superpowers/specs/2026-06-30-top-view-design.md`.

## Acceptance Criteria

```gherkin
@story:fullscreen-top-view-0025 @priority:high @ui
Feature: Full-screen single-host top view

  # --- Keybind swap (p = processes, t = top) ---

  Scenario: Pressing p opens the renamed process table
    Given the dashboard is showing the fleet grid
    When the operator presses "p"
    Then the process table opens labelled "processes"

  Scenario: Pressing t opens the full-screen top view
    Given the dashboard has a focused host
    When the operator presses "t"
    Then the full-screen single-host top view opens for that host

  Scenario: Pressing esc or q exits the top view
    Given the operator is in the full-screen top view
    When the operator presses "esc" or "q"
    Then the dashboard returns to the fleet grid

  # --- Panel content (reads existing collected data) ---

  Scenario: The top view shows every system panel for the focused host
    Given the operator opens the top view for a focused host
    Then it shows per-core CPU bars, a CPU utilisation sparkline, and CPU frequency
    And it shows memory used and swap with a memory-bandwidth sparkline
    And it shows power for package, cpu, gpu, and npu with a power sparkline
    And it shows GPU and NPU utilisation, VRAM, and temperature
    And it shows network and disk sparklines
    And it shows the process list

  Scenario: Time-series graphs are drawn from the existing history buffers
    Given the focused host has recorded metric history
    When the top view renders its braille sparklines
    Then the sparklines are drawn from the existing per-host history buffers, not a new collector

  Scenario: The accelerator panel uses NPU terminology instead of ANE
    Given the operator opens the top view
    When the accelerator panel renders
    Then it is labelled "NPU" rather than "ANE"

  # --- Screen-size awareness (reuses the existing responsive pattern) ---

  Scenario: A wide terminal shows the two-column panel grid with full graphs
    Given a focused host shown in the top view
    When the terminal is at least 100 columns wide
    Then panels render in a two-column grid with full sparklines and a multi-column per-core matrix

  Scenario: A medium terminal stacks panels in a single column
    Given a focused host shown in the top view
    When the terminal is between 60 and 99 columns wide
    Then panels stack in a single column, sparklines are kept, and per-core bars wrap to fewer columns

  Scenario: A narrow terminal collapses per-core to an aggregate
    Given a focused host shown in the top view
    When the terminal is between 40 and 59 columns wide
    Then panels stack, sparklines are shortened, and per-core collapses to an aggregate bar with a core-count summary

  Scenario: A tiny terminal shows key numbers only
    Given a focused host shown in the top view
    When the terminal is narrower than 40 columns
    Then only key numbers (cpu%, mem%, power, temp) are shown, one value per line, with no graphs and no per-core bars

  Scenario: No rendered line exceeds the terminal width
    Given a focused host shown in the top view at any width tier
    When the top view renders
    Then every rendered line fits within the terminal width

  Scenario: Content taller than the terminal scrolls within a fixed header and footer
    Given the top view content is taller than the terminal height
    When the operator scrolls with the up and down keys
    Then the header and footer stay fixed, the body scrolls, and the frame never exceeds the terminal height

  Scenario: The top view stays usable at iPhone-portrait width in Termius
    Given the operator is connected over SSH from Termius on a phone in portrait
    When the top view renders at a width narrower than 40 columns
    Then it shows the key-numbers-only layout with no clipped lines

  # --- Graceful degradation: unavailable metrics render a dash ---

  Scenario Outline: A metric the platform cannot supply renders as a dash
    Given a "<platform>" host focused in the top view
    When the top view renders "<metric>"
    Then it shows "—" rather than an error or a fabricated 0

    Examples:
      | platform | metric    |
      | windows  | cpu.load  |
      | linux    | mem.bw    |
      | linux    | npu.util  |
      | windows  | cpu.freq  |

  Scenario Outline: A metric the platform supports renders a real value
    Given a "<platform>" host focused in the top view
    When the top view renders "<metric>"
    Then it shows a real value rather than a dash

    Examples:
      | platform | metric    |
      | macos    | cpu.load  |
      | linux    | mem.swap  |
      | macos    | mem.bw    |
      | macos    | npu.util  |

  # --- NPU rename, backward compatible ---

  Scenario: The legacy power.ane key is normalised to power.npu at ingest
    Given a daemon that reports accelerator power under the legacy key "power.ane"
    When the dashboard ingests that metric
    Then it is stored under the canonical key "power.npu"
    And the top view renders it under the NPU label

  Scenario: A canonical power.npu key is rendered directly
    Given a daemon that reports accelerator power under "power.npu"
    When the dashboard ingests that metric
    Then it is rendered under the NPU label without further translation

  # --- Backward compatibility in a mixed fleet ---

  Scenario: The top view works against an older daemon that lacks the new keys
    Given a mixed fleet where the focused host runs an older daemon without the new metric keys
    When the operator opens the top view for that host
    Then the panels render, the missing metrics show "—", and no error is raised

  Scenario: New metric keys are additive on the wire
    Given a v2.2 daemon emitting the new metric keys
    When an older hub or dashboard receives the stream
    Then it ignores the unknown keys and keeps working
```

## Definition of Done

- [ ] All Gherkin scenarios have passing behave step implementations
- [ ] Unit tests green (layout plan per width; sparkline golden tests; alias
      normalisation table; keybind press tests; per-line width bound)
- [ ] `make test-all` green (no regressions)
- [ ] Cross-builds green: linux + windows (CGO-free), darwin (cgo + CGO-free)
- [ ] Wireframe approved (required when `ui: true`)
- [ ] Mockup approved (required when `ui: true`)
- [ ] A11y audit passed (required when `ui: true`)
- [ ] CHANGELOG entry added under [2.2.0]
- [ ] PR description references this story

## Non-goals

- No new binary; the view lives inside `heimdall-dashboard`.
- No fleet aggregation in the top view — it shows the single focused host only.
- No persistence/history beyond the existing in-memory ring buffers.
- No new responsive mechanism; this reuses `layout(width)` +
  `window()`/`scrollWindow` + `tea.WindowSizeMsg`.

## Design notes

- New package `app/internal/tui/topview/` — a pure render + local key-state
  Bubble Tea sub-model the dashboard switches into on `t`. It takes a snapshot
  and a history slice in; it never reaches into adapters.
- `topview.layout(width, height)` returns the densest plan that fits (wide ≥100 /
  medium 60–99 / narrow 40–59 / tiny <40), mirroring the fleet grid's approach
  from `dashboard-responsive-columns-0023`.
- Vertical overflow reuses the `window()`/`scrollWindow` clamp with a fixed
  header and footer from `dashboard-small-screen-viewport-and-filter-terms-0022`,
  driven by `tea.WindowSizeMsg`.
- NPU rename is backward-compatible: `power.ane` is accepted as a read alias and
  normalised to `power.npu` in one place (dashboard metric intake), so older
  daemons keep rendering.
- Availability matrix (Unavailable renders `—`, never an error or fake 0):
  `cpu.load` (mac+linux), `mem.swap` (all), `cpu.freq` per-core (linux/mac
  best-effort, windows often unavailable), `mem.bw` (mac only), `npu.util`
  (mac only).

## ADR Links

<!-- populated by architect agent when adr_required: true -->
