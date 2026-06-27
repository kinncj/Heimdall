# Event Contract — MetricBus and Stream Cadence

Companion to `common/proto/monitoring/v1/monitoring.proto`. Defines the internal
**MetricBus** event shapes, the on-wire cadence (deltas, keyframes, heartbeats),
sizing for low-bandwidth links, and the liveness thresholds. The proto is the wire
contract; this document is the behavioural contract for producers and subscribers.

## 1. MetricBus overview

The hub ingests `Snapshot` frames, updates the host registry and ring buffers,
then publishes **bus events** to subscribers (dashboards, federation relay, the
self-metrics sink). The bus is in-memory pub/sub. Subscribers are decoupled from
ingest: a slow subscriber is back-pressured or dropped, never blocking ingest.

```
daemon --Snapshot--> hub.ingest --> registry + ring buffers --> MetricBus --> subscribers
```

Bus events are derived from domain types, not raw proto. The transport layer maps
`monitoring.v1` messages into domain types at ingest; the bus carries domain
events. This keeps the boundary intact (subscribers depend on domain, not proto).

## 2. Event types

| Event | Trigger | Payload | Subscribers act |
|---|---|---|---|
| `snapshot.applied` | A `Snapshot` (keyframe or delta) ingested | `host_id`, `seq`, `ts`, changed samples, merged view | Render / relay |
| `host.state_changed` | Liveness FSM transition | `host_id`, `from`, `to`, `ts`, `last_seen` | Re-render badge |
| `host.enrolled` | Enroll accepted | `host_id`, `HostContext` | Add host tile |
| `host.removed` | Host deleted | `host_id` | Drop host; purge subscription state |
| `adapter.removed` | Adapter retired for a host | `host_id`, `metric_keys[]` | Drop those series |
| `control.audit` | Control request resolved | `request_id`, `actor`, `host_id`, `cmd`, `args`, `exit`, `decision` | Append to audit log |
| `log.line` | Opt-in tail line (E6, separate stream) | `host_id`, `source`, `ts`, `line`, `level`, `rate_limited` | Append to logs pane |

### 2.1 Event envelope (domain shape)

```
BusEvent {
  type:      string   // "snapshot.applied" | "host.state_changed" | ...
  host_id:   string
  ts_unix_ms:int64
  seq:       uint64   // present for snapshot.applied; 0 otherwise
  payload:   <one of the typed payloads above>
}
```

### 2.2 `host.state_changed` payload

```
{ host_id, from: "ONLINE", to: "STALE", ts_unix_ms, last_seen_unix_ms }
```

States: `ENROLLING | ONLINE | STALE | OFFLINE`. Transitions are defined by the
liveness FSM in `architecture.md` §7.

## 3. Snapshot cadence

Producer (daemon) cadence on the `MetricStreamService.Stream` channel:

| Frame | When | `keyframe` | `delta` | Contents |
|---|---|---|---|---|
| **Keyframe** | On connect, on reconnect, on hub `KeyframeRequest`, and every `keyframe_interval` | `true` | `false` | Full set: every metric with current value + `MetricStatus` |
| **Delta** | Every `sample_interval` when something changed | `false` | `true` | Only metrics whose value or status changed since the last frame |
| **Heartbeat** | Every `sample_interval` when nothing changed | `false` | `false` | Empty `samples[]`, current `seq` + `ts` only |

Rules:

- `seq` is monotonic per host and increments on **every** frame (including
  heartbeats). The hub acks the highest durably-applied `seq` via
  `StreamControl.AckResume`.
- A heartbeat resets the hub's `stale`/`offline` timers exactly like a data frame.
  This decouples "host is alive" from "a metric changed", so an idle but healthy
  host never goes `STALE`.
- After reconnect the daemon **must** send a keyframe before any delta, so the hub
  and all subscribers resync from a known-complete state.
- The hub may lower the daemon's cadence at runtime via `StreamControl.CadenceUpdate`
  (e.g. on metered links or hub load).

### 3.1 Recommended defaults

| Parameter | Default | Notes |
|---|---|---|
| `sample_interval_ms` | 2000 | Operator-tunable per host; raise on low-bandwidth links |
| `keyframe_interval` | 30 × `sample_interval` | Bounds resync cost and corrects any drift |
| `stale_after_s` | 3 × `sample_interval` (min 6s) | Hub-issued at enroll |
| `offline_after_s` | 10 × `sample_interval` (min 30s) | Hub-issued at enroll |
| `heartbeat` | every `sample_interval` | Implicit: a frame is sent each interval regardless of change |

`stale_after < offline_after` is an invariant. Both are returned in
`EnrollResponse` and are runtime-tunable.

## 4. Low-bandwidth sizing

Design target: a single host's steady-state uplink stays small enough for
constrained links (cellular, remote sites). Levers:

| Lever | Mechanism | Effect |
|---|---|---|
| Delta encoding | `Snapshot.delta=true`, changed samples only | Steady-state frame carries a handful of samples, not the full set |
| Scalar timestamps | `int64 ts_unix_millis` (no per-sample sub-message) | ~8 bytes/frame vs a `Timestamp` message per sample |
| Packed per-core | `PerCore.values [packed=true]` | One length-delimited block for N cores, not N fields |
| Counters + rate | `Counter{total, rate}` for disk/net | Ship monotonic deltas, not recomputed gauges |
| Compression | gRPC message compression (gzip/zstd) | Repetitive metric keys compress well |
| Cadence | tunable `sample_interval`, hub `CadenceUpdate` | Trade resolution for bandwidth on the fly |
| Heartbeat | empty `samples[]` | Liveness costs ~tens of bytes when idle |

**Order-of-magnitude budget** (illustrative, not a guarantee):

| Frame | Approx serialized size |
|---|---|
| Keyframe (~40 metrics, per-core CPU on 16 cores) | ~1–2 KB |
| Steady-state delta (3–6 changed metrics) | ~80–250 B |
| Heartbeat (no change) | ~20–40 B |

At `sample_interval=2s` a mostly-idle host averages well under 1 KB/s before
compression; a busy host is bounded by keyframe size × keyframe cadence plus delta
churn. Operators raise `sample_interval` to cut this further on metered links.

## 5. Liveness thresholds and last-known retention

- `ONLINE → STALE`: no frame (data or heartbeat) within `stale_after_s`.
- `STALE → OFFLINE`: no frame within `offline_after_s`.
- `STALE`/`OFFLINE` hosts retain **last-known values**, rendered with **symbol +
  text + timestamp** (`last_seen`). The dashboard never presents frozen values as
  live.
- Any frame returns the host to `ONLINE` under the **same** `host_id` (no duplicate
  registration).
- De-duplication is by `(host_id, seq)`; a replayed frame after reconnect is
  idempotent.

## 6. Federation event semantics (E4)

A child hub subscribes to its own bus and relays `snapshot.applied` upstream as a
`RelayEnvelope`:

- `origin_hub_id` = the first hub that ingested the snapshot.
- `path[]` = ordered hub ids traversed. A hub **drops** any envelope whose `path`
  already contains its own id (**loop prevention**).
- Upstream de-duplication uses `(origin_hub_id, host_id, seq)` (**duplication
  prevention**).
- The parent hub republishes onto its own bus; parent-side dashboards subscribe
  exactly as local ones do. Fan-out to multiple dashboards is concurrent.

## 7. Back-pressure and subscriber lifecycle

- Each subscriber has a bounded queue. On overflow the bus drops the **oldest**
  snapshot for that subscriber (monitoring favours latest state) and increments a
  `dropped` self-metric; ingest is never blocked.
- `host.removed` / `adapter.removed` events instruct subscribers to purge state,
  guaranteeing the **delete-reassignment** property: no orphan series, buffers, or
  subscriptions. This is verified by the delete-reassignment integration test.
- Unsubscribe tears down the queue and goroutine deterministically (no leak).

## 8. Self-metrics (Heimdall monitors itself)

The hub and daemon publish internal counters as ordinary Heimdall metrics on the
same bus:

| Metric key | Unit | Meaning |
|---|---|---|
| `self.adapter.collect_ms` | ms | Per-adapter collect latency |
| `self.adapter.timeouts` | count | Adapter timeouts |
| `self.adapter.panics` | count | Recovered adapter panics |
| `self.stream.reconnects` | count | Daemon reconnects |
| `self.bus.queue_depth` | count | Per-subscriber queue depth |
| `self.bus.dropped` | count | Back-pressure drops |
| `self.ring.occupancy` | percent | Ring-buffer fill |
| `self.logs.rate_limited` | count | Dropped log lines (E6) |
