# Privileged Metrics — Linux + NVIDIA

GPU, power, and thermal on a Linux host with an NVIDIA GPU (workstation, DGX,
Grace-Blackwell like a GB10). For the shared helper + systemd deployment, see
[Privileged Metrics (overview)](04-privileged-metrics.md); this guide is the
NVIDIA specifics.

## What you get, and from where

| Metric | Source | Needs root? |
|---|---|---|
| `gpu.util`, `gpu.vram`, `gpu.temp`, `power.gpu` | **`nvidia-smi`** | no |
| `power.pkg` | RAPL (`/sys/class/powercap/intel-rapl`) via the helper | yes |
| `temp.pkg` | hwmon (`coretemp` / `k10temp` / `zenpower`) via the helper | yes |

GPU metrics are **unprivileged** — `nvidia-smi` is readable by any user, so a
normal daemon reports the full GPU panel with no helper. The daemon runs:

```sh
nvidia-smi --query-gpu=utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw \
  --format=csv,noheader,nounits
```

Verify:

```sh
nvidia-smi -L                                  # confirm the driver sees the GPU
./bin/heimdall-daemon --once | grep -E "gpu|power"
# e.g. gpu.util=0%  gpu.vram=1%  gpu.temp=33C  power.gpu=4.95W
```

`gpu.vram` is reported as a percentage of total, with the absolute `used / total
GB` in the detail view.

## CPU package power & temperature (helper)

NVIDIA covers the GPU; **CPU** package power and package temperature still come
from the Linux privileged sources served by `heimdall-helper`:

- **`power.pkg`** from RAPL, sampled as an energy-counter delta.
- **`temp.pkg`** from a trusted hwmon chip.

Either may be absent (no powercap, no recognised sensor) — that metric reads
`unavailable` and the daemon keeps running. Set the helper up with the systemd
layout in
[Run both as systemd services](04-privileged-metrics.md#run-both-as-systemd-services-helper-root-daemon-as-you).

## Multi-GPU

`nvidia-smi` returns one row per GPU; the daemon currently reads the first row.
Multi-GPU per-device breakout is not wired yet.
