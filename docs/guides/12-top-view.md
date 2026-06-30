# Top View (Hliðskjálf) — One Host, mactop-Style

Press **`t`** on a focused host to open the full-screen system view: a
mactop/btop-style read-out of one machine — per-core CPU, memory, power,
GPU/NPU, network, disk, and the top processes. Press **`esc`** to go back to the
fleet, **`q`** to quit.

> Named **Hliðskjálf** — Odin's high seat, from which he sees into all realms.

## Keys

| Key | Action |
|---|---|
| `t` | open the top view for the focused host |
| `p` | open the process table (this used to be `t`) |
| `↑/↓` · `k/j` | scroll the body |
| `pgup/pgdn` | page the body |
| `esc` | back to the fleet/detail |
| `q` · `ctrl+c` | quit |

## What it shows

- **CPU** — utilisation, clock, load average, a trend sparkline, and per-core bars
  (coloured by load).
- **MEMORY** — used and swap gauges, a usage trend, and memory bandwidth.
- **POWER** — package / CPU / GPU / NPU watts and a power trend.
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
- **GPU** — via `nvidia-smi`; AMD GPUs aren't read yet, so that panel is sparse
  on non-NVIDIA hosts.
- **power** — package/CPU power needs RAPL (x86 Linux) or SMC (macOS); ARM and
  Windows show **`—`**.
- **npu.util / mem.bw** — not collected yet; they render **`—`** today.

See the [Metrics Reference](../metrics.md) for every metric and its units.
