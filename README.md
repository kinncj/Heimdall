<p align="center">
  <img src="assets/LOGO_NO_BG.png" alt="Heimdall" width="320">
</p>

<h1 align="center">Heimdall</h1>

<p align="center"><em>Watch Over All Realms</em></p>

<p align="center">
  <img alt="license" src="https://img.shields.io/badge/license-AGPL--3.0-blue">
  <img alt="go" src="https://img.shields.io/badge/go-1.26-00ADD8">
  <img alt="ui" src="https://img.shields.io/badge/ui-terminal%20(Bubble%20Tea)-5fd7ff">
</p>

---

Heimdall is a lightweight, cross-platform **hardware monitoring** system with a real-time
**terminal dashboard**. Unprivileged daemons stream metrics from every host over a low-bandwidth
gRPC link to a central hub; a btop/mactop-class Go TUI renders the fleet live. It **sees** (metrics)
and **hears** (opt-in logs) across every realm.

## Capabilities

- **Cross-platform daemon** — Windows/macOS/Linux × amd64/arm64; enrolls over TLS, auto-reconnects, low bandwidth.
- **SOLID metric adapters** — CPU (with per-core), memory, disk, network throughput, temperature, GPU, power, internet + per-NIC gateway latency, uptime, host context. New signals add without touching existing adapters; a failing adapter is isolated, never dropping the host.
- **Unprivileged by default** — Apple Silicon GPU/power via IOReport (no sudo) and an optional privileged helper for full thermal/CPU/ANE power; adapters self-report `unavailable` / `needs-helper` rather than failing.
- **Real-time TUI** — live grid, per-host detail, gradient gauges, stale/offline with last-known values; high-fidelity and graceful-degradation render modes.
- **Bifröst federation** — a hub relays upstream (local → cloud); multiple dashboards subscribe to one hub.
- **Read-only control plane** — allow-listed, no-sudo remote queries, audited.
- **Opt-in log streaming** — rate-limited, on its own channel.

## Quick start

Monitor your own machine — build, then run all three pieces locally:

```sh
make build-tui
./bin/heimdall-hub &                                               # central server (:9090)
./bin/heimdall-daemon --hub localhost:9090 --name "$(hostname)" &  # collector
./bin/heimdall-dashboard                                           # the TUI (subscribes to :9090)
```

Just want to see the interface? `./bin/heimdall-dashboard --demo` renders a
simulated fleet — no hub or daemon needed.

→ Full walkthrough: **[Quickstart](docs/guides/01-quickstart.md)**.

## How it works

Heimdall separates **collection** from **presentation**:

```
 host(s):  heimdall-daemon  ─┐
           (+ heimdall-helper, optional, root)
                             ├─►  heimdall-hub  ─►  heimdall-dashboard (×N)
                             │     (station)        (pure presentation)
```

| Binary | Runs on | Job |
|---|---|---|
| `heimdall-hub` | monitoring station | receives metrics, fans out to dashboards |
| `heimdall-dashboard` | monitoring station (any number) | renders the fleet — collects nothing itself |
| `heimdall-daemon` | every host | collects + streams this host's metrics |
| `heimdall-helper` | a host (optional, root) | serves privileged metrics to the local daemon |

Clean Architecture over a single Go module. The versioned gRPC contract in
[`common/proto/monitoring/v1`](common/proto/monitoring/v1) is the single source of truth shared by
every binary. Full design in [`docs/specs/current/architecture.md`](docs/specs/current/architecture.md)
and the [ADRs](docs/architecture/); architecture & operations overview with diagrams in
[`docs/deployment.md`](docs/deployment.md).

## Documentation

Full docs live in **[`docs/`](docs/README.md)**.

**Start guides — by what you want to do:**

| Guide | Use it when you want to… |
|---|---|
| [Quickstart](docs/guides/01-quickstart.md) | watch a single machine (all-in-one) |
| [Monitor a Fleet](docs/guides/02-monitor-a-fleet.md) | watch many hosts from one station |
| [Secure Deployment](docs/guides/03-secure-deployment.md) | require TLS + an enrollment token |
| [Privileged Metrics](docs/guides/04-privileged-metrics.md) | unlock power, GPU, and full thermal |
| [Federation (Bifröst)](docs/guides/05-federation.md) | span multiple sites / networks |
| [Control Plane](docs/guides/06-control-plane.md) | run read-only remote diagnostics |
| [Log Streaming](docs/guides/07-log-streaming.md) | tail host logs in the dashboard |
| [Demo Mode](docs/guides/08-demo-mode.md) | explore the UI with no setup |

**Reference:** [Installation](docs/installation.md) ·
[Configuration](docs/configuration.md) ·
[Architecture & Operations](docs/deployment.md) ·
[Troubleshooting](docs/troubleshooting.md)

## Install

Prebuilt binaries are published to [GitHub Releases](https://github.com/kinncj/Heimdall/releases)
for Linux/macOS × amd64/arm64 (Windows assets too). Install only what each machine needs —
`heimdall-hub` + `heimdall-dashboard` on the station, `heimdall-daemon` on each host:

```sh
# Dashboard (on the monitoring station)
curl -fsSL https://raw.githubusercontent.com/kinncj/Heimdall/main/scripts/install.sh | sh -s -- dashboard

# Daemon (on each host)
curl -fsSL https://raw.githubusercontent.com/kinncj/Heimdall/main/scripts/install.sh | sh -s -- daemon
```

Or build from source (`make build-tui`). See **[Installation](docs/installation.md)** for
version pinning, install dirs, `go install`, and Windows.

## Build & test from source

```sh
make build-tui        # builds bin/heimdall-{dashboard,daemon,hub,helper}
make run-tui          # dashboard subscribing to a hub (localhost:9090)
make run-demo         # dashboard with a simulated fleet (no hub needed)
make test             # unit tests
make lint             # gofmt + go vet
make test-acceptance  # behave acceptance suite (drives the real binaries)
```

## License

[AGPL-3.0-or-later](LICENSE) © Kinn Coelho Juliao &lt;kinncj@gmail.com&gt;
