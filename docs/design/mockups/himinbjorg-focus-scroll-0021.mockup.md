---
story_id: "himinbjorg-focus-scroll-0021"
story_file: "docs/stories/himinbjorg-focus-scroll-20260702235455-0021/Story.md"
wireframe: "docs/design/wireframes/himinbjorg-focus-scroll-0021.wireframe.md"
design: "docs/specs/himinbjorg-focus-scroll/design.md"
target: tui
status: draft
approved_by: null
approved_at: null
created_at: "2026-07-02"
---

# Mockup — himinbjorg-focus-scroll-0021 (Himinbjörg)

Target **tui**. High-fidelity, lipgloss-annotated terminal render for the one shared
`pane.Pane` / `pane.Group` primitive. Colour comes from theme **roles** in
`docs/design/identity/terminal-theme.json`; every focus signal is also carried by a
non-colour channel so the UI survives `NO_COLOR`.

## Role → lipgloss binding (applies to every frame)

| Element | Theme role | lipgloss treatment | Non-colour signal |
|---|---|---|---|
| Focused pane frame | `focus` | `lipgloss.ThickBorder()` (`┏━┓`) | heavy border weight |
| Unfocused pane frame | `border` | `lipgloss.NormalBorder()` (`┌─┐`) | light border weight |
| Active row caret `▸` | `selection` | `Bold(true)` on the row | leading `▸` glyph (space when unfocused) |
| Scroll affordance `▲ more · y/total` | `caption` | `Faint(true)` | glyph + numeric counter |
| Footer keybind legend | `keybinding` (keys) + `text_muted` (labels) | keys `Bold`, labels `Faint` | capability-driven token set |
| Panel heading (`CPU`, `PROCESSES`) | `heading` | `Bold` | — |
| Metric value / row text | `value` | plain | — |
| Metric label / unit | `label` / `unit` | `Faint` | — |
| Brand sigil `⬢`, sort glyph `▼`/`▲` | `accent` | plain | glyph carries meaning |
| Filter input line | `focus` (prompt) + `value` (text) | prompt `Bold` | leading `/` |

**Load-bearing rule:** exactly one pane per Group is focused. Focus = ThickBorder + `▸`
caret + capability footer. Three redundant signals; none is colour-only.

Chrome mirrors the existing `app/internal/tui/topview` header/footer:
`⬢ HEIMDALL · <view> · <host> · <os> <arch> · up <uptime>` left, `● ONLINE` badge right;
footer = keybind legend. Header + footer are fixed; only the pane Group scrolls.

---

## Frame 1 — Top screen, fits, one pane focused

```
⬢ HEIMDALL · top · nidhogg-01 · darwin arm64 · up 6d 04:12          ● ONLINE

┏━ CPU ━━━━━━━━━━━━━━━━━━━━━━━━┓   ┌─ MEMORY ───────────────────┐
┃ ▸ P-cores  ▇▇▇▇▇▆▁▁  62%     ┃   │   used   11.4G / 16.0G  71% │
┃   E-cores  ▇▇▁▁▁▁▁▁  24%     ┃   │   swap    0.2G /  4.0G   5% │
┃   LP-core  ▁▁▁▁▁▁▁▁   3%     ┃   └────────────────────────────┘
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛   ┌─ POWER ────────────────────┐
┌─ GPU / NPU ─────────────────┐   │   pkg    14.2 W             │
│   gpu     ▇▇▇▁▁▁▁▁  38%      │   │   cpu     9.1 W            │
│   npu     ▁▁▁▁▁▁▁▁   0%      │   └────────────────────────────┘
└─────────────────────────────┘

  ↑/↓ scroll · Tab focus · esc back · q quit
```

- Focused pane **CPU**: `focus`/ThickBorder + `▸` on the active row (`P-cores`).
- All other panes: `border`/NormalBorder, rows prefixed with a space.
- CPU is static/neutral (already-classified P/E/LP rows) → footer shows **no** `/`, **no** `s`.
- Gauges reuse `render.Gauge` in `value`; headings in `heading`; units in `unit`.

---

## Frame 2 — Small terminal: whole-page scroll, focused row visible

```
⬢ HEIMDALL · top · nidhogg-01 · darwin arm64 · up 6d 04:12   ● ONLINE

  ▲ more above · page 3/7                        « caption, Faint »
┌─ MEMORY ───────────────────────────────────────────────────────┐
│   used   11.4G / 16.0G  71%                                     │
└────────────────────────────────────────────────────────────────┘
┏━ POWER ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ▸ pkg    14.2 W                                                ┃
┃   cpu     9.1 W                                                ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
  ▼ more below · page 3/7

  ↑/↓ scroll · Tab focus · esc back · q quit
```

- Page taller than the terminal → `Group.pageScroll` engaged. `▲/▼ more · page y/total`
  in `caption`/Faint on the body edges.
- **Invariant:** POWER is focused and the page auto-scrolled so its `▸` active row is on
  screen. The focused active row is never hidden behind the fold.

---

## Frame 3 — A pane taller than the viewport scrolls its own overflow

```
  ▲ more above · page 2/4

┏━ PROCESSES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃     PID    PPID   CPU%   MEM%  COMMAND      CPU% ▼             ┃
┃   ▲ more · 14/212                                             ┃  « in-pane caption »
┃ ▸  4821       1   38.2%   6.1%  /usr/bin/coreaudiod           ┃
┃    9013    4821   12.7%   2.4%  WindowServer                  ┃
┃     771       1    8.0%   1.1%  launchd                       ┃
┃   ▼ more · 14/212                                             ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  ↑/↓ scroll · Tab focus · / filter · s sort · esc back
```

- **Two distinct counters:** page `page 2/4` (Group) on the body edge vs in-pane
  `14/212` (Pane) inside the box. They cooperate — the `▸` row stays visible.
- Sort glyph `CPU% ▼` in `accent`. Footer now shows `/ filter · s sort` because the
  focused pane's source is `Filterable`+`Sortable`.

---

## Frame 4 — Tab moved focus to the next pane

```
┌─ CPU ───────────────────────┐   ┏━ MEMORY ━━━━━━━━━━━━━━━━━━━━┓
│   P-cores  ▇▇▇▇▇▆▁▁  62%     │   ┃ ▸ used   11.4G / 16.0G  71% ┃
│   E-cores  ▇▇▁▁▁▁▁▁  24%     │   ┃   swap    0.2G /  4.0G   5% ┃
│   LP-core  ▁▁▁▁▁▁▁▁   3%     │   ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
└─────────────────────────────┘

  ↑/↓ scroll · Tab focus · esc back · q quit
```

- Ring moved CPU → MEMORY. Only the border weight, caret, and footer changed; layout
  is stable. Tab order = visual reading order; wraps at both ends (Shift-Tab reverses).

---

## Frame 5 — Selection list (command picker), cursor near bottom, window followed

```
┏━ COMMAND — nidhogg-01 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃   ▲ more · 6/11                                               ┃
┃    restart-helper                                            ┃
┃    rotate-logs                                               ┃
┃    flush-metrics                                             ┃
┃ ▸  tail-journal                                              ┃   « selection role, Bold »
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  ↑/↓ pick · ⏎ run · / filter · esc back
```

- `Selectable` source. Cursor pushed to the last visible row → the pane scrolled so the
  `▸` selection stays inside the window (selection-follows-cursor). No more off-screen
  highlight. `▲ more · 6/11` shows there are earlier entries above.

---

## Frame 6 — Mouse: wheel over a non-focused pane, click to focus

```
┌─ CPU ───────────────────────┐   ┏━ MEMORY ━━━━━━━━━━━━━━━━━━━━┓
│   P-cores  ▇▇▇▇▇▆▁▁  62% ◄┐  │   ┃ ▸ used   11.4G / 16.0G  71% ┃
│   E-cores  ▇▇▁▁▁▁▁▁  24%  │  │   ┃   swap    0.2G /  4.0G   5% ┃
│   LP-core  ▁▁▁▁▁▁▁▁   3%  ▓  │   ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
└──────────────────────────┼──┘        ▓ = mouse pointer
                           └ wheel here scrolls CPU (pane under pointer),
                             NOT the focused MEMORY pane
```

- `Group` hit-tests `tea.MouseMsg` against each pane's `box` (x,y,w,h captured at render):
  - **wheel** → scroll the pane under the pointer (CPU), leaving keyboard focus on MEMORY.
  - **click** → focus the pane under the pointer (would move the ring to CPU).
- Same path drives the full-screen top view, fixing today's wheel-does-nothing bug.

---

## Frame 7 — `/` filter active on the journalctl log pane (full-text)

```
┏━ LOG — nidhogg-01 / journalctl ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ /timeout▏                                    23 shown / 4189   ┃   « / prompt: focus role; text: value »
┃   14:02:11  sshd[8123]: connection timeout from 10.0.0.4      ┃
┃ ▸ 14:03:40  kube-proxy: dial tcp timeout after 5s            ┃
┃   14:05:02  etcd: request timeout, retrying                  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  type to filter · ⏎ keep · esc clear · ↑/↓ scroll
```

- Source is `Filterable`; `/` opens an input line (blinking `▏`). `SetFilter("timeout")`
  narrows rows to full-text matches. Footer swaps to the **search** legend. `23 shown / 4189`
  in `caption`. esc clears the filter and restores the normal legend.

---

## Frame 8 — `/` filter on the command list (name match)

```
┏━ COMMAND — nidhogg-01 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ /log▏                                          2 shown / 11    ┃
┃ ▸  rotate-logs                                               ┃
┃    tail-journal                                             ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  type to filter · ⏎ run · esc clear · ↑/↓ pick
```

- Same `/` UX, different source semantics: name-substring match (`log`), not full-text.
  Selection stays valid and in-view after the list narrows.

---

## Frame 9 — `top` process pane, in-pane sort cycle

```
┏━ PROCESSES — nidhogg-01 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃     PID    PPID   CPU%   MEM%  COMMAND      CPU% ▼             ┃   « sort glyph: accent »
┃ ▸  4821       1   38.2%   6.1%  /usr/bin/coreaudiod           ┃
┃    9013    4821   12.7%   2.4%  WindowServer                  ┃
┃     771       1    8.0%   1.1%  launchd                       ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  ↑/↓ scroll · s sort · / filter · esc back
```

- `Sortable` source. `s` cycles keys: **CPU%↓ → MEM%↓ → PID↑ → COMMAND↑** (existing
  `sortedProcesses` defaults). Active key shows its glyph in the header (`CPU% ▼`).
- The old separate sort **modal is gone** — folded into this in-pane cycle.
- `/` filter matches the **COMMAND** column text.

---

## Frame 10 — Empty filter state

```
┏━ LOG — nidhogg-01 / journalctl ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ /zzznomatch▏                                    0 shown / 4189 ┃
┃                                                              ┃
┃        no lines match the search                             ┃   « text_muted, Faint »
┃                                                              ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  type to filter · esc clear
```

- Filter matches nothing → one placeholder row in `text_muted`; selection resets to 0;
  scroll/selection keys are no-ops until the filter narrows to ≥1 row.

---

## Frame 11 — Detail view: single whole-page pane

```
⬢ HEIMDALL · detail · nidhogg-01 · darwin arm64 · up 6d 04:12    ● ONLINE

  ▲ more above · 12/38
  Host          nidhogg-01
  State         ● online
  CPU model     Apple M3 Pro (11 cores)
  P-cores       6    E-cores  5    LP  0
  Memory        11.4G / 16.0G (71%)
  Power (pkg)   14.2 W
  ▼ more below · 12/38

  ⇧↑/↓ scroll · esc back · q quit
```

- Confirmed decision: the detail view is **one** whole-page pane (a 1-pane Group), not
  multiple Tab-focusable regions. No ring, no Tab — just whole-page scroll with the
  standard `▲/▼ more · y/total` affordance in `caption`. Behaviour parity with today's
  detail view, now on the shared primitive.

---

## a11y notes for the auditor

- Focus never depends on colour: ThickBorder weight + `▸` caret + capability footer are
  all colour-independent. Verify each renders under `NO_COLOR=1`.
- Every scrollable/selectable region is keyboard-reachable (Tab/Shift-Tab + arrows/pgup/
  pgdn/home/end); mouse is additive, never required.
- Affordance counters (`y/total`) give a non-visual position cue for screen scrapers.
- Static P/E/LP core panes correctly omit `/` and `s` (ADR-0022 render-only); no dead
  affordances advertised.
