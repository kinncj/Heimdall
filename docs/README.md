# Heimdall Documentation

> *Watch Over All Realms* — a lightweight, cross-platform hardware monitoring
> system with a real-time terminal dashboard.

Start here, then jump to the guide for what you want to do.

## New here?

1. [Installation](installation.md) — get the binaries
2. [Quickstart](guides/01-quickstart.md) — monitor your own machine in ~2 minutes
3. [Architecture & Operations](deployment.md) — how the pieces fit, with diagrams

## Start guides by modality

| Guide | Use it when you want to… |
|---|---|
| [Quickstart](guides/01-quickstart.md) | watch a single machine (all-in-one) |
| [Monitor a Fleet](guides/02-monitor-a-fleet.md) | watch many hosts from one station |
| [Secure Deployment](guides/03-secure-deployment.md) | require TLS + an enrollment token |
| [Privileged Metrics](guides/04-privileged-metrics.md) | unlock power, GPU, and full thermal |
| [Federation (Bifröst)](guides/05-federation.md) | span multiple sites / networks |
| [Process View & Commands](guides/06-control-plane.md) | see a host's process table (top) and run allow-listed diagnostics |
| [Log Streaming](guides/07-log-streaming.md) | tail host logs in the dashboard |
| [Demo Mode](guides/08-demo-mode.md) | explore the UI with no setup |
| [Metrics Export (Mímir)](guides/09-metrics-export.md) | scrape Heimdall from Prometheus / Grafana |
| [Alerting (Gjallarhorn)](guides/10-alerting.md) | fire threshold alerts to a webhook |
| [`heimdall-cli` (programmatic & agents)](guides/11-hub-cli.md) | query the fleet from scripts, CI/CD, or an AI agent |

## Reference

| Doc | Contents |
|---|---|
| [Configuration](configuration.md) | every flag, env var, and the config file + first-run wizard, per binary |
| [Metrics](metrics.md) | every metric collected, with units and meaning |
| [Glossary](glossary.md) | codenames (Bifröst, Ratatoskr, …) and how to pronounce them |
| [Architecture & Operations](deployment.md) | topology, what-runs-where, sequence diagrams, ports |
| [Troubleshooting](troubleshooting.md) | common issues and host-liveness states |

## The four binaries

| Binary | Runs on | Privilege | Job |
|---|---|---|---|
| `heimdall-hub` | monitoring station | normal | receives metrics, fans out to dashboards |
| `heimdall-dashboard` | monitoring station (any number) | normal | renders the fleet — pure presentation |
| `heimdall-daemon` | every host | unprivileged | collects + streams this host's metrics |
| `heimdall-helper` | a host (optional) | root | serves privileged metrics to the local daemon |

Data flow: **daemon → hub → dashboard(s)**; the helper is a local sidecar to a
daemon (**helper → daemon** over a unix socket).

## Design & decisions

- [System architecture](specs/current/architecture.md)
- [Architecture Decision Records](architecture/) — enrollment, adapters, helper,
  GPU/power, federation, control plane, storage
- [User stories](stories/) — the Gherkin specs each feature implements
- [Brand & terminal identity](design/identity/tui-brand.md)

## Contributing / development

```sh
make build-tui        # build all four binaries
make test             # unit tests
make lint             # gofmt + go vet
make test-acceptance  # behave acceptance suite (drives the real binaries)
```
