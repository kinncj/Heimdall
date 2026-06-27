# Metrics Export (Mímir) — Prometheus & Grafana

The hub can expose every host's live readings in Prometheus text-exposition
format, so an existing observability stack scrapes Heimdall like any other target.
A bounded in-memory history makes short-range trends queryable too.

Named after [Mímir](../glossary.md), the keeper of the
well of memory beneath Yggdrasil.

## Turn it on

The export is **off** until you give the hub an address to serve it on:

```sh
./bin/heimdall-hub --listen :9090 --metrics-listen :9091 &
```

`--metrics-listen` serves two endpoints; an empty value disables both.

| Endpoint | Returns |
|---|---|
| `GET /metrics` | Prometheus/OpenMetrics text for every host |
| `GET /history?host=<id>&metric=<name>` | JSON trend samples for one series |

## Scrape it

```sh
curl -s localhost:9091/metrics
```

```text
# TYPE heimdall_cpu_util gauge
heimdall_cpu_util{host="web-01",hub="station-1",env="prod"} 42
# TYPE heimdall_host_up gauge
heimdall_host_up{host="web-01",hub="station-1",env="prod",state="online"} 1
```

Each series carries the host id, the origin hub as `hub`, and all of the host's
effective [tags (Realms)](../glossary.md). Dotted metric
names become valid identifiers (`cpu.util` → `heimdall_cpu_util`). Per-core
metrics fan out to one series per core with a `core="N"` label. Non-OK readings
and info-only metrics are **not** emitted as numeric series.

## Prometheus scrape_config

```yaml
scrape_configs:
  - job_name: heimdall
    static_configs:
      - targets: ["station-1:9091"]
```

## Grafana

Point a Grafana dashboard at the Prometheus data source above and graph the
`heimdall_*` series. Group or filter by the `hub` and tag labels (e.g.
`env="prod"`) to slice the fleet without changing the daemons.

## Trends without Prometheus

For a quick trend without a scraper, hit `/history` directly:

```sh
curl -s 'localhost:9091/history?host=web-01&metric=cpu.util'
```

The history is **bounded and in-memory**: it is lost on hub restart, by design.
Live values rebuild within seconds from the next keyframe, so a restart costs you
recent trend depth, not current state.

## Federation

A parent hub's `/metrics` includes the hosts relayed up to it from child hubs.
Scrape at **one level** — either each site hub or the parent — or dedupe by the
`hub` label so a federated host is not counted twice. The `hub` label is the
host's authoritative origin hub, which makes the whole federated fleet groupable
from a single scrape.

## Background

See [ADR 0011 — OpenMetrics export and bounded history](../architecture/) and the
[Federation guide](05-federation.md) for how the `hub` label is stamped.

## Next steps

- Raise alarms off these thresholds → [Alerting (Gjallarhorn)](10-alerting.md)
- Full flag reference → [Configuration](../configuration.md)
