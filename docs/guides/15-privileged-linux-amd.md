# Privileged Metrics — Linux + AMD

GPU, power, and thermal on a Linux host with an AMD GPU — discrete Radeon, or the
**Strix Halo** iGPU (Ryzen AI Max, e.g. an HP ZBook Ultra G1a). For the shared
helper + systemd deployment, see
[Privileged Metrics (overview)](04-privileged-metrics.md); this guide is the AMD
specifics.

## What you get, and from where

| Metric | Source | Needs root? |
|---|---|---|
| `gpu.util`, `gpu.vram`, `gpu.temp`, `power.gpu` | **`amd-smi`** if installed, else **amdgpu sysfs** | no |
| `npu.util` (XDNA) | — (no stable counter yet) | — |
| `power.pkg` | RAPL via the helper | yes |
| `temp.pkg` | hwmon (`k10temp` / `zenpower`) via the helper | yes |

Both GPU sources are **unprivileged**, so you get the GPU panel out of the box —
installing `amd-smi` only adds richness and cross-version stability.

## Two sources, tried in order

1. **`amd-smi`** (preferred) — AMD's tool, shipped with ROCm. The daemon runs:

   ```sh
   amd-smi metric --usage --power --temperature --mem-usage --csv
   ```

   Columns are matched by token (e.g. `gfx_activity`, `socket_power`,
   `edge_temperature`, `used_vram`/`total_vram`), so the parser survives the
   header renames that differ across `amd-smi` releases.

2. **amdgpu sysfs** (automatic fallback) — the in-tree `amdgpu` driver always
   exposes these, no package required. Fills any field `amd-smi` did not provide,
   or the whole set when `amd-smi` is absent:

   | sysfs node | Metric |
   |---|---|
   | `…/card*/device/gpu_busy_percent` | `gpu.util` |
   | `…/device/mem_info_vram_used` + `…_total` | `gpu.vram` |
   | `…/device/hwmon/hwmon*/power1_average` (µW) | `power.gpu` |
   | `…/device/hwmon/hwmon*/temp1_input` (m°C) | `gpu.temp` |

   The daemon picks the first DRM card whose `device/uevent` says
   `DRIVER=amdgpu`.

## Install `amd-smi`

```sh
# Arch / CachyOS
sudo pacman -S rocm-smi-lib                     # provides amd-smi
# Debian / Ubuntu (after adding the AMD ROCm apt repo)
sudo apt install amd-smi-lib
# Fedora
sudo dnf install amd-smi

amd-smi version                                  # confirm it is on PATH
./bin/heimdall-daemon --once | grep -E "gpu|power|npu"
# e.g. gpu.util=37%  gpu.vram=25%  gpu.temp=48C  power.gpu=12.5W  npu.util=unavailable
```

`amd-smi` needs **no root** for these read-only queries — run the daemon as your
normal user. The helper is only for the CPU extras (RAPL `power.pkg`, hwmon
`temp.pkg`).

## Unified memory caveat

On a Strix Halo iGPU, `mem_info_vram_*` reflects the GPU's GTT/VRAM carve-out,
**not** system RAM. Heimdall reports it as `gpu.vram` (percent of that budget) and
keeps system memory on the separate `mem.*` keys — the two are never conflated.

## NPU (XDNA / Ryzen AI)

The XDNA NPU (`amdxdna` driver) has no stable utilisation counter yet, so
`npu.util` reads `unavailable` on AMD hosts — the accelerator is advertised but
never faked. It will light up when the driver exposes a residency figure (the
same situation as the Apple ANE `npu.util` gap).
