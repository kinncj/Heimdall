# Log Streaming — Pushed, In-Dashboard

A daemon can tail explicitly configured log sources and **push** their lines to the
hub on its existing stream. The dashboard reads them from the hub and shows them
under the host detail view's **`l`** key. The daemon never listens; logs ride the
same outbound connection as metrics but as distinct payloads, so a noisy source
can never starve metrics.

See [ADR 0017](../architecture/0017-heimdallr-sight-in-dashboard-observability.md).

## Opt-in by design

Logs stay **off** until you register a source alias. A daemon started with no
`--log-source` pushes nothing.

## Enable sources on a host

Sources are `alias=path` pairs:

```sh
heimdall-daemon --hub station:9090 --name "$(hostname)" \
  --log-source "app=/var/log/app.log,sys=/var/log/system.log"
```

## View them in the dashboard

Select a host with `↑/↓` and `⏎`, then press **`l`**. Pick a source from the list
with `↑/↓` and `⏎`; the same modal then streams that source live:

```text
  LOG — web-01 / app

  15:05:43  boot sequence started
  15:05:43  watch over all realms
  15:05:44  guarding the bifrost
  ↑/↓ scroll  esc sources
```

`esc` steps back to the source list; `esc` again returns to the host detail view —
the app's universal back button. `l` only appears for hosts that push logs.

## Rate limiting & bandwidth

Each push carries only the lines tailed since the last snapshot, capped per push so
a noisy source cannot inflate a frame; when lines are dropped to honour the cap, the
next delivered line is flagged `[rate-limited]`. The hub keeps a bounded ring per
host; the dashboard tails live from connect time.

## Migration

Log streaming was a daemon-served gRPC stream reached by
`heimdall-dashboard --control HOST --tail app`. That path is **removed in v1.6.0**
(daemons no longer listen):

| Removed | Replacement |
|---|---|
| `heimdall-daemon --log-source … ` *served on `--control-listen`* | `--log-source …` *pushed to the hub* |
| `heimdall-dashboard --control HOST --tail app` | press `l` in the host detail view |

## Next steps

- The process view (top) works the same way → [Process View](06-control-plane.md)
- Full flag reference → [Configuration](../configuration.md)
