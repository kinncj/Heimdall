# Secure Deployment — TLS + Enrollment Token

By default the hub is **plaintext and unauthenticated** — fine for a laptop, wrong
for a network. For any shared or networked deployment, require an enrollment token
and enable TLS so daemons authenticate and the channel is encrypted.

## What this protects

- **Token**: a daemon (or dashboard, or relaying hub) presenting a missing or
  invalid token is rejected at enrollment and never registered.
- **TLS**: the gRPC channel is encrypted; clients verify the hub's certificate.

## 1. Generate a certificate

For local testing, a self-signed cert:

```sh
make dev-certs    # writes certs/hub.crt and certs/hub.key (dev only)
```

> **Production**: issue per-host certificates from a real CA (or your internal PKI)
> instead of `make dev-certs`. The dev cert is for local use only.

Pick a shared enrollment token:

```sh
export HEIMDALL_TOKEN=$(openssl rand -hex 16)
```

## 2. Monitoring station

```sh
./bin/heimdall-hub --listen :9090 \
  --tls-cert certs/hub.crt --tls-key certs/hub.key \
  --token "$HEIMDALL_TOKEN" &

./bin/heimdall-dashboard --hub localhost:9090 \
  --tls --tls-ca certs/hub.crt \
  --token "$HEIMDALL_TOKEN"
```

## 3. Each host

```sh
export HEIMDALL_TOKEN=<same-token-as-the-station>

./bin/heimdall-daemon --hub station-host:9090 --name "$(hostname)" \
  --tls --tls-ca certs/hub.crt
# token is read from $HEIMDALL_TOKEN (or pass --token)
```

## Token precedence

Both `--token` and the `HEIMDALL_TOKEN` environment variable are accepted; the
flag wins when both are set. Using the environment variable keeps the secret out
of process listings and shell history.

## TLS flags

| Flag | Applies to | Meaning |
|---|---|---|
| `--tls-cert` / `--tls-key` | hub | server certificate + key (enables TLS) |
| `--tls` | daemon, dashboard | connect to the hub over TLS |
| `--tls-ca` | daemon, dashboard | PEM bundle to trust (omit to use system roots) |
| `--tls-server-name` | daemon, dashboard | override the verified server name (SNI) |
| `--tls-insecure` | daemon, dashboard | **dev only** — skip certificate verification |

## Reconnect behaviour

A daemon that drops and reconnects **re-authenticates** before any data is
accepted, and resumes under the same stable HostID — no duplicate hosts, no
unauthenticated gap.

## Federation links

Cross-hub relay uses the same model with an `--upstream-` prefix
(`--upstream-token`, `--upstream-tls`, `--upstream-tls-ca`). See
[Federation](05-federation.md).

## Next steps

- Span multiple sites securely → [Federation](05-federation.md)
- Full flag reference → [Configuration](../configuration.md)
