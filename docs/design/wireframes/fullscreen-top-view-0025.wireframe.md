---
id: fullscreen-top-view-0025
story: docs/stories/fullscreen-top-view-20260630152849-0025/story.md
design: docs/superpowers/specs/2026-06-30-top-view-design.md
target: tui
status: draft
created_at: 2026-06-30
---

# Wireframe — fullscreen-top-view-0025 (Hliðskjálf)

Target: **tui**. Low-fidelity, structural only. Colour, weight, and lipgloss roles are
deferred to the mockup. This is the full-screen single-host "top" view entered with `t`
from the fleet grid (`p` opens the renamed process table). `esc`/`q` returns to the grid.

The view **renders only** — it reads `model.history` (sparklines) and `domain.Metric.PerCore`
(per-core bars) for the focused host. It never collects. Responsive behaviour reuses the
established pattern: `topview.layout(width)` picks the densest plan that fits (stories 0022/0023),
and vertical overflow uses the same `scrollWindow`/`window()` clamp with a fixed header/footer
as `detailView` (story 0022), driven by `tea.WindowSizeMsg`.

## Chrome (fixed, every tier)

- **Header** (top, fixed): `⬢ HEIMDALL · top · <host> · <os> <arch> · up <uptime>` on the left,
  `● ONLINE` host-state badge on the right. Mirrors `brand.SkinnyHeader`.
- **Footer** (bottom, fixed): keybind legend — `↑/↓ scroll · pgup/pgdn page · esc back · q quit`.
  Shortens as width drops (`↑/↓ scroll · esc back` on narrow, `↑/↓ · esc` on tiny).
- **Body** (between, scrolls): the panel region. Header + footer never scroll and the frame
  never exceeds the terminal height.

## Width tiers (per `layout(width)`)

| Tier   | Width    | Per-core            | Sparklines | Panel layout        |
|--------|----------|---------------------|------------|---------------------|
| WIDE   | ≥100     | multi-column matrix | full       | two-column grid     |
| MEDIUM | 60–99    | wraps to fewer cols | full       | single column       |
| NARROW | 40–59    | aggregate bar + N   | shortened  | single column       |
| TINY   | <40      | none                | none       | key numbers, 1/line |

## Legend (state = symbol + text, never colour alone)

```
● ONLINE   ◐ degraded   ○ offline   ⏱ stale       ← host-state badge
⣀⣤⣶⣷⣿   braille sparkline (low → high)            █ ▌  per-core fill bar
—          metric unavailable on this platform (never an error, never a fake 0)
▲ / ▼      more content above / below (scroll affordance)
```

## Unavailability (`—`) examples shown below

Per the design availability matrix, a metric the platform cannot supply renders `—`:

- WIDE/MEDIUM/NARROW POWER panel: `npu —` (NPU power residency unavailable on this host).
- TINY: `freq  —` (per-core frequency unavailable, e.g. a Windows host).
- These are the same `—` affordance `detailView` already uses for non-OK metrics.

## NPU terminology

The accelerator panel is labelled **NPU** (was ANE). The legacy `power.ane` key is
normalised to `power.npu` at ingest, so older daemons still render under the NPU label.
The wireframe never shows the string "ANE" as a live label.

## Focus / key order (no mouse path — keyboard reachable)

The body has a single scroll focus; there are no in-panel interactive fields in this view.

1. `t` (from grid) — enter top view for the focused host.
2. `↑` / `↓` — scroll body one line.
3. `pgup` / `pgdn` — scroll body one page (wide/medium only; narrow/tiny use `↑/↓`).
4. `p` — (from grid) open the renamed process table instead.
5. `esc` or `q` — exit to the fleet grid.

Reading order within the body is top-to-bottom, and within a WIDE grid row left-to-right
(CPU → MEMORY, POWER → GPU/NPU, NET&DISK → LOAD/UPTIME, then PROCESSES full width).

## Accessibility risk flags (for the a11y auditor)

- **focus-visible**: this view has one scroll focus and no per-element focus ring; the
  scroll position is shown by the `▲/▼ … scroll 12/31` affordance, not colour. Verify the
  scroll indicator is text, not colour-only.
- **color-only-signaling**: host-state badge pairs a glyph (`●`) with the word `ONLINE`;
  keep both in the mockup. Sparkline trend must not be the only carrier of a value — each
  panel shows the numeric value alongside the braille graph.
- **min-width-resize**: TINY tier (<40) must not clip; every block here is width-bounded
  (WIDE=100, MEDIUM=78, NARROW=56, TINY=36) — see the per-tier max-width note under each.
- **no-color-support**: layout reads under `NO_COLOR`; nothing here depends on colour to
  parse — borders, glyphs, and labels carry the structure.

## Open questions for product-owner / design

1. PROCESSES sort key in the top view — fixed to cpu%, or does it inherit the process
   table's current sort? (Wireframe assumes top-by-cpu.)
2. On WIDE, the second column of the third row uses a LOAD / UPTIME panel to balance the
   grid. The Gherkin lists load under CPU; confirm whether load belongs in CPU only or
   also as its own panel. (Not inventing a state — surfacing a layout choice.)

---

## State 1 — WIDE (>=100 cols): two-column panel grid

```text
╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation · macOS 14.4 arm64 · up 6d 4h                            ● ONLINE ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ┌────────────────────────────────────────────┐  ┌────────────────────────────────────────────┐   ║
║ │ CPU                                        │  │ MEMORY                                     │   ║
║ ├────────────────────────────────────────────┤  ├────────────────────────────────────────────┤   ║
║ │ util 72%    freq 3.20 GHz    load 2.41     │  │ used 61%    18.4 / 32 GB                   │   ║
║ │ util  ⣀⣤⣶⣷⣿⣿⣿⣷⣶⣤⣀⣤⣶⣷⣿⣷⣶  72%               │  │ swap  2%     0.6 / 8 GB                    │   ║
║ │ per-core (10):                             │  │ bw   ⣀⣤⣶⣷⣿⣿⣷⣶⣤⣶⣷⣿  41 GB/s                 │   ║
║ │ c0 ███▌71  c1 ██▌ 52  c2 ████ 80           │  │                                            │   ║
║ │ c3 ██  41  c4 ███ 63  c5 █▌  33            │  │                                            │   ║
║ │ c6 ███▌70  c7 ██▌ 55  c8 ████ 88           │  │                                            │   ║
║ │ c9 ██  44                                  │  │                                            │   ║
║ └────────────────────────────────────────────┘  └────────────────────────────────────────────┘   ║
║                                                                                                  ║
║ ┌────────────────────────────────────────────┐  ┌────────────────────────────────────────────┐   ║
║ │ POWER                                      │  │ GPU / NPU                                  │   ║
║ ├────────────────────────────────────────────┤  ├────────────────────────────────────────────┤   ║
║ │ pkg 22.4 W                                 │  │ gpu  util 64%   vram 47%   temp 58°C       │   ║
║ │ cpu 14.1 W   gpu 6.0 W   npu  —            │  │ npu  util 38%   (NPU label, was ANE)       │   ║
║ │ pwr  ⣀⣀⣤⣶⣷⣿⣷⣶⣤⣀⣤⣶  22 W                    │  │ vram ⣀⣤⣶⣷⣿⣷⣶⣤  7.5 / 16 GB                 │   ║
║ │ note: npu power residency unavailable → —  │  │                                            │   ║
║ └────────────────────────────────────────────┘  └────────────────────────────────────────────┘   ║
║                                                                                                  ║
║ ┌────────────────────────────────────────────┐  ┌────────────────────────────────────────────┐   ║
║ │ NET & DISK                                 │  │ LOAD / UPTIME                              │   ║
║ ├────────────────────────────────────────────┤  ├────────────────────────────────────────────┤   ║
║ │ net ↓ ⣀⣤⣶⣷⣿⣷⣶   3.20 MB/s                  │  │ load  2.41  1.98  1.55  (1/5/15m)          │   ║
║ │ net ↑ ⣀⣀⣤⣶⣷⣶    0.80 MB/s                  │  │ uptime 6d 4h    seen 14:22:08              │   ║
║ │ disk r ⣀⣤⣶⣷⣿    12.4 MB/s                  │  │                                            │   ║
║ │ disk w ⣀⣀⣤⣶      2.10 MB/s                 │  │                                            │   ║
║ └────────────────────────────────────────────┘  └────────────────────────────────────────────┘   ║
║                                                                                                  ║
║ ┌──────────────────────────────────────────────────────────────────────────────────────────────┐ ║
║ │ PROCESSES (top by cpu)                                                                       │ ║
║ ├──────────────────────────────────────────────────────────────────────────────────────────────┤ ║
║ │ PID     USER      CPU%    MEM%   COMMAND                                                     │ ║
║ │ 1234    kinncj    42.1     3.2   heimdall-dashboard                                          │ ║
║ │  880    root      11.7     1.1   heimdall-daemon                                             │ ║
║ │ 2051    kinncj     8.4     6.0   firefox                                                     │ ║
║ │ 3120    kinncj     3.2     2.4   ghostty                                                     │ ║
║ └──────────────────────────────────────────────────────────────────────────────────────────────┘ ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · pgup/pgdn page · esc back · q quit                                 host workstation ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝
```

## State 2 — MEDIUM (60–99 cols): single column, sparklines kept

```text
╔════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation                                    ● ONLINE ║
║ macOS 14.4 arm64 · up 6d 4h                                                ║
╠════════════════════════════════════════════════════════════════════════════╣
║ ┌────────────────────────────────────────────────────────────────────────┐ ║
║ │ CPU                                                                    │ ║
║ ├────────────────────────────────────────────────────────────────────────┤ ║
║ │ util 72%   freq 3.20 GHz   load 2.41                                   │ ║
║ │ util ⣀⣤⣶⣷⣿⣿⣿⣷⣶⣤⣀⣤⣶⣷⣿⣷⣶  72%                                            │ ║
║ │ per-core (10):                                                         │ ║
║ │ c0 ███▌71   c1 ██▌ 52   c2 ████ 80                                     │ ║
║ │ c3 ██  41   c4 ███ 63   c5 █▌  33                                      │ ║
║ │ c6 ███▌70   c7 ██▌ 55   c8 ████ 88                                     │ ║
║ │ c9 ██  44                                                              │ ║
║ └────────────────────────────────────────────────────────────────────────┘ ║
║                                                                            ║
║ ┌────────────────────────────────────────────────────────────────────────┐ ║
║ │ MEMORY                                                                 │ ║
║ ├────────────────────────────────────────────────────────────────────────┤ ║
║ │ used 61%  18.4/32 GB    swap 2%  0.6/8 GB                              │ ║
║ │ bw  ⣀⣤⣶⣷⣿⣿⣷⣶⣤⣶⣷⣿  41 GB/s                                              │ ║
║ └────────────────────────────────────────────────────────────────────────┘ ║
║                                                                            ║
║ ┌────────────────────────────────────────────────────────────────────────┐ ║
║ │ POWER                                                                  │ ║
║ ├────────────────────────────────────────────────────────────────────────┤ ║
║ │ pkg 22.4 W   cpu 14.1 W   gpu 6.0 W   npu —                            │ ║
║ │ pwr ⣀⣀⣤⣶⣷⣿⣷⣶⣤⣀⣤⣶  22 W                                                 │ ║
║ └────────────────────────────────────────────────────────────────────────┘ ║
║                                                                            ║
║ ┌────────────────────────────────────────────────────────────────────────┐ ║
║ │ GPU / NPU                                                              │ ║
║ ├────────────────────────────────────────────────────────────────────────┤ ║
║ │ gpu util 64%  vram 47%  temp 58°C                                      │ ║
║ │ npu util 38%  (NPU label, was ANE)                                     │ ║
║ └────────────────────────────────────────────────────────────────────────┘ ║
║                                                                            ║
║ ┌────────────────────────────────────────────────────────────────────────┐ ║
║ │ NET & DISK                                                             │ ║
║ ├────────────────────────────────────────────────────────────────────────┤ ║
║ │ net ↓ ⣀⣤⣶⣷⣿⣷  3.20 MB/s   net ↑ ⣀⣤⣶ 0.80 MB/s                          │ ║
║ │ disk r ⣀⣤⣶⣷ 12.4 MB/s   disk w ⣀⣤ 2.10 MB/s                            │ ║
║ └────────────────────────────────────────────────────────────────────────┘ ║
║                                                                            ║
║ ┌────────────────────────────────────────────────────────────────────────┐ ║
║ │ PROCESSES (top by cpu)                                                 │ ║
║ ├────────────────────────────────────────────────────────────────────────┤ ║
║ │ PID    USER     CPU%   MEM%  COMMAND                                   │ ║
║ │ 1234   kinncj   42.1    3.2  heimdall-dashboard                        │ ║
║ │  880   root     11.7    1.1  heimdall-daemon                           │ ║
║ │ 2051   kinncj    8.4    6.0  firefox                                   │ ║
║ └────────────────────────────────────────────────────────────────────────┘ ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · esc back · q quit                                             ║
╚════════════════════════════════════════════════════════════════════════════╝
```

## State 3 — NARROW (40–59 cols): per-core collapses to an aggregate bar

```text
╔══════════════════════════════════════════════════════╗
║ ⬢ top · workstation                         ● ONLINE ║
╠══════════════════════════════════════════════════════╣
║ ┌──────────────────────────────────────────────────┐ ║
║ │ CPU                                              │ ║
║ ├──────────────────────────────────────────────────┤ ║
║ │ util 72%   freq 3.20 GHz                         │ ║
║ │ load 2.41  (1/5/15m 2.41 1.98 1.55)              │ ║
║ │ util ⣀⣤⣶⣷⣿⣷⣶⣤⣶⣷⣿  72%                            │ ║
║ │ cores ████████▌░░  10 cores  avg 58  max 88      │ ║
║ └──────────────────────────────────────────────────┘ ║
║                                                      ║
║ ┌──────────────────────────────────────────────────┐ ║
║ │ MEMORY                                           │ ║
║ ├──────────────────────────────────────────────────┤ ║
║ │ used 61%  18.4/32 GB   swap 2%                   │ ║
║ │ bw ⣀⣤⣶⣷⣿⣷⣶  41 GB/s                              │ ║
║ └──────────────────────────────────────────────────┘ ║
║                                                      ║
║ ┌──────────────────────────────────────────────────┐ ║
║ │ POWER                                            │ ║
║ ├──────────────────────────────────────────────────┤ ║
║ │ pkg 22 W  cpu 14 W  gpu 6 W  npu —               │ ║
║ │ pwr ⣀⣤⣶⣷⣿⣷⣶  22 W                                │ ║
║ └──────────────────────────────────────────────────┘ ║
║                                                      ║
║ ┌──────────────────────────────────────────────────┐ ║
║ │ GPU / NPU                                        │ ║
║ ├──────────────────────────────────────────────────┤ ║
║ │ gpu 64%  vram 47%  temp 58°C                     │ ║
║ │ npu 38%  (NPU)                                   │ ║
║ └──────────────────────────────────────────────────┘ ║
║                                                      ║
║ ┌──────────────────────────────────────────────────┐ ║
║ │ NET & DISK                                       │ ║
║ ├──────────────────────────────────────────────────┤ ║
║ │ net ↓ ⣀⣤⣶ 3.2   ↑ ⣀⣤ 0.8 MB/s                    │ ║
║ │ disk r ⣀⣤⣶ 12.4  w ⣀ 2.1 MB/s                    │ ║
║ └──────────────────────────────────────────────────┘ ║
║                                                      ║
║ ┌──────────────────────────────────────────────────┐ ║
║ │ PROCESSES                                        │ ║
║ ├──────────────────────────────────────────────────┤ ║
║ │ PID    CPU%  COMMAND                             │ ║
║ │ 1234   42.1  heimdall-dashboard                  │ ║
║ │  880   11.7  heimdall-daemon                     │ ║
║ └──────────────────────────────────────────────────┘ ║
║                                                      ║
╠══════════════════════════════════════════════════════╣
║ ↑/↓ scroll · esc back                                ║
╚══════════════════════════════════════════════════════╝
```

## State 4 — TINY (<40 cols, iPhone portrait in Termius): key numbers only

```text
╔══════════════════════════════════╗
║ ⬢ top · ws                  ● ON ║
╠══════════════════════════════════╣
║ cpu   72%                        ║
║ mem   61%                        ║
║ swap   2%                        ║
║ pwr   22 W                       ║
║ temp  58°C                       ║
║ gpu   64%                        ║
║ npu   38%                        ║
║ load  2.41                       ║
║ freq  —                          ║
╠══════════════════════════════════╣
║ ↑/↓ · esc                        ║
╚══════════════════════════════════╝
```

## State 5 — SCROLL: body taller than terminal, header & footer fixed

```text
╔══════════════════════════════════════════════════════════════════════════════════════════════════╗
║ ⬢ HEIMDALL · top · workstation · macOS 14.4 arm64 · up 6d 4h                            ● ONLINE ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ▲ more above (POWER · GPU/NPU panels scrolled off)                                  scroll 12/31 ║
║ ┌────────────────────────────────────────────┐                                                   ║
║ │ NET & DISK                                 │                                                   ║
║ ├────────────────────────────────────────────┤                                                   ║
║ │ net ↓ ⣀⣤⣶⣷⣿⣷⣶   3.20 MB/s                  │                                                   ║
║ │ net ↑ ⣀⣀⣤⣶⣷⣶    0.80 MB/s                  │                                                   ║
║ │ disk r ⣀⣤⣶⣷⣿    12.4 MB/s                  │                                                   ║
║ │ disk w ⣀⣀⣤⣶      2.10 MB/s                 │                                                   ║
║ └────────────────────────────────────────────┘                                                   ║
║                                                                                                  ║
║ ┌──────────────────────────────────────────────────────────────────────────────────────────────┐ ║
║ │ PROCESSES (top by cpu)                                                                       │ ║
║ ├──────────────────────────────────────────────────────────────────────────────────────────────┤ ║
║ │ PID     USER      CPU%    MEM%   COMMAND                                                     │ ║
║ │ 1234    kinncj    42.1     3.2   heimdall-dashboard                                          │ ║
║ │  880    root      11.7     1.1   heimdall-daemon                                             │ ║
║ │ 2051    kinncj     8.4     6.0   firefox                                                     │ ║
║ │ 3120    kinncj     3.2     2.4   ghostty                                                     │ ║
║ └──────────────────────────────────────────────────────────────────────────────────────────────┘ ║
║ ▼ more below                                                                        scroll 12/31 ║
╠══════════════════════════════════════════════════════════════════════════════════════════════════╣
║ ↑/↓ scroll · pgup/pgdn page · esc back · q quit                                                  ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝
```

