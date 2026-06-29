# Hub CLI — Machine & AI-Friendly Fleet Queries

`heimdall-hub cli` turns the hub binary into a one-shot, JSON-emitting client for a
running hub. It subscribes like a dashboard, gathers the current fleet state,
prints JSON to stdout, and exits — so scripts and AI agents can consume the fleet
without the TUI.

## Usage

```sh
heimdall-hub cli [--hub addr] [--token t] [--tls …] <command> [args]
```

Defaults to `--hub localhost:9090`. All output is JSON on **stdout**; errors are
JSON on **stderr** with a non-zero exit.

## Commands

| Command | Output |
|---|---|
| `fleet` | summary: total, counts `by_state`, `host_ids` |
| `hosts` | array of every host: state, user labels, metrics, `has_logs`/`has_processes`, `log_sources` |
| `host <id>` | one host in full, including its `processes` |
| `top <id>` | the host's latest process table |
| `logs <id> [source]` | the host's buffered log lines (optionally one source) |

Reserved (`_`-prefixed) capability labels are not exposed as labels; they surface
as the `has_logs` / `has_processes` / `log_sources` fields instead.

## Examples

```sh
# every offline host id
heimdall-hub cli hosts | jq -r '.[] | select(.state=="offline").id'

# hottest processes on a host
heimdall-hub cli top dgx-spark | jq '.processes | sort_by(-.cpu_pct)[:5]'

# error lines from a source
heimdall-hub cli logs web-01 app | jq -r '.lines[] | select(.line|test("error")).line'

# fleet health at a glance
heimdall-hub cli fleet
```

## Notes

- The data is what the hub already has: logs/processes appear only for hosts that
  push them (`heimdall-daemon --log-source` / `--process-interval`). See the
  [process view](06-control-plane.md) and [log streaming](07-log-streaming.md)
  guides.
- `--wait` (default 800ms) bounds how long the CLI gathers state before printing;
  raise it on a slow link or a large fleet.
- It is read-only — it only subscribes; it never mutates the hub or any host.

## Next steps

- Full flag reference → [Configuration](../configuration.md)
