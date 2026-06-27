# Alerting (Gjallarhorn)

The hub can evaluate threshold rules against the same live snapshots that drive
the dashboard, and POST an event to a webhook when a rule fires and again when it
clears. A breach must persist before it fires, so brief spikes stay quiet.

Named after [Gjallarhorn](../glossary.md),
the horn Heimdall sounds to warn the gods that danger is coming.

## 1. Write a rules file

Rules are a JSON array. Each rule is one threshold over one metric:

```json
[
  {
    "name": "cpu-hot",
    "metric": "cpu.util",
    "op": ">",
    "threshold": 90,
    "for": "5m",
    "match": {"env": "prod"}
  }
]
```

| Field | Meaning |
|---|---|
| `name` | rule id, used in the event payload |
| `metric` | metric name to watch, e.g. `cpu.util` |
| `op` | comparison: `>`, `<`, `>=`, `<=` |
| `threshold` | numeric value to compare against |
| `for` | duration the breach must hold (`5m`, `30s`); empty = fire immediately |
| `match` | optional tag selector — all keys must match the host's effective tags |

## 2. Start the hub with alerting on

```sh
./bin/heimdall-hub --listen :9090 \
  --alert-rules ./rules.json \
  --alert-webhook https://hooks.example.com/heimdall &
```

`--alert-rules` empty disables alerting entirely. `--alert-webhook` is where fire
and clear events are POSTed.

## Webhook payload

Each transition is a JSON POST — once when the alert starts firing, once when it
resolves:

```json
{
  "rule": "cpu-hot",
  "host": "web-01",
  "state": "firing",
  "metric": "cpu.util",
  "value": 95.4,
  "at": "2026-06-27T19:05:43Z"
}
```

`state` is `"firing"` on breach and `"resolved"` when the value comes back. The
shape is generic enough for a Slack or PagerDuty bridge.

## For-duration hysteresis

A rule with `"for": "5m"` only fires once the breach has held continuously for
five minutes. If the value dips back under the threshold before then, the timer
resets and nothing fires. This stops a single noisy sample from paging anyone. An
empty `for` fires on the first breaching sample.

## Tag scoping

`match` restricts a rule to hosts whose [tags (Realms)](../glossary.md)
contain every listed key/value. The example above only watches `env=prod` hosts;
staging and dev are ignored by the same rule. Tags come from the daemon's
`--tags` and any inherited hub tags — see
[Monitor a Fleet → Tag your hosts](02-monitor-a-fleet.md).

## Alternative: alert in Grafana

If you already scrape the [Mímir export](09-metrics-export.md), you can write the
same thresholds as Grafana/Prometheus alert rules off the `heimdall_*` series
instead of running Gjallarhorn on the hub. Use the in-hub rules when you want
alerting with no extra stack; use Grafana when you already have one.

## Background

See [ADR 0012 — Threshold alerting on the hub](../architecture/).

## Next steps

- Where the metrics come from → [Metrics Export (Mímir)](09-metrics-export.md)
- Full flag reference → [Configuration](../configuration.md)
