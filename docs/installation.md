# Installation

Heimdall ships four small static binaries: `heimdall-hub`, `heimdall-dashboard`,
`heimdall-daemon`, and `heimdall-helper`. Install only what each machine needs:

- **Monitoring station** → `heimdall-hub` + `heimdall-dashboard`
- **Each host** → `heimdall-daemon` (+ optional `heimdall-helper`)

## Option 1 — Prebuilt binaries (curl)

Prebuilt binaries are published to [GitHub Releases](https://github.com/kinncj/Heimdall/releases)
for Linux/macOS × amd64/arm64 (Windows assets too).

**Dashboard** (on the monitoring station):
```sh
curl -fsSL https://raw.githubusercontent.com/kinncj/Heimdall/main/scripts/install.sh | sh -s -- dashboard
```

**Daemon** (on each host):
```sh
curl -fsSL https://raw.githubusercontent.com/kinncj/Heimdall/main/scripts/install.sh | sh -s -- daemon
```

You can install several at once and pin a version:
```sh
curl -fsSL https://raw.githubusercontent.com/kinncj/Heimdall/main/scripts/install.sh | sh -s -- hub dashboard
HEIMDALL_VERSION=v0.1.0 sh install.sh daemon helper
```

Overrides:

| Variable | Default | Meaning |
|---|---|---|
| `HEIMDALL_VERSION` | latest release | tag to install (e.g. `v0.1.0`) |
| `HEIMDALL_BIN_DIR` | `/usr/local/bin` (or `~/.local/bin` if not writable) | install directory |
| `HEIMDALL_REPO` | `kinncj/Heimdall` | source `owner/repo` |

The installer verifies each binary against the release `SHA256SUMS`.

> **Apple Silicon power/GPU**: release binaries are CGO-free and fall back to
> `powermetrics` (needs sudo). For no-sudo GPU power via IOReport, build from
> source on macOS (CGO is on by default there). See
> [Privileged Metrics](guides/04-privileged-metrics.md).

## Option 2 — Build from source

Requires Go 1.26+.

```sh
git clone https://github.com/kinncj/Heimdall.git
cd Heimdall
make build-tui        # builds bin/heimdall-{dashboard,daemon,hub,helper}
```

Install them onto `PATH`:

```sh
sudo install bin/heimdall-* /usr/local/bin/
```

## Option 3 — go install

```sh
go install heimdall/app/cmd/dashboard@latest
go install heimdall/app/cmd/daemon@latest
go install heimdall/app/cmd/hub@latest
go install heimdall/app/cmd/helper@latest
```

## Windows

Download the `*_windows_amd64.exe` / `*_windows_arm64.exe` assets from the
[Releases](https://github.com/kinncj/Heimdall/releases) page, or build from source
with `go build ./app/cmd/...`.

## Verify

```sh
heimdall-dashboard --help
heimdall-daemon --once         # prints this machine's metrics
```

## Next steps

- [Quickstart — Monitor your own machine](guides/01-quickstart.md)
- [Monitor a fleet](guides/02-monitor-a-fleet.md)
