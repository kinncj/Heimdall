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
- **SOLID metric adapters** — CPU, memory, disk, network, temperature, GPU, power, ping/reachability, host context. New signals add without touching existing adapters; a failing adapter is isolated, never dropping the host.
- **Unprivileged by default** — an optional privileged helper unlocks power/RAPL/full-thermal; adapters self-report `unavailable` / `needs-helper` rather than failing.
- **Real-time TUI** — live grid, per-host detail, gradient gauges, stale/offline with last-known values; high-fidelity and graceful-degradation render modes.
- **Bifröst federation** — a hub relays upstream (local → cloud); multiple dashboards subscribe.
- **Read-only control plane** — allow-listed, no-sudo remote queries, audited.
- **Opt-in log streaming** — rate-limited, on its own channel.

## Architecture

Clean Architecture over a single Go module. The versioned gRPC contract in
[`common/proto/monitoring/v1`](common/proto/monitoring/v1) is the single source of truth shared by
every binary. Full design: [`docs/specs/current/architecture.md`](docs/specs/current/architecture.md);
decisions in [`docs/architecture/`](docs/architecture); stories in [`docs/stories/`](docs/stories).

```
app/cmd/{daemon,helper,hub,dashboard}   common/proto/monitoring/v1   infra/   tests/   docs/
```

Binaries: `heimdall-daemon` · `heimdall-helper` · `heimdall-hub` · `heimdall-dashboard`.

## Install

Prebuilt binaries are published to GitHub Releases for Linux/macOS × amd64/arm64 (Windows assets too).

**Dashboard** (the TUI):
```sh
curl -fsSL https://raw.githubusercontent.com/kinncj/heimdall/main/scripts/install.sh | sh -s -- dashboard
```

**Daemon** (per-host collector):
```sh
curl -fsSL https://raw.githubusercontent.com/kinncj/heimdall/main/scripts/install.sh | sh -s -- daemon
```

Override the source repo or install dir with `HEIMDALL_REPO=owner/repo` and `HEIMDALL_BIN_DIR=~/bin`.
Windows users: grab the `*_windows_*.exe` assets from the [Releases](https://github.com/kinncj/heimdall/releases) page.

## Build & run from source

> **New here? See the [Deployment & Operations Guide](docs/deployment.md)** — what
> to run on the monitoring station vs. each host, the helper, multiple dashboards,
> and architecture diagrams.

```sh
make build-tui     # builds bin/heimdall-{dashboard,daemon,hub,helper}
make run-hub       # central gRPC server (:9090)
make run-daemon    # daemon: print this machine's metrics
make run-tui       # dashboard subscribing to a hub (localhost:9090)
make run-demo      # dashboard with a simulated multi-host fleet (no hub needed)
make test && make lint
```

### Monitor one machine (daemon + hub + dashboard)

The dashboard is a presentation surface — it renders what daemons publish to a
hub, and never collects metrics itself. To watch a single machine, run all three
locally (the daemon, hub, and dashboard can share one host):

```sh
./bin/heimdall-hub --listen :9090 &
./bin/heimdall-daemon --hub localhost:9090 --name $(hostname) &
./bin/heimdall-dashboard            # subscribes to localhost:9090 by default
```

### Run a fleet (daemon → hub → dashboard)

```sh
./bin/heimdall-hub --listen :9090 &
./bin/heimdall-daemon --hub localhost:9090 --name $(hostname) &
./bin/heimdall-dashboard --hub localhost:9090
```

Each daemon streams its host's metrics to the hub over the versioned gRPC contract
(auto-reconnecting); the dashboard subscribes for live fan-out.

### Secure mode (TLS + enrollment token)

The hub is unauthenticated and plaintext by default for local development. For any
shared or networked deployment, require an enrollment token and enable TLS so daemons
authenticate and the channel is encrypted:

```sh
make dev-certs                      # self-signed cert -> certs/ (dev only)
export HEIMDALL_TOKEN=$(openssl rand -hex 16)

./bin/heimdall-hub --listen :9090 \
  --tls-cert certs/hub.crt --tls-key certs/hub.key --token "$HEIMDALL_TOKEN" &

./bin/heimdall-daemon --hub localhost:9090 --name $(hostname) \
  --tls --tls-ca certs/hub.crt &      # token read from $HEIMDALL_TOKEN

./bin/heimdall-dashboard --hub localhost:9090 --tls --tls-ca certs/hub.crt
```

A daemon presenting a missing or invalid token is rejected during enrollment and never
registered. A reconnecting daemon re-presents its stable HostID and is matched to the
existing host — no duplicate entries. Both `--token` and the `HEIMDALL_TOKEN` environment
variable are accepted (the flag wins). Issue real per-host certificates from a proper CA
in production; `make dev-certs` is for local use only.

### Privileged helper (power, GPU, full thermal)

Power, GPU, and full thermal metrics need elevated access (Apple Silicon
`powermetrics`, NVIDIA `nvidia-smi`/NVML). Rather than run the daemon as root,
Heimdall ships an optional `heimdall-helper` that runs as its own privileged unit
and serves those readings to the unprivileged daemon over a local unix socket:

```sh
sudo ./bin/heimdall-helper          # serves power/GPU/thermal on the local socket
./bin/heimdall-daemon --hub localhost:9090   # daemon stays unprivileged; reads via the socket
```

Without the helper, power and GPU metrics show a **needs-helper** affordance (`⚿`)
instead of erroring — the daemon keeps running and reports every other metric. To
preview the populated values without root, run the helper in demo mode and read a
sample from the daemon:

```sh
./bin/heimdall-helper --demo &
./bin/heimdall-daemon --once        # GPU and POWER now show sample values
```

Alternatively, on Apple Silicon the daemon reads **GPU power (and CPU/ANE power
where the SoC exposes it) natively via Apple's IOReport energy counters — no sudo
required**, so `heimdall-daemon` shows power unprivileged. Running it as root
additionally fills GPU utilisation via `powermetrics`. Some M-series chips do not
expose a CPU package-power counter at all (neither IOReport nor powermetrics), so
CPU power reads as unavailable there. On Linux, `nvidia-smi` provides GPU metrics
unprivileged.

> IOReport requires a CGO build. The local `make build-tui` enables CGO by default
> on macOS; the CGO-free release binaries fall back to `powermetrics` (which needs
> sudo).

The helper is read-only: the daemon sends nothing and cannot influence what is
collected, so there is no argument-injection or command-execution surface. Power
profile is displayed as read-only information — the dashboard offers no control to
change it.

### Federation (Bifröst) — relay hubs upstream

A site hub can relay its hosts to a parent hub, so a central dashboard sees every
host across sites. Relay is one-directional (child → parent); each hub appends its
id to the snapshot's path and drops any envelope that already contains its own id,
so cross-linked hubs never loop.

```sh
./bin/heimdall-hub --id site-a --listen :9090 &                           # parent
./bin/heimdall-hub --id edge-1 --listen :9091 --upstream localhost:9090 &  # child relays up
./bin/heimdall-daemon --hub localhost:9091 --name edge-host &              # feeds the child
./bin/heimdall-dashboard --hub localhost:9090                              # sees edge-host via the parent
```

The cross-hub link reuses the same auth as any client (`--upstream-token`,
`--upstream-tls`, `--upstream-tls-ca`) and re-authenticates on reconnect; hosts
resume by stable HostID without duplication.

### Remote control plane (read-only, allow-listed)

The daemon can serve a read-only control plane: a fixed allow-list of safe
commands (process list, disk usage, uptime, directory listing under allow-listed
roots). Commands run as the unprivileged daemon user — never `sudo`, never a
shell — and every invocation is audit-logged.

```sh
heimdall-daemon --control-listen :9100 --control-token "$HEIMDALL_CONTROL_TOKEN" &
heimdall-dashboard --control localhost:9100 --token "$HEIMDALL_CONTROL_TOKEN" --run process.list
heimdall-dashboard --control localhost:9100 --token "$HEIMDALL_CONTROL_TOKEN" --run "dir.list /var/log"
```

Anything off the allow-list — `sudo`, arbitrary binaries, paths outside the
allow-listed roots — is refused with `insufficient_permission` and never
executed. The daemon writes an audit line per invocation (actor, command, args,
decision).

### Opt-in log streaming

A daemon can tail explicitly configured log sources and stream them on a
separate gRPC service — independent of the metric stream and rate-limited. Logs
stay off until a source alias is registered:

```sh
heimdall-daemon --control-listen :9100 --control-token "$T" \
  --log-source "app=/var/log/app.log,sys=/var/log/system.log" &
heimdall-dashboard --control localhost:9100 --token "$T" --tail app
```

Only registered aliases are tail-able; an unknown alias — or a daemon started
with no `--log-source` — streams nothing. The server caps the line rate to
protect the low-bandwidth link and flags lines that followed dropped ones.

### Daemon logging

The daemon writes structured JSON logs — one event per line, the format Splunk,
Datadog, and Kibana ingest directly. `--log-file` governs the daemon's entire
output (operational logs, the control-plane audit trail, and metric samples):

| `--log-file` | Behaviour |
|---|---|
| unset (default) | metric samples on stdout, JSON logs on stderr (the terminal) |
| `false` (or `off`/`none`/`0`) | no output at all — fully silent |
| a path | everything appended to that file as JSON lines |

```sh
heimdall-daemon --hub localhost:9090 --log-file /var/log/heimdall/daemon.json
```

In file mode metric samples are emitted as JSON too (with the per-core array),
so a single `--log-file` path hands an aggregator the full operational, audit,
and metric trail.

Cut a release by pushing a tag — the [`release` workflow](.github/workflows/release.yml)
cross-compiles both binaries and uploads them to the GitHub Release:

```sh
git tag v0.1.0 && git push origin v0.1.0
```

## Brand

Steel wordmark, electric-blue signature (`#00d7ff`), near-black surface. Logos and headers live in
[`assets/`](assets) — PNGs for docs, ASCII-art `.txt` for the terminal splash. See
[`docs/design/identity/tui-brand.md`](docs/design/identity/tui-brand.md).

## License

[AGPL-3.0-or-later](LICENSE) © Kinn Coelho Juliao &lt;kinncj@gmail.com&gt;
