# v2.2.0 — Full-screen "top" view (mactop/btop-style)

Status: approved design (2026-06-30)
Target release: v2.2.0 (next minor)
Design target: tui

## Context

Today the `t` key opens `modalTop` — a process table sorted by cpu/mem. That is
really a *process list*, not a system "top". v2.2 reassigns keys and adds a real
mactop/btop-style full-screen system view:

- `p` → **processes** (the existing process table, renamed).
- `t` → **top** (new full-screen system dashboard).

The view is built mostly from data Heimdall already collects. `model.history`
(`map[HostID]map[string][]float64`) holds per-host, per-metric time series;
`domain.Metric.PerCore []float64` holds per-core values. The new view *reads*
these — it does not collect.

## Goals

- A full-screen, single-host "top" view entered with `t`, exited with `esc`/`q`.
- Braille/sparkline time-series graphs for CPU, memory, power, network, disk.
- Per-core CPU bars; power breakdown incl. a generic **NPU** (was ANE).
- New cross-platform metrics: load average, swap, per-core frequency, memory
  bandwidth, NPU utilisation — each degrading gracefully where unavailable.
- Generic terminology: ANE → NPU across UI and metric keys.

## Non-goals

- No new binary; the view lives inside `heimdall-dashboard`.
- No fleet aggregation in the top view — it shows the single focused host.
- No persistence/history beyond the existing in-memory ring buffers.

## Hard constraints (carried from prior decisions)

1. **Backward compatible.** A v2.2 dashboard must keep working against older
   daemons/hubs in a mixed fleet (e.g. the Mac mini), and vice versa. The wire
   stays additive: new metric keys are optional; `power.ane` is preserved as a
   read alias for `power.npu`.
2. **Cross-platform: macOS, Linux, Windows.** Every new collector returns
   `StatusUnavailable` (not an error, not a fake 0) where the platform can't
   supply it. The UI renders `—` for unavailable metrics.
3. **Screen-size aware.** Must follow the existing responsive pattern, not a new
   one: `layout(width)` picks the densest plan that fits, dropping/compacting as
   it narrows; `scrollWindow`/`window()` clamp to height with fixed header/footer
   and vertical scroll; driven by `tea.WindowSizeMsg`. Must remain usable in
   **Termius on iPhone/iPad** (very narrow widths, ~35–50 cols).

## Architecture

New package `app/internal/tui/topview/` — a Bubble Tea sub-model the dashboard
switches into on `t`.

- Dashboard owns metric streaming + history. On `t` it constructs
  `topview.Model` from the focused host's latest snapshot and its history slice,
  and delegates `Update`/`View` while active. `esc`/`q` returns to the fleet grid.
- `topview` is pure render + local key state (scroll, panel focus). It takes data
  in; it never reaches into adapters. This keeps it independent and testable, and
  off the already-large `modal.go`.

### Responsive layout (the core of the screen-size constraint)

`topview` defines `layout(width, height)` returning a plan, mirroring the grid's
approach:

- **Wide (≥100 cols):** two-column panel grid, full sparklines, per-core bars in
  a multi-column matrix.
- **Medium (~60–99):** single column of stacked panels, sparklines kept, per-core
  bars wrap to fewer columns.
- **Narrow (~40–59, iPad split / landscape phone):** stacked panels, sparklines
  shortened, per-core collapses to an aggregate bar + "N cores avg/max".
- **Tiny (<40, iPhone portrait):** key numbers only (cpu%, mem%, power, temp),
  no graphs, no per-core; one value per line.

Vertical overflow uses the existing `scrollWindow`/`window()` with a fixed header
and footer, exactly like `detailView`. No panel assumes a minimum height; each
renders what fits and the whole body scrolls.

## Units

Each unit is independently testable and lands behind the responsive UI.

### 1. Keybind swap
- `t` → enter topview; `p` → open the renamed process table.
- Rename `modalTop`→ processes wording in `modal.go`, `detail.go` footer,
  `help.go`; update docs/manpages.
- Tests: model press tests (`press(m,"p")`, `press(m,"t")`) like `modal_test.go`.

### 2. NPU rename (bc-safe)
- Canonical metric key `power.npu`; `npu.util` for utilisation.
- `power.ane` accepted as a read alias, normalised to `power.npu` at ingest in
  one place (dashboard metric intake), so old daemons keep rendering.
- Collectors (darwin `collect.go`) emit `power.npu`; `adapters` Describe lists it;
  `fake/fleet.go` and UI labels say "NPU".
- Tests: alias normalisation table test; Describe contains `power.npu`.

### 3. New collectors (graceful-degrade)
Availability matrix (Unavailable = renders `—`, never an error):

| Metric        | macOS                         | Linux                  | Windows         |
|---------------|-------------------------------|------------------------|-----------------|
| `cpu.load`    | sysctl loadavg                | `/proc/loadavg`        | Unavailable     |
| `mem.swap`    | sysctl/vm_stat                | `/proc/meminfo`        | WMI/perf        |
| `cpu.freq`    | best-effort (IOReport/pmetr.) | `/sys` cpufreq         | WMI / Unavail.  |
| `mem.bw`      | IOReport AMC Stats            | Unavailable            | Unavailable     |
| `npu.util`    | IOReport ANE residency        | Unavailable            | Unavailable     |

- Each is a small adapter/helper function with a parser unit-tested on captured
  sample text (the SMC-fix pattern); Linux/Windows verified by parser tests +
  cross-compile, not runtime, on the dev machine.

### 4. `render/sparkline.go`
- Braille line graph over `[]float64` → string of configurable width/height, with
  min/max/current annotations; handles empty/short series and a target width.
- Golden unit tests for known inputs (incl. width clamping for narrow screens).

### 5. `topview` UI
- Panels: CPU (per-core + util graph + freq), Memory (used/swap + bw graph),
  Power (pkg/cpu/gpu/npu + graph), GPU/NPU (util/vram/temp), Net & Disk (graphs),
  Processes (reuse existing table). Footer: `↑/↓ scroll  esc back`.
- Tests: layout plan selection per width; render snapshot at wide/medium/narrow/
  tiny widths; unavailable metrics render `—`.

## Data flow

```
adapters/helper → transport → dashboard (model.metrics + model.history)
                                   │  press "t"
                                   ▼
                         topview.Model(snapshot, history)  ── reads only
```

New metrics ride the same additive pipeline; no transport/proto breaking change.

## Sequencing

1. Keybind swap + NPU rename (alias) — small, lands first, fully bc.
2. `render/sparkline.go` + `topview` UI over **existing** metrics — working view.
3. New collectors one at a time, each degrading gracefully — UI improves as they land.

## Testing & verification

- TDD throughout (`tdd-workflow`): parser tests, sparkline golden tests, alias
  normalisation, keybind press tests, layout/snapshot tests.
- `make test-all` green; `go vet`; `gofmt`.
- Cross-builds: linux + windows (CGO-free) and darwin (cgo + CGO-free).
- macOS verified at runtime locally; Linux/Windows collectors verified via parser
  tests + cross-compile (stated explicitly where runtime is unverified).
- Branch `feat/top-view-2.2`; humanized commits; no `Co-Authored-By`; stop before
  release for human go-ahead.

## Risks

- `npu.util` / `mem.bw` may be unavailable even on macOS Pro/Max (same IOReport
  gaps as the power fix) — acceptable; renders `—`.
- Tiny-width (iPhone portrait) legibility: mitigated by the "key numbers only"
  tier and snapshot tests at <40 cols.
- Per-core freq on Windows likely Unavailable — acceptable per the matrix.
