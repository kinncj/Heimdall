// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package hub is the central gRPC server: it ingests metric streams from
// daemons into a host registry and fans snapshots out to dashboard subscribers
// (the pub/sub seam that federation later rides on).
package hub

import (
	"context"
	"errors"
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

	mu   sync.Mutex
	subs map[chan *v1.Snapshot]struct{}

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
	h.reg.Enroll(domain.Host{ID: domain.HostID(id), Hostname: host.GetHostname(), DisplayName: name}, time.Now())
	return &v1.EnrollResponse{
		HostId:           id,
		Accepted:         true,
		SampleIntervalMs: 2000,
		StaleAfterS:      int32(h.staleAfter.Seconds()),
		OfflineAfterS:    int32(h.offlineAfter.Seconds()),
	}, nil
}

// Stream ingests a daemon's snapshots until the stream closes.
func (h *Hub) Stream(stream v1.MetricStreamService_StreamServer) error {
	for {
		snap, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		hostID, ms, labels := transport.FromSnapshot(snap)
		h.reg.Observe(domain.HostID(hostID), ms, labels, time.Now())
		h.recordOrigin(domain.HostID(hostID), h.idOrDefault(), nil)
		h.publish(snap)
	}
}

// Subscribe streams the current state then live snapshots to a dashboard.
func (h *Hub) Subscribe(_ *v1.SubscribeRequest, stream v1.FederationService_SubscribeServer) error {
	ch := make(chan *v1.Snapshot, 64)
	h.addSub(ch)
	defer h.removeSub(ch)

	for _, hv := range h.reg.Hosts() {
		if err := stream.Send(transport.ToSnapshot(string(hv.Host.ID), hv.LastSnapshot, hv.Host.Context.Labels, 0, hv.LastSeen)); err != nil {
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
}

func (h *Hub) removeSub(ch chan *v1.Snapshot) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}
