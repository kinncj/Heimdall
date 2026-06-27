# Changelog

All notable changes to Heimdall are recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
