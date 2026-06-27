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
| `gpu.util` | percent | GPU utilization |
| `gpu.vram` | percent | GPU memory used |

## Memory & disk

| Metric | Unit | Meaning |
|---|---|---|
| `mem.used` | percent | memory used; used/total absolutes shown in the detail view |
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
| `power.pkg` | W | whole-package power |
| `power.cpu` | W | CPU power |
| `power.gpu` | W | GPU power |
| `power.ane` | W | Apple Neural Engine power (Apple Silicon) |
