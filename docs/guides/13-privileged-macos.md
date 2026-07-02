# Privileged Metrics — macOS / Apple Silicon

Power, GPU, and thermal on a Mac. For the shared helper + launchd deployment
mechanics, see [Privileged Metrics (overview)](04-privileged-metrics.md); this
guide is the macOS specifics.

## What you get, and from where

| Metric | Source | Needs root? |
|---|---|---|
| `gpu.util`, `power.gpu` | **IOReport** energy counters | no |
| `power.total` (whole-system) | **SMC `PSTR`** ("System Total Power") | no |
| `power.cpu`, `power.npu` | IOReport per-domain, where the SoC exposes it | no |
| full thermal, gaps IOReport misses | `powermetrics` (via the helper) | yes |

Power standardizes on `power.cpu` / `power.gpu` / `power.total` across the fleet.
On macOS `power.total` is the SMC whole-system figure (it already includes the
GPU, so it is **not** a sum of the rails).

Apple Silicon is the one platform where GPU power **and** whole-system power are
available with **no sudo** — IOReport and SMC are both unprivileged. A normal
daemon reads them in-process:

```sh
./bin/heimdall-daemon --once | grep -E "gpu|power"
# e.g. power.gpu=1.19W  power.total=18.4W  gpu.util=54%
```

## Source precedence for `power.total`

`power.total` is read in a strict order so a phantom reading never shadows a real
one:

1. **SMC `PSTR`** — the authoritative whole-system rail (what mactop/btop use).
2. **`powermetrics`** — fills the package figure when SMC is absent.
3. **IOReport per-domain sum** — last resort only.

This ordering exists because of the Pro/Max quirk below.

> **M-series Pro/Max 0 W note**: on Apple Silicon **Pro/Max** SoCs the IOReport
> "Energy Model" per-domain CPU/ANE channels read **0**, and only GPU shows a
> sub-watt figure — so the energy-sum collapses to ~0 W. Reading SMC `PSTR`
> first is what keeps an M3 Max from reporting 0 W. The variable is the **SoC,
> not the Heimdall version**. See [ADR 0020](../architecture/0020-hlidskjalf-top-view-and-npu-rename.md).

> **CPU power on Pro/Max**: IOReport and `powermetrics` report `0` for CPU on
> Pro/Max/Ultra, but the raw SMC per-cluster keys still carry it — Heimdall reads
> those (see the next section). `power.cpu` only stays `unavailable` on a chip
> whose SMC keys aren't mapped yet.

## How CPU/GPU power is read on Apple Silicon (base vs Pro/Max)

Heimdall reads Apple power from two unprivileged sources and layers them:

1. **IOReport** "Energy Model" — per-domain CPU/GPU/ANE. On **base** dies (M1–M4)
   this populates `power.cpu`. On **Pro/Max/Ultra** the CPU and ANE channels read
   **0** (both IOReport *and* `powermetrics` return `CPU Power: 0 mW` with the
   cores pegged) — so IOReport alone can't give CPU power there.
2. **Raw SMC keys** — the same source Stats reads. Pro/Max dies still expose CPU
   power under the per-cluster power keys (`PC02` + `PC42` = the two 6-core
   P-clusters on an M3 Max), even though the energy model reports 0. When IOReport
   gives no CPU figure, Heimdall sums those SMC cluster keys and reports
   `power.cpu` tagged `P-cores (SMC)`.

So `power.cpu` is **real on both tiers** now — from IOReport on base dies, from
the SMC cluster keys on Pro/Max. It only falls back to
`no per-domain CPU power (IOReport + SMC)` when *neither* source responds (a chip
whose SMC cluster keys we haven't mapped — the keys are reverse-engineered and can
differ per generation, exactly like the temperature keys). `power.total` (SMC
`PSTR`) always carries the whole-machine draw regardless.

**GPU** power is a valid IOReport channel on every Apple Silicon SoC, so it is
always reported — including a few milliwatts at idle (since v2.4.3 the idle GPU
shows as `0 W` instead of being dropped).

Verified live across two SoCs and macOS builds:

| Machine | SoC | macOS (Darwin) | `power.cpu` source | value |
|---|---|---|---|---|
| Mac mini | Apple **M4** (base) | 26.6 / Darwin 25.6 | IOReport | ~0.5 W idle |
| MacBook Pro | Apple **M3 Max** | 27.0 / Darwin 27.0 | **SMC `PC02`+`PC42`** | ~11 W under load |

## Build note — IOReport needs CGO

IOReport (and SMC) are reached through cgo. The local `make build-tui` enables
**CGO on macOS by default**, so a locally built daemon gets IOReport. The
CGO-free cross-compiled binaries fall back to `powermetrics` (which needs sudo).
If you want unprivileged GPU/power, **build on the Mac**.

## GPU VRAM

Apple Silicon is unified memory — the GPU has no discrete VRAM to report. Since
v2.2.5 `gpu.vram` reads `unavailable` with the detail `unified memory (no
discrete VRAM)`, so the panel explains the dash instead of silently omitting the
metric. GPU **utilisation** (`gpu.util`) and **power** (`power.gpu`) are still
reported from IOReport as above.

## NPU (ANE)

Accelerator power is the vendor-neutral `power.npu` (Apple's ANE). The legacy
`power.ane` key is still accepted and normalised on ingest, so a mixed fleet
with older daemons keeps rendering. `npu.util` reads `unavailable` — Apple
exposes no ANE utilisation counter.

## Run as launchd services

For the always-on layout (helper as a root LaunchDaemon, daemon as you, sharing
a `heimdall` group and a socket under `/usr/local/var/heimdall`), follow
[Run both as launchd services](04-privileged-metrics.md#run-both-as-launchd-services-macos-helper-root-daemon-as-you)
in the overview guide. The helper only adds full thermal and any power IOReport
missed — GPU/util and whole-system power already work without it.
