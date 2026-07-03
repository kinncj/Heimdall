# Himinbjörg — unified focus + scroll for the TUI

**Status:** design (approved shape, pending spec review)
**Branch:** `feature/himinbjorg-focus-scroll`
**Codename:** Himinbjörg — Heimdall's hall at the top of Bifröst, the vantage from
which he watches every realm. A single place from which you navigate and see all
views. (Add a `docs/glossary.md` entry.)

## 1. Context

Scrolling in the TUI is implemented three separate times, and interaction is
inconsistent:

- **Dashboard detail / log view / cmd-result** — body rendered as `[]string`,
  windowed by `modalScroll`/`detailScroll` with a max-scroll clamp. Pure scroll.
  Shared `scrollWindow` lives in `app/internal/tui/dashboard/modal.go`.
- **Full-screen top view** (`app/internal/tui/topview`) — its own *separate*
  `scrollWindow(m Model, …)` copy; scrolls the whole panel body as one flat list.
- **Selection lists** — the command picker (`cmdSel`) and log-source/sort lists
  (`modalSel`) move a selection index that is **not** wired to the scroll offset.
  When the list is taller than the screen the highlighted row can move off-screen
  and the viewport does not follow it.

Additional gaps:

- Mouse wheel routes by *mode*, not pointer position (`MouseMsg.X/Y` ignored), and
  **does nothing in the full-screen top view** — a wheel event there falls through
  to moving the hidden grid cursor.
- Search/filter and sort exist only as one-off code bolted onto specific modals
  (`logQuery`/`matchesLogQuery`, `sortedProcesses`/`topSortBody`).

The top view is tall on small terminals, so it needs whole-page scroll *and*
per-region scroll at the same time.

## 2. Goals / Non-goals

**Goals**
- One shared, tested scroll/focus primitive. Delete both `scrollWindow` copies.
- Selection lists keep the cursor in view (selection follows scroll).
- Identical keys and affordances (`▲/▼ more`, `scroll x/y`) across every view.
- Tab / Shift-Tab focus between regions on multi-region screens, with a focus ring.
- Two-level scroll: whole page scrolls on small screens, focused region scrolls its
  own overflow, and the focused region's active row is never hidden.
- Mouse: wheel scrolls the pane under the pointer; click focuses it. Fixes
  wheel-in-top.
- Per-pane `/` search/filter and sort, contextual and opt-in.

**Non-goals**
- No change to what any metric *means*; this is pure TUI interaction (respects the
  metric-standardization boundary in ADR-0022 — panes render, they do not classify).
- No new persisted settings beyond what already exists (top sort default).
- No mouse drag-select, no text reflow.

## 3. Architecture

New package `app/internal/tui/pane`. Clean Architecture: it depends only on
`theme`/`render`, never on `dashboard`/`topview` (they depend on it).

### 3.1 `Pane` — one focusable, scrollable region

```go
// Pane is one focusable, scrollable region that knows its on-screen box so the
// mouse can target it.
type Pane struct {
    src            Source
    scroll, sel    int    // sel = -1 when the source is not selectable
    filtering      bool
    query, sortKey string
    box            Rect   // x,y,w,h; set at render, read for mouse hit-testing
    focused        bool
}
```

`Pane` owns — and is the single tested home for:
- the scroll window and clamp,
- the `▲ more above` / `▼ more below` + `scroll x/y` affordances,
- selection-follows-scroll (cursor kept inside the window),
- the `/` filter input UX and the sort-key cycle,
- the focus ring.

### 3.2 Capability interfaces (ISP — a pane gets only what its source opts into)

```go
type Source     interface { Rows() []string }               // base: styled lines
type Selectable interface { Source; RowCount() int }          // cursor + selection
type Filterable interface { SetFilter(q string) }             // enables "/"
type Sortable   interface { SortKeys() []string; SetSort(string) } // enables sort
```

The **source** decides *what* rows are and *what* filter/sort mean — full-text for
`journalctl` logs, column sort for `top` processes, name-match for the command
list. A static panel (P/E/LP cores) implements only `Source`: no cursor, no `/`,
no sort. Existing `logQuery`/`sortedProcesses` logic moves into the respective
sources unchanged in meaning.

### 3.3 `Group` — focus + two-level scroll

```go
type Group struct { panes []*Pane; focus, pageScroll int; box Rect }
```

- **Tab / Shift-Tab** cycle focus; the focused pane draws the ring.
- Scroll/select/`/`/sort keys route to the focused pane.
- **Small-screen invariant:** when the stacked panes exceed the terminal, the Group
  page-scrolls the canvas so the focused pane's active row is visible; a pane taller
  than the viewport also scrolls internally. Page scroll and pane scroll cooperate —
  **the focused pane's active row is never hidden.**
- Single-pane screens (detail view) are a 1-pane Group and get whole-page scroll for
  free.

### 3.4 Mouse

- Group hit-tests `tea.MouseMsg` against each pane's `box`:
  - **wheel** → scroll the pane under the pointer,
  - **click** → focus the pane under the pointer.
- Removes the mode-based routing in `dashboard.scroll(dir)` and fixes wheel doing
  nothing in the top view.

## 4. Interaction / keymap (consistent everywhere)

| Key | Action |
|---|---|
| ↑/↓ · k/j | scroll / move selection in focused pane |
| pgup/pgdn | page in focused pane |
| home/end | top / bottom of focused pane |
| Tab / Shift-Tab | next / previous pane (multi-pane screens) |
| `/` | filter focused pane (if `Filterable`); esc clears |
| `s` (or sort key cycle) | next sort key (if `Sortable`) |
| wheel | scroll pane under pointer |
| click | focus pane under pointer |
| esc | leave filter, else leave view |

Footers render only the affordances the focused pane actually supports.

## 5. Adoption map

| Surface | Becomes |
|---|---|
| Top screen panels | `Group` of panes (P/E/LP cores, processes, …) |
| Detail view sections | 1-pane Group (whole-page scroll). *Assumption:* sections stay one scrollable body, not separate Tab panes. |
| Log-source list + log view | 2-pane Group; log-view source is `Filterable` |
| `top`-in-modal / cmd list + result | Group; process source `Filterable`+`Sortable`; cmd list `Selectable` |
| Sort modal | folded into the `Sortable` sort-key cycle on the process pane |

Deleted after migration: `dashboard.scrollWindow`, `topview.scrollWindow`,
`dashboard.scroll(dir)` mode routing, the bespoke `topSort` modal.

## 6. Error handling / edge cases

- Empty source → one muted placeholder row; scroll/selection are no-ops.
- Live refresh (new snapshot) preserves scroll and selection, then re-clamps on the
  next key (matches today's `topview.Refresh`).
- Resize re-clamps page and pane scroll; focus index preserved.
- A pane shorter than the viewport shows no affordances and cannot scroll.
- `/` filter that matches nothing → placeholder row, selection resets to 0.

## 7. Testing

- **Pane** (table tests): scroll clamp; selection-follows-scroll at both edges;
  filter narrows rows and resets selection; affordance appears/disappears at the
  fit boundary.
- **Group**: Tab order and wrap; focus ring on the right pane; page-auto-scroll
  keeps the focused active row visible at small heights; mouse hit-testing by Y
  picks the correct pane for wheel and click.
- **Parity**: keep existing `dashboard/viewport_test.go`, `topview_test.go`,
  `coretopo_render_test.go` green (behavior parity) before extending.

## 8. Trade-offs & risks

- **FinOps/SRE:** none — client-side TUI only. No new deps, no runtime cost.
- **Risk:** two-level (page + pane) nested scroll is the hard part; the invariant
  "focused active row always visible" is the contract to test hardest.
- **Risk:** migrating every modal at once is a wide diff. Mitigation: land the
  primitive + tests first *within the branch*, migrate surface-by-surface behind
  green parity tests, delete duplicates last. One branch, incremental commits.
- **Codename:** Himinbjörg is a proposal; swap freely.

## 9. Process

- Built on `feature/himinbjorg-focus-scroll`; tested in-branch before merge.
- Implementation touches `app/` and `tests/` → routes through `/pipeline-runner`
  with a Gherkin story (project `CLAUDE.md` gate). This is one feature, not
  decomposed into separate stories.
- Add a `docs/glossary.md` entry for Himinbjörg.
