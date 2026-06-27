# Heimdall — Deployment & Operations Guide

How to run Heimdall: what goes on the **monitoring station**, what goes on each
**host**, and how metrics stream to **one or more dashboards**.

## The four binaries

| Binary | Runs on | Privilege | Responsibility |
|---|---|---|---|
| `heimdall-hub` | monitoring station | normal user | Receives metric streams from daemons and fans them out to dashboards. The one process every other piece connects to. |
| `heimdall-dashboard` | monitoring station (any number) | normal user | Pure presentation. Subscribes to a hub and renders the fleet. Collects nothing itself. |
| `heimdall-daemon` | every monitored host | **unprivileged** | Collects this host's metrics (CPU, per-core, memory, disk, temperature, network throughput, internet + per-NIC gateway latency, uptime, GPU/power where available) and streams them to the hub. |
| `heimdall-helper` | a monitored host (optional) | **root** | Exposes privileged metrics (full thermal, CPU/ANE power, extra GPU detail) to the local unprivileged daemon over a unix socket, so the daemon never runs as root. |

Data flow: **daemon → hub → dashboard(s)**. The helper is a local sidecar to the
daemon: **helper → daemon** over a unix socket on the same host.

## Topology

```mermaid
graph TB
    subgraph STATION["Monitoring Station"]
        HUB["heimdall-hub<br/>listens :9090"]
        DASH1["heimdall-dashboard<br/>operator A"]
        DASH2["heimdall-dashboard<br/>operator B / NOC wall"]
    end

    subgraph HOSTA["Host A — Apple Silicon"]
        DA["heimdall-daemon<br/>unprivileged"]
        HA["heimdall-helper<br/>root, optional"]
    end

    subgraph HOSTB["Host B — Linux + NVIDIA"]
        DB["heimdall-daemon<br/>unprivileged"]
    end

    subgraph HOSTC["Host C — Windows"]
        DC["heimdall-daemon<br/>unprivileged"]
    end

    HA -->|"unix socket<br/>privileged metrics"| DA
    DA -->|"gRPC :9090<br/>enroll + stream"| HUB
    DB -->|"gRPC :9090"| HUB
    DC -->|"gRPC :9090"| HUB
    HUB -->|"subscribe<br/>fan-out"| DASH1
    HUB -->|"subscribe<br/>fan-out"| DASH2
```

The hub fans the same live data out to **every** subscribed dashboard, so you can
run as many dashboards as you like (a laptop, a NOC wall display, a teammate's
terminal) against one hub.

## What runs where — at a glance

```mermaid
graph LR
    subgraph "On the monitoring station"
        S1["heimdall-hub"]
        S2["heimdall-dashboard (1..N)"]
    end
    subgraph "On each monitored host"
        H1["heimdall-daemon (required)"]
        H2["heimdall-helper (optional, root)"]
    end
    H1 -->|"streams to"| S1
    S1 -->|"renders in"| S2
    H2 -->|"local socket"| H1
```

## Quick start

### 1. On the monitoring station

```sh
# Receive metrics from hosts (listens on :9090)
./bin/heimdall-hub &

# Watch the fleet (defaults to localhost:9090)
./bin/heimdall-dashboard
```

Run `heimdall-dashboard` again — on the same or another machine — to open a second
view of the same fleet:

```sh
# from a teammate's laptop, pointed at the station
./bin/heimdall-dashboard --hub 192.168.1.50:9090
```

### 2. On each host you want to monitor

```sh
# Replace 192.168.1.50 with the monitoring station's IP
./bin/heimdall-daemon --hub 192.168.1.50:9090 --name "$(hostname)"
```

Each daemon appears as its own row in every dashboard, keyed by `--name`.

### 3. (Optional) Privileged metrics on a host

Only if a metric shows the needs-helper affordance (`⚿`) and you want it without
running the daemon as root:

```sh
sudo ./bin/heimdall-helper &     # serves privileged metrics on a local socket
# the daemon auto-detects the socket — no daemon flag needed
```

### Monitor just your own machine

The station can also be a monitored host — run all three locally:

```sh
./bin/heimdall-hub &
./bin/heimdall-daemon --hub localhost:9090 --name "$(hostname)" &
./bin/heimdall-dashboard
```

## End-to-end sequence

How a host enrolls, streams, and reaches multiple dashboards:

```mermaid
sequenceDiagram
    autonumber
    participant HLP as helper (host, root)
    participant DMN as daemon (host)
    participant HUB as hub (station)
    participant DA as dashboard A
    participant DB as dashboard B

    DA->>HUB: Subscribe
    DB->>HUB: Subscribe
    DMN->>HUB: Enroll (host id + token)
    HUB-->>DMN: Accepted (interval, thresholds)
    loop every sample interval
        HLP-->>DMN: privileged metrics (optional, via socket)
        DMN->>DMN: collect adapters (cpu, mem, disk, net, gateway, ...)
        DMN->>HUB: Snapshot
        HUB-->>DA: Snapshot
        HUB-->>DB: Snapshot
    end
    Note over HUB,DB: One stream in, fanned out to every dashboard
```

## Do I need the helper?

Usually **no** — start with just the daemon and add the helper only if a metric
you want shows `⚿`.

| Platform | GPU / power without helper? | Helper adds |
|---|---|---|
| Apple Silicon (macOS) | Yes — GPU power + utilisation via IOReport, no root | Full thermal, CPU/ANE power (where the SoC exposes it) |
| Linux + NVIDIA | Yes — `nvidia-smi` is readable unprivileged | Vendor-specific extras |
| Other | Depends on the platform tool | Whatever needs root |

> Note: some Apple Silicon SoCs do not expose a CPU package-power counter at all
> (neither IOReport nor `powermetrics`); CPU power reads as unavailable there.

## Networking & ports

- The hub listens on **`:9090`** by default (`--listen` to change). Open this port
  on the monitoring station so hosts can reach it.
- Daemons make **outbound** connections to the hub; dashboards make outbound
  connections to the hub. No inbound ports are needed on the hosts.
- The helper uses a **local unix socket** only — nothing is exposed on the network.

## Security (production)

Unauthenticated plaintext is the default for local development. For any shared or
networked deployment, require a token and enable TLS on all three:

```sh
# Monitoring station
export HEIMDALL_TOKEN=$(openssl rand -hex 16)
make dev-certs            # self-signed cert -> certs/ (dev only; use a real CA in prod)
./bin/heimdall-hub --tls-cert certs/hub.crt --tls-key certs/hub.key --token "$HEIMDALL_TOKEN" &
./bin/heimdall-dashboard --tls --tls-ca certs/hub.crt --token "$HEIMDALL_TOKEN"

# Each host
./bin/heimdall-daemon --hub station:9090 --tls --tls-ca certs/hub.crt --token "$HEIMDALL_TOKEN"
```

A daemon presenting a missing or invalid token is rejected at enrollment and never
registered. See the README "Secure mode" section for details.

## Advanced: multiple sites (federation / Bifröst)

For multiple sites or networks, a site-local hub can relay its hosts to a parent
hub. A central dashboard then sees every host across sites, and loop prevention
keeps cross-linked hubs safe.

```mermaid
graph TB
    subgraph SITE1["Site A"]
        D1["daemons"] --> H1["hub A<br/>:9090"]
    end
    subgraph SITE2["Site B"]
        D2["daemons"] --> H2["hub B<br/>:9090"]
    end
    subgraph CENTRAL["Central / Cloud"]
        HP["parent hub<br/>:9090"]
        DASH["dashboard(s)"]
    end
    H1 -->|"relay upstream"| HP
    H2 -->|"relay upstream"| HP
    HP -->|"subscribe / fan-out"| DASH
```

```sh
# Parent (central) hub
./bin/heimdall-hub --id central --listen :9090 &

# Each site hub relays its hosts upstream to the parent
./bin/heimdall-hub --id site-a --listen :9090 --upstream central-host:9090 &

# Dashboards subscribe to the parent to see every site
./bin/heimdall-dashboard --hub central-host:9090
```

## Dashboard keys

`↑/↓` select host · `⏎` host detail · `r` refresh · `?` help · `q` quit.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| Dashboard shows no hosts | No hub reachable, or no daemons connected | Confirm the hub is running and `--hub` points at it; check the station firewall allows `:9090`. |
| A host never appears | Daemon can't reach the hub | Verify the station IP/port from the host: `nc -vz station 9090`. |
| Metrics show `⚿` (needs-helper) | Privileged metric without a source | Run `sudo heimdall-helper` on that host, or ignore if you don't need it. |
| `Unauthenticated` on the daemon | Token mismatch | Use the same `--token` / `HEIMDALL_TOKEN` on hub, daemon, and dashboard. |
| GPU/CPU power blank on macOS | SoC exposes no counter | Expected on some chips; not a misconfiguration. |
