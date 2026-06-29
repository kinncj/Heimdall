// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package hub is the central gRPC server: it ingests metric streams from
// daemons into a host registry and fans snapshots out to dashboard subscribers
// (the pub/sub seam that federation later rides on).
package hub

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/secure"
	"heimdall/app/internal/transport"
	v1 "heimdall/common/proto/monitoring/v1"
)

// Hub implements the enrollment, metric-stream, and federation services.
type Hub struct {
	v1.UnimplementedEnrollmentServiceServer
	v1.UnimplementedMetricStreamServiceServer
	v1.UnimplementedFederationServiceServer

	reg          *domain.HostRegistry
	staleAfter   time.Duration
	offlineAfter time.Duration
	auth         secure.Authenticator
	id           string

	mu           sync.Mutex
	subs         map[chan *v1.Snapshot]struct{}
	daemons      map[chan *v1.StreamControl]struct{}      // active daemon control sinks (v2 demand-driven push)
	daemonByHost map[domain.HostID]chan *v1.StreamControl // hostID -> sink, for routing on-demand commands (v2 Phase 2)

	fmu sync.Mutex
	fed map[domain.HostID]relayMeta
}

// relayMeta tracks a host's federation provenance: the hub that first ingested
// it and the ordered hub path it has traversed, used to build upstream relay
// envelopes and to prevent loops.
type relayMeta struct {
	origin string
	path   []string
}

// New builds a hub with the given liveness thresholds. Authorization defaults to
// allow-all; call SetToken to require an enrollment token.
func New(staleAfter, offlineAfter time.Duration) *Hub {
	return &Hub{
		reg:          domain.NewHostRegistry(staleAfter, offlineAfter),
		staleAfter:   staleAfter,
		offlineAfter: offlineAfter,
		auth:         secure.NewAuthenticator(""),
		subs:         make(map[chan *v1.Snapshot]struct{}),
		daemons:      make(map[chan *v1.StreamControl]struct{}),
		daemonByHost: make(map[domain.HostID]chan *v1.StreamControl),
		fed:          make(map[domain.HostID]relayMeta),
	}
}

// SetToken requires the given enrollment token on every RPC. An empty token
// restores unauthenticated dev mode.
func (h *Hub) SetToken(token string) { h.auth = secure.NewAuthenticator(token) }

// SetID sets this hub's federation identity, recorded as the origin of locally
// ingested hosts and appended to relay paths.
func (h *Hub) SetID(id string) { h.id = id }

func (h *Hub) idOrDefault() string {
	if h.id == "" {
		return "hub"
	}
	return h.id
}

func (h *Hub) recordOrigin(id domain.HostID, origin string, path []string) {
	h.fmu.Lock()
	h.fed[id] = relayMeta{origin: origin, path: path}
	h.fmu.Unlock()
}

// Registry exposes the host registry (e.g. for an embedded dashboard).
func (h *Hub) Registry() *domain.HostRegistry { return h.reg }

// withOrigin returns a copy of labels with the reserved "hub" tag set to the
// host's origin hub — the Yggdrasil grouping axis. The origin is authoritative
// and overrides any incoming "hub" value.
func withOrigin(labels map[string]string, origin string) map[string]string {
	out := make(map[string]string, len(labels)+1)
	for k, v := range labels {
		out[k] = v
	}
	out["hub"] = origin
	return out
}

// Enroll registers a daemon's host identity and returns stream policy.
func (h *Hub) Enroll(_ context.Context, req *v1.EnrollRequest) (*v1.EnrollResponse, error) {
	host := req.GetHost()
	id := host.GetHostId()
	name := host.GetDisplayName()
	if name == "" {
		name = host.GetHostname()
	}
	if name == "" {
		name = id
	}
	h.reg.Enroll(domain.Host{
		ID: domain.HostID(id), Hostname: host.GetHostname(), DisplayName: name,
		Context: domain.HostContext{
			OS:     host.GetContext().GetOs(),
			Arch:   host.GetContext().GetArch(),
			Labels: host.GetContext().GetLabels(),
		},
	}, time.Now())
	return &v1.EnrollResponse{
		HostId:           id,
		Accepted:         true,
		SampleIntervalMs: 2000,
		StaleAfterS:      int32(h.staleAfter.Seconds()),
		OfflineAfterS:    int32(h.offlineAfter.Seconds()),
	}, nil
}

// Stream ingests a daemon's snapshots until the stream closes. It also drives the
// daemon's demand window: directives ride the downstream side of this same bidi
// stream (ADR 0018), so the daemon needs no inbound port. The send goroutine runs
// concurrently with the Recv loop — one reader, one writer, as gRPC allows.
func (h *Hub) Stream(stream v1.MetricStreamService_StreamServer) error {
	ctrl := make(chan *v1.StreamControl, 4)
	h.addDaemon(ctrl)
	var boundHost domain.HostID
	defer func() {
		h.removeDaemon(ctrl)
		if boundHost != "" {
			h.unbindDaemon(boundHost, ctrl)
		}
	}()
	go func() {
		for {
			select {
			case <-stream.Context().Done():
				return
			case c := <-ctrl:
				if err := stream.Send(c); err != nil {
					return
				}
			}
		}
	}()
	select { // tell the daemon the current demand as soon as it connects
	case ctrl <- h.observabilityWindow():
	default:
	}

	for {
		snap, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		hostID, ms, labels := transport.FromSnapshot(snap)
		id := domain.HostID(hostID)
		if boundHost != id { // bind this stream to its host for command routing (v2)
			boundHost = id
			h.bindDaemon(id, ctrl)
		}
		h.reg.Observe(id, ms, withOrigin(labels, h.idOrDefault()), time.Now())
		// Heimdallr's sight (ADR 0017): buffer the host's pushed process table and
		// append its tailed log lines to the bounded ring, for the dashboard modals.
		procs, procsAt, logs := transport.ObservabilityFromSnapshot(snap)
		h.reg.RecordPush(id, procs, procsAt, logs)
		h.recordOrigin(id, h.idOrDefault(), nil)
		h.publish(h.enrich(id, snap))
	}
}

// enrich rebuilds a snapshot from the registry view so live updates carry the
// same merged labels (origin hub, inherited tags) and alert state as the initial
// subscribe state — without this, the live path would forward the raw daemon
// snapshot and the dashboard would lose the hub label and alerts between frames.
// Falls back to the raw snapshot if the host is unknown.
func (h *Hub) enrich(id domain.HostID, raw *v1.Snapshot) *v1.Snapshot {
	hv, ok := h.reg.Host(id)
	if !ok {
		return raw
	}
	out := transport.ToSnapshot(string(hv.Host.ID), hv.LastSnapshot, hv.Host.Context.Labels, raw.GetSeq(), time.Now())
	out.Alerts = hv.Alerts
	// Forward a fresh process table only when the daemon just pushed one (no
	// re-send between pushes — bandwidth), and pass the freshly tailed log lines
	// straight through. The dashboard keeps the last table and its own log ring.
	out.Processes = raw.GetProcesses()
	out.ProcessesAtUnixMillis = raw.GetProcessesAtUnixMillis()
	out.LogLines = raw.GetLogLines()
	out.CommandResult = raw.GetCommandResult() // forward on-demand command results (v2 Phase 2)
	return out
}

// bindDaemon / unbindDaemon map a host to its control sink for command routing.
func (h *Hub) bindDaemon(id domain.HostID, ctrl chan *v1.StreamControl) {
	h.mu.Lock()
	h.daemonByHost[id] = ctrl
	h.mu.Unlock()
}

func (h *Hub) unbindDaemon(id domain.HostID, ctrl chan *v1.StreamControl) {
	h.mu.Lock()
	if h.daemonByHost[id] == ctrl { // only clear if still ours (a reconnect may have replaced it)
		delete(h.daemonByHost, id)
	}
	h.mu.Unlock()
}

// RunCommand routes an on-demand allow-listed command to the owning daemon over
// its outbound stream (v2 Phase 2, ADR 0018). The ack reports only whether the
// host was reachable; the result returns asynchronously on the host's snapshot
// (command_result), matched by request_id. The daemon enforces the allow-list.
func (h *Hub) RunCommand(_ context.Context, req *v1.ControlRequest) (*v1.CommandAck, error) {
	ack := &v1.CommandAck{RequestId: req.GetRequestId()}
	if req.GetHostId() == "" || req.GetAllowlistedCmd() == "" {
		ack.Error = "host_id and allowlisted_cmd are required"
		return ack, nil
	}
	h.mu.Lock()
	ctrl, ok := h.daemonByHost[domain.HostID(req.GetHostId())]
	h.mu.Unlock()
	if !ok {
		ack.Error = fmt.Sprintf("host %q is not connected to this hub", req.GetHostId())
		return ack, nil
	}
	select {
	case ctrl <- &v1.StreamControl{Control: &v1.StreamControl_Run{Run: req}}:
		ack.Accepted = true
	default:
		ack.Error = "host control channel is busy; retry"
	}
	return ack, nil
}

// Subscribe streams the current state then live snapshots to a dashboard.
func (h *Hub) Subscribe(_ *v1.SubscribeRequest, stream v1.FederationService_SubscribeServer) error {
	ch := make(chan *v1.Snapshot, 64)
	h.addSub(ch)
	defer h.removeSub(ch)

	for _, hv := range h.reg.Hosts() {
		snap := transport.ToSnapshot(string(hv.Host.ID), hv.LastSnapshot, hv.Host.Context.Labels, 0, hv.LastSeen)
		snap.Alerts = hv.Alerts
		// Seed a newly-connected subscriber (dashboard or `hub cli`) with the latest
		// process table and the buffered log ring, so the top/log views and the CLI
		// have history immediately instead of only lines pushed after connect.
		transport.AttachObservability(snap, hv.Processes, hv.ProcessesAt, string(hv.Host.ID), hv.Logs)
		if err := stream.Send(snap); err != nil {
			return err
		}
	}
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case snap := <-ch:
			if err := stream.Send(snap); err != nil {
				return err
			}
		}
	}
}

// EvaluateLoop recomputes liveness every second so stale/offline propagate.
func (h *Hub) EvaluateLoop(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.reg.Evaluate(time.Now())
		}
	}
}

func (h *Hub) publish(snap *v1.Snapshot) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- snap:
		default: // drop for a slow subscriber rather than block ingest
		}
	}
}

func (h *Hub) addSub(ch chan *v1.Snapshot) {
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	h.broadcastWindow() // a new watcher may open the observability window (v2)
}

func (h *Hub) removeSub(ch chan *v1.Snapshot) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
	h.broadcastWindow() // the last watcher leaving closes the window (v2)
}

// addDaemon / removeDaemon register a daemon's control sink so the hub can push
// StreamControl directives down its outbound stream (ADR 0018 — no daemon port).
func (h *Hub) addDaemon(ch chan *v1.StreamControl) {
	h.mu.Lock()
	h.daemons[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) removeDaemon(ch chan *v1.StreamControl) {
	h.mu.Lock()
	delete(h.daemons, ch)
	h.mu.Unlock()
}

// observabilityWindow is the current demand directive: push logs + a process table
// only while at least one dashboard is subscribed.
func (h *Hub) observabilityWindow() *v1.StreamControl {
	h.mu.Lock()
	on := len(h.subs) > 0
	h.mu.Unlock()
	return &v1.StreamControl{Control: &v1.StreamControl_Observability{
		Observability: &v1.ObservabilityWindow{Logs: on, Processes: on},
	}}
}

// broadcastWindow pushes the current demand window to every connected daemon.
// Non-blocking: a slow daemon sink misses this edge and gets the next one.
func (h *Hub) broadcastWindow() {
	w := h.observabilityWindow()
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.daemons {
		select {
		case ch <- w:
		default:
		}
	}
}
