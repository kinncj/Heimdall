---
id: "dashboard-responsive-columns-0023"
# ui:false — reflow of the existing fleet grid. No new screen, component, layout
# concept, or palette; columns are dropped/condensed to fit narrow terminals.
# Gated by tests (display-width bounds + layout table).
title: "Dashboard responsive columns for narrow / portrait terminals"
epic: "yggdrasil-tui-grouping"
priority: "high"
ui: false
adr_required: false
milestone: "v1.5.2"
phase: done
labels:
  - "type:bug"
  - "priority:high"
status: done
issue_number: null
issue_url: null
created_at: "2026-06-29T03:51:17+0000"
---

# Dashboard responsive columns for narrow / portrait terminals

## Story

**As an** operator viewing the fleet on a narrow screen (iPad mini / phone in
portrait, SSH via Termius),
**I want** the grid to drop and condense columns to the terminal width,
**so that** rows are never clipped off the right edge and the data stays readable.

## Context

v1.5.1 fixed vertical overflow. Horizontally the grid still forces content to
width ≥ 88 and renders fixed-width columns, so on a narrow terminal the rows are
clipped at the right edge (DISK → `D`, TEMP/GPU/PWR gone) and the chrome's right
border falls off-screen. Reproduced at 64 columns.

## Acceptance Criteria

```gherkin
@story:dashboard-responsive-columns-0023 @epic:yggdrasil-tui-grouping @priority:high
Feature: Responsive fleet-grid columns

  Scenario: No rendered line exceeds the terminal width
    Given a fleet shown on a narrow terminal
    When the dashboard renders the grid
    Then every rendered line fits within the terminal width

  Scenario: Columns drop right-to-left as the terminal narrows
    Given a wide terminal showing all metric columns
    When the terminal narrows
    Then the least-essential columns (power, gpu, temp, …) drop first
    And host name, state, and CPU remain visible longest

  Scenario: State condenses to a glyph on very narrow terminals
    Given a terminal too narrow for the full state badge plus a metric column
    When the dashboard renders the grid
    Then the state shows as a compact glyph instead of the full badge

  Scenario: The header, status, and footer fit the terminal width
    Given a narrow terminal
    When the dashboard renders
    Then the chrome borders are not clipped and the footer keys are not cut off

  Scenario: A wide terminal still shows every column
    Given a terminal at least as wide as the full column set
    When the dashboard renders the grid
    Then all columns (host, state, cpu, mem, disk, temp, gpu, power) are shown
```

## Definition of Done

- [ ] Unit tests green (layout breakpoint table + per-line display-width bounds)
- [ ] `make test-all` green (no regressions)
- [ ] Columns implemented as a registry/strategy (no `switch` on width)
- [ ] CHANGELOG entry added under [1.5.2]
- [ ] Released as v1.5.2

## Non-goals

- Per-cell shrinking of gauges/values (columns are dropped whole, not resized).
- A dedicated single-column "card" layout; this story keeps the tabular grid.

## Design notes

- Metric columns are a `gridColumn` registry (title, width, renderer), ordered
  most→least essential. `fitColumns` greedily includes the prefix that fits — the
  drop order is the registry order, not a conditional.
- `gridLayout` is the per-width plan: name width, full-badge vs compact-glyph
  state, and the visible columns. Computed once per frame and shared by the
  column header and every row so they stay aligned.
- Chrome (header/status) renders at the actual terminal width; the ≥88 floor is
  removed. The footer falls back to a glyph-only form when the labelled form
  would not fit.
