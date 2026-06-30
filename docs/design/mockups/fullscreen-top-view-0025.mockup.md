---
id: fullscreen-top-view-0025
story: docs/stories/fullscreen-top-view-20260630152849-0025/story.md
wireframe: docs/design/wireframes/fullscreen-top-view-0025.wireframe.md
tokens: docs/design/identity/tokens.json
theme: docs/design/identity/terminal-theme.json
target: tui
mode: dark
status: approved
approved_by: null
approved_at: null
created_at: 2026-06-30
---

# Mockup — fullscreen-top-view-0025 (Hliðskjálf)

High-fidelity terminal render of the full-screen single-host `top` view, brand applied
(dark mode; light mirrors via `terminal-theme.json` → `modes.light`). All colour comes
from `tokens.json` / `terminal-theme.json` — no raw hex outside the token set.

The render below reproduces the approved wireframe geometry exactly (same glyphs, same
per-tier width budget). High fidelity here means: every region is bound to a concrete
lipgloss role with its dark-mode hex/ansi value (Styles section), per-core bars and the
fill/spark glyphs are coloured by the severity ramp keyed on the metric value, and the
NPU label / `—` unavailable affordance are pinned.

Markdown cannot carry ANSI inside a code fence, so the renders stay monochrome box-art and
the **Styles** section is the authoritative colour map. A region marked `⟦role⟧` in the
callouts maps 1:1 to a row in the Styles table.

Frame uses the app double-border (`border.app`, lipgloss `DoubleBorder`). Panels use the
single border (`border.panel`, lipgloss `NormalBorder`). Colour only ever reinforces
meaning that is also carried by glyph + label + numeric value → NO_COLOR-safe.

Per-tier width budget (total frame width incl. the `║…║` double border):

| Tier   | Width | Per-core            | Sparklines | Panel layout    |
|--------|-------|---------------------|------------|-----------------|
| WIDE   | 100   | multi-column matrix | full       | two-column grid |
| MEDIUM | 78    | wraps to fewer cols | full       | single column   |
| NARROW | 56    | aggregate bar + N   | shortened  | single column   |
| TINY   | 36    | none                | none       | 1 value / line  |

---

## State 1 — WIDE (100 cols): two-column panel grid

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

Region callouts (WIDE):

- `⬢ HEIMDALL` ⟦type.title⟧ steel + bold; the `⬢` sigil tints `structure.accent` electric-blue.
- `· top · workstation · macOS 14.4 arm64 · up 6d 4h` ⟦type.muted⟧; the `·` separators ⟦structure.border⟧.
- `● ONLINE` ⟦status.online⟧ glyph `●` + word, severity-independent host-state badge (right-aligned).
- Panel titles `CPU` / `MEMORY` / `POWER` / `GPU / NPU` / `NET & DISK` / `LOAD / UPTIME` / `PROCESSES` ⟦type.heading⟧ bold.
- Metric labels `util` `freq` `load` `used` `swap` `bw` `pkg` `cpu` `gpu` `npu` `vram` `temp` `net ↓` `disk r` `uptime` `seen` ⟦type.label⟧ faint.
- Numeric values `72%` `3.20 GHz` `18.4 / 32 GB` `22 W` `58°C` `2.41` ⟦type.value⟧ bold; unit suffixes (`%` `GHz` `GB` `W` `°C` `MB/s`) ⟦type.unit⟧.
- Braille sparklines `⣀⣤⣶⣷⣿…` ⟦render.Sparkline → structure.text_secondary⟧, tinted by the latest sample's severity tier (see Severity ramp).
- Per-core bars `███▌` ⟦render.Gauge⟧: filled `█`/`▌` in the severity tier for that core's %, track in `structure.text_muted`; trailing number ⟦type.value⟧.
- `npu  —` and the `note: … → —` line ⟦status.unavailable⟧ glyph `—` + faint — never an error, never a fake 0.
- Footer keybinds `↑/↓` `pgup/pgdn` `esc` `q` ⟦type.keybinding⟧ underline+bold; surrounding text ⟦type.muted⟧; `host workstation` right tag ⟦type.muted⟧.
- App frame ⟦border.app⟧ DoubleBorder in `structure.border`. Panel frames ⟦border.panel⟧ NormalBorder in `structure.border`.

---

## State 2 — MEDIUM (78 cols): single column, sparklines kept

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

MEDIUM deltas vs WIDE (same roles, denser plan):

- Header wraps to two lines: line 1 `⬢ HEIMDALL · top · workstation` + right `● ONLINE`; line 2 `macOS 14.4 arm64 · up 6d 4h` ⟦type.muted⟧. Same fixed header treatment.
- Panels stack single column; per-core matrix wraps to 3 columns; sparklines kept full.
- Footer drops `pgup/pgdn` and the `host` tag (`↑/↓ scroll · esc back · q quit`); roles unchanged.

---

## State 3 — NARROW (56 cols): per-core collapses to an aggregate bar

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

NARROW deltas:

- Header shortens to `⬢ top · workstation` (no `HEIMDALL` wordmark, no os/arch line); `⬢` keeps `structure.accent`, badge `● ONLINE` ⟦status.online⟧ stays.
- Per-core matrix collapses to one aggregate `cores ████████▌░░` ⟦render.Gauge⟧ filled to `avg` and tinted by `avg` (58 → moderate), track `░` ⟦structure.text_muted⟧, with `10 cores  avg 58  max 88` summary ⟦type.label⟧ + ⟦type.value⟧.
- Sparklines shortened (fewer samples); roles unchanged.
- PROCESSES drops USER + MEM% columns. Footer is `↑/↓ scroll · esc back`.

---

## State 4 — TINY (36 cols): key numbers only

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

TINY callouts:

- Header `⬢ top · ws` ⟦type.muted⟧ + `⬢` ⟦structure.accent⟧; badge abbreviates to `● ON` ⟦status.online⟧ — glyph + short word preserved (still not colour-only).
- One metric per line: label `cpu` `mem` `swap` `pwr` `temp` `gpu` `npu` `load` `freq` ⟦type.label⟧; value ⟦type.value⟧ + unit ⟦type.unit⟧.
- No sparklines, no per-core bars.
- `freq  —` ⟦status.unavailable⟧: per-core frequency unsupplied on this host (e.g. a Windows host) → `—`, faint, never `0`.
- Footer `↑/↓ · esc` ⟦type.keybinding⟧ — page keys are not bound at this tier.

---

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

SCROLL callouts:

- `▲ more above …` / `▼ more below` scroll affordances ⟦type.caption⟧ (italic→faint fallback): text glyph, not colour — the off-screen panel names are spelled out.
- `scroll 12/31` position indicator ⟦type.value⟧ position + ⟦type.muted⟧ total — a number, never colour-only (a11y `focus-visible` risk flag).
- Header (`╠…╣` rule line 2) and footer stay fixed; only the body region between them scrolls (reuses `window()`/`scrollWindow` clamp). Frame height never exceeds the terminal.

---

## Host-state badge — variants (top-right chrome)

The badge is the only host-state carrier in the chrome; it pairs glyph + word and never relies
on colour. Same render at every tier (word abbreviates to `ON` at TINY).

```text
● ONLINE      degraded → ◐ DEGRADED      offline → ○ OFFLINE      stale → ⏱ STALE
```

| Badge | Status token | Foreground (dark) | Glyph | Notes |
|---|---|---|---|---|
| `● ONLINE` | `status.online` | #5fd75f (77) | ● | default in every render above |
| `◐ DEGRADED` | `status.degraded` | #ffaf00 (214) | ◐ | host reporting but unhealthy |
| `○ OFFLINE` | `status.offline` | #a8a8a8 (248) | ○ | no recent sample |
| `⏱ STALE` | `status.stale` | #d7af87 (180) faint | ⏱ | last sample aged out |

---

## Styles (lipgloss)

Each region maps to a role in `terminal-theme.json` (`modes.dark`). Hex/ansi are the concrete
dark-mode values; light mirrors via `modes.light`. No colour drift from `tokens.json` / `palette.json`.

| Region | Theme role | Foreground | Attrs | Border / BG |
|---|---|---|---|---|
| App frame | `border.app` / `structure.border` | #767676 (243) | — | DoubleBorder ╔═╗║╚╝ |
| Panel frame | `border.panel` / `structure.border` | #767676 (243) | — | NormalBorder ┌─┐│└┘, panel bg #1c1c1c (234) |
| Wordmark `⬢ HEIMDALL` | `type.title` (`structure.title`) | #d0d0d0 steel (252) | bold | leading `⬢` |
| `⬢` sigil tint | `structure.accent` | #00d7ff (45) | — | brand eye / accent |
| Header trail `· top · host · os arch · up …` | `type.muted` (`structure.text_muted`) | #949494 (246) | — | `·` separators in `structure.border` |
| Host badge `● ONLINE` | `status.online` | #5fd75f (77) | — | glyph + word, right-aligned |
| Panel title `CPU` … `PROCESSES` | `type.heading` (`structure.heading`) | #eeeeee (255) | bold | — |
| Metric label `util`/`freq`/`bw`/`pkg`/`npu`/… | `type.label` (`structure.label`) | #949494 (246) | faint | — |
| Metric value `72%`/`3.20`/`22`/`58`/… | `type.value` (`structure.value`) | #eeeeee (255) | bold | — |
| Unit suffix `%`/`GHz`/`GB`/`W`/`°C`/`MB/s` | `type.unit` (`structure.unit`) | #949494 (246) | — | — |
| Sparkline `⣀⣤⣶⣷⣿…` | `render.Sparkline` → `structure.text_secondary` | #c6c6c6 (251) | — | latest-sample severity tint (see ramp) |
| Per-core / aggregate bar fill `█▌` | `render.Gauge` → `color.severity.*` | per value (ramp) | — | filled cells coloured by % |
| Bar track `░` | `structure.text_muted` | #949494 (246) | — | unfilled gauge remainder |
| Unavailable `—` (`npu —`, `freq —`, note) | `status.unavailable` | #a8a8a8 (248) | faint | glyph `—` + word "unavailable" semantics |
| PROCESSES header row `PID USER CPU% …` | `type.label` (`structure.label`) | #949494 (246) | faint | column headers |
| PROCESSES data rows | `type.body` (`structure.body`) | #eeeeee (255) | — | CPU%/MEM% values stay `type.value` bold |
| Scroll affordance `▲ more above` / `▼ more below` | `type.caption` (`structure.caption`) | #949494 (246) | italic→faint | text, never colour-only |
| Scroll position `scroll 12/31` | `type.value` + `type.muted` | #eeeeee / #949494 | bold / — | numeric indicator |
| Footer keybinds `↑/↓`/`pgup`/`esc`/`q` | `type.keybinding` (`structure.keybinding`) | #eeeeee (255) | underline+bold | surrounding text `type.muted` |

### Severity ramp — per-core bars, fill gauges, sparkline tint

`render.Gauge` colours filled cells by `Mode.SeverityFor(pct)`; the sparkline is tinted by its
latest sample's tier. Five tiers grouped into the ok / warn / critical reading the a11y auditor checks.

| Band | Tier | Foreground (dark) | Group | Example in render |
|---|---|---|---|---|
| 0–39% | `severity.1` nominal | #5fd7af (79) | ok | `c5 █▌ 33`, `npu 38` |
| 40–59% | `severity.2` moderate | #87d75f (113) | ok | `c1 ██▌ 52`, `c7 55`, aggregate `avg 58` |
| 60–74% | `severity.3` elevated | #ffd75f (221) | warn | `c0 ███▌71`, `c6 70`, `gpu util 64`, `util 72` |
| 75–89% | `severity.4` high | #ff875f (209) | warn | `c2 ████ 80`, `c8 ████ 88` (`max 88`) |
| 90–100% | `severity.5` critical | #ff5f5f (203) | critical | none in sample data; e.g. a core at 95 |

Light-mode ramp (mirror): nominal #006d00 (22) · moderate #5a7400 (64) · elevated #806600 (94) ·
high #9c5400 (130) · critical #c81414 (160).

### Glyph sets (pinned)

| Use | Glyphs | Source |
|---|---|---|
| Braille sparkline (low → high) | `⣀ ⣤ ⣶ ⣷ ⣿` | wireframe legend — design glyph set for this view |
| Gauge fill / track | `█ ▌` filled · `░` track | `decoration.gauge_blocks` (`█▓▒░`); `▌` = half-cell partial |
| Trend arrows | `↑ → ↓` | `decoration.trend` (rising/steady/falling); net uses `↓`/`↑` |
| Brand sigil | `⬢` | `structure.title.glyph` |
| Host-state | `● ◐ ○ ⏱` | `status.{online,degraded,offline,stale}.glyph` |
| Unavailable | `—` | `status.unavailable.glyph` |
| Scroll | `▲ ▼` | scroll affordance (more above / below) |

## NPU terminology

The accelerator panel is labelled **NPU** in every tier (`GPU / NPU`, `npu util 38%`,
`npu —`). The string "ANE" never appears as a live label — the parenthetical
`(NPU label, was ANE)` is mockup annotation, not rendered chrome. Backward compatibility
(`power.ane` → `power.npu`) is handled at ingest; the renderer only ever sees `power.npu`.

## Unavailable `—` affordance

A metric the platform cannot supply renders `—` ⟦status.unavailable⟧ — faint `#a8a8a8`, glyph
only, no fabricated `0`, no error styling. Shown here as `npu —` (WIDE/MEDIUM/NARROW POWER) and
`freq —` (TINY). This is the identical affordance `detailView` uses for non-OK metrics.

## Accessibility

- **focus-visible**: single scroll focus, no per-element ring. Position is the text indicator
  `scroll 12/31` + `▲/▼` glyphs — not colour. (Wireframe a11y flag honoured.)
- **color-only-signaling**: host-state badge is glyph `●` + word `ONLINE`; every sparkline and
  bar is accompanied by its numeric value; severity is reinforcement, never the sole carrier.
- **min-width-resize**: every tier is width-bounded (WIDE 100 / MEDIUM 78 / NARROW 56 / TINY 36).
  No line exceeds its tier width incl. the double border. Braille and box glyphs are single-cell.
- **no-color-support**: under `NO_COLOR`/`TERM=dumb`, colours collapse to terminal default fg/bg;
  meaning survives via glyph + label + value + gauge block density (`█▓▒░`) per `mono_fallback`.

## Implementation notes — stubbed / deferred / decisions

- **Render-only**: no business logic in scope. The view reads `model.history` (sparklines) and
  `domain.Metric.PerCore` (bars) for the focused host; it never collects. Layout selection
  (`topview.layout(width,height)`) and vertical clamp (`window()`/`scrollWindow`) reuse the
  existing 0022/0023 mechanism — not reimplemented here.
- **Sparkline glyph set**: `render.Sparkline` today emits the block ramp `▁▂▃▄▅▆▇█` in
  `structure.text_secondary`. This view's design glyph set is the **braille** ramp `⣀⣤⣶⣷⣿`.
  Implementation must extend the sparkline helper with a braille ramp variant (or a glyph-set
  param) for this view. Flagged for the implementing engineer — the colour role is unchanged.
- **Sparkline severity tint**: the live helper is monochrome `text_secondary`. This mockup tints
  the line by its latest sample's severity tier for at-a-glance trend health. **Decision needed**:
  keep monochrome (matches today's helper, value carries severity) vs. adopt the tint. Defaulted
  to tint in the Styles map; revert to plain `text_secondary` if product/design prefers parity.
- **Open layout questions carried from the wireframe** (for product-owner): (1) PROCESSES sort
  key — fixed `cpu%` or inherits the process table's sort; mockup assumes top-by-cpu. (2) the
  WIDE third-row `LOAD / UPTIME` panel balances the grid; confirm load lives in CPU only or also
  as its own panel.

## Approval

- [ ] Matches approved wireframe geometry (WIDE/MEDIUM/NARROW/TINY/SCROLL)
- [ ] Design tokens applied correctly (roles + severity ramp from `terminal-theme.json`)
- [ ] All tiers + states present (width tiers, scroll, unavailable `—`, host-state badge variants)
- [ ] NPU labelling correct; no live "ANE" string
- [ ] Sparkline glyph-set + tint decisions resolved
- [ ] Approved by product owner / UX lead
