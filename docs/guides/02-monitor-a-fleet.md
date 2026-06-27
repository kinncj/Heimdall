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

## Resilience

- **Daemon restarts / network blips**: the daemon auto-reconnects with backoff and
  resumes under the same stable HostID — no duplicate rows.
- **A host goes away**: it transitions `ONLINE → STALE → OFFLINE` in the dashboard,
  keeping its last-known values (tune with the hub's `--stale-after` /
  `--offline-after`).

## Running daemons as a service

For always-on hosts, run the daemon under your init system so it survives reboots.

**systemd (Linux)** — `/etc/systemd/system/heimdall-daemon.service`:

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

**launchd (macOS)**: wrap the same command in a LaunchDaemon plist with
`KeepAlive=true`.

## Next steps

- Encrypt and authenticate the link → [Secure Deployment](03-secure-deployment.md)
- Span multiple sites/networks → [Federation](05-federation.md)
- Ship logs alongside metrics → [Log Streaming](07-log-streaming.md)
