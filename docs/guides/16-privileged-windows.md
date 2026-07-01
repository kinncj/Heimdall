# Privileged Metrics — Windows

Thermal and GPU on Windows. For the cross-platform model, see
[Privileged Metrics (overview)](04-privileged-metrics.md); this guide is the
Windows specifics.

## What you get, and from where

| Metric | Source | Needs admin? |
|---|---|---|
| `temp.pkg` | **WMI** `MSAcpi_ThermalZoneTemperature` via PowerShell | sometimes |
| `gpu.util`, `gpu.vram`, `gpu.temp`, `power.gpu` (NVIDIA) | `nvidia-smi` | no |
| `power.cpu` | **Scaphandre** (its driver reads RAPL; scraped over HTTP) | no |

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
ring-0 driver, which Heimdall does not ship. Rather than sign our own driver,
Heimdall reads CPU power from **[Scaphandre](https://github.com/hubblo-org/scaphandre)**,
which installs the signed Hubblo RAPL driver and runs as a Windows service.

Set it up once:

1. **(only if the driver won't load — dev/unsigned builds)** enable test-signing
   as administrator, then **reboot**:

   ```
   bcdedit.exe -set TESTSIGNING ON
   bcdedit.exe -set nointegritychecks on
   ```

   The tell is `Failed to open device : HANDLE(-1)` from `scaphandre stdout`.

2. Download the latest **Scaphandre** Windows `.exe` installer (≥ 1.0.0) and run
   it **as administrator** — it installs the Hubblo RAPL driver (Intel and AMD
   since Ryzen).

3. Register it as a service running the **`prometheus`** exporter (the *pull*
   exporter that listens locally — **not** `prometheus-push`, which pushes to a
   remote gateway):

   ```
   sc.exe create Scaphandre binPath="C:\Program Files (x86)\scaphandre\scaphandre.exe prometheus -a 127.0.0.1 -p 8080" DisplayName=Scaphandre start=auto
   sc.exe start Scaphandre
   ```

4. Verify — **in PowerShell** (the `&` call operator and `Invoke-WebRequest` are
   PowerShell; don't paste these into `cmd.exe`), one per line:

   ```powershell
   driverquery /v | Select-String capha
   & 'C:\Program Files (x86)\scaphandre\scaphandre.exe' stdout
   (Invoke-WebRequest http://127.0.0.1:8080/metrics -UseBasicParsing).Content | Select-String scaph_socket_power
   ```

   The cmd.exe equivalents are `driverquery /v | findstr /i capha`,
   `"C:\Program Files (x86)\scaphandre\scaphandre.exe" stdout`, and
   `curl http://127.0.0.1:8080/metrics | findstr scaph_socket_power`. If step 2
   prints `Failed to open device : HANDLE(-1)`, the driver isn't loaded — do the
   test-signing step above and reboot.

Heimdall scrapes `http://127.0.0.1:8080/metrics` and reports the summed per-socket
power (`scaph_socket_power_microwatts`) as `power.cpu` — pure Go, no cgo. If
Scaphandre listens elsewhere, point Heimdall at it with `HEIMDALL_SCAPHANDRE_URL`;
restart `heimdall-daemon` once Scaphandre is up.

When Scaphandre isn't reachable, `power.cpu` reads `unavailable` with
`no RAPL on Windows — run Scaphandre for CPU power` rather than failing.

## Run the daemon as a service

The daemon is a console program, not a native Windows service, so registering it
directly with `sc.exe` trips the Service Control Manager (error 1053). Use the
built-in **Task Scheduler** to run it at boot instead — no third-party wrapper.
In an **Administrator** PowerShell, substituting your own binary path, hub
address, and `--name`:

```powershell
schtasks /create /tn "Heimdall Daemon" `
  /tr "C:\Heimdall\heimdall-daemon.exe --hub HUB_HOST:9090 --name %COMPUTERNAME% --process-interval 5s" `
  /sc onstart /ru SYSTEM /rl HIGHEST /f
schtasks /run /tn "Heimdall Daemon"
```

`%COMPUTERNAME%` fills in the host name automatically. Running as `SYSTEM` is
enough — the daemon's Windows sources (`nvidia-smi`, WMI temperature, and the
Scaphandre scrape) are all unprivileged, so **no `heimdall-helper` is needed on
Windows**. Add `--log-file C:\Heimdall\daemon.log` if you want a log to tail.

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
| `power.cpu` | no (scrapes HTTP) | needs **Scaphandre** running (installs the RAPL driver, exposes Prometheus); else `unavailable` |

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
