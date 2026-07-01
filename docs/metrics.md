# Metrics Reference

Every metric a `heimdall-daemon` collects, with its name, unit, and meaning. Each
reading also carries a status — `ok`, `unavailable`, `needs-helper`, or `error`.
A non-OK reading carries a short reason instead of a value (a failing adapter is
isolated; it never drops the host). See the daemon's print mode in
[Configuration](configuration.md#heimdall-daemon).

## Host context

Gathered once at startup; the value is a string.

| Metric | Meaning |
|---|---|
| `host.os` | operating system / kernel family and release |
| `host.arch` | CPU architecture (e.g. `arm64`, `amd64`) |
| `host.kernel` | kernel version |
| `host.cpu` | CPU model and thread count |
| `host.gpu` | GPU model |
| `host.version` | Heimdall daemon version |
| `host.uptime` | time since boot |

## Compute

| Metric | Unit | Meaning |
|---|---|---|
| `cpu.util` | percent | overall CPU utilization |
| `cpu.cores` | percent (per-core) | per-core utilization; the dashboard shows avg and max |
| `cpu.load` | load | 1/5/15-minute load average (macOS/Linux; unavailable on Windows) |
| `cpu.freq` | MHz | CPU clock; per-core where available. `/sys` cpufreq on Linux; unavailable where no clock is exposed (e.g. Apple Silicon) |
| `gpu.util` | percent | GPU utilization |
| `gpu.vram` | percent | GPU memory used |
| `gpu.mem.util` | percent | GPU memory-controller utilization (distinct from `gpu.vram` occupancy) |
| `gpu.clock` | MHz | GPU graphics/shader clock |
| `gpu.fan` | percent | GPU fan speed (where a fan + sensor exist) |

> **GPU sources**: Apple Silicon via IOReport; NVIDIA via `nvidia-smi`; AMD via
> `amd-smi` when installed, otherwise the `amdgpu` sysfs nodes (both
> unprivileged). `gpu.clock` / `gpu.mem.util` / `gpu.fan` are best-effort and
> render `—` where the chip or query doesn't expose them (e.g. fan on a passively
> cooled DGX). See [Privileged Metrics](guides/04-privileged-metrics.md).

## Memory & disk

| Metric | Unit | Meaning |
|---|---|---|
| `mem.used` | percent | memory used; used/total absolutes shown in the detail view |
| `mem.swap` | percent | swap used; used/total absolutes shown in the detail view |
| `disk.used` | percent | disk used; used/total absolutes shown in the detail view |
| `disk.read` | MB/s | disk read throughput |
| `disk.write` | MB/s | disk write throughput |

## Network

| Metric | Unit | Meaning |
|---|---|---|
| `net.rx` | MB/s | total receive throughput |
| `net.tx` | MB/s | total transmit throughput |
| `net.rx.<nic>` | MB/s | per-NIC receive throughput |
| `net.tx.<nic>` | MB/s | per-NIC transmit throughput |
| `net.latency` | ms | round-trip latency to the internet ping target (`--ping-target`) |
| `net.gateway` | ms | round-trip latency to the default gateway |
| `net.gateway.<nic>` | ms | per-NIC gateway latency |

## Thermal & power

These often need elevated access. Without it, the daemon reports `needs-helper`
rather than failing. See [Privileged Metrics](guides/04-privileged-metrics.md).

| Metric | Unit | Meaning |
|---|---|---|
| `temp.pkg` | °C | CPU package temperature |
| `gpu.temp` | °C | GPU temperature |
| `power.total` | W | whole-machine power — `cpu + gpu (+ npu)`, or the SMC whole-system total on macOS |
| `power.cpu` | W | CPU power — RAPL package (Linux), IOReport CPU (macOS), Scaphandre (Windows, when running); Unavailable-with-reason where no source exists (Apple Pro/Max, Windows without Scaphandre, GB10) |
| `power.gpu` | W | GPU power |
| `power.npu` | W | NPU / accelerator power (Apple Silicon ANE today). The legacy `power.ane` key is accepted and normalised to `power.npu` on ingest. |
| `npu.util` | percent | NPU utilisation — real on Intel NPUs (`intel_vpu`); Unavailable-with-reason on AMD XDNA and Apple ANE, which expose no counter |
