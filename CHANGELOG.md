# Changelog

All notable changes to Heimdall are recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.1.0]: https://github.com/kinncj/Heimdall/compare/v1.0.7...v1.1.0
[1.0.7]: https://github.com/kinncj/Heimdall/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/kinncj/Heimdall/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/kinncj/Heimdall/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/kinncj/Heimdall/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/kinncj/Heimdall/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/kinncj/Heimdall/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/kinncj/Heimdall/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/kinncj/Heimdall/releases/tag/v1.0.0
