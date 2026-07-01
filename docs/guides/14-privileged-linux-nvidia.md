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

If `gpu.*` is blank but `power.pkg` still reads, run the query by hand:

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
privileged sources served by `heimdall-helper`:

- **`power.pkg`** — the CPU **package** (cores + uncore), from the RAPL package
  domain, sampled as an energy-counter delta.
- **`power.cpu`** — the CPU **cores** alone, from the RAPL `core` (pp0)
  subdomain. Absent on CPUs that don't expose it; the field then stays blank.
- **`temp.pkg`** from a trusted hwmon chip.

> **Why `power.pkg` can read *below* `power.gpu`.** Both `power.pkg` and
> `power.cpu` are the **CPU socket only** (`power.cpu` is the cores, a subset of
> the whole `power.pkg` package). A discrete NVIDIA card is a separate power rail
> reported as `power.gpu`, so a workstation happily shows e.g. `power.cpu 12W`,
> `power.pkg 17W`, `power.gpu 61W` — the GPU is not part of the CPU package
> figure. **`power.total`** is the source-aware whole-machine sum the top view
> headlines: on a discrete-NVIDIA host it is `pkg + gpu (+ npu)`; on Apple it is
> `pkg` alone (SMC `PSTR` already covers everything); on an AMD APU it is `pkg`
> alone (the iGPU is inside the package).

### GB10 (Grace/ARM) has no CPU power sensor

A GB10 exposes **no** RAPL, INA, or SoC power sensor to the OS — only the GPU
rail via `nvidia-smi`. So `power.pkg` reads `unavailable` (`no RAPL power sensor
(SoC/ARM)`) and `power.total` is the GPU power alone. The Grace CPU / module draw
is simply not measurable from userspace on this platform.
>
> `power.npu` stays `unavailable` on Intel/AMD hosts — their NPUs expose no
> power counter yet, same as Apple's ANE.

Either may be absent (no powercap, no recognised sensor) — that metric reads
`unavailable` and the daemon keeps running. Set the helper up with the systemd
layout in
[Run both as systemd services](04-privileged-metrics.md#run-both-as-systemd-services-helper-root-daemon-as-you).

## Multi-GPU

`nvidia-smi` returns one row per GPU; the daemon currently reads the first row.
Multi-GPU per-device breakout is not wired yet.
