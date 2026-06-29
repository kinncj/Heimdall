---
id: "dashboard-small-screen-viewport-0022"
title: "Dashboard small-screen viewport + term-scoped fleet filter"
epic: "yggdrasil-tui-grouping"
# ui:false — bugfix to an existing TUI surface. No new screen, component, layout,
# or palette is introduced (the only new glyph is a terminal-native "↑/↓ N more"
# row), so the wireframe/mockup design gates do not apply. Gated by tests.
priority: "high"
ui: false
adr_required: false
milestone: "v1.5.1"
phase: implement
labels:
  - "type:bug"
  - "type:feature"
  - "priority:high"
status: in_progress
issue_number: null
issue_url: null
created_at: "2026-06-29T03:21:48+0000"
---

# Dashboard small-screen viewport + term-scoped fleet filter

## Story

**As an** operator viewing the fleet on a small screen (SSH from an iPad mini /
phone via Termius),
**I want** the dashboard grid to fit the terminal height and the `/` filter to
search hosts and groups by scoped terms,
**so that** filtering is visibly effective on short screens and I can narrow the
fleet by host, tag, hub, OS, or state without grouping first.

## Context

Two defects, one render path (`app/internal/tui/dashboard`):

1. **Viewport overflow (the reported bug).** `GridView` renders every matched
   row plus fixed chrome and never clamps to the terminal height (`m.height` is
   captured but unused). On a screen shorter than the frame, content overflows
   the alt-screen: the header scrolls off and the visible slice barely moves when
   filtering ungrouped, so the filter *looks* inert. Grouping reorders the top,
   so it *looks* like filtering only works when grouped. Reproduced at 100×14.

2. **Filter matching is unintuitive.** `matchesFilter` does a raw substring match
   over `name+id` and each literal `key=value`, with no field scoping. `env=fo`
   fails to match `environment=foo`; you cannot scope a term to the host name vs.
   a tag.

## Acceptance Criteria

```gherkin
@story:dashboard-small-screen-viewport-0022 @epic:yggdrasil-tui-grouping @priority:high
Feature: Small-screen viewport and term-scoped fleet filter

  Scenario: The grid never renders taller than the terminal
    Given a fleet with more hosts than fit on a short terminal
    When the dashboard renders the grid at that terminal height
    Then the rendered frame has no more lines than the terminal height

  Scenario: The selected host stays visible when the list overflows
    Given a fleet with more hosts than fit on a short terminal
    When the cursor moves to a host below the visible window
    Then the rendered frame still includes the selected host's row

  Scenario: An overflowing list shows how many rows are hidden
    Given a fleet with more hosts than fit on a short terminal
    When the dashboard renders the grid
    Then an indicator reports how many rows are hidden above or below

  Scenario: A bare term searches every field
    Given a fleet with hosts "bar" and "baz" and hosts tagged env=bar
    When the operator filters by the bare term "ba"
    Then the hosts named "bar" and "baz" and the env=bar hosts all remain visible

  Scenario: A host-scoped term searches only the host name
    Given a fleet with hosts "bar" and "baz" and a host tagged env=bar
    When the operator filters by "host=ba"
    Then only the hosts named "bar" and "baz" remain visible

  Scenario: A tag-scoped term searches only that tag
    Given a fleet grouped by an "env" tag with values "foo" and "bar"
    When the operator filters by "env=fo"
    Then only hosts tagged env=foo remain visible

  Scenario: Hub, OS, and state are scopable fields
    When the operator filters by "hub=home", "os=linux", or "state=offline"
    Then only hosts matching that field's value remain visible

  Scenario: The group alias scopes to the active grouping dimension
    Given the dashboard is grouped by the "env" tag
    When the operator filters by "group=foo"
    Then only hosts in the "foo" group remain visible

  Scenario: Multiple terms narrow conjunctively
    Given a fleet with mixed env and state
    When the operator filters by "env=prod state=online"
    Then only hosts that match both terms remain visible

  Scenario: An empty filter shows the whole fleet
    When the operator clears the filter
    Then every host is visible

  Scenario: An unknown field is treated as a literal value
    When the operator filters by "zzz=qqq" with no such field or tag
    Then the dashboard shows the empty state instead of any host

  Scenario: Filtering works identically with grouping off and on
    Given a fleet shown ungrouped
    When the operator applies a filter that matches a subset
    Then the same subset is shown whether grouping is off or on
```

## Definition of Done

- [ ] Unit tests green (filter grammar table + viewport windowing + height bound)
- [ ] `make test-all` green (no regressions)
- [ ] Filter scopes implemented as a strategy/adapter registry (no `switch` on field)
- [ ] CHANGELOG entry added under [1.5.1]
- [ ] Released as v1.5.1

## Non-goals

- Responsive horizontal column layout for very narrow terminals (portrait phones)
  is out of scope for 1.5.1; the grid keeps its fixed column widths. Tracked as a
  follow-up.
- Fuzzy subsequence matching; this story uses case-insensitive substring matching.

## Design notes

- Each searchable field is a `fieldMatcher` (adapter over `HostView`), registered
  in a list exactly like `groupDim` in `group.go` — Open/Closed: adding a scope is
  registering a matcher, not editing a conditional.
- `group` is an adapter that delegates to the active `groupDim`, reusing the
  grouping strategy rather than duplicating it.
- The viewport is a pure `window(lines, cursor, max)` function: clamps to `max`
  lines, keeps the cursor in view, and overwrites the edge lines with
  "↑/↓ N more" indicators so the frame height stays bounded.
