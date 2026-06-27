# Demo Mode — Try the UI With No Setup

Want to see the dashboard without standing up a hub or daemons? Demo mode renders a
simulated multi-host fleet so you can explore the interface immediately.

## Run it

```sh
./bin/heimdall-dashboard --demo
```

Or via Make:

```sh
make run-demo
```

You'll see a synthetic fleet (workstation, dgx-spark, rpi-5, …) with hosts in
various states — `ONLINE`, `STALE`, `OFFLINE` — and live-moving gauges, so you can
try every part of the UI.

## What to try

| Key | Action |
|---|---|
| `↑` / `↓` | move the selection |
| `⏎` | open per-host detail — per-core CPU strip, sparklines, network, gateway, uptime |
| `r` | refresh now |
| `?` | toggle the help overlay |
| `q` | quit |

## Render a single frame (no TTY)

Useful for screenshots, docs, or piping:

```sh
./bin/heimdall-dashboard --demo --snapshot          # one grid frame to stdout
./bin/heimdall-dashboard --demo --detail --snapshot # one host-detail frame
./bin/heimdall-dashboard --splash                   # the brand splash frame
```

## Theme

```sh
./bin/heimdall-dashboard --demo --mode light   # or: dark (default)
```

## Important

Demo data is **simulated** — it is not your machine and not a real fleet. To
monitor real hardware, start with the [Quickstart](01-quickstart.md).

## Next steps

- Monitor this machine for real → [Quickstart](01-quickstart.md)
- Monitor many machines → [Monitor a Fleet](02-monitor-a-fleet.md)
