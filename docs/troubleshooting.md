# Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| Dashboard shows no hosts | No hub reachable, or no daemons connected | Confirm the hub is running and the dashboard's `--hub` points at it; check the station firewall allows `:9090`. |
| A host never appears | Daemon can't reach the hub | From the host: `nc -vz STATION 9090`. Check the address, port, and firewall. |
| `Unauthenticated` on the daemon | Token mismatch | Use the same `--token` / `HEIMDALL_TOKEN` on hub, daemon, and dashboard. |
| TLS handshake errors | Cert/name mismatch | Ensure clients trust the hub cert (`--tls-ca`) and the name matches (`--tls-server-name` or the cert SAN). |
| Metrics show `⚿` (needs-helper) | Privileged metric without a source | Run `sudo heimdall-helper` on that host, or ignore if you don't need it. See [Privileged Metrics](guides/04-privileged-metrics.md). |
| GPU/CPU power blank on macOS | SoC exposes no counter | Expected on some Apple Silicon chips — not a misconfiguration. |
| Power/GPU only with `sudo` | Using a CGO-free release binary | Build from source on macOS for no-sudo IOReport, or run the helper. |
| `dir.list` refused | Path outside an allow-listed root | The control plane only lists `/var/log`, `/tmp`, etc. by design. |
| Control command refused | Off the allow-list, or bad token | Only built-in keys run; check `--control-token`. |

## Host liveness states

A host moves through these states based on how recently the hub received an update:

| State | Meaning |
|---|---|
| `● ONLINE` | updates arriving within `--stale-after` (default 10s) |
| `⏱ STALE` | no update for longer than `--stale-after`, showing last-known values |
| `○ OFFLINE` | no update for longer than `--offline-after` (default 30s) |

Tune the thresholds on the **hub** (`--stale-after`, `--offline-after`). The
dashboard reflects the hub's authoritative liveness.

## Inspecting the daemon

```sh
heimdall-daemon --once          # one human-readable sample
heimdall-daemon --once --json   # one JSON object per metric
```

## Logs

Point the daemon's output at a file as structured JSON and tail it:

```sh
heimdall-daemon --hub station:9090 --log-file /var/log/heimdall/daemon.json &
tail -f /var/log/heimdall/daemon.json
```

See [Configuration → Daemon logging](configuration.md#daemon-logging).
