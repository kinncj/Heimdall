# `heimdall-cli` — Programmatic & AI-Friendly Fleet Access

`heimdall-cli` is a one-shot, JSON-emitting client for a running hub. It subscribes
like a dashboard, gathers the current fleet state, prints JSON to **stdout**, and
exits — so shell scripts, CI/CD, and AI agent harnesses can consume the fleet
without the TUI. (The same tool is also reachable as `heimdall-hub cli`.)

It is **read-only**: it only subscribes; it never mutates the hub or any host.

## Install

```sh
# one binary, like the others
curl -fsSL https://github.com/kinncj/Heimdall/releases/latest/download/heimdall-cli_linux_amd64 -o heimdall-cli
chmod +x heimdall-cli && sudo mv heimdall-cli /usr/local/bin/
# or: paru -S heimdall-cli-bin     # Arch (AUR)
```

## Usage

```sh
heimdall-cli [--hub addr] [--token t] [--tls …] [--wait 800ms] <command> [args]
```

Defaults to `--hub localhost:9090`. Output is JSON on **stdout**; errors are JSON
on **stderr** with a non-zero exit.

| Command | Output |
|---|---|
| `fleet` | summary: `total`, counts `by_state`, `host_ids` |
| `hosts` | array of hosts: state, labels, metrics, `has_logs`/`has_processes`, `log_sources` |
| `host <id>` | one host in full, including `processes` |
| `top <id>` | the host's latest process table |
| `logs <id> [source]` | the host's buffered log lines (optionally one source) |
| `run <id> <cmd> [args]` | run an **allow-listed, read-only** command on a host (v2) |

`run` commands (read-only, allow-listed, audited on the daemon): `process.list`,
`disk.df`, `uptime`, `os.info`, `dir.list <dir>` (bounded to safe roots), plus the
privileged `dmesg` and `journal.tail`. They are hub-mediated — the dashboard/CLI
asks the **hub**, which routes the request down the host's outbound stream; the
daemon runs unprivileged commands itself and delegates **privileged** ones to the
root **helper** over the local socket (a host with no helper returns
`insufficient_permission`). Anything off the list is refused and never executed.
Requires the daemon to be started with `--allow-commands`.

```sh
heimdall-cli --hub "$HUB" run web-01 disk.df  | jq -r .stdout
heimdall-cli --hub "$HUB" run web-01 os.info   | jq -r .stdout
# exit non-zero if a command was refused
test "$(heimdall-cli --hub "$HUB" run web-01 uptime | jq -r .status)" = ok
```

> `top`/`logs` show data only for hosts that push it (`heimdall-daemon
> --process-interval` / `--log-source`). On a quiet fleet, raise `--wait` (e.g.
> `--wait 3s`) so the CLI catches the next push.

## Bash parsing (jq)

```sh
HUB=hub.internal:9090
export HEIMDALL_TOKEN=…   # if the hub requires a token

# ids of every offline host
heimdall-cli --hub "$HUB" hosts | jq -r '.[] | select(.state=="offline").id'

# one host's CPU%
heimdall-cli --hub "$HUB" host web-01 | jq -r '.metrics["cpu.util"]'

# top 5 processes by CPU on a host
heimdall-cli --hub "$HUB" top dgx-spark | jq '.processes | sort_by(-.cpu_pct)[:5]'

# exit non-zero if any host is offline (handy as a health gate)
test "$(heimdall-cli --hub "$HUB" fleet | jq '.by_state.offline // 0')" -eq 0
```

Because every value is typed JSON, you never screen-scrape a table.

## CI/CD — wait for a host to come online after a release

A common deploy gate: after shipping, block until the target host re-registers and
reports `online`, then continue (or fail the job if it never does).

```yaml
# .github/workflows/verify-online.yml
name: Verify host online after deploy
on:
  release:
    types: [published]

jobs:
  verify:
    runs-on: ubuntu-latest
    env:
      HUB: ${{ secrets.HEIMDALL_HUB }}            # e.g. hub.internal:9090 (reachable from CI)
      HEIMDALL_TOKEN: ${{ secrets.HEIMDALL_TOKEN }}
      HOST: web-01
    steps:
      - name: Install heimdall-cli
        run: |
          curl -fsSL https://github.com/kinncj/Heimdall/releases/latest/download/heimdall-cli_linux_amd64 -o heimdall-cli
          chmod +x heimdall-cli && sudo mv heimdall-cli /usr/local/bin/

      - name: Wait until ${{ env.HOST }} is online
        run: |
          for i in $(seq 1 60); do
            state=$(heimdall-cli --hub "$HUB" host "$HOST" 2>/dev/null | jq -r '.state // "unknown"')
            echo "attempt $i — $HOST: $state"
            if [ "$state" = "online" ]; then
              echo "::notice::$HOST is online"; exit 0
            fi
            sleep 5
          done
          echo "::error::$HOST did not come online within 5 minutes"; exit 1
```

The same loop works in GitLab CI, a Jenkins stage, or a plain `cron` — it's just a
binary and `jq`.

## Pipe logs to Datadog (or anything)

Stream a host's error lines and forward them as Datadog events. This uses the
`dog` CLI from [`datadogpy`](https://github.com/DataDog/datadogpy) (`pip install
datadog`); swap in `datadog-ci`, `curl` to the Events API, or your log shipper.

```sh
HUB=hub.internal:9090

heimdall-cli --hub "$HUB" logs web-01 app \
  | jq -r '.lines[] | select(.line | test("error|ERROR|panic")) | .line' \
  | while IFS= read -r line; do
      dog event post "heimdall web-01 error" "$line" \
        --tags "service:web,host:web-01,source:heimdall" --alert_type error
    done
```

Raw API variant (no extra CLI):

```sh
heimdall-cli --hub "$HUB" logs web-01 app \
  | jq -c '.lines[] | {ddsource:"heimdall", host:.host, service:.source, message:.line}' \
  | while IFS= read -r evt; do
      curl -sS -X POST "https://http-intake.logs.datadoghq.com/api/v2/logs" \
        -H "DD-API-KEY: $DD_API_KEY" -H "Content-Type: application/json" -d "$evt" >/dev/null
    done
```

Run it on a timer (cron/systemd) to tail continuously; each invocation forwards the
hub's buffered window.

---

## Use it from an AI agent / harness (copy-paste)

The files below make an AI harness query the fleet out-of-the-box. The Claude Code
trio (AGENT/SKILL/COMMAND) comes first; equivalents for **GitHub Copilot** and
**any other harness (Hermes, OpenAI-compatible, custom)** follow. Copy each to the
path shown; they assume `heimdall-cli` is on `$PATH` and read `HEIMDALL_HUB` /
`HEIMDALL_TOKEN` from the environment.

### AGENT — `.claude/agents/fleet.md`

```markdown
---
name: fleet
description: >-
  Read-only Heimdall fleet inspector. Use to answer questions about host
  health, state (online/stale/offline), per-host metrics, top processes, or
  recent logs. Calls the heimdall-cli binary; never guesses.
tools: Bash
---

You answer questions about a Heimdall fleet using the `heimdall-cli` binary, which
prints JSON from a running hub. Always run the CLI — never invent values.

Connection (read from the environment, with fallbacks):
- hub: `${HEIMDALL_HUB:-localhost:9090}`
- token: `$HEIMDALL_TOKEN` (omit `--token` if unset)

Commands (read-only, JSON on stdout):
- `heimdall-cli --hub "$HEIMDALL_HUB" fleet`            — counts by state
- `heimdall-cli --hub "$HEIMDALL_HUB" hosts`            — every host + metrics + capabilities
- `heimdall-cli --hub "$HEIMDALL_HUB" host <id>`        — one host in full
- `heimdall-cli --hub "$HEIMDALL_HUB" top <id>`         — process table
- `heimdall-cli --hub "$HEIMDALL_HUB" logs <id> [src]`  — recent log lines
- `heimdall-cli --hub "$HEIMDALL_HUB" run <id> <cmd>`   — allow-listed diagnostic
  (process.list | disk.df | uptime | os.info | dir.list <dir>); read-only, audited

Parse with `jq`. Be concise and factual: name host ids and states explicitly.
If a host is missing, say so — do not assume it exists. For "is X healthy?",
check `.state` and the relevant `.metrics` (cpu.util, mem.used, disk.used) and any
`.alerts`.
```

### SKILL — `.claude/skills/fleet/SKILL.md`

```markdown
---
name: fleet
description: >-
  Query a Heimdall fleet (host state, metrics, top processes, logs) via the
  heimdall-cli JSON client. Use when asked about fleet or host health, what is
  online/offline, what is using CPU, or to read a host's recent logs.
---

# SKILL: fleet

Run `heimdall-cli` (JSON, read-only) against the hub at
`${HEIMDALL_HUB:-localhost:9090}` (token: `$HEIMDALL_TOKEN`).

## Recipes

- Health snapshot: `heimdall-cli --hub "$HEIMDALL_HUB" fleet`
- Offline hosts: `heimdall-cli --hub "$HEIMDALL_HUB" hosts | jq -r '.[]|select(.state=="offline").id'`
- Hot processes: `heimdall-cli --hub "$HEIMDALL_HUB" top <id> | jq '.processes|sort_by(-.cpu_pct)[:5]'`
- Errors in logs: `heimdall-cli --hub "$HEIMDALL_HUB" logs <id> <src> | jq -r '.lines[]|select(.line|test("error";"i")).line'`
- Run a diagnostic: `heimdall-cli --hub "$HEIMDALL_HUB" run <id> disk.df | jq -r .stdout`
  (allow-listed read-only only: process.list, disk.df, uptime, os.info, dir.list)

Always parse the JSON; report host ids, states, and concrete numbers.
```

### COMMAND — `.claude/commands/fleet.md`

```markdown
---
description: Summarize the Heimdall fleet, or one host if an id is given.
---

Inspect the Heimdall fleet with `heimdall-cli` (hub `${HEIMDALL_HUB:-localhost:9090}`).

- If `$ARGUMENTS` is empty: run `heimdall-cli --hub "$HEIMDALL_HUB" fleet` and
  `heimdall-cli --hub "$HEIMDALL_HUB" hosts`, then report total/online/stale/offline
  and call out any host with alerts or high cpu/mem/disk.
- If `$ARGUMENTS` is a host id: run `heimdall-cli --hub "$HEIMDALL_HUB" host
  $ARGUMENTS` and summarize its state, key metrics, and any alerts.

Keep it to a short, factual brief with explicit host ids and numbers.
```

### Copilot — `.github/copilot-instructions.md`

GitHub Copilot Chat reads repo-wide custom instructions from this file. Append the
block below (it coexists with any existing instructions):

```markdown
## Heimdall fleet (read-only)

When asked about host or fleet health, run the `heimdall-cli` binary — it prints
JSON from a running hub. Never invent values; always call the CLI and parse with `jq`.

- Hub: `${HEIMDALL_HUB:-localhost:9090}` · token: `$HEIMDALL_TOKEN` (omit `--token` if unset)
- `heimdall-cli --hub "$HEIMDALL_HUB" fleet` — counts by state
- `heimdall-cli --hub "$HEIMDALL_HUB" hosts` — every host + metrics + capabilities
- `heimdall-cli --hub "$HEIMDALL_HUB" host <id>` — one host in full
- `heimdall-cli --hub "$HEIMDALL_HUB" top <id>` — process table
- `heimdall-cli --hub "$HEIMDALL_HUB" logs <id> [src]` — recent log lines
- `heimdall-cli --hub "$HEIMDALL_HUB" run <id> <cmd>` — allow-listed read-only
  diagnostic (process.list | disk.df | uptime | os.info | dir.list <dir>)

Report host ids and states explicitly; if a host is missing, say so.
```

### Any other harness (Hermes, OpenAI-compatible, custom)

There's no standard file, so do two things: (1) drop the AGENT text above into the
harness's **system / developer message** verbatim — it is the system prompt; and
(2) for tool-calling models, register **one** tool that shells out to the CLI.

A portable wrapper (works from any language that can exec a process):

```sh
#!/usr/bin/env sh
# heimdall-tool <subcommand> [args...]  ->  JSON on stdout, JSON error + non-zero on failure
exec heimdall-cli --hub "${HEIMDALL_HUB:-localhost:9090}" ${HEIMDALL_TOKEN:+--token "$HEIMDALL_TOKEN"} "$@"
```

A matching tool/function schema (OpenAI-style; trim to your harness):

```json
{
  "name": "heimdall_fleet",
  "description": "Query a Heimdall fleet (read-only JSON): host state, metrics, top processes, logs, allow-listed diagnostics.",
  "parameters": {
    "type": "object",
    "properties": {
      "command": {
        "type": "string",
        "enum": ["fleet", "hosts", "host", "top", "logs", "run"],
        "description": "fleet|hosts need no args; host|top|logs|run take a host id (and logs a source, run a command)."
      },
      "args": { "type": "array", "items": { "type": "string" }, "description": "e.g. [\"web-01\"] or [\"web-01\", \"disk.df\"]" }
    },
    "required": ["command"]
  }
}
```

Wire the tool handler to run `heimdall-tool "$command" "${args[@]}"` and return its
stdout. The model gets the same read-only, JSON-only surface every other harness uses.

That's the whole loop: a human pipes JSON through `jq`, CI gates on a host's state,
a log shipper consumes `logs`, and an agent — Claude, Copilot, Hermes, or your own —
answers fleet questions, all from the same read-only binary.

## Notes

- `--wait` (default 800ms) bounds how long the CLI gathers before printing; raise
  it on a slow link, a large fleet, or right after an idle period.
- Data is whatever the hub holds: logs/processes appear only for hosts configured
  to push them. See [process view](06-control-plane.md) and
  [log streaming](07-log-streaming.md).

## Next steps

- Full flag reference → [Configuration](../configuration.md)
