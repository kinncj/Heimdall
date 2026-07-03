# ADR-0023 — Himinbjörg: one focus + scroll primitive for the TUI

- Status: accepted
- Date: 2026-07-02
- Story: `docs/stories/himinbjorg-focus-scroll-20260702235455-0021/Story.md`
- Design: `docs/specs/himinbjorg-focus-scroll/design.md`
- Relates to: [ADR-0022](0022-metric-standardization-boundary.md) (render-only consumers)

## 1. Context

Scrolling is implemented three times in the TUI, and the implementations diverge:

- `app/internal/tui/dashboard/modal.go` — `scrollWindow(lines, offset, max)`,
  edge rows replaced by `↑/↓ N more`. Drives detail, log view, cmd result.
- `app/internal/tui/topview/topview.go` — its own `scrollWindow(m, lines, offset,
  vis)`, edge rows replaced by a `scroll x/total` caption. Drives the top screen.
- Selection lists (`cmdSel`, `modalSel`) move a selection index that is **not**
  wired to the scroll offset, so a highlighted row can leave the viewport.

Mouse wheel routes by mode, ignores pointer position, and does nothing in the
full-screen top view. Search/filter and sort exist only as one-off code on
specific modals. On small terminals the top view is taller than the screen and
needs page scroll and per-region scroll at once.

## 2. Goals / Non-goals

**Goals:** one tested scroll/focus primitive; selection follows the cursor;
identical keys and affordances everywhere; Tab focus between regions with a focus
ring; two-level (page + pane) scroll that never hides the focused active row;
pointer-targeted mouse; opt-in per-pane search and sort.

**Non-goals:** no change to metric meaning (ADR-0022 holds — panes render, they do
not classify); no new persisted settings; no mouse drag-select; no text reflow.

## 3. Proposal

New framework-free package `app/internal/tui/pane` depending only on `theme` and
`render`. `dashboard` and `topview` depend on it; it depends on neither.

### `Pane` — one focusable, scrollable region

Holds a content `Source`, a scroll offset, an optional selection index, filter and
sort state, and its on-screen `Rect` (captured at render for mouse hit-testing).
`Pane` owns the scroll window + clamp, the `▲/▼ more · y/total` affordances,
selection-follows-scroll, the `/` filter input, the sort-key cycle, and the focus
ring. This is the single tested home for that behaviour; both existing
`scrollWindow` copies are deleted.

### Capability interfaces (ISP — a pane gets only what its source opts into)

```go
type Source     interface { Rows() []string }
type Selectable interface { Source; RowCount() int }
type Filterable interface { SetFilter(q string) }
type Sortable   interface { SortKeys() []string; SetSort(string) }
```

The source decides what rows are and what filter/sort mean (full-text for logs,
column sort for processes, name-match for the command list). Static P/E/LP panels
implement only `Source`.

### `Group` — focus + two-level scroll + mouse

Ordered panes + a focus index + a page-scroll offset. Tab/Shift-Tab cycle focus and
draw the ring; scroll/select/`/`/sort keys route to the focused pane; the page
auto-scrolls so the focused pane's active row is visible; `MouseMsg` is hit-tested
against each pane's `Rect` (wheel scrolls, click focuses the pane under the pointer).

## 4. Alternatives

- **Shared viewport only, per-screen focus.** Rejected: focus + page-auto-scroll
  logic gets reimplemented per screen — the duplication we are removing, one layer up.
- **Minimal fix (wire selection to scroll, unify keys, no Tab).** Rejected: does not
  deliver Tab-between-regions, an explicit requirement.

## 5. Trade-offs and Risks

- Two-level nested scroll is the hard part; the invariant "focused active row always
  visible" is the contract tested hardest.
- Migrating every modal at once is a wide diff. Mitigation: land the primitive with
  tests first, migrate surface-by-surface behind green parity tests, delete the
  duplicates last, all on `feature/himinbjorg-focus-scroll`.

## 6. Impact

- **FinOps:** none — client-side TUI, no new dependencies, no runtime cost.
- **SRE:** none — no daemon, transport, or adapter change. Failure surface is render
  only; a bad offset clamps, it does not crash.
- **Security:** none.
- **Team:** removes two divergent scroll copies; new interaction code has one home.

## 7. Decision

Adopt the `pane` package (`Pane` + `Group` + capability interfaces) as the single
focus/scroll primitive for the TUI. Migrate topview and all dashboard modals onto
it and delete the duplicate `scrollWindow` implementations.

## 8. Next Steps

- Build `pane` with unit tests (scroll clamp, selection-follows, filter, affordance
  edges, Tab order, page-auto-scroll, mouse hit-testing).
- Migrate `topview`, then dashboard modals; keep existing view tests green.
- a11y audit: focus signalled by border weight + caret + capability footer (survives
  `NO_COLOR`).
