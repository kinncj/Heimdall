# Privileged Metrics — Windows

Thermal and GPU on Windows. For the cross-platform model, see
[Privileged Metrics (overview)](04-privileged-metrics.md); this guide is the
Windows specifics.

## What you get, and from where

| Metric | Source | Needs admin? |
|---|---|---|
| `temp.pkg` | **WMI** `MSAcpi_ThermalZoneTemperature` via PowerShell | sometimes |
| `gpu.util`, `gpu.vram`, `gpu.temp`, `power.gpu` (NVIDIA) | `nvidia-smi` | no |
| `power.cpu` | — (RAPL is inaccessible without a kernel driver) | — |

CPU/memory/disk/network/uptime are always available unprivileged, as on every
platform. The Windows-specific privileged path adds **package temperature** from
the ACPI thermal zone:

```powershell
Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature `
  | Select-Object -ExpandProperty CurrentTemperature
```

The value is ACPI deci-Kelvin; the daemon converts it to °C. Some firmware
exposes no usable thermal zone, in which case `temp.pkg` reads `unavailable`.

## GPU

If the host has an NVIDIA GPU with the driver installed, `nvidia-smi` is on PATH
and the GPU panel works **unprivileged** — same as
[Linux + NVIDIA](14-privileged-linux-nvidia.md). AMD GPU collection on Windows is
**not** wired (the amdgpu sysfs path is Linux-only, and `amd-smi` parsing is not
yet exercised on Windows).

## CPU power

Windows exposes no RAPL to user space — reading the CPU energy MSRs needs a
ring-0 driver, which Heimdall does not ship (that's exactly what tools like
Scaphandre and WinPowerMonitor install). Rather than sign our own driver,
Heimdall reads CPU power from a **driver-backed monitor you already run**:

- **[LibreHardwareMonitor](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor)**
  — start it (its signed driver reads the RAPL MSRs on Intel **and** AMD) and
  leave "Remote Web Server / WMI" enabled. Heimdall queries its WMI provider
  (`root/LibreHardwareMonitor`) and reports the CPU-package sensor as `power.cpu`.

When no such monitor is running, `power.cpu` reads `unavailable` with
`no RAPL on Windows — run LibreHardwareMonitor for CPU power` rather than failing.

## How the helper works on Windows

The helper runs on Windows — its unix-domain socket works on Windows 10 1803+
(AF_UNIX). When it's reachable, the daemon uses it as the **source** for
privileged metrics, so on a host with the helper connected your GPU and thermal
readings flow *through* it.

What matters in practice is whether the helper collects anything the **bare
daemon** can't. Here's what each Windows source needs:

| Windows source | Needs elevation? | Notes |
|---|---|---|
| `nvidia-smi` (`gpu.util/vram/temp`, `power.gpu`) | normally **no** | richest panel on Windows when an NVIDIA GPU is present |
| WMI `temp.pkg` (`MSAcpi_ThermalZoneTemperature`) | usually **yes** | only if the firmware exposes an ACPI thermal zone — many laptops don't |
| `power.cpu` | no (reads WMI) | needs **LibreHardwareMonitor** running (its driver reads RAPL); else `unavailable` |

`nvidia-smi` is normally unprivileged, so a daemon running as your user *should*
collect the GPU panel without the helper. But account/PATH setups vary — don't
assume, **measure it on your host**:

```powershell
nssm stop HeimdallHelper
# wait ~10s for the next daemon poll, then from the monitoring station:
heimdall-cli --hub YOUR_HUB:9090 host $env:COMPUTERNAME | ConvertFrom-Json | % metrics
nssm start HeimdallHelper
```

- `gpu.*` **still present** → the daemon collects it itself; the helper is
  redundant on this host, run daemon-only.
- `gpu.*` **gone** → on this host the elevated helper *is* required for the GPU —
  keep it.

The helper's one Windows-exclusive contribution is WMI `temp.pkg`, which needs an
elevated process **and** a firmware thermal zone. If a host shows `gpu.*` but
never `temp.pkg`, it has no thermal zone and the helper adds nothing there.

> On Linux/macOS the helper does more — RAPL `power.cpu`, hwmon `temp.pkg`, full
> Apple thermal — so this "is it redundant?" question is Windows-specific.

## Run as Windows services

`heimdall-daemon` and `heimdall-helper` are plain console programs — they don't
implement the Windows Service Control Manager (SCM) protocol. Register each with a
supervisor that adapts a console app into a service. [NSSM](https://nssm.cc) is
the simplest; [WinSW](https://github.com/winsw/winsw) works too.

### Daemon-only (the common case)

No thermal zone, so the helper adds nothing — just run the daemon:

```powershell
nssm install Heimdall "C:\Program Files\Heimdall\heimdall-daemon.exe" --hub YOUR_HUB:9090 --name $env:COMPUTERNAME --allow-commands
nssm set Heimdall Start SERVICE_AUTO_START
nssm start Heimdall
```

### With the helper, for `temp.pkg` (only if your host exposes a thermal zone)

The socket path is the thing to get right. The default socket lives under the
process owner's **`%TEMP%`**, so the helper and daemon connect on the default path
**only when they resolve to the same temp dir** — i.e. when both run **as the same
user**. Two layouts:

**Same user, helper elevated (simplest — default socket Just Works).** Run both
services under *your* account. A service started under a user in the
Administrators group gets the **full (elevated) token**, so the helper can read
WMI thermal, and because it's the same user, `%TEMP%` matches the daemon's — no
`--socket`, no env var:

```powershell
nssm install HeimdallHelper "C:\Program Files\Heimdall\heimdall-helper.exe"
nssm set HeimdallHelper ObjectName ".\YOUR_USER" "YOUR_PASSWORD"   # an admin user
nssm set HeimdallHelper Start SERVICE_AUTO_START

nssm install Heimdall "C:\Program Files\Heimdall\heimdall-daemon.exe" --hub YOUR_HUB:9090 --name $env:COMPUTERNAME --allow-commands
nssm set Heimdall ObjectName ".\YOUR_USER" "YOUR_PASSWORD"          # same user
nssm set Heimdall Start SERVICE_AUTO_START

nssm start HeimdallHelper
nssm start Heimdall
# both land on C:\Users\YOUR_USER\AppData\Local\Temp\heimdall-helper.sock
```

> This is the layout that "running the helper as admin and the daemon as the same
> user" actually produces — both processes share your user `%TEMP%`, so the
> default socket connects with no extra config.

**Different accounts (e.g. helper as `LocalSystem`).** Now the temp dirs differ
(`LocalSystem` → `C:\Windows\Temp`), so the default paths **won't** match — you
**must** pin an explicit shared `--socket` and tell the daemon about it:

```powershell
$sock = "C:\ProgramData\Heimdall\helper.sock"
New-Item -ItemType Directory -Force "C:\ProgramData\Heimdall" | Out-Null

nssm install HeimdallHelper "C:\Program Files\Heimdall\heimdall-helper.exe" --socket $sock
nssm set HeimdallHelper ObjectName LocalSystem
nssm install Heimdall "C:\Program Files\Heimdall\heimdall-daemon.exe" --hub YOUR_HUB:9090 --name $env:COMPUTERNAME --allow-commands
nssm set Heimdall ObjectName ".\YOUR_USER" "YOUR_PASSWORD"
nssm set Heimdall AppEnvironmentExtra HEIMDALL_HELPER_SOCKET=$sock   # required across accounts
```

(Your user also needs read/write on `C:\ProgramData\Heimdall` to open the socket.)

To change flags later: `nssm edit <name>`. To remove: `nssm stop <name>; nssm
remove <name> confirm`.

### Verify

```powershell
Get-Service Heimdall, HeimdallHelper
# from the monitoring station:
heimdall-cli --hub YOUR_HUB:9090 host $env:COMPUTERNAME | ConvertFrom-Json
```

Expect `gpu.*` if an NVIDIA GPU is present. `temp.pkg` populates only where the
firmware exposes a thermal zone **and** the helper runs elevated. `power.cpu`
stays `unavailable`. If `gpu.*` shows but `temp.pkg` does not, that host has no
ACPI thermal zone and the helper is contributing nothing — you can drop it.
