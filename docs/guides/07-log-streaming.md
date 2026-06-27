# Log Streaming — Opt-In, Rate-Limited

A daemon can tail explicitly configured log sources and stream them to an operator
on a **separate** gRPC service — independent of the metric stream, so log volume
can never starve metrics.

## Opt-in by design

Logs stay **off** until you register a source alias. A daemon started with no
`--log-source` streams nothing, and only registered aliases are tail-able. An
unknown alias yields no data.

## Enable sources on a host

Sources are `alias=path` pairs, served on the daemon's `--control-listen`
endpoint and protected by its token:

```sh
heimdall-daemon --hub station:9090 --name "$(hostname)" \
  --control-listen :9100 --control-token "$T" \
  --log-source "app=/var/log/app.log,sys=/var/log/system.log"
```

## Tail from the dashboard

```sh
heimdall-dashboard --control HOST:9100 --token "$T" --tail app
```

Lines stream until you press `Ctrl-C`:

```text
19:05:43 app  boot sequence started
19:05:43 app  watch over all realms
19:05:44 app  guarding the bifrost
```

## Rate limiting

The server caps the line rate to protect a low-bandwidth link. When lines are
dropped to honour the cap, the next delivered line is flagged
`[rate-limited]` so you know there was a gap rather than silence.

## Independence from metrics

Log tailing runs on its own gRPC service (`LogStreamService`) on the same daemon
endpoint as the control plane, but on a distinct channel. A noisy log source
cannot back-pressure or delay metric snapshots.

## Background

See the story
[`docs/stories/opt-in-log-streaming-…`](../stories) and the
[control plane guide](06-control-plane.md), which shares the same endpoint and
token.

## Next steps

- Forward the daemon's own logs to an aggregator → [Configuration → Logging](../configuration.md#daemon-logging)
- Full flag reference → [Configuration](../configuration.md)
