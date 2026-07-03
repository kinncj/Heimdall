# Top View (Hliðskjálf) — One Host, mactop-Style

Press **`t`** on a focused host to open the full-screen system view: a
mactop/btop-style read-out of one machine — per-core CPU, memory, power,
GPU/NPU, network, disk, and the top processes. Press **`esc`** to go back to the
fleet, **`q`** to quit.

> Named **Hliðskjálf** — Odin's high seat, from which he sees into all realms.
> Its focus and scrolling are **Himinbjörg**, the one focus+scroll primitive
> shared across the whole TUI.

## Keys

| Key | Action |
|---|---|
| `t` | open the top view for the focused host |
| `p` | open the process table (this used to be `t`) |
| `↑/↓` · `k/j` | scroll — the **whole view** by default, or the focused panel |
| `tab` / `shift+tab` | move the focus ring to the next / previous panel; wraps back to the whole view |
| `pgup/pgdn` | page |
| `home/end` | jump to the top / bottom |
| mouse wheel | scroll the panel under the pointer (or the whole view) |
| mouse click | focus the panel under the pointer |
| `esc` | back to the fleet/detail |
| `q` · `ctrl+c` | quit |

## Focus and scrolling (Himinbjörg)

The view is a grid of panels, and it can be taller than your terminal — especially
over SSH on a phone. So it scrolls two ways, and you never lose your place:

- **The whole view is focused first.** `↑/↓` and the mouse wheel scroll the entire
  screen, so it is always scrollable at any resolution — nothing is stranded below
  the fold.
- **`tab` steps the focus ring into each panel** (CPU → MEMORY → POWER → …), then
  wraps back around to the whole-view stop. The focused panel is drawn with a
  **heavier border** (not just a colour, so it survives `NO_COLOR`), and a focused
  panel scrolls its own overflow while the page keeps its active row on screen.
- **The mouse targets the panel under the pointer** — wheel scrolls it, click focuses
  it. (Before Himinbjörg the wheel did nothing here.)

The `▲ more above` / `▼ more below` markers with a `y/total` (or `page y/total`)
count tell you where you are.

## What it shows

- **CPU** — utilisation, clock, load average, a trend sparkline, and per-core bars
  (coloured by load). On hybrid silicon (Apple P/E, Intel P/E) the bars are
  grouped under `P-cores (N):` / `E-cores (N):` headers with their real core
  ids; uniform CPUs and older daemons keep the plain `per-core (N):` grid.
- **MEMORY** — used and swap gauges, a usage trend, and memory bandwidth.
- **POWER** — total / CPU / GPU / NPU watts and a power trend (headlined by
  `power.total`). On Apple Pro/Max a `cpu clusters: P … W · E … W` line
  decomposes the CPU figure into its SMC cluster rails.
- **GPU / NPU** — utilisation, VRAM, temperature, NPU utilisation.
- **NET & DISK** — rx/tx and read/write with trends.
- **PROCESSES** — top by CPU; the list grows to fill the screen.

The trends reuse the dashboard's in-memory history — the view only renders, it
never collects.

## It fits any screen

The top view (and the host detail view) reflow to the terminal width, so they
stay readable over SSH from **Termius on a phone or tablet**:

| Width | Layout |
|---|---|
| ≥ 100 cols | two-column panel grid, full sparklines, per-core matrix |
| 60–99 | single column, sparklines kept |
| 40–59 | single column, per-core collapses to an aggregate bar |
| < 40 | key numbers only (cpu, mem, power, temp…), one per line |

## Cross-platform & graceful degradation

A metric the host can't supply renders **`—`** — never a fabricated `0`. So the
panels look different per machine, and that's expected:

- **load** — macOS/Linux; **`—`** on Windows (no OS load average).
- **swap** — everywhere.
- **cpu.freq** — `/sys` cpufreq on Linux (real per-core clock); **`—`** where no
  clock is exposed (e.g. Apple Silicon).
- **GPU** — NVIDIA via `nvidia-smi`; AMD via `amd-smi` / amdgpu sysfs on Linux.
  AMD on Windows isn't wired yet, so that panel is sparse there.
- **power** — `power.cpu` needs RAPL (x86 Linux), SMC (macOS), or a running
  Scaphandre (Windows); ARM/GB10 has no CPU power sensor and shows **`—`**.
- **npu.util / mem.bw** — not collected yet; they render **`—`** today.

See the [Metrics Reference](../metrics.md) for every metric and its units.
