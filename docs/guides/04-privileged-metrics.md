# Privileged Metrics — Power, GPU, Full Thermal

CPU/memory/disk/network/uptime are always available unprivileged. Some metrics —
full thermal, package/CPU/ANE power, detailed GPU — need elevated access. Heimdall
gets them without running the daemon as root.

## How power/GPU metrics are sourced

The daemon tries sources in order and degrades gracefully:

| Platform | Without any helper / root | With `heimdall-helper` (root) |
|---|---|---|
| **Apple Silicon** | GPU power + GPU utilisation via **IOReport** (no sudo) | adds full thermal, CPU/ANE power where the SoC exposes it |
| **Linux** | CPU/mem/disk/net/uptime (no sudo) | adds CPU package power from RAPL + package temperature from hwmon |
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

## Linux power & thermal via the helper

On Linux the helper reads two privileged sources and exposes them to the
unprivileged daemon:

- **CPU package power** from RAPL (`/sys/class/powercap/intel-rapl`), sampled as
  an energy counter delta and reported as `power.pkg`.
- **Package temperature** from a trusted hwmon chip — Intel `coretemp`, AMD
  `k10temp` or `zenpower` — reported as `temp.pkg`.

Either source may be absent. A host with no powercap or no recognised sensor
reports that metric as `unavailable` (the `⚿` affordance) and keeps running —
unsupported hardware degrades, it does not fail.

```sh
sudo ./bin/heimdall-helper &                 # serves RAPL power + hwmon temp
./bin/heimdall-daemon --hub localhost:9090   # auto-detects the socket
```

## Option B — the privileged helper

Run `heimdall-helper` as root on a host; it serves privileged readings to the
**unprivileged** daemon over a local unix socket. To try it in a shell, run both as
the **same user** (e.g. both under `sudo`, or both as you) and the daemon
auto-detects the socket — no daemon flag needed:

```sh
sudo ./bin/heimdall-helper &                 # serves privileged metrics locally
./bin/heimdall-daemon --hub localhost:9090   # auto-detects /tmp/heimdall-helper.sock
```

The helper's socket is `0660` (owner + group only — never world-readable, since it
also runs delegated privileged commands). So when the **helper runs as root and the
daemon as your user** — the normal service layout — they must share a group and a
socket path. That's the setup below.

### Why a helper instead of `sudo heimdall-daemon`?

Privilege isolation. The daemon — which talks to the network — stays unprivileged.
Only the small, local, **read-only** helper holds root. The daemon sends the helper
nothing it can't validate, and the helper enforces its **own** allow-list — so there
is no argument-injection surface, and a delegated command runs only if the helper
itself permits it. Power is presented read-only; the dashboard offers no control to
change a power profile.

### Run both as systemd services (helper root, daemon as you)

The robust layout for an always-on host: the **helper as root**, the **daemon as
your user**, talking over a socket in `/run/heimdall` that a shared `heimdall` group
gates. Tested on CachyOS (Arch); works on any systemd distro.

**1. Shared group** so your user can reach the root helper's `0660` socket:

```sh
sudo groupadd -f heimdall
sudo usermod -aG heimdall "$USER"      # log out/in (or `newgrp heimdall`) to pick it up
```

**2. Helper unit** — `/etc/systemd/system/heimdall-helper.service`:

```ini
[Unit]
Description=Heimdall privileged helper (power/GPU/thermal + delegated privileged commands)
After=network.target

[Service]
Type=simple
User=root
Group=heimdall
# systemd creates /run/heimdall as root:heimdall; 0710 lets the group traverse it.
RuntimeDirectory=heimdall
RuntimeDirectoryMode=0710
ExecStart=/usr/bin/heimdall-helper --socket /run/heimdall/helper.sock
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target
```

`Group=heimdall` is the key line: it makes the socket the helper creates land as
`root:heimdall 0660`, so the group — not the world — can talk to it.

**3. Daemon unit** — `/etc/systemd/system/heimdall-daemon.service` (set `User=` and
`--hub`):

```ini
[Unit]
Description=Heimdall daemon (per-host metrics -> hub)
Wants=network-online.target heimdall-helper.service
After=network-online.target heimdall-helper.service

[Service]
Type=simple
User=YOUR_USER
SupplementaryGroups=heimdall
Environment=HEIMDALL_HELPER_SOCKET=/run/heimdall/helper.sock
ExecStart=/usr/bin/heimdall-daemon --hub YOUR_HUB:9090 --name %H --allow-commands --process-interval 5s
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target
```

`Wants=` (not `Requires=`) the helper, so the daemon still runs and reports
unprivileged metrics if the helper is down. Drop `--allow-commands` if you don't
want the on-demand command plane.

**4. Enable & start** — helper first:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now heimdall-helper.service
sudo systemctl enable --now heimdall-daemon.service

ls -l /run/heimdall/helper.sock          # expect:  srw-rw---- root heimdall
heimdall-cli --hub YOUR_HUB:9090 host "$(hostname)" \
  | jq '{state, power: .metrics["power.pkg"], gpu: .metrics["gpu.util"]}'
```

If `power`/`gpu` show `needs-helper`, the daemon isn't reaching the socket — check
that your user is in `heimdall` and that both units use the same socket path.

> **Install via the AUR** on Arch/CachyOS: `paru -S heimdall-helper-bin
> heimdall-daemon-bin heimdall-cli-bin` (binaries land in `/usr/bin`, with manpages).

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

| Flag / env | Meaning |
|---|---|
| `--socket <path>` | unix socket the helper serves on (default `/tmp/heimdall-helper.sock`) |
| `--demo` | serve canned sample metrics (no root needed; for trying the UI) |
| `HEIMDALL_HELPER_SOCKET` | env the **daemon** reads to find the helper socket — set it to the helper's `--socket` path when they differ (e.g. the service layout above) |

## Next steps

- Full flag reference → [Configuration](../configuration.md)
- Architecture of privilege tiers → [ADR 0004](../architecture/0004-optional-privileged-helper-and-privilege-tiers.md)
