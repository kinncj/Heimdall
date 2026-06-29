# Changelog

All notable changes to Heimdall are recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
