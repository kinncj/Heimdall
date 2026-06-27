# Quickstart — Monitor Your Own Machine

The fastest way to see Heimdall working: run all three pieces on one computer and
watch its live metrics. Takes about two minutes.

## What you'll get

A live terminal dashboard showing this machine's CPU (with per-core detail),
memory, disk, temperature, network throughput, internet + gateway latency,
uptime, and — on supported hardware — GPU and power.

## Prerequisites

- Go 1.26+ (to build from source), or a prebuilt binary set (see
  [Installation](../installation.md)).
- macOS, Linux, or Windows.

## Steps

```sh
# 1. Build the binaries (skip if you installed prebuilt ones)
make build-tui

# 2. Start the hub — the central process every piece talks to
./bin/heimdall-hub &

# 3. Start a daemon on this machine, pointed at the local hub
./bin/heimdall-daemon --hub localhost:9090 --name "$(hostname)" &

# 4. Open the dashboard (it subscribes to localhost:9090 by default)
./bin/heimdall-dashboard
```

You should see one host — this machine — appear as `● ONLINE` with live gauges.

## Why three processes for one machine?

Heimdall separates **collection** from **presentation** (see
[Architecture](../deployment.md)):

- **daemon** collects metrics on a host
- **hub** receives them and fans them out
- **dashboard** renders — it never collects anything itself

The dashboard can run on the same box as the daemon and hub, or on a completely
different machine. This quickstart just puts all three on one host.

## Dashboard keys

| Key | Action |
|---|---|
| `↑` / `↓` | select a host |
| `⏎` | open per-host detail (per-core CPU, sparklines, network, gateway) |
| `r` | refresh now |
| `?` | toggle the help overlay |
| `q` | quit |

## Verify without the TUI

Print one sample straight from the daemon (no dashboard needed):

```sh
./bin/heimdall-daemon --once          # human-readable
./bin/heimdall-daemon --once --json   # one JSON object per metric
```

## Next steps

- Watch more than one machine → [Monitor a Fleet](02-monitor-a-fleet.md)
- Lock it down for a network → [Secure Deployment](03-secure-deployment.md)
- Unlock power/GPU/thermal → [Privileged Metrics](04-privileged-metrics.md)
- Just want to see the UI with no setup → [Demo Mode](08-demo-mode.md)
