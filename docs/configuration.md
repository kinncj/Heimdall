# Configuration Reference

Every Heimdall binary is configured by command-line flags, a few environment
variables, and a saved JSON config file. Run any binary with `--help` for its
built-in usage. This page is the complete reference.

Settings resolve in this order, highest precedence last:

```
defaults  <  config file  <  environment  <  flags
```

The first run on a terminal offers a wizard; see
[Configuration file & first-run wizard](#configuration-file--first-run-wizard).

Every binary also accepts `--version` (print the binary name and version, then
exit) and `--help`.

> **Secrets**: prefer the environment variables (`HEIMDALL_TOKEN`,
> `HEIMDALL_CONTROL_TOKEN`, `HEIMDALL_UPSTREAM_TOKEN`) over flags so tokens stay out
> of process listings and shell history. Where both are set, the flag wins.

---

## heimdall-hub

The central server. Receives metric streams from daemons and fans snapshots out to
dashboards. Optionally relays upstream to a parent hub (federation).

| Flag | Default | Meaning |
|---|---|---|
| `--listen` | `:9090` | gRPC listen address |
| `--id` | hostname | federation id — origin of local hosts, appended to relay paths |
| `--tags` | — | hub tags `k=v,k2=v2` (Realms); inherited by this hub's hosts (a host's own tag wins), e.g. `region=apac,tier=edge` |
| `--token` | env `HEIMDALL_TOKEN` | required enrollment token; empty disables auth |
| `--tls-cert` | — | PEM server certificate (enables TLS with `--tls-key`) |
| `--tls-key` | — | PEM server private key |
| `--stale-after` | `10s` | mark a host stale after no updates for this long |
| `--offline-after` | `30s` | mark a host offline after no updates for this long |
| `--purge-after` | `15m` | drop a host from the registry after it has been unseen this long (`0` disables) |
| `--upstream` | — | parent hub address to relay this hub's hosts to |
| `--upstream-token` | env `HEIMDALL_UPSTREAM_TOKEN` | enrollment token for the parent |
| `--upstream-tls` | `false` | relay to the parent over TLS |
| `--upstream-tls-ca` | — | PEM CA bundle to trust for the parent |
| `--upstream-tls-server-name` | — | override the verified server name |
| `--upstream-tls-insecure` | `false` | dev only — skip parent verification |
| `--relay-interval` | `2s` | how often to relay hosts upstream |
| `--metrics-listen` | — | serve Prometheus/OpenMetrics + history on this address (e.g. `:9091`); empty = off (Mímir) |
| `--discoverable` | `false` | advertise this hub over mDNS so daemons can auto-discover it (Ratatoskr) |
| `--alert-rules` | — | path to a JSON alert-rules file; empty = alerting off (Gjallarhorn) |
| `--alert-webhook` | — | URL to POST alert events to on fire/clear (Gjallarhorn) |

---

## heimdall-daemon

Runs on each monitored host. Collects metrics and, with `--hub`, streams them.
Without `--hub`, it prints samples locally (see print mode below).

### Core

| Flag | Default | Meaning |
|---|---|---|
| `--hub` | — | hub address to stream to; **empty prints locally**; `auto` = discover via mDNS (Ratatoskr) |
| `--discover` | `false` | auto-discover the hub via mDNS when `--hub` is unset (explicit `--hub` always wins) |
| `--discover-seed` | — | fallback hub address for discovery on overlay networks (Tailscale, etc.) with no multicast |
| `--name` | hostname | host display name shown in the dashboard |
| `--interval` | `2s` | sample interval |
| `--ping-target` | `1.1.1.1` | internet host pinged for `net.latency` |
| `--tags` | — | host tags `k=v,k2=v2` (Realms), e.g. `env=prod,role=db`; shown in the dashboard and exported as labels |

### Print mode (no `--hub`)

| Flag | Default | Meaning |
|---|---|---|
| `--once` | `false` | collect a single sample and exit |
| `--json` | `false` | emit one JSON object per metric |

### Transport security (to the hub)

| Flag | Default | Meaning |
|---|---|---|
| `--token` | env `HEIMDALL_TOKEN` | enrollment token presented to the hub |
| `--tls` | `false` | connect to the hub over TLS |
| `--tls-ca` | system roots | PEM CA bundle to trust |
| `--tls-server-name` | — | override the verified server name |
| `--tls-insecure` | `false` | dev only — skip hub verification |

### Control plane & logs (optional)

| Flag | Default | Meaning |
|---|---|---|
| `--control-listen` | — | serve the read-only control plane on this address (e.g. `:9100`) |
| `--control-token` | env `HEIMDALL_CONTROL_TOKEN` | token required to invoke control commands |
| `--control-tls-cert` | — | PEM server cert for the control plane |
| `--control-tls-key` | — | PEM server key for the control plane |
| `--log-source` | — | opt-in log sources `alias=path,…` (served on `--control-listen`; empty = logs off) |

See [Control Plane](guides/06-control-plane.md) and
[Log Streaming](guides/07-log-streaming.md).

### Daemon logging

`--log-file` governs the daemon's **entire** output — operational logs, the
control-plane audit trail, and metric samples — as structured JSON (one event per
line; the format Splunk, Datadog, and Kibana ingest directly).

| `--log-file` | Behaviour |
|---|---|
| unset (default) | metric samples on stdout, JSON logs on stderr (the terminal) |
| `false` (or `off` / `none` / `0`) | no output at all — fully silent |
| a path | everything appended to that file as JSON lines |

```sh
heimdall-daemon --hub station:9090 --log-file /var/log/heimdall/daemon.json
```

---

## heimdall-dashboard

The TUI. Pure presentation — it subscribes to a hub and renders; it never collects
metrics itself.

### Display

| Flag | Default | Meaning |
|---|---|---|
| `--hub` | `localhost:9090` | hub address to subscribe to; `auto` = discover via mDNS (Ratatoskr) |
| `--discover` | `false` | auto-discover the hub via mDNS when `--hub` is auto/unset |
| `--discover-seed` | — | fallback hub address for discovery on overlay networks (Tailscale, etc.) |
| `--demo` | `false` | render a simulated fleet (no hub needed) |
| `--mode` | `dark` | theme mode: `dark` or `light` |
| `--purge-after` | `15m` | drop a host from the grid after it has been unseen this long (`0` disables) |

### One-shot rendering (no TTY)

| Flag | Default | Meaning |
|---|---|---|
| `--snapshot` | `false` | render one grid frame to stdout and exit |
| `--detail` | `false` | render the host-detail frame (with `--snapshot`) |
| `--splash` | `false` | render the brand splash frame and exit |

### Transport security (to the hub)

Same `--token`, `--tls`, `--tls-ca`, `--tls-server-name`, `--tls-insecure` as the
daemon.

### Control plane client

| Flag | Meaning |
|---|---|
| `--control <addr>` | daemon control-plane address |
| `--run <cmd>` | run an allow-listed command (e.g. `process.list`, `"dir.list /var/log"`) |
| `--tail <alias>` | tail an opt-in log source alias; streams until `Ctrl-C` |

### Dashboard keys

`↑`/`↓` select · `⏎` detail · `r` refresh · `?` help · `q` quit.

---

## heimdall-helper

Optional privileged sidecar on a host. Serves power/GPU/thermal to the local
daemon over a unix socket. See [Privileged Metrics](guides/04-privileged-metrics.md).

| Flag | Default | Meaning |
|---|---|---|
| `--socket` | OS temp dir | unix socket to serve on (the daemon uses the same default) |
| `--demo` | `false` | serve canned sample metrics (no root needed; for trying the UI) |

---

## Configuration file & first-run wizard

Each binary persists its settings to a JSON file named after the binary —
`daemon.json`, `hub.json`, `dashboard.json`, `helper.json` — in the Heimdall
config directory, resolved in this order:

| Lookup | Path |
|---|---|
| `$HEIMDALL_CONFIG_DIR` (if set) | the value verbatim |
| `$XDG_CONFIG_HOME` (if set) | `$XDG_CONFIG_HOME/heimdall` |
| Linux / macOS | `~/.config/heimdall` |
| Windows | `%AppData%\heimdall` |

The file is one JSON object, one key per setting (toggles as booleans). It is
written `0600` (owner-only) because it can hold tokens.

```json
{
  "hub": "station:9090",
  "interval": "2s",
  "ping-target": "1.1.1.1",
  "tls": true
}
```

**When the file is written.** A binary saves its resolved settings when the
first-run wizard runs, or when you pass any setting flag — so
`heimdall-daemon --hub station:9090` both runs once and records `hub` for next
time. It prints `saved config to <path>` on stderr when it does.

**First-run wizard.** On the first run, if no config file exists, no flags were
passed, and stdin is a terminal, each binary walks through its main settings,
showing the current value as the `[default]` — press Enter to accept. The wizard
never runs under pipes, CI, or one-shot modes (`--once`, `--snapshot`), so
scripted use is unaffected. Delete the config file to run it again.

Precedence still applies after the file exists: a flag or environment variable
overrides the saved value for that run without rewriting the file (unless a
setting flag triggers a save).

## Environment variables

| Variable | Used by | Equivalent flag |
|---|---|---|
| `HEIMDALL_TOKEN` | hub, daemon, dashboard | `--token` |
| `HEIMDALL_CONTROL_TOKEN` | daemon | `--control-token` |
| `HEIMDALL_UPSTREAM_TOKEN` | hub | `--upstream-token` |
| `HEIMDALL_HELPER_SOCKET` | helper, daemon | `--socket` |
| `HEIMDALL_CONFIG_DIR` | all | overrides the config directory (see above) |

## Ports

| Port | Process | Direction |
|---|---|---|
| `9090` (default) | hub | daemons and dashboards connect **in** |
| `9100` (example) | daemon control/logs | dashboards connect **in** (only if `--control-listen` is set) |
| unix socket | helper ↔ daemon | local only, never on the network |

Daemons and dashboards make **outbound** connections; hosts need no inbound ports
unless you enable the control plane.
