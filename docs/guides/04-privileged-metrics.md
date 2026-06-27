# Privileged Metrics — Power, GPU, Full Thermal

CPU/memory/disk/network/uptime are always available unprivileged. Some metrics —
full thermal, package/CPU/ANE power, detailed GPU — need elevated access. Heimdall
gets them without running the daemon as root.

## How power/GPU metrics are sourced

The daemon tries sources in order and degrades gracefully:

| Platform | Without any helper / root | With `heimdall-helper` (root) |
|---|---|---|
| **Apple Silicon** | GPU power + GPU utilisation via **IOReport** (no sudo) | adds full thermal, CPU/ANE power where the SoC exposes it |
| **Linux + NVIDIA** | GPU via `nvidia-smi` (unprivileged) | vendor extras |
| **Other** | depends on the platform tool | whatever needs root |

When a metric has no source, the dashboard shows a **needs-helper** affordance
(`⚿`) instead of an error — the daemon keeps running and reports everything else.

> **Apple Silicon note**: some M-series SoCs do not expose a CPU package-power
> counter at all — neither IOReport nor `powermetrics` reports it. CPU power reads
> as `unavailable` there. That is a hardware limit, not a misconfiguration.

## Option A — IOReport (Apple Silicon, no sudo)

Nothing to do. A normal daemon already reads GPU power and utilisation from Apple's
IOReport energy counters:

```sh
./bin/heimdall-daemon --once | grep -E "gpu|power"
# e.g. power.gpu=1.19W  power.pkg=1.19W  gpu.util=54%
```

> IOReport requires a **CGO build**. The local `make build-tui` enables CGO by
> default on macOS. The CGO-free release binaries fall back to `powermetrics`
> (which needs sudo) — build locally if you want IOReport.

## Option B — the privileged helper

Run `heimdall-helper` as root on a host; it serves privileged readings to the
**unprivileged** daemon over a local unix socket. The daemon auto-detects the
socket — no daemon flag needed.

```sh
sudo ./bin/heimdall-helper &                 # serves privileged metrics locally
./bin/heimdall-daemon --hub localhost:9090   # auto-detects the socket
```

### Why a helper instead of `sudo heimdall-daemon`?

Privilege isolation. The daemon — which talks to the network — stays unprivileged.
Only the small, local, **read-only** helper holds root. The daemon sends the helper
nothing and cannot influence what it collects, so there is no argument-injection or
command-execution surface. Power is presented read-only; the dashboard offers no
control to change a power profile.

## Option C — preview without root

To see the populated values without elevation (useful for screenshots/testing):

```sh
./bin/heimdall-helper --demo &     # canned sample readings on the socket
./bin/heimdall-daemon --once       # GPU and POWER now show sample values
```

## Do I need the helper?

Usually **no** — start with just the daemon. Add the helper only if a metric you
want shows `⚿`. On Apple Silicon and Linux+NVIDIA, GPU metrics already work
without it.

## Helper flags

| Flag | Meaning |
|---|---|
| `--socket <path>` | unix socket to serve on (daemon uses the same default) |
| `--demo` | serve canned sample metrics (no root needed; for trying the UI) |

## Next steps

- Full flag reference → [Configuration](../configuration.md)
- Architecture of privilege tiers → [ADR 0004](../architecture/0004-optional-privileged-helper-and-privilege-tiers.md)
