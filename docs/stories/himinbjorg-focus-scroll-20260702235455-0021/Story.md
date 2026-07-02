---
id: "himinbjorg-focus-scroll-0021"
title: "Himinbjörg: Unified Focus + Scroll for the TUI"
epic: "himinbjorg-focus-scroll"
priority: "high"
ui: true
adr_required: true
milestone: null
phase: discover
labels:
  - "type:feature"
  - "priority:high"
  - "phase:discover"
issue_number: null
issue_url: null
created_at: "2026-07-02T23:54:55+0000"
---

# Himinbjörg: Unified Focus + Scroll for the TUI

## Story

**As an** operator navigating the Heimdall TUI,
**I want** one consistent focus-and-scroll system across every view, with mouse,
keyboard, per-pane search, and sort behaving the same everywhere,
**so that** I can move through tall panels, logs, process lists, and command
results without losing the row I care about — and so the codebase stops carrying
three divergent scroll implementations.

## Background / Constraints

- One shared primitive lives in a new `app/internal/tui/pane` package: a `Pane`
  (scroll, selection-that-follows-cursor, `/` filter, sort) and a `Group`
  (Tab/Shift-Tab focus, focus ring, two-level page + pane scroll).
- Both duplicate `scrollWindow` copies (`dashboard` and `topview`) are deleted;
  the mode-based mouse routing in `dashboard.scroll(dir)` and the bespoke sort
  modal are folded into the primitive.
- **Metric-standardization boundary (ADR-0022):** panes **render only**. A pane
  never classifies, buckets, or switches on platform. Static core panels (P/E/LP)
  display neutral, already-classified rows produced upstream in `domain`; they get
  no filter and no sort. This story adds interaction, not metric meaning.
- Target: `tui`. This is one feature, not decomposed into sub-stories.

## Acceptance Criteria

```gherkin
@story:himinbjorg-focus-scroll-0021 @epic:himinbjorg-focus-scroll @priority:high
Feature: Unified focus and scroll for the TUI

  @story @priority:high
  Scenario: Whole-page scroll on a small terminal keeps the focused row visible
    Given the top view is open on a terminal too short to show every panel
    When the operator scrolls down past the bottom of the focused pane
    Then the page scrolls the canvas to follow the focused pane
    And the focused pane's active row is never hidden

  @story @priority:high
  Scenario: A pane taller than the viewport scrolls its own overflow
    Given the top view is open with a focused pane taller than the terminal
    When the operator scrolls within that pane
    Then the pane scrolls its own body
    And the page scroll and pane scroll cooperate so the active row stays visible

  @story @priority:high
  Scenario: Tab moves focus between panes and draws the focus ring
    Given a multi-pane screen with more than one pane
    When the operator presses Tab
    Then focus moves to the next pane
    And only the focused pane draws the focus ring
    And Shift-Tab moves focus to the previous pane

  @story @priority:high
  Scenario: Tab focus wraps around at the ends
    Given a multi-pane screen with focus on the last pane
    When the operator presses Tab
    Then focus wraps to the first pane

  @story @priority:high
  Scenario: Selection follows the cursor and stays in view
    Given a selection list taller than the visible window
    When the operator moves the selection toward an edge of the window
    Then the pane scrolls so the highlighted row stays inside the window
    And the highlighted row is never off-screen

  @story @priority:high
  Scenario: Mouse wheel scrolls the pane under the pointer
    Given a multi-pane screen with panes stacked on screen
    When the operator turns the mouse wheel over a specific pane
    Then that pane scrolls
    And other panes do not scroll

  @story @priority:high
  Scenario: Mouse wheel works in the full-screen top view
    Given the full-screen top view is open
    When the operator turns the mouse wheel over a panel
    Then that panel scrolls
    And the wheel no longer falls through to move a hidden grid cursor

  @story @priority:high
  Scenario: Clicking a pane focuses it
    Given a multi-pane screen with focus on one pane
    When the operator clicks a different pane
    Then focus moves to the clicked pane
    And the clicked pane draws the focus ring

  @story @priority:high
  Scenario: Slash filters journalctl logs by full text
    Given the log view is focused and shows journalctl lines
    When the operator presses the slash key and types a filter term
    Then only log lines matching the term full-text remain visible
    And pressing escape clears the filter and restores every line

  @story @priority:high
  Scenario: Slash filters the command list by name match
    Given the command list is focused
    When the operator presses the slash key and types a name fragment
    Then only commands whose name matches remain visible

  @story @priority:high
  Scenario: Column sort cycles the top process list
    Given the top process pane is focused
    When the operator cycles the sort key
    Then the processes reorder by the active sort column
    And the footer shows the active sort key

  @story @priority:high
  Scenario: The sort modal is replaced by the in-pane sort-key cycle
    Given the top process pane is focused
    When the operator changes the sort order
    Then no separate sort modal opens
    And the sort changes in place on the focused pane

  @story @priority:high
  Scenario: Static core panels offer neither filter nor sort
    Given a static P/E/LP core panel is focused
    When the operator presses the slash key or the sort key
    Then nothing happens
    And the footer shows no filter or sort affordance for that panel

  @story @priority:high
  Scenario: A filter that matches nothing shows an empty state
    Given a filterable pane is focused
    When the operator enters a filter term that matches no row
    Then the pane shows a single muted placeholder row
    And the selection resets to the top

  @story @priority:high
  Scenario: Live refresh preserves scroll and selection
    Given a pane is scrolled and has a selected row
    When a new metrics snapshot arrives
    Then the scroll offset and selection are preserved
    And the pane re-clamps on the next key press

  @story @priority:high
  Scenario: Resize re-clamps scroll and preserves focus
    Given a multi-pane screen scrolled partway down
    When the terminal is resized smaller
    Then the page and pane scroll offsets are re-clamped into range
    And the focused pane index is preserved

  @story @priority:medium
  Scenario: Affordances appear and disappear at the fit boundary
    Given a pane whose content exactly fits the viewport
    When the content grows to exceed the viewport
    Then the "▲/▼ more" and "scroll x/y" affordances appear
    And they disappear again when the content fits

  @story @priority:medium
  Scenario: The detail view is a single whole-page pane
    Given the detail view is open
    When the operator scrolls
    Then the whole detail body scrolls as one pane
    And there are no separate Tab-focusable sections
```

## Definition of Done

- [ ] Unit tests green (Pane: scroll clamp, selection-follows-scroll at both edges,
      filter narrows rows and resets selection, affordance at fit boundary)
- [ ] Unit tests green (Group: Tab order and wrap, focus ring on correct pane,
      page-auto-scroll keeps focused active row visible, mouse hit-testing by Y)
- [ ] Integration tests green
- [ ] Cucumber/Behave scenarios green
- [ ] Parity tests kept green (`dashboard/viewport_test.go`, `topview_test.go`,
      `coretopo_render_test.go`) before extension
- [ ] Both `scrollWindow` copies deleted (dashboard + topview)
- [ ] Wireframe approved (required when `ui: true`)
- [ ] Mockup approved (required when `ui: true`)
- [ ] A11y audit passed (required when `ui: true`)
- [ ] ADRs linked where required (new `pane` primitive; ADR-0022 boundary respected)
- [ ] CHANGELOG entry added
- [ ] `docs/glossary.md` entry for Himinbjörg added
- [ ] PR description references this story

## ADR Links

<!-- populated by architect agent when adr_required: true -->
<!-- New architectural primitive: app/internal/tui/pane (Pane + Group). -->
<!-- Must respect ADR-0022 metric-standardization boundary: panes render only. -->
