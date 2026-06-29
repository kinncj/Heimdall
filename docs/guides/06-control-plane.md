# Process View (top) — Pushed, In-Dashboard

Heimdall surfaces a host's process table inside the dashboard, under the detail
view's **`t`** key. It is push-based: the daemon collects the table locally and
pushes it to the hub on its existing stream, and the dashboard reads it from the
hub. The daemon never listens and never runs a command on request.

See [ADR 0017](../architecture/0017-heimdallr-sight-in-dashboard-observability.md).

## Enable it on a host

The daemon collects and pushes a process table when you set a non-zero interval:

```sh
heimdall-daemon --hub station:9090 --name "$(hostname)" \
  --process-interval 5s
```

The table rides the same outbound stream as metrics. No inbound port is opened.

## View it in the dashboard

Connect the dashboard to the hub as usual, select a host with `↑/↓` and `⏎`, then
press **`t`**:

```text
  TOP — dgx-spark   updated 15:04:33

      PID    PPID   CPU%   MEM%  COMMAND
      100       1  49.3%   5.5%  heimdall-daemon
      107       1  33.1%   8.5%  systemd
      ...
  ↑/↓ scroll  esc back
```

`t` only appears for hosts that push a table. `↑/↓` scrolls; `esc` closes.

## Cross-platform

| OS | Source command |
|---|---|
| Linux / macOS | `ps -eo pid,ppid,pcpu,pmem,comm` |
| Windows | `tasklist /FO CSV /NH` (pid + command; cpu/ppid unavailable) |

Collection is read-only and unprivileged. A privileged `heimdall-helper`-backed
source for fuller detail is a planned enrichment (the collector is an interface).

## Bandwidth

The table is pushed only at `--process-interval` (default **off**), independent of
the faster metric tick, and the hub forwards it to dashboards only when fresh. Pick
an interval that matches how closely you watch a host.

## Migration from the removed control plane

The direct daemon-served control plane was **removed in v1.6.0** — daemons are
outbound-only and must not listen (only hubs do). These flags no longer exist:

| Removed | Replacement |
|---|---|
| `heimdall-daemon --control-listen / --control-token / --control-tls-*` | `--process-interval` (push) |
| `heimdall-dashboard --control HOST --run process.list` | press `t` in the host detail view |

On-demand, arbitrary allow-listed commands return with the v2 socket model
(`feature/sockets`); v1 ships the push-only process view. See ADR 0017 §3.9.

## Next steps

- Stream logs the same way → [Log Streaming](07-log-streaming.md)
- Full flag reference → [Configuration](../configuration.md)
