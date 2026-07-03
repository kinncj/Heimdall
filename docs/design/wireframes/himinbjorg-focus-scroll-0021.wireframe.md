---
story_id: "himinbjorg-focus-scroll-0021"
story_file: "docs/stories/himinbjorg-focus-scroll-20260702235455-0021/Story.md"
design: "docs/specs/himinbjorg-focus-scroll/design.md"
target: tui
status: draft
approved_by: null
approved_at: null
created_at: "2026-07-02"
---

# Wireframe — himinbjorg-focus-scroll-0021 (Himinbjörg)

Target: **tui**. Low-fidelity, **structural only**. Colour, weight, and lipgloss roles are
deferred to the mockup. This wireframe defines layout, focus, scroll, and interaction states for
the one shared `pane.Pane` / `pane.Group` primitive that replaces the three divergent scroll
implementations (dashboard, topview, selection lists).

The panes **render only** (ADR-0022). This story adds *interaction* — focus, scroll, filter, sort —
never metric meaning. Static P/E/LP core panels display already-classified neutral rows and get
**no** filter and **no** sort.

## Focus-ring convention (read this first — it is load-bearing, not color)

Focus is carried by **border weight**, never colour alone:

```
┏━━━━━━━━━━━━┓      heavy / double-struck border  = FOCUSED pane (draws the focus ring)
┃  FOCUSED   ┃      the focused pane also owns the ▸ active-row caret and the footer affordances
┗━━━━━━━━━━━━┛

┌────────────┐      light single border           = unfocused pane
│ unfocused  │
└────────────┘
```

- Exactly **one** pane in a Group is focused at a time.
- The focused pane's **active row** is prefixed `▸`; other rows are prefixed with a space.
- Footer renders **only** the affordances the focused pane actually supports (capability-driven).

## Chrome (fixed, every state)

- **Header** (top, fixed): `⬢ HEIMDALL · <view> · <host> · <os> <arch> · up <uptime>` left,
  `● ONLINE` host-state badge right. Mirrors the existing `topview` header.
- **Footer** (bottom, fixed): the keybind legend. It swaps by focused-pane capability (see below).
- **Body** (between, scrolls): the pane Group. Header + footer never scroll; the frame never
  exceeds the terminal height.

## Keybind legend (footer) — capability-driven

The full vocabulary is consistent everywhere; the footer only shows what applies:

```
↑/↓ scroll · Tab focus · / filter · s sort · esc        ← full (focused pane is Filterable+Sortable)
↑/↓ scroll · Tab focus · / filter · esc                 ← log/cmd pane   (Filterable only)
↑/↓ scroll · Tab focus · esc                            ← static core pane (Source only: no /, no s)
↑/↓ scroll · esc                                        ← single-pane Group (detail view: no Tab)
type to filter · enter apply · esc clear                ← while `/` filter input is active
```

Rule: a static P/E/LP core pane, when focused, shows **no** `/ filter` and **no** `s sort` token.
Pressing `/` or `s` on it is a no-op (Scenario: "Static core panels offer neither filter nor sort").

## Two-level scroll model (the hard part)

- **Page scroll** (Group): when stacked panes exceed the terminal, the Group scrolls the whole
  canvas. Affordance: `▲ more above` / `▼ more below` + `scroll y/total` on the body edge rows.
- **Pane scroll** (Pane): a pane taller than the viewport scrolls its **own** body. Affordance:
  the same `▲/▼ more` + `scroll y/total`, rendered **inside** that pane's box.
- **Invariant:** page scroll and pane scroll cooperate so the **focused pane's active row is never
  hidden**. When the active row would fall off the bottom, the Group page-scrolls to reveal it.

## Legend (state = symbol + text, never colour alone)

```
● ONLINE   ◐ degraded   ○ offline   ⏱ stale     ← host-state badge (glyph + word)
▸ row                                            ← active / selected row caret
▲ / ▼ more                                       ← scroll affordance (page edge or in-pane)
CPU% ▼                                           ← active sort column + direction (▼ desc / ▲ asc)
/term█                                           ← active filter input line (█ = cursor)
(no matches)                                     ← empty-filter placeholder (muted)
—                                                ← metric unavailable (never a fake 0)
☟ (r,c)                                          ← mouse pointer marker + row,col for these frames
```

## Focus / tab order (keyboard-reachable — no mouse-only path)

Tab order follows visual reading order: **top-to-bottom**, and within a WIDE grid row
**left-to-right**. On the top screen the order is:

1. CPU  →  2. MEMORY  →  3. POWER  →  4. GPU/NPU  →  5. NET&DISK  →  6. LOAD/UPTIME  →  7. PROCESSES

- `Tab` advances to the next pane; `Shift-Tab` to the previous; both **wrap** at the ends
  (PROCESSES → CPU on Tab from the last pane).
- Within the focused pane: `↑/↓`·`k/j` scroll/move selection, `pgup/pgdn` page, `home/end` jump.
- `/` enters filter (Filterable panes only); `s` cycles the sort key (Sortable panes only);
  `esc` leaves the filter first, else leaves the view.
- Mouse is **additive, never required**: wheel scrolls the pane under the pointer, click focuses it.
  Every mouse action has a keyboard equivalent above.

## Accessibility risk flags (for the a11y auditor)

- **focus-visible**: focus is drawn as a heavier border **and** the `▸` active-row caret **and** the
  footer affordance set — three non-colour signals. Verify the mockup keeps the border-weight
  difference (not just a colour swap) so focus survives `NO_COLOR`.
- **color-only-signaling**: sort direction is a glyph (`▼`/`▲`) beside the column name, not colour.
  Filter match is conveyed by row removal + count, not highlight colour. Keep both.
- **keyboard-reachable**: mouse wheel/click (States 6) are conveniences; Tab/`↑↓`/`/`/`s` cover the
  same actions. No action is mouse-only. Flag if the mockup introduces a mouse-only affordance.
- **min-width-resize**: States 2 and small frames must not clip at the minimum supported width;
  every block here is width-bounded. Resize re-clamps both page and pane scroll (State via note).
- **color-only-signaling (empty)**: the empty-filter state shows the **text** `(no matches)`, not an
  empty coloured box.

## Open questions for product-owner (surfaced, not invented)

1. Sort key `s` cycles CPU% → MEM% → PID → COMMAND and back. Confirm the cycle order and whether a
   second key reverses direction (▼/▲), or whether each press toggles direction on re-selection.
   (Wireframe assumes: `s` advances column, re-pressing the current column flips direction.)
2. Does `/` filter on the PROCESSES pane match COMMAND only, or COMMAND+USER? (Wireframe assumes
   COMMAND substring; log pane is full-text; command list is name-match — per the design.)

---

## State 1 — Top screen, multi-pane, FITS the terminal (CPU focused)

WIDE tier. Everything fits; no page-scroll affordance. **CPU** pane is focused (heavy border +
`▸` active row); all other panes are unfocused (light border). Footer shows the static-core legend
because CPU here is a render-only core panel: `Tab focus` present, **no** `/ filter`, **no** `s sort`.

```text
╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation · macOS 14.4 arm64 · up 6d 4h                            ● ONLINE ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┌────────────────────────────────────────────┐   ║
║ ┃ CPU                              ◂ focused ┃  │ MEMORY                                     │   ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ├────────────────────────────────────────────┤   ║
║ ┃▸util 72%   freq 3.20 GHz   load 2.41       ┃  │ used 61%    18.4 / 32 GB                   │   ║
║ ┃ util ⣀⣤⣶⣷⣿⣿⣿⣷⣶⣤⣀⣤⣶⣷⣿⣷⣶  72%              ┃  │ swap  2%     0.6 / 8 GB                    │   ║
║ ┃ per-core (10):                             ┃  │ bw   ⣀⣤⣶⣷⣿⣿⣷⣶⣤⣶⣷⣿  41 GB/s                 │   ║
║ ┃ c0 ███▌71  c1 ██▌ 52  c2 ████ 80           ┃  │                                            │   ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  └────────────────────────────────────────────┘   ║
║                                                                                                  ║
║ ┌────────────────────────────────────────────┐  ┌────────────────────────────────────────────┐   ║
║ │ POWER                                      │  │ GPU / NPU                                  │   ║
║ ├────────────────────────────────────────────┤  ├────────────────────────────────────────────┤   ║
║ │ pkg 22.4 W   cpu 14.1 W   gpu 6.0 W        │  │ gpu util 64%   vram 47%   temp 58°C        │   ║
║ │ pwr ⣀⣀⣤⣶⣷⣿⣷⣶⣤⣀⣤⣶  22 W   npu —            │  │ npu util 38%                               │   ║
║ └────────────────────────────────────────────┘  └────────────────────────────────────────────┘   ║
║                                                                                                  ║
║ ┌──────────────────────────────────────────────────────────────────────────────────────────────┐ ║
║ │ PROCESSES (top by cpu)                                                                       │ ║
║ ├──────────────────────────────────────────────────────────────────────────────────────────────┤ ║
║ │ PID     USER      CPU%    MEM%   COMMAND                                                     │ ║
║ │ 1234    kinncj    42.1     3.2   heimdall-dashboard                                          │ ║
║ │  880    root      11.7     1.1   heimdall-daemon                                             │ ║
║ └──────────────────────────────────────────────────────────────────────────────────────────────┘ ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · Tab focus · esc back · q quit                                      host workstation ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝
```

Note: `◂ focused` is an annotation for this wireframe, not chrome. The real signal is the heavy
border + `▸` caret. Because CPU is a static core panel, the footer omits `/` and `s`.

---

## State 2 — Small terminal: WHOLE-PAGE scroll keeps the focused row visible

The stacked panes are taller than this short terminal, so the **Group page-scrolls**. Edge rows carry
`▲ more above` / `▼ more below` + `scroll y/total`. Focus is on **PROCESSES**; the page has scrolled
so the focused pane's active row (`▸ 1234 … heimdall-dashboard`) stays on screen. PROCESSES is
Filterable+Sortable, so the footer shows the full legend.

```text
╔════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation                                    ● ONLINE ║
╠════════════════════════════════════════════════════════════════════════════╣
║ ▲ more above (CPU · MEMORY · POWER panels scrolled off)         scroll 14/33 ║
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ PROCESSES (top by cpu)                                       ◂ focused ┃ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ ┃ PID    USER     CPU%   MEM%  COMMAND                                   ┃ ║
║ ┃▸1234   kinncj   42.1    3.2  heimdall-dashboard        ← active, kept  ┃ ║
║ ┃  880   root     11.7    1.1  heimdall-daemon                          ┃ ║
║ ┃ 2051   kinncj    8.4    6.0  firefox                                  ┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
║ ▼ more below                                                    scroll 14/33 ║
╠════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · Tab focus · / filter · s sort · esc back                       ║
╚════════════════════════════════════════════════════════════════════════════╝
```

Scenario covered: *"Whole-page scroll on a small terminal keeps the focused row visible."* Scrolling
down past the bottom of the focused pane moves the canvas to follow it; the active row is never hidden.

---

## State 3 — A pane taller than the viewport scrolls its OWN overflow inside the page

PROCESSES holds more rows than fit. The **page** is scrolled (edge affordances at the body edges) and
the focused pane **also** scrolls internally — note the in-pane `▲/▼ more` rows **inside** the box,
distinct from the page-level ones. The two cooperate; the active row `▸ 2051 firefox` stays visible.

```text
╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation · macOS 14.4 arm64 · up 6d 4h                            ● ONLINE ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ▲ more above (core panels scrolled off)                                             scroll 9/40 ║
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ PROCESSES (top by cpu)                                                            ◂ focused ┃ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ ┃ PID     USER      CPU%    MEM%   COMMAND                                                     ┃ ║
║ ┃ ▲ more above (12 higher-cpu processes)                                     in-pane 13/128   ┃ ║
║ ┃  611    root      15.2     0.9   containerd                                                  ┃ ║
║ ┃▸2051    kinncj     8.4     6.0   firefox                     ← active, kept inside the pane  ┃ ║
║ ┃ 3120    kinncj     3.2     2.4   ghostty                                                     ┃ ║
║ ┃ 4410    kinncj     2.1     1.2   Code Helper (Renderer)                                      ┃ ║
║ ┃ ▼ more below (110 more)                                                    in-pane 13/128   ┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
║ ▼ more below                                                                        scroll 9/40 ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · Tab focus · / filter · s sort · esc back · q quit                                   ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝
```

Scenario covered: *"A pane taller than the viewport scrolls its own overflow"* + the cooperation
invariant. Two scroll counters are intentionally shown: page `scroll 9/40` (outer) and
`in-pane 13/128` (inner).

---

## State 4 — Tab moves focus between panes and draws the focus ring

Two frames of the same screen. **Before**: focus on MEMORY (pane A). **After Tab**: focus on POWER
(pane B). Only the focused pane draws the heavy ring; `Shift-Tab` reverses; Tab from the last pane
wraps to the first.

### 4a — before Tab (MEMORY focused)

```text
║ ┌──────────────────────┐  ┏━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ │ CPU                  │  ┃ MEMORY    ◂ focused ┃ ║
║ ├──────────────────────┤  ┣━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ │ util 72%  freq 3.2G  │  ┃▸used 61%  18.4/32GB ┃ ║
║ │ per-core (10) …      │  ┃ swap  2%   0.6/8 GB ┃ ║
║ └──────────────────────┘  ┗━━━━━━━━━━━━━━━━━━━━━━┛ ║
```

### 4b — after Tab (focus advanced to POWER)

```text
║ ┌──────────────────────┐  ┌──────────────────────┐ ║
║ │ CPU                  │  │ MEMORY               │ ║
║ ├──────────────────────┤  ├──────────────────────┤ ║
║ │ util 72%  freq 3.2G  │  │ used 61%  18.4/32GB  │ ║
║ │ per-core (10) …      │  │ swap  2%   0.6/8 GB  │ ║
║ └──────────────────────┘  └──────────────────────┘ ║
║ ┏━━━━━━━━━━━━━━━━━━━━━━┓  ┌──────────────────────┐ ║
║ ┃ POWER     ◂ focused ┃  │ GPU / NPU            │ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━┫  ├──────────────────────┤ ║
║ ┃▸pkg 22.4 W          ┃  │ gpu 64% vram 47%     │ ║
║ ┃ cpu 14.1 W  gpu 6 W ┃  │ npu 38%              │ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━┛  └──────────────────────┘ ║
```

Scenarios covered: *"Tab moves focus between panes and draws the focus ring"* and
*"Tab focus wraps around at the ends."* Tab order for this grid: CPU → MEMORY → POWER → GPU/NPU → … .

---

## State 5 — Selection list (command picker): cursor near the bottom edge, window followed it

The command list is taller than its window. The selection `▸` has moved to the last visible row; the
pane has scrolled so the highlighted row stays inside the window (never off-screen). `▲ more above`
shows earlier commands scrolled off. Selection-follows-scroll.

```text
╔════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · commands · workstation                       ● ONLINE ║
╠════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ COMMANDS                                             ◂ focused ┃ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ ┃ ▲ more above (6 commands)                          sel 12/18   ┃ ║
║ ┃  restart daemon                                               ┃ ║
║ ┃  tail journalctl                                             ┃ ║
║ ┃  disk usage (df -h)                                          ┃ ║
║ ┃  network sockets (ss -tulpn)                                 ┃ ║
║ ┃▸ top processes                        ← cursor at bottom edge ┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
╠════════════════════════════════════════════════════════════════════╣
║ ↑/↓ move · Tab focus · / filter · enter run · esc back              ║
╚════════════════════════════════════════════════════════════════════╝
```

Scenario covered: *"Selection follows the cursor and stays in view."* Moving the selection toward an
edge scrolls the window so the highlighted row is never off-screen. `enter` runs the selected command.

---

## State 6 — Mouse: wheel scrolls the pane UNDER the pointer; click focuses it

Keyboard focus is on **CPU** (top-left). The mouse pointer `☟` hovers **PROCESSES** (a non-focused
pane). Turning the wheel scrolls **PROCESSES** — not the focused CPU pane, and no other pane.
A **click** at the same spot would move focus (draw the ring) to PROCESSES. Group hit-tests the
pointer's row/col against each pane's box.

```text
╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation · macOS 14.4 arm64 · up 6d 4h                            ● ONLINE ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┌────────────────────────────────────────────┐   ║
║ ┃ CPU  ◂ keyboard-focused (ring stays here)  ┃  │ MEMORY                                     │   ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ├────────────────────────────────────────────┤   ║
║ ┃▸util 72%   freq 3.20 GHz                   ┃  │ used 61%   18.4 / 32 GB                    │   ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  └────────────────────────────────────────────┘   ║
║                                                                                                  ║
║ ┌──────────────────────────────────────────────────────────────────────────────────────────────┐ ║
║ │ PROCESSES (top by cpu)              wheel here scrolls THIS pane, not CPU        in-pane 3/128 │ ║
║ ├──────────────────────────────────────────────────────────────────────────────────────────────┤ ║
║ │ PID     USER      CPU%    MEM%   COMMAND                                                     │ ║
║ │  880    root      11.7     1.1   heimdall-daemon        ☟ (pointer row 11, col 40)           │ ║
║ │ 2051    kinncj     8.4     6.0   firefox                                                     │ ║
║ │ 3120    kinncj     3.2     2.4   ghostty                                                     │ ║
║ └──────────────────────────────────────────────────────────────────────────────────────────────┘ ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · Tab focus · / filter · s sort · esc back · q quit          wheel: pane under pointer ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝
```

Scenarios covered: *"Mouse wheel scrolls the pane under the pointer,"* *"Mouse wheel works in the
full-screen top view"* (wheel no longer falls through to a hidden grid cursor), and *"Clicking a pane
focuses it."* Keyboard path is unchanged, so no mouse-only dependency.

---

## State 7 — `/` filter active on the LOG pane (journalctl, full-text)

Focus on the log pane; `/` opened the filter input. Typing `oom` narrows to full-text matches only.
The footer legend **swaps** to the filter-input legend while typing. `esc` clears the filter and
restores every line.

```text
╔════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · logs · workstation · journalctl                                  ● ONLINE ║
╠════════════════════════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ LOG · journalctl                                                       ◂ focused   ┃ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ ┃ /oom█                                                       filter · 3 of 812 lines ┃ ║
║ ┃▸Jul 02 14:19:07 ws kernel: Out of memory: Killed process 5567 (chrome) oom_score…  ┃ ║
║ ┃ Jul 02 14:19:07 ws systemd: session-4.scope: Consumed … oom-kill triggered         ┃ ║
║ ┃ Jul 02 11:02:55 ws kernel: oom_reaper: reaped process 5567 (chrome)                ┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
╠════════════════════════════════════════════════════════════════════════════════════════╣
║ type to filter · enter apply · esc clear                                                ║
╚════════════════════════════════════════════════════════════════════════════════════════╝
```

Scenario covered: *"Slash filters journalctl logs by full text."* Only matching lines remain;
`esc` clears and restores all 812 lines. The `filter · 3 of 812 lines` count is text, not colour.

---

## State 8 — `/` filter on the COMMAND list (name match)

Same filter UX, different **meaning** decided by the source: the command list matches on command
**name**. Typing `net` keeps only name-matching commands.

```text
╔════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · commands · workstation                       ● ONLINE ║
╠════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ COMMANDS                                             ◂ focused ┃ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ ┃ /net█                                        filter · 2 of 18   ┃ ║
║ ┃▸ network sockets (ss -tulpn)                                   ┃ ║
║ ┃  network throughput (iftop)                                   ┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
╠════════════════════════════════════════════════════════════════════╣
║ type to filter · enter run · esc clear                              ║
╚════════════════════════════════════════════════════════════════════╝
```

Scenario covered: *"Slash filters the command list by name match."* Selection resets to the top of
the narrowed list.

---

## State 9 — Sort cycle on the PROCESSES pane (no modal; in-pane)

Focus on PROCESSES. Pressing `s` cycles the sort key **in place** — no separate sort modal opens
(the old `topSort` modal is folded away). The active sort column carries a direction glyph
(`CPU% ▼` = descending) in the header row, and the footer echoes the active sort key.

```text
╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation · macOS 14.4 arm64 · up 6d 4h                            ● ONLINE ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ PROCESSES · sort CPU% ▼                                                            ◂ focused ┃ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ ┃ PID     USER      CPU% ▼   MEM%    COMMAND                                                   ┃ ║
║ ┃▸1234    kinncj    42.1      3.2    heimdall-dashboard                                         ┃ ║
║ ┃  880    root      11.7      1.1    heimdall-daemon                                            ┃ ║
║ ┃ 2051    kinncj     8.4      6.0    firefox                                                    ┃ ║
║ ┃ 3120    kinncj     3.2      2.4    ghostty                                                    ┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · Tab focus · / filter · s sort:CPU% ▼ · esc back · q quit                            ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝
```

Next `s` → `MEM% ▼` (rows reorder by MEM%); the header glyph and footer token move to MEM%.
Scenarios covered: *"Column sort cycles the top process list"* and *"The sort modal is replaced by
the in-pane sort-key cycle."*

---

## State 10 — Empty filter state (no matches)

A filter term matching no row shows a **single muted placeholder row**; the selection resets to the
top. No error, no empty coloured box — the word `(no matches)` carries the state.

```text
╔════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · logs · workstation · journalctl                                  ● ONLINE ║
╠════════════════════════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ LOG · journalctl                                                       ◂ focused   ┃ ║
║ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ ║
║ ┃ /zzxqq█                                                     filter · 0 of 812 lines ┃ ║
║ ┃ (no matches)                                                                       ┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
╠════════════════════════════════════════════════════════════════════════════════════════╣
║ type to filter · enter apply · esc clear                                                ║
╚════════════════════════════════════════════════════════════════════════════════════════╝
```

Scenario covered: *"A filter that matches nothing shows an empty state."* `esc` restores all rows;
selection is already reset to 0.

---

## State 11 — Detail view: a single whole-page pane (confirmed, not multi-pane)

The detail view is a **1-pane Group**. There are no separate Tab-focusable sections — the whole body
scrolls as one pane. Because it is single-pane, the footer omits `Tab focus`. Scrolling moves the
whole detail body; `▲/▼ more` + `scroll y/total` marks position.

```text
╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · detail · workstation · macOS 14.4 arm64                                    ● ONLINE ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ║
║ ┃ ▲ more above (Identity · Metrics sections scrolled off)                            scroll 18/47┃ ║
║ ┃                                                                                              ┃ ║
║ ┃ Power                                                                                        ┃ ║
║ ┃   package        22.4 W                                                                      ┃ ║
║ ┃   cpu            14.1 W                                                                       ┃ ║
║ ┃   gpu             6.0 W                                                                       ┃ ║
║ ┃   npu             —      (unavailable on this host)                                          ┃ ║
║ ┃                                                                                              ┃ ║
║ ┃ Network                                                                                      ┃ ║
║ ┃   rx            3.20 MB/s                                                                     ┃ ║
║ ┃   tx            0.80 MB/s                                                                     ┃ ║
║ ┃ ▼ more below (Disk · Processes · Journal sections)                                scroll 18/47┃ ║
║ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · pgup/pgdn page · esc back · q quit                                                  ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝
```

Scenario covered: *"The detail view is a single whole-page pane."* One focus ring wraps the entire
body; no in-body Tab targets. (`Assumption` in the design §5 adoption map — confirmed here.)

---

## Live-refresh & resize behaviour (no separate frame — behaviour notes)

- **Live refresh** (new metrics snapshot): scroll offset and selection are **preserved** — the frame
  numbers update in place, the `▸` row and `scroll y/total` do not jump. The pane re-clamps on the
  next key press. (Scenario: *"Live refresh preserves scroll and selection."*)
- **Resize smaller**: page scroll and pane scroll are **re-clamped** into range; the focused pane
  index is preserved. Affordances (`▲/▼ more`, `scroll x/y`) appear/disappear at the fit boundary —
  present only when content exceeds the viewport. (Scenarios: *"Resize re-clamps…"*,
  *"Affordances appear and disappear at the fit boundary."*)

## Interaction Notes (summary)

- **Tab order**: CPU → MEMORY → POWER → GPU/NPU → NET&DISK → LOAD/UPTIME → PROCESSES, wrapping.
  Detail view is single-pane (no Tab). Reading order top-to-bottom, left-to-right within a grid row.
- **Primary action per screen**: on the top screen, move focus + scroll to read a pane; on the
  command list, `enter` runs the highlighted command. One primary action, others subordinate.
- **Error / empty states are first-class**: empty-filter placeholder (State 10) and `—` unavailable
  metrics are explicit, not afterthoughts.
- **Focus signal is non-colour**: heavy border + `▸` caret + capability-driven footer (three signals).
- **Mouse is additive**: every wheel/click action has a keyboard equivalent; no mouse-only path.

### Approval

- [ ] Approved by product owner
- [ ] Approved by UX lead (if applicable)
