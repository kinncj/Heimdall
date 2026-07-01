# Changelog

All notable changes to Heimdall are recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Standardized power on three metrics everywhere: `power.cpu`, `power.gpu`,
  `power.total`** — the same CPU / GPU / total split btop, mactop, and top show.
  `power.pkg` is retired; the CPU rail is always `power.cpu`:
  - **Linux**: `power.cpu` = the RAPL **package** (whole CPU socket — what btop
    calls "CPU"); the unreliable RAPL `core` subdomain is dropped.
  - **macOS**: `power.cpu` = IOReport CPU; `power.total` = SMC `PSTR`
    (whole-system) — the SMC figure is no longer mislabelled `power.pkg`.
  - `power.total` = `cpu + gpu + npu` on every non-Apple host (Apple already
    reports a whole-system total). This fixes **AMD APUs (Strix Halo)**
    under-counting: RAPL package and amdgpu are separate rails and are now summed
    (matching btop's CPU + GPU).
- Both the **top view** and the **host detail view** now show CPU, GPU, and total
  power.
- `power.cpu` is reported Unavailable-with-reason where no source exists: GB10
  (no RAPL on Grace/ARM) and Windows (RAPL inaccessible).

## [2.2.9] - 2026-07-01

### Added
- **`power.total` — a source-aware whole-machine power figure**, now headlined by
  the top-view POWER panel. The headline used to be `power.pkg` (CPU package
  only), so a workstation pulling 506 W on the GPU read as ~32 W. `power.total`
  sums the right rails per platform **without double-counting**: Apple =
  `power.pkg` (SMC `PSTR` is already whole-system); NVIDIA / GB10 =
  `pkg + gpu + npu` (the GPU is a separate rail); AMD APU / integrated = `pkg`
  (the iGPU is already inside the package). rtx-pro now reads ~538 W, GB10 ~50 W.
- **`power.pkg` on a no-RAPL SoC now says why.** GB10 (Grace/ARM) exposes no
  RAPL, INA, or any CPU/module power sensor, so `power.pkg` reads Unavailable with
  `no RAPL power sensor (SoC/ARM)` instead of a silent blank; there `power.total`
  is the GPU rail alone.

### Changed
- POWER panel layout: a `total` headline, then `cpu / gpu / npu`, then `CPU pkg`
  — making the CPU-cores (`cpu`) vs CPU-package (`pkg`) distinction explicit
  (`cpu` is the cores, a subset of the whole `pkg` socket).

## [2.2.8] - 2026-06-30

### Added
- **`heimdall-cli` now exposes OK-metric detail strings.** Info metrics
  (`host.version`, `host.gpu`, `host.os`, `host.cpu`, `host.kernel`, `host.arch`)
  carry their human value in the metric detail with a zero gauge, so the CLI
  previously showed them as `0` and dropped the string entirely. A new `details`
  object (metric name → detail text) surfaces them, plus notes like `gpu.vram`'s
  `43/122 GB shared`. `metrics` stays value-only, so existing scripts are
  unaffected. Now `heimdall-cli host <id> | jq -r '.details["host.version"]'`
  works.

## [2.2.7] - 2026-06-30

### Fixed
- **Top view: the GB10 `gpu.vram` detail wrapped and broke the panel box.** The
  `(shared)` detail was longer than a discrete card's, so lipgloss wrapped the
  line and mangled the GPU panel. The detail is now clipped to the space left on
  the line (never wraps) and shortened to `42/122 GB shared` (was
  `42.6 / 121.6 GB (shared)`).
- **`heimdall-cli` hid unavailable metrics.** It only emitted OK metrics as
  `metrics` (name → value), so a metric that is deliberately Unavailable — Apple
  `gpu.vram`, `npu.util`, anything needing the helper — vanished from CLI output
  with no reason. A new `unavailable` object (name → `{status, detail}`) now
  carries them. `metrics` is unchanged, so existing scripts keep working.

### Added
- **`npu.util` is Unavailable-with-reason everywhere, not a bare dash.** NPUs
  (Apple ANE, Intel AI Boost, AMD XDNA) expose no utilisation counter; hosts
  without an NPU source now report `npu.util` Unavailable with
  `no NPU utilisation counter` (the AMD path keeps its own specific reason).

### Build
- `scripts/release.sh` now builds macOS with CGO (so IOReport/SMC power is
  compiled in) and refuses to emit a CGO-free darwin binary when run off a Mac —
  `make release` on a Mac is no longer a footgun. CI is unaffected.
- The Makefile stamps `-X main.version` from `git describe`, so `make build-tui`
  binaries report a real version instead of `dev`.

## [2.2.6] - 2026-06-30

### Added
- **`power.cpu` on Linux from the RAPL `core` subdomain.** The Linux privileged
  path only read the RAPL *package* (`power.pkg`), so the top view's CPU power
  column was blank while the GPU showed 60W+ and the package ~17W — which read as
  "where's the CPU, and why is the package below the GPU?". It now also reads the
  `core` (pp0) subdomain as `power.cpu` (cores only). The two CPU rails are
  labelled — `power.pkg` = "CPU package", `power.cpu` = "CPU cores" — to make
  clear they are the CPU socket only; a discrete GPU is a separate rail
  (`power.gpu`), which is why `power.pkg` can sit well below `power.gpu`. Absent
  on CPUs with no `core` subdomain. `power.npu` stays unavailable — Intel/AMD
  NPUs expose no power counter (same posture as Apple ANE / AMD XDNA).

## [2.2.5] - 2026-06-30

### Added
- **GPU VRAM on unified-memory NVIDIA (GB10 Grace-Blackwell).** `nvidia-smi`
  reports `memory.used`/`memory.total` as `[N/A]` on GB10 because there is no
  discrete VRAM — so hosts like `promaxgb10` showed GPU util but no `gpu.vram`.
  When the aggregate counter is `[N/A]`, `gpu.vram` is now derived from the sum
  of per-process GPU memory (`nvidia-smi --query-compute-apps=used_memory`) over
  the system RAM total, e.g. `gpu.vram  34%  41.6 / 121.6 GB (shared)`. Discrete
  NVIDIA cards (with a real aggregate counter) are unchanged.

### Fixed
- **A broken `nvidia-smi` silently blanked GPU instead of saying why.** When
  `nvidia-smi` is installed but exits non-zero — most often a
  `Driver/library version mismatch` after the NVIDIA driver is upgraded without a
  reboot — the collector dropped every `gpu.*` key, so the host looked like
  Heimdall was failing. `gpu.util` and `gpu.vram` are now reported Unavailable
  carrying the `nvidia-smi` reason (e.g. `nvidia-smi: Failed to initialize NVML:
  Driver/library version mismatch`), so the fix (reboot the host) is obvious. The
  underlying blank on such hosts is a host driver state, not a Heimdall bug.
- **Apple Silicon `gpu.vram` now explains the dash.** Unified memory has no
  discrete VRAM, so `gpu.vram` was simply absent on Macs. It is now reported
  Unavailable with `unified memory (no discrete VRAM)` rather than silently
  omitted.

## [2.2.4] - 2026-06-30

### Fixed
- **Running the helper could blank out power/GPU it was meant to provide.** The
  privileged adapter trusted the helper's reply on any successful socket call —
  even one with no `ok` metrics — so a reachable-but-empty helper *shadowed* the
  daemon's own in-process reading. On Apple Silicon, where IOReport/SMC are
  unprivileged, bootstrapping the helper made `power.*` and `gpu.*` disappear on
  a Mac that worked fine without it. The adapter now uses the helper only when it
  returns an `ok` metric, otherwise falls back to in-process collection — so the
  helper is additive, never subtractive. Also covers Windows hosts where the
  helper would otherwise shadow the daemon's `nvidia-smi` GPU read.

## [2.2.3] - 2026-06-30

### Fixed
- **Windows process table showed every process at 0% CPU and 0% memory.** The
  collector used `tasklist`, which has no CPU column, and the parser also dropped
  its memory column — so `p` and the top view listed real processes with fake
  zeros. Windows now reads `Win32_PerfFormattedData_PerfProc_Process` for real
  instantaneous `PercentProcessorTime` and working-set memory (as a percent of
  total RAM), falling back to `tasklist` (pid + name only) when that query isn't
  available. macOS/Linux (`ps`) are unchanged.

## [2.2.2] - 2026-06-30

### Fixed
- **First-run wizard stored `process-interval` as `0s`** so the top view (`t`)
  process table was empty out of the box — the daemon never pushed a table. The
  wizard now prompts for it and **suggests `2s`**, storing what you accept; the
  runtime/flag default stays `0s` (off) for headless services, so the ADR 0017
  opt-in posture is unchanged. Existing daemons started with `--process-interval`
  are unaffected. Set `0` in the wizard to keep it off.
- **systemd `ExecStart` paths** in the privileged-metrics guide pointed at
  `/usr/bin`; the manual install dir is `/usr/local/bin`. Fixed, with a note that
  AUR installs use `/usr/bin`.

## [2.2.1] - 2026-06-30

### Added
- **AMD GPU metrics** — `gpu.util`, `gpu.vram`, `gpu.temp`, and `power.gpu` now
  work on AMD hardware (discrete Radeon and the Strix Halo iGPU, e.g. an HP ZBook
  Ultra G1a). GPU collection was `nvidia-smi`-only, so AMD hosts showed a blank
  accelerator panel. The daemon now prefers `amd-smi` (ROCm) and falls back to
  the in-tree `amdgpu` sysfs nodes (`gpu_busy_percent`, `mem_info_vram_*`, hwmon
  `power1_average` / `temp1_input`) — both unprivileged, so it works out of the
  box with no extra install. `npu.util` is advertised as `unavailable` on XDNA
  until the `amdxdna` driver exposes a utilisation counter.
- **Richer GPU panel** — `gpu.clock` (graphics clock), `gpu.mem.util` (memory-
  controller utilisation), and `gpu.fan` (fan speed) on both NVIDIA (`nvidia-smi`)
  and AMD (`amd-smi` / amdgpu hwmon `freq1_input`). The top-view GPU panel now
  shows clock / mem / fan; anything the chip doesn't expose renders `—` rather
  than a fabricated value.
- **Per-platform privileged-metrics guides** — split the one big guide into
  [macOS](docs/guides/13-privileged-macos.md),
  [Linux + NVIDIA](docs/guides/14-privileged-linux-nvidia.md),
  [Linux + AMD](docs/guides/15-privileged-linux-amd.md), and
  [Windows](docs/guides/16-privileged-windows.md) (with `amd-smi` install steps
  and Windows service setup). The original stays as the cross-platform overview.

## [2.2.0] - 2026-06-30

### Added
- **Full-screen `top` view (Hliðskjálf)** — press `t` on a focused host for a
  mactop/btop-style single-host dashboard: per-core CPU bars, braille sparklines
  for CPU/memory/power, GPU/NPU, network & disk, and a process table. It refreshes
  live and works on macOS, Linux, and Windows. Responsive across four width tiers
  down to an iPhone-portrait "key numbers" column, so it stays readable in Termius
  on a phone. `esc`/`q` exits.
- **`mem.swap`, `cpu.load`, and `cpu.freq` collectors** — swap usage, 1/5/15m load
  average, and CPU clock. Each degrades to Unavailable (rendered `—`) where the
  platform can't supply it (e.g. load average on Windows, clock on Apple Silicon)
  rather than reporting a fabricated 0.

### Changed
- **`t` / `p` keys** — the process table now opens on `p` ("processes"); `t` is the
  new full-screen top view.
- **ANE → NPU** — accelerator power is now the vendor-neutral `power.npu`. The
  legacy `power.ane` key is accepted and normalised on ingest, so a mixed fleet
  with older daemons keeps working.

## [2.1.1] - 2026-06-30

### Fixed
- **macOS power read 0 W on Apple Silicon Pro/Max** (e.g. M3 Max). Those chips
  report 0 for the IOReport per-domain CPU/ANE energy counters and only a
  sub-watt GPU figure, so the package total collapsed to ~0 W — while base
  M-series Macs (which do populate the counters) were fine. The helper now reads
  the SMC `PSTR` ("System Total Power") rail — the same unprivileged source
  mactop/btop use — as the package figure on macOS. Source precedence is now SMC
  → powermetrics → IOReport sum, so a phantom sub-watt IOReport total can no
  longer shadow a real reading. Linux (RAPL/hwmon) and Windows (WMI) paths are
  unchanged; the SMC reader is darwin/cgo-only with a no-op stub elsewhere.

## [2.1.0] - 2026-06-29

> Quality-of-life follow-up to v2.0.0: ephemeral runs that never touch your config,
> and broader agent/CLI onboarding. No wire or behavior changes to the transport.
> Full notes: [docs/releases/v2.1.0.md](docs/releases/v2.1.0.md).

### Added
- **`--no-save` / `--ephemeral`** on every binary — run with the given flags but
  leave the config file untouched, so a one-off (e.g. `heimdall-hub --listen
  :19090 --no-save`) never sticks as the new default.
- **Agent harness snippets for Copilot and other harnesses** in the `heimdall-cli`
  guide — a `.github/copilot-instructions.md` block, plus a portable CLI wrapper
  and an OpenAI-style tool schema for Hermes / OpenAI-compatible / custom harnesses
  (alongside the existing Claude Code AGENT/SKILL/COMMAND files).

### Documentation
- **Run-as-a-service guide** — the Privileged Metrics guide now has a complete
  systemd setup (helper as root, daemon as your user) using a shared `heimdall`
  group and a `/run/heimdall` socket, matching the v2 `0660` helper socket. The
  fleet guide cross-links it.
- **Real agent-session transcripts** in the `heimdall-cli` guide — three unedited
  `claude -p` sessions (free-form, the `/fleet` command, a targeted risk question)
  showing what fleet Q&A looks like in practice, plus a "Scriptable & agent-friendly"
  bullet in the README.

## [2.0.0] - 2026-06-29

> The *everything socket* release. On-demand interaction across the fleet with no
> inbound port on any host. Full notes: [docs/releases/v2.0.0.md](docs/releases/v2.0.0.md).

The release that brings on-demand interaction back to the fleet without giving any
daemon an inbound port. Daemons stay outbound-only; the hub remains the sole
listener and mediates every directive over the daemon's existing stream.

### Added
- **Hub-mediated socket transport (ADR 0018).** The daemon's outbound metric
  stream is reused as a bidirectional control channel, so the hub can push
  directives down to a daemon that never listens:
  - **Demand-driven push (Phase 1).** The hub opens a log/process window on a host
    only while a dashboard or CLI is subscribed, and closes it on the last
    unsubscribe — daemons push observability data on demand instead of always-on.
  - **On-demand commands (Phase 2).** A dashboard or the CLI asks the hub to run an
    allow-listed, **read-only** command on a host; the hub routes it down the
    host's stream, the daemon runs it as its unprivileged user, and the result
    returns correlated by request id. Nothing arbitrary is ever executed.
  - **Helper-delegated privileged commands (Phase 2b).** Commands that need root
    (e.g. `dmesg`, `journal.tail`) are delegated by the daemon to the local
    privileged helper over a unix socket; the helper enforces its **own**
    allow-list and never trusts the daemon. The daemon itself stays least-
    privilege.
- **`heimdall-cli` — a first-class, agent-friendly binary.** Machine-readable JSON
  for `fleet`, `hosts`, `host`, `top`, `logs`, and `run`, built for scripts,
  CI/CD, and AI harnesses. `--hub auto` discovers the hub via zeroconf and, when
  more than one hub is present, reports them and instructs the operator to pick one
  with `--hub <name>`. Shipped with a programmatic/agent guide (bash parsing, a
  CI/CD GitHub Action that waits for a host to come online, Datadog log piping, and
  copy-paste AGENT/SKILL/COMMAND files).
- **On-demand command modal in the dashboard (`c`).** From a host's detail view,
  press `c` to run an allow-listed command and read its result inline. The
  affordance appears only for hosts that advertise the `_cmd` capability — the same
  capability-gated model as logs (`_logs`) and top (`_proc`).
- **Log search and top sorting (ADR 0019).** Inside the log modal, `/` filters the
  buffered and live lines to case-insensitive substring matches, keeping
  timestamps and scoped to the modal. In the top modal, the process table sorts
  CPU-descending by default; `s` opens a sort picker, and the choice re-sorts live
  and **persists to the dashboard config** as the new default.
- **Zeroconf multi-hub discovery (Ratatoskr).** When more than one hub is
  advertised, the dashboard shows a picker and the CLI reports the candidates;
  with a single hub, both connect transparently.
- **Manpages** for every binary (roff `.1` + plain text), generated from each
  binary's `--help`.
- **Real-time online → offline on disconnect.** The hub now acts on a daemon's
  stream ending: it flips the host Offline immediately and pushes a `disconnected`
  snapshot to subscribers, so the dashboard reflects the change at once instead of
  waiting out the freshness window. This covers any detectable socket end — the
  daemon's clean `CloseSend` on SIGTERM *and* an abrupt process death (the OS
  closes the fd either way). The timeout path is retained as the fallback for
  disconnects the hub can't observe (SIGKILL with a frozen network, power loss,
  partition). Additive wire field `Snapshot.disconnected = 13`; old subscribers
  ignore it and fall back to the timeout.
- **Socket-hygiene verification.** A `socket-hygiene.feature` acceptance suite
  proves the model against the *running processes'* real sockets (via `ss`): a
  daemon listens on nothing (no inbound surface), the hub is the sole listener,
  the daemon holds exactly one outbound connection — to the hub — and an on-demand
  command opens **no new socket** (it rides the existing stream). `scripts/
  verify-sockets.sh` runs the same audit on a live host.
- **Documentation screenshot tooling** (`make screenshots`). The `--snapshot`
  path now honours `COLUMNS`/`LINES`, so the generator captures the dashboard at a
  matrix of sizes/views/themes headlessly — including the wide vs. narrow grid
  that shows the responsive column drop — as ANSI (always), styled HTML (`aha`),
  and animated GIFs of the modal flows (`vhs` tapes).
- **Capability gating.** Daemons opt in to what they expose — `--log-source`
  advertises `_logs`, `--process-interval` advertises `_proc`, and
  `--allow-commands` advertises `_cmd`. Reserved `_`-prefixed labels are filtered
  from user tags and grouping. A host shows an affordance only for what it
  advertises.

### Changed
- The helper protocol is now request-based (`collect` | `exec`) with backward
  compatibility: a silent old client falls back to `collect` after a short read
  deadline.

### Wire
- `StreamControl` gains `ObservabilityWindow` and `ControlRequest` directives;
  `Snapshot` gains additive `processes`, `processes_at`, `log_lines`, and
  `command_result` fields; `FederationService.RunCommand` returns a `CommandAck`.
  All additive — old daemons/hubs ignore the new fields, so there is no lockstep
  upgrade.

## [1.6.0] - 2026-06-29

The *Heimdallr's sight* release — host logs and a live process view, inside the
dashboard, with no inbound port on any daemon.

### Added
- **In-dashboard logs (`l`) and top (`t`).** From a host's detail view, press `l`
  to pick a log source and stream it live (scrollable), or `t` for a scrollable,
  refreshing process table. `esc` is the universal back button, unwinding one
  level at a time. The affordances appear only for hosts that expose the
  capability. Explorable in `--demo`.
- **Push-based, hub-mediated observability (ADR 0017).** The daemon — a pure
  outbound producer — tails `--log-source` files and collects a process table on
  `--process-interval`, pushing both to the hub on its existing stream. The hub
  buffers them per host and serves them to dashboards. Nothing connects to a
  daemon; the dashboard talks only to the hub. Cross-platform process collection
  (`ps` on Linux/macOS, `tasklist` on Windows).
- Wire: additive `Snapshot.{processes, processes_at, log_lines}` and a `ProcessRow`
  message. Old daemons/hubs ignore them — no lockstep upgrade.

### Removed
- **BREAKING — the daemon no longer acts as a server.** Daemons are outbound-only
  and must not listen (only hubs do), so the direct daemon-served control plane is
  removed:
  - daemon flags `--control-listen`, `--control-token`, `--control-tls-cert`,
    `--control-tls-key` (use `--process-interval` / `--log-source` to push);
  - dashboard flags `--control`, `--run`, `--tail` (use `l` / `t` in the TUI).
  - `--log-source` is kept, but now configures what the daemon **pushes** rather
    than a served stream.
  On-demand command execution returns with the v2 socket model
  (`feature/sockets`); see ADR 0017 §3.9.

## [1.5.2] - 2026-06-29

### Fixed
- **Dashboard grid clipped on narrow terminals.** The fleet grid forced content
  to ≥ 88 columns and rendered fixed-width columns, so on a narrow screen
  (portrait tablet/phone over SSH) rows were clipped at the right edge and the
  chrome borders fell off-screen. The grid is responsive now: it renders at the
  actual terminal width, drops metric columns right-to-left as it narrows
  (power → gpu → temp → disk → mem, keeping host, state, and CPU longest),
  condenses the state badge to a glyph when very narrow, and falls back to a
  glyph-only footer. No rendered line exceeds the terminal width. Columns are a
  registry (mirroring the grouping dimensions), so adding one is registering a
  column, not editing a width conditional.

## [1.5.1] - 2026-06-28

### Fixed
- **Dashboard overflowed short terminals.** The fleet grid rendered every row
  with no height clamp, so on a screen shorter than the fleet (SSH from a tablet
  or phone) the frame overran the terminal — the header scrolled off and
  ungrouped filtering looked inert because the visible rows never moved. The grid
  now windows host rows to the terminal height, keeps the selected host in view,
  and shows an `↑/↓ N more` indicator for hidden rows. Filtering is visibly
  effective now regardless of grouping or screen size.

### Changed
- **Term-scoped fleet filter.** `/` parses space-separated terms. A bare term
  (`ba`) matches any field — host name, tag value, hub, OS, or state — so it
  surfaces both hosts named `bar`/`baz` and hosts tagged `env=bar`. A scoped term
  narrows to one field: `host=ba`, `env=fo`, `hub=home`, `os=linux`,
  `state=offline`, or `group=<value>` (the active grouping dimension). Multiple
  terms narrow conjunctively; an empty filter shows everything. Searchable fields
  are a strategy registry mirroring the grouping dimensions, so adding a scope is
  registering a matcher, not editing a conditional.

## [1.5.0] - 2026-06-27

### Added
- **Arch Linux packages on the AUR.** The release workflow now publishes each
  binary as a prebuilt (`-bin`) AUR package, so Arch users install with
  `paru -S heimdall-dashboard-bin` (and `heimdall-hub-bin`, `heimdall-daemon-bin`,
  `heimdall-helper-bin`). PKGBUILDs are generated per release from the published
  checksums (`packaging/aur/gen-pkgbuild.sh`); the publish step is gated behind the
  `ENABLE_AUR` repo variable. Packages cover `x86_64` and `aarch64`, and the daemon
  declares the helper as an optional dependency.

## [1.4.0] - 2026-06-27

### Added
- **Mímir durable sink + hub state recovery.** Point the hub at any
  Prometheus-compatible TSDB with `--tsdb <url>`: it persists the fleet via
  Prometheus remote-write *and* restores its last-known state from the TSDB on
  restart — repainting the dashboard (including offline hosts, with their real
  last-seen age) before daemons reconnect, instead of starting blank. Restore is
  best-effort (scalar values + labels + last-seen; info strings and alert state
  reconverge live). Heimdall still embeds no database; the sink is off by default.

### Changed
- Factored the metric→series mapping into `observe.SeriesOf`, shared by the
  OpenMetrics export and the durable sink.

## [1.3.1] - 2026-06-27

### Fixed
- Dashboard *group by OS* showed every host as `(unknown)` on real data. The
  dashboard only receives OS as the `host.os` metric (not in `Context`, which
  never crosses the wire), so grouping now reads the metric's OS family and falls
  back to `Context.OS` (how `--demo` carries it).

## [1.3.0] - 2026-06-27

### Added
- **Yggdrasil — interactive dashboard grouping.** Group the fleet by origin hub,
  OS, or any tag with `g`, and filter/search by host name or tag with `/`. Works
  live and in `--demo`, with per-group section headers.
- **Gjallarhorn — alerts in the dashboard.** Firing alerts now surface as a per-row
  `⚠` badge (the host turns red) and a fleet alert count in the header — not just
  webhook + log.
- **Windows privileged metrics.** The helper reports CPU/zone temperature via WMI
  (`MSAcpi_ThermalZoneTemperature`); `power.pkg` stays unavailable on Windows
  (RAPL needs a kernel driver). Build-tagged, parser unit-tested.
- **Richer demo mode.** The simulated fleet now carries tags, spans three origin
  hubs (`home`, `remote-work-station`, `central`), and runs one host hot enough to
  fire an alert — so grouping, filtering, and the badge are all explorable with no
  setup.

### Changed
- The hub now sends dashboards registry-derived **enriched snapshots** (merged
  labels + alert state) on every update, so live frames match the initial state —
  the origin `hub` label and inherited tags no longer vanish between updates.
- `Snapshot` gains an additive `alerts` field (wire-compatible; old peers ignore it).

## [1.2.2] - 2026-06-27

### Added
- The dashboard can auto-discover its hub over mDNS: `heimdall-dashboard --hub auto`
  (or `--discover`, with `--discover-seed` for overlay networks) — the same Ratatoskr
  discovery the daemon uses. Discovery only resolves the address; the enrollment token
  and TLS still gate the connection.

## [1.2.1] - 2026-06-27

### Fixed
- `<binary> update` now elevates when the install directory needs root — `sudo`
  on Linux/macOS, a UAC prompt on Windows — instead of failing with a permission
  error in a system path like `/usr/local/bin`. Affects all four binaries.

### Docs
- New guides: **Metrics export (Mímir)** — scrape Heimdall from Prometheus/Grafana —
  and **Alerting (Gjallarhorn)**.
- Fleet, privileged-metrics, and federation guides updated for discovery (Ratatoskr),
  tags (Realms), and Linux power/thermal.
- README: clearer "what & why", v1.2.0 capabilities, a Contributing section, and a
  changelog link.

## [1.2.0] - 2026-06-27

The *Watch Over All Realms* release — discovery, tags, alerting, export, and
cross-platform power/thermal. Feature codenames are Norse; see the
[glossary](docs/glossary.md) for what they mean and how to say them.

### Added
- **Ratatoskr — zeroconf discovery.** The hub advertises over mDNS
  (`--discoverable`) and a daemon finds it with `--hub auto` or `--discover`
  (`--discover-seed` covers overlay networks with no multicast). Discovery only
  resolves the address — the enrollment token and TLS still gate trust, and an
  explicit `--hub` always wins.
- **Realms — host & hub tags.** `heimdall-daemon --tags env=prod,role=db` and
  `heimdall-hub --tags region=apac`. Tags ride the stream and inherit across the
  Bifröst relay; a host's own tag wins over an inherited hub tag.
- **Mímir — Prometheus/OpenMetrics export + history.** `heimdall-hub
  --metrics-listen :9091` serves `/metrics` (the whole fleet, labeled by host,
  origin hub, and tags) and `/history`, backed by a bounded in-memory ring
  (lost on restart, by design).
- **Gjallarhorn — alerting.** `heimdall-hub --alert-rules rules.json` evaluates
  threshold rules (e.g. `cpu.util > 90 for 5m`) with for-duration hysteresis so
  spikes don't flap, scoped by tag, and POSTs to `--alert-webhook` on fire and
  clear.
- **Yggdrasil — topology grouping.** Each host's Bifröst origin hub is surfaced
  as a `hub` label and the hub stores enrolled OS/arch context, so the fleet is
  groupable by origin hub, OS, and tags in Prometheus/Grafana.
- **Cross-platform helper parity.** The privileged helper reads CPU package
  power from RAPL and temperatures from hwmon on Linux, so non-Mac hosts get
  `power.pkg`/`temp.pkg` like Apple Silicon.

### Changed
- `Snapshot` gains an additive `labels` field so tags inherit across the relay
  (wire-compatible — old peers ignore it).

### Docs
- ADRs 0009–0012 (Ratatoskr, Realms & Yggdrasil, Mímir, Gjallarhorn), v1.2.0
  Gherkin stories 0012–0017, and a codename pronunciation glossary.

### Notes
- Interactive in-dashboard grouping/filtering (the TUI side of Yggdrasil) is a
  follow-up; the grouping data is available via labels and export today.

## [1.1.1] - 2026-06-27

### Fixed
- Add the missing `scripts/gen-dev-certs.sh` that `make dev-certs` and the TLS
  acceptance test call. It writes a self-signed `hub.crt`/`hub.key` (SANs for
  `localhost` and `127.0.0.1`) usable as both the hub server cert and the client
  CA bundle.
- Make the acceptance suite hermetic: each scenario gets its own
  `HEIMDALL_CONFIG_DIR` and each simulated host its own config dir. Scenarios
  were leaking saved daemon settings (a stale `control-listen`) into each other,
  so co-located daemons collided on a port and only one host came online.

### Changed
- Replace the `switch m.Unit` in the daemon print formatter with a per-unit
  strategy table (no behaviour change).
- Bump GitHub Actions off the deprecated Node 20 runtime: checkout v7, setup-go
  v6, setup-node v6, upload-artifact v7, download-artifact v8, gh-release v3.

## [1.1.0] - 2026-06-27

First release with a consolidated changelog. Rolls up the v1.0.x line and fixes
a print-mode rendering bug in the host inventory metrics.

### Fixed
- `--print` and `--json` modes rendered host inventory info metrics
  (`host.os`, `host.cpu`, `host.gpu`, `host.kernel`, `host.arch`, `host.version`)
  as `=0`. These carry their value as a string in `Detail` with no gauge, and the
  formatter fell through to the numeric default. Print mode now shows the string
  (e.g. `host.os=darwin 27.0`); JSON output carries a `detail` field.

### Docs
- Document the layered config system and first-run wizard.
- Document `--purge-after`, `--install-location`, and disk I/O metrics.
- Add `--version` and `--purge-after` to the configuration reference.

## [1.0.7] - 2026-06-27

### Fixed
- Hub now shuts down on SIGTERM whether or not a dashboard is connected.

## [1.0.6] - 2026-06-27

### Added
- Hub, dashboard, and helper resolve settings through the layered options system.
- Daemon resolves settings through the options system with a first-run wizard.
- Layered config resolution (flags > env > file > defaults) with a first-run wizard.

### Fixed
- Departed hosts are purged to bound memory under churn.

### Docs
- Credit Maple in the README.

## [1.0.5] - 2026-06-27

### Added
- Per-binary installers; system bin as the default install location; `--install-location`.
- Self-update via `<binary> update`.

## [1.0.4] - 2026-06-27

### Added
- Per-NIC network breakdown.
- `--version` on every binary.
- Dashboard shows used/total absolutes for memory, disk, and VRAM.

### Fixed
- Status bar reports the true connection state instead of an optimistic guess.

### Docs
- Add the dashboard demo gif to the README.

## [1.0.3] - 2026-06-27

### Added
- Host inventory (OS, CPU, GPU, kernel, arch, version) in the detail view.

### Changed
- Split the adapters package into one adapter per file.

### Fixed
- macOS release binaries build with CGO so IOReport-backed metrics work.

## [1.0.2] - 2026-06-27

### Added
- Disk read/write throughput.

### Fixed
- Dashboard surfaces GPU power, temperature, and VRAM.
- Detail sparklines are capped to their column width.

### Docs
- Drop maintainer release and brand notes from the README.

## [1.0.1] - 2026-06-27

### Fixed
- Hosts age from the hub snapshot timestamp, not dashboard wall-clock time.

### CI
- Release workflow publishes on the release event with idempotent asset upload.

## [1.0.0] - 2026-06-27

Initial release. A fleet metrics daemon, an aggregating hub, and a terminal
dashboard, streaming over mTLS gRPC.

### Added
- Daemon: collects host metrics (CPU, memory, disk, network, temperature, power,
  GPU) and streams them to a hub, or prints them locally (`--print`, `--once`, `--json`).
- Hub: aggregates snapshots from many daemons and fans out to dashboards.
- Dashboard: terminal UI with a fleet grid and per-host detail drilldown.
- Optional privileged helper for metrics that need elevated access.
- Demo mode with a synthetic fleet.
- Install script, release script, and release workflow.
- Modality start guides and reference docs.

[Unreleased]: https://github.com/kinncj/Heimdall/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/kinncj/Heimdall/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/kinncj/Heimdall/compare/v1.6.0...v2.0.0
[1.6.0]: https://github.com/kinncj/Heimdall/compare/v1.5.2...v1.6.0
[1.5.2]: https://github.com/kinncj/Heimdall/compare/v1.5.1...v1.5.2
[1.5.1]: https://github.com/kinncj/Heimdall/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/kinncj/Heimdall/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/kinncj/Heimdall/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/kinncj/Heimdall/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/kinncj/Heimdall/compare/v1.2.2...v1.3.0
[1.2.2]: https://github.com/kinncj/Heimdall/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/kinncj/Heimdall/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/kinncj/Heimdall/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/kinncj/Heimdall/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/kinncj/Heimdall/compare/v1.0.7...v1.1.0
[1.0.7]: https://github.com/kinncj/Heimdall/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/kinncj/Heimdall/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/kinncj/Heimdall/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/kinncj/Heimdall/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/kinncj/Heimdall/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/kinncj/Heimdall/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/kinncj/Heimdall/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/kinncj/Heimdall/releases/tag/v1.0.0
