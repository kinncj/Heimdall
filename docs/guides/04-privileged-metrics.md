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
| **Linux + AMD** | GPU via `amd-smi` if installed, else **amdgpu sysfs** (both unprivileged) | vendor extras |
| **Other** | depends on the platform tool | whatever needs root |

When a metric has no source, the dashboard shows a **needs-helper** affordance
(`⚿`) instead of an error — the daemon keeps running and reports everything else.

### Per-platform guides

This page covers the cross-platform model and the helper/systemd/launchd
deployment. For the GPU/power specifics of your hardware, jump to:

| Platform | Guide |
|---|---|
| **macOS / Apple Silicon** | [Privileged Metrics — macOS](13-privileged-macos.md) |
| **Linux + NVIDIA** | [Privileged Metrics — Linux + NVIDIA](14-privileged-linux-nvidia.md) |
| **Linux + AMD** (Radeon / Strix Halo) | [Privileged Metrics — Linux + AMD](15-privileged-linux-amd.md) |
| **Windows** | [Privileged Metrics — Windows](16-privileged-windows.md) |

> **Apple Silicon note**: some M-series SoCs do not expose a CPU package-power
> counter at all — neither IOReport nor `powermetrics` reports it. CPU power reads
> as `unavailable` there. That is a hardware limit, not a misconfiguration.

## Option A — IOReport (Apple Silicon, no sudo)

Nothing to do. A normal daemon already reads GPU power and utilisation from Apple's
IOReport energy counters:

```sh
./bin/heimdall-daemon --once | grep -E "gpu|power"
# e.g. power.gpu=1.19W  power.total=1.19W  gpu.util=54%
```

> IOReport requires a **CGO build**. The local `make build-tui` enables CGO by
> default on macOS. The CGO-free release binaries fall back to `powermetrics`
> (which needs sudo) — build locally if you want IOReport.

## Linux power & thermal via the helper

On Linux the helper reads two privileged sources and exposes them to the
unprivileged daemon:

- **CPU package power** from RAPL (`/sys/class/powercap/intel-rapl`), sampled as
  an energy counter delta and reported as `power.cpu`; the daemon then adds
  `power.gpu` to it for `power.total`.
- **Package temperature** from a trusted hwmon chip — Intel `coretemp`, AMD
  `k10temp` or `zenpower` — reported as `temp.pkg`.

Either source may be absent. A host with no powercap or no recognised sensor
reports that metric as `unavailable` (the `⚿` affordance) and keeps running —
unsupported hardware degrades, it does not fail.

```sh
sudo ./bin/heimdall-helper &                 # serves RAPL power + hwmon temp
./bin/heimdall-daemon --hub localhost:9090   # auto-detects the socket
```

## Linux + AMD GPU (Radeon / Strix Halo)

AMD GPUs — discrete Radeon and the Strix Halo iGPU (e.g. an HP ZBook Ultra G1a)
— report `gpu.util`, `gpu.vram`, `gpu.temp`, and `power.gpu`. The daemon tries
two sources, both **unprivileged**:

1. **`amd-smi`** (preferred) — AMD's tool, shipped with ROCm. Richer and
   version-stable. The daemon runs:

   ```sh
   amd-smi metric --usage --power --temperature --mem-usage --csv
   ```

2. **amdgpu sysfs** (automatic fallback) — the in-tree `amdgpu` driver always
   exposes these, no package required:
   `gpu_busy_percent`, `mem_info_vram_used/total`, and the hwmon
   `power1_average` / `temp1_input`. Used for any field `amd-smi` did not
   provide, or whenever `amd-smi` is absent.

So **you get GPU metrics out of the box** on any amdgpu host. Installing
`amd-smi` only adds richness and cross-version stability:

```sh
# Arch / CachyOS
sudo pacman -S rocm-smi-lib                   # provides amd-smi
# Debian / Ubuntu (after adding the AMD ROCm repo)
sudo apt install amd-smi-lib
# Fedora
sudo dnf install amd-smi

amd-smi version                                # verify it is on PATH
./bin/heimdall-daemon --once | grep -E "gpu|power"
# e.g. gpu.util=37%  gpu.vram=25%  gpu.temp=48C  power.gpu=12.5W
```

`amd-smi` does **not** need root for these read-only queries — run the daemon as
your normal user. The helper is only needed for the Linux extras above (RAPL
package power, hwmon package temp).

> **NPU (XDNA) note**: the Ryzen AI / XDNA NPU (`amdxdna` driver) has no stable
> utilisation counter yet, so `npu.util` reads as `unavailable` on AMD hosts —
> the accelerator is advertised but not faked. Same situation as the Apple
> `npu.util` gap. This will light up when the driver exposes a residency figure.

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
ExecStart=/usr/local/bin/heimdall-helper --socket /run/heimdall/helper.sock
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
ExecStart=/usr/local/bin/heimdall-daemon --hub YOUR_HUB:9090 --name %H --allow-commands --process-interval 5s
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
  | jq '{state, power: .metrics["power.total"], gpu: .metrics["gpu.util"]}'
```

If `power`/`gpu` show `needs-helper`, the daemon isn't reaching the socket — check
that your user is in `heimdall` and that both units use the same socket path.

> **Install via the AUR** on Arch/CachyOS: `paru -S heimdall-helper-bin
> heimdall-daemon-bin heimdall-cli-bin` (the AUR packages install to `/usr/bin`,
> with manpages — set the units' `ExecStart` to `/usr/bin/...` if you go that
> route; the examples above assume a manual install under `/usr/local/bin`).

### Run both as launchd services (macOS, helper root, daemon as you)

The macOS equivalent of the layout above: the **helper as a root LaunchDaemon**, the
**daemon as an unprivileged LaunchDaemon** running as you, sharing a `heimdall` group
and a socket under `/usr/local/var/heimdall`. Both start at **boot** (a LaunchDaemon,
not a per-login LaunchAgent).

> **Apple Silicon build note**: GPU power/util come from **IOReport**, which needs a
> **CGO build** (`make build-tui` locally). The CGO-free release binary falls back to
> `powermetrics` (root). Full thermal and CPU/ANE power come only from the helper.
> Some M-series SoCs expose no CPU package-power counter at all — `power.total` reads
> `unavailable` there. That is a hardware limit, not a misconfiguration.

**1. Shared group** and a stable socket dir the group can traverse:

```sh
sudo dseditgroup -o create heimdall
sudo dseditgroup -o edit -a "$(whoami)" -t user heimdall
sudo mkdir -p /usr/local/var/heimdall
sudo chgrp heimdall /usr/local/var/heimdall
sudo chmod 0710 /usr/local/var/heimdall
```

**2. Helper LaunchDaemon** — `/Library/LaunchDaemons/com.heimdall.helper.plist`. It
runs as root; `GroupName` makes the socket land `root:heimdall 0660`, so the group —
not the world — can reach it:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.heimdall.helper</string>
  <key>ProgramArguments</key><array>
    <string>/usr/local/bin/heimdall-helper</string>
    <string>--socket</string><string>/usr/local/var/heimdall/helper.sock</string>
  </array>
  <key>GroupName</key><string>heimdall</string>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardErrorPath</key><string>/var/log/heimdall-helper.log</string>
</dict></plist>
```

**3. Daemon LaunchDaemon** — `/Library/LaunchDaemons/com.heimdall.daemon.plist`. It
stays unprivileged via `UserName`, joins `heimdall` to reach the socket, and points
`HEIMDALL_HELPER_SOCKET` at the helper:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.heimdall.daemon</string>
  <key>ProgramArguments</key><array>
    <string>/usr/local/bin/heimdall-daemon</string>
    <string>--hub</string><string>HUB:9090</string>
    <string>--name</string><string>my-mac</string>
  </array>
  <key>UserName</key><string>YOUR_USER</string>
  <key>GroupName</key><string>heimdall</string>
  <key>EnvironmentVariables</key><dict>
    <key>HEIMDALL_TOKEN</key><string>YOUR_TOKEN</string>
    <key>HEIMDALL_HELPER_SOCKET</key><string>/usr/local/var/heimdall/helper.sock</string>
  </dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardErrorPath</key><string>/var/log/heimdall-daemon.log</string>
</dict></plist>
```

**4. Own them `root:wheel`, lock the token file, load helper first:**

```sh
sudo chown root:wheel /Library/LaunchDaemons/com.heimdall.{helper,daemon}.plist
sudo chmod 644 /Library/LaunchDaemons/com.heimdall.helper.plist
sudo chmod 600 /Library/LaunchDaemons/com.heimdall.daemon.plist   # token inside
sudo launchctl bootstrap system /Library/LaunchDaemons/com.heimdall.helper.plist
sudo launchctl bootstrap system /Library/LaunchDaemons/com.heimdall.daemon.plist
```

**5. Verify:**

```sh
ls -l /usr/local/var/heimdall/helper.sock        # expect: srw-rw---- root heimdall
sudo launchctl print system/com.heimdall.daemon | grep -E 'state|pid'
heimdall-cli --hub HUB:9090 host my-mac \
  | jq '{state, power: .metrics["power.total"], gpu: .metrics["gpu.util"]}'
```

If `power`/`gpu` show `⚿` (`needs-helper`), either the daemon isn't reaching the
socket — check your user is in `heimdall` and both plists use the same socket path —
or you're on a CGO-free release binary (install a local `make build-tui` daemon for
IOReport). Reload after editing a plist with
`sudo launchctl bootout system/com.heimdall.daemon`, then `bootstrap` again.

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
