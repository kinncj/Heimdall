# Privileged Metrics — Linux + NVIDIA

GPU, power, and thermal on a Linux host with an NVIDIA GPU (workstation, DGX,
Grace-Blackwell like a GB10). For the shared helper + systemd deployment, see
[Privileged Metrics (overview)](04-privileged-metrics.md); this guide is the
NVIDIA specifics.

## What you get, and from where

| Metric | Source | Needs root? |
|---|---|---|
| `gpu.util`, `gpu.vram`, `gpu.temp`, `power.gpu` | **`nvidia-smi`** | no |
| `power.cpu` | RAPL (`/sys/class/powercap/intel-rapl`) via the helper | yes |
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

### Unified memory (GB10 Grace-Blackwell)

On a GB10 the GPU shares the system LPDDR5X pool — there is no discrete VRAM, so
`nvidia-smi` reports `memory.used`/`memory.total` as `[N/A]` and the aggregate
counter is empty. Heimdall falls back to per-process GPU memory over system RAM:

```sh
nvidia-smi --query-compute-apps=used_memory --format=csv,noheader,nounits
# sum of these ÷ total system RAM → gpu.vram, e.g. "41.6 / 121.6 GB (shared)"
```

The detail is tagged `(shared)` to mark it as the unified pool. An idle GPU (no
resident contexts) reads a stable `0%` rather than dropping the metric. Discrete
NVIDIA cards keep using the aggregate counter unchanged.

## Troubleshooting — GPU metrics missing

If `gpu.*` is blank but `power.cpu` still reads, run the query by hand:

```sh
nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader,nounits
```

- **`Failed to initialize NVML: Driver/library version mismatch`** — the driver
  package was upgraded but the running kernel module is still the old version.
  **Reboot the host** (or reload the modules). Confirm with
  `cat /proc/driver/nvidia/version` vs `nvidia-smi --version`; a pending
  `/var/run/reboot-required` is the tell. Since v2.2.5 the daemon surfaces this
  as `gpu.util` / `gpu.vram` `unavailable` with the `nvidia-smi` reason in the
  detail, instead of a silent blank.
- **`command not found`** — the NVIDIA driver/utilities are not installed, or
  not on the daemon's `PATH`.

## CPU power & temperature (helper)

NVIDIA covers the GPU; **CPU** power and temperature come from the Linux
privileged sources served by `heimdall-helper`. Power standardizes on three
metrics — the same CPU / GPU / total split btop and top show:

- **`power.cpu`** — the CPU **package** (whole socket), from the RAPL package
  domain, sampled as an energy-counter delta.
- **`power.gpu`** — the GPU, from `nvidia-smi` (a separate rail).
- **`power.total`** — `power.cpu + power.gpu (+ power.npu)`, the whole-machine
  figure the top view headlines.
- **`temp.pkg`** from a trusted hwmon chip.

> **Why `power.cpu` can read *below* `power.gpu`.** They are separate rails: the
> CPU package vs. a discrete NVIDIA card. A workstation happily shows e.g.
> `cpu 33W`, `gpu 458W`, `total 491W`. `power.total` is what reflects real draw.

### GB10 (Grace/ARM) has no CPU power sensor

A GB10 exposes **no** RAPL, INA, or SoC power sensor to the OS — only the GPU
rail via `nvidia-smi`. So `power.cpu` reads `unavailable` (`no RAPL power sensor
(SoC/ARM)`) and `power.total` is the GPU power alone. The Grace CPU / module draw
is simply not measurable from userspace on this platform. `power.npu` likewise
stays `unavailable` on Intel/AMD hosts — their NPUs expose no power counter yet,
same as Apple's ANE.

Any of these may be absent (no powercap, no recognised sensor); that metric reads
`unavailable` and the daemon keeps running. Set the helper up with the systemd
layout in
[Run both as systemd services](04-privileged-metrics.md#run-both-as-systemd-services-helper-root-daemon-as-you).

## Multi-GPU

`nvidia-smi` returns one row per GPU; the daemon currently reads the first row.
Multi-GPU per-device breakout is not wired yet.
