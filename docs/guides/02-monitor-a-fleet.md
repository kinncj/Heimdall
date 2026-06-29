# Monitor a Fleet — One Station, Many Hosts

Run a central **monitoring station** and point any number of **hosts** at it. Each
host appears as its own row in the dashboard.

## Topology

```
 host A ──┐
 host B ──┼──►  hub  ──►  dashboard(s)
 host C ──┘   (station)
```

- **Monitoring station**: runs `heimdall-hub` (receives metrics) and one or more
  `heimdall-dashboard` instances (render them).
- **Each host**: runs `heimdall-daemon`, pointed at the station.

See [Architecture & Operations](../deployment.md) for the full picture and
diagrams.

## 1. On the monitoring station

```sh
# Receive metrics from hosts (listens on :9090)
./bin/heimdall-hub --listen :9090 &

# Watch the fleet
./bin/heimdall-dashboard --hub localhost:9090
```

Note the station's IP address — the hosts need it. The hub listens on **:9090**;
open that port in the station's firewall so hosts can reach it.

## 2. On each host you want to monitor

```sh
# Replace 192.168.1.50 with the monitoring station's IP
./bin/heimdall-daemon --hub 192.168.1.50:9090 --name "$(hostname)"
```

Run this on as many hosts as you like. Each `--name` becomes a row in every
dashboard. Daemons make **outbound** connections only — no inbound ports are
needed on the hosts.

## 3. (Optional) Extra dashboards

The hub fans the same live data out to every dashboard. Open as many as you want —
a laptop, a NOC wall display, a teammate's terminal:

```sh
# From anywhere that can reach the station
./bin/heimdall-dashboard --hub 192.168.1.50:9090
```

## Zero-config discovery (Ratatoskr)

Skip hard-coding the station's IP on every host. Make the hub advertise itself
over mDNS, then let daemons find it.

```sh
# On the station: advertise as _heimdall-hub._tcp on the LAN
./bin/heimdall-hub --listen :9090 --discoverable &

# On each host: discover the hub instead of naming it
./bin/heimdall-daemon --hub auto --name "$(hostname)"   # or: --discover
```

- `--hub auto` forces discovery; `--discover` discovers only when no `--hub` is set.
- An explicit `--hub host:port` **always wins** over discovery.
- mDNS needs multicast. On an overlay network with none (Tailscale, etc.), give a
  fallback address:

  ```sh
  ./bin/heimdall-daemon --discover --discover-seed 100.64.0.1:9090 --name "$(hostname)"
  ```

Discovery only finds the **address**. Trust is still gated — the enrollment token
and TLS apply exactly as if you had typed the hub in by hand. See
[Ratatoskr](../glossary.md) in the
glossary.

## Tag your hosts (Realms)

Tags are `k=v` pairs that ride along with a host's metrics and show up as labels
in the [metrics export](09-metrics-export.md) and as
[alert](10-alerting.md) selectors.

```sh
./bin/heimdall-daemon --hub 192.168.1.50:9090 --name web-01 --tags env=prod,role=db
```

A hub can stamp tags onto **every** host it reports, so a whole site inherits a
common label without touching each daemon:

```sh
./bin/heimdall-hub --listen :9090 --tags region=apac,tier=edge &
```

On a key conflict the **host's own tag wins** over the hub's. See
[Realms](../glossary.md).

## Resilience

- **Daemon restarts / network blips**: the daemon auto-reconnects with backoff and
  resumes under the same stable HostID — no duplicate rows.
- **A host goes away**: it transitions `ONLINE → STALE → OFFLINE` in the dashboard,
  keeping its last-known values (tune with the hub's `--stale-after` /
  `--offline-after`).

## Running as a service

For always-on hosts, run Heimdall under your init system so it survives reboots.

### Daemon — systemd (Linux)

`/etc/systemd/system/heimdall-daemon.service`:

```ini
[Unit]
Description=Heimdall metrics daemon
After=network-online.target

[Service]
ExecStart=/usr/local/bin/heimdall-daemon --hub 192.168.1.50:9090 --name %H --log-file /var/log/heimdall/daemon.json
Restart=always
RestartSec=5
User=heimdall

[Install]
WantedBy=multi-user.target
```

```sh
sudo systemctl enable --now heimdall-daemon
```

**Pairing with the root helper** (power/GPU/full thermal): run the daemon as your
user and the helper as root, sharing a `heimdall` group and a `/run/heimdall`
socket. Add `SupplementaryGroups=heimdall` and
`Environment=HEIMDALL_HELPER_SOCKET=/run/heimdall/helper.sock` to the daemon unit —
the complete two-service setup is in
[Privileged Metrics → Run both as systemd services](04-privileged-metrics.md#run-both-as-systemd-services-helper-root-daemon-as-you).

### Hub — systemd (Linux)

The hub is the same pattern with the hub binary. It needs no socket and no shared
group — it talks to nothing local — so a plain system user is enough. Pass the
enrollment token through the environment so it stays out of `ps` and shell history.

```sh
sudo useradd --system --no-create-home --shell /usr/sbin/nologin heimdall
```

`/etc/systemd/system/heimdall-hub.service`:

```ini
[Unit]
Description=Heimdall hub (receives daemon streams, fans out to dashboards)
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=heimdall
# token via env keeps it out of `ps`; see Secure Deployment
Environment=HEIMDALL_TOKEN=<your-token>
ExecStart=/usr/local/bin/heimdall-hub --listen :9090 --id station
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target
```

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now heimdall-hub
journalctl -u heimdall-hub -f          # watch it come up
ss -ltnp 'sport = :9090'               # expect heimdall-hub LISTEN
```

Set `--id` to the host's name so the hub label reads cleanly in the dashboard. Add
`--tls-cert/--tls-key` ([Secure Deployment](03-secure-deployment.md)) when the hub
leaves a trusted network. To put the hub box itself in the grid, run a daemon on it
too, pointed at `--hub localhost:9090` (pair it with the helper as in
[Privileged Metrics](04-privileged-metrics.md#run-both-as-systemd-services-helper-root-daemon-as-you)).

### Daemon — launchd (macOS)

To start the daemon at **boot** — before anyone logs in — install a
**LaunchDaemon**. (A LaunchAgent in `~/Library/LaunchAgents` starts only at *your*
login; use a LaunchDaemon when you want boot.) The daemon stays unprivileged via
`UserName`.

`/Library/LaunchDaemons/com.heimdall.daemon.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.heimdall.daemon</string>
  <key>ProgramArguments</key><array>
    <string>/usr/local/bin/heimdall-daemon</string>
    <string>--hub</string><string>station:9090</string>
    <string>--name</string><string>my-mac</string>
  </array>
  <!-- runs as you, not root -->
  <key>UserName</key><string>YOUR_USER</string>
  <key>EnvironmentVariables</key><dict>
    <key>HEIMDALL_TOKEN</key><string>YOUR_TOKEN</string>
  </dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardErrorPath</key><string>/var/log/heimdall-daemon.log</string>
</dict></plist>
```

LaunchDaemon plists must be owned `root:wheel`; this one holds the token, so lock it
to `0600` before loading:

```sh
sudo chown root:wheel /Library/LaunchDaemons/com.heimdall.daemon.plist
sudo chmod 600 /Library/LaunchDaemons/com.heimdall.daemon.plist
sudo launchctl bootstrap system /Library/LaunchDaemons/com.heimdall.daemon.plist
sudo launchctl print system/com.heimdall.daemon | grep -E 'state|pid'
```

Reload after editing the plist with
`sudo launchctl bootout system/com.heimdall.daemon`, then `bootstrap` again. For
power/GPU/full thermal on Apple Silicon, pair the daemon with the root helper — see
[Privileged Metrics → Run both as launchd services](04-privileged-metrics.md#run-both-as-launchd-services-macos-helper-root-daemon-as-you).

## Next steps

- Encrypt and authenticate the link → [Secure Deployment](03-secure-deployment.md)
- Span multiple sites/networks → [Federation](05-federation.md)
- Ship logs alongside metrics → [Log Streaming](07-log-streaming.md)
