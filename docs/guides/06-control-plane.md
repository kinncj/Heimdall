# Remote Control Plane — Read-Only, Allow-Listed

A daemon can expose a small, **read-only** control plane: a fixed allow-list of
safe diagnostic commands an operator can run remotely. Everything is audited, runs
as the unprivileged daemon user, and nothing off the list is ever executed.

## Safety model

- **Allow-list only**: logical command keys map to fixed argument vectors. There is
  no shell, no `sudo`, and no way to run an arbitrary binary.
- **Unprivileged**: commands run as the daemon's user.
- **Audited**: every invocation — allowed or refused — is logged with the actor,
  command, arguments, and decision.
- **Bounded**: output is size-capped and time-limited.

## Built-in commands

| Key | Runs | Notes |
|---|---|---|
| `process.list` | process list | pid, ppid, cpu%, mem%, command |
| `disk.df` | disk usage | human-readable |
| `uptime` | load average + uptime | |
| `dir.list <path>` | directory listing | only under allow-listed roots (e.g. `/var/log`, `/tmp`) |

## Enable it on a host

The control plane shares the daemon's `--control-listen` endpoint and is protected
by a token:

```sh
export HEIMDALL_CONTROL_TOKEN=$(openssl rand -hex 16)

heimdall-daemon --hub station:9090 --name "$(hostname)" \
  --control-listen :9100 \
  --control-token "$HEIMDALL_CONTROL_TOKEN"
```

For TLS on the control endpoint, add `--control-tls-cert` / `--control-tls-key`.

## Run a command from the dashboard

```sh
heimdall-dashboard --control HOST:9100 --token "$HEIMDALL_CONTROL_TOKEN" --run process.list
heimdall-dashboard --control HOST:9100 --token "$HEIMDALL_CONTROL_TOKEN" --run "dir.list /var/log"
```

## What refusal looks like

Anything off the allow-list — `sudo`, an unknown binary, a path outside the
allow-listed roots, or a missing/invalid token — is refused with
`insufficient_permission` and **never executed**:

```text
heimdall-dashboard: refused (METRIC_STATUS_INSUFFICIENT_PERMISSION): command "sudo" is not allow-listed
```

The daemon records a structured audit line for it either way:

```json
{"level":"INFO","msg":"control audit","actor":"alice","command":"sudo","args":["reboot"],"decision":"refuse","exit_code":-1}
```

## Background

See [ADR 0007 — Unprivileged remote control plane](../architecture/0007-unprivileged-remote-control-plane.md).

## Next steps

- Stream logs from the same endpoint → [Log Streaming](07-log-streaming.md)
- Full flag reference → [Configuration](../configuration.md)
