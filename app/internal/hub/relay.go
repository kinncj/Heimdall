// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package hub

import (
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/transport"
	v1 "heimdall/common/proto/monitoring/v1"
)

// Relay ingests snapshots relayed from a child hub. Each envelope carries the
// ordered path of hubs it has traversed; if this hub already appears in that
// path the envelope is dropped, which prevents relay loops between hubs. Ingest
// keys on stable HostID, so a reconnecting child resumes without duplication.
func (h *Hub) Relay(stream v1.FederationService_RelayServer) error {
	self := h.idOrDefault()
	for {
		env, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if containsString(env.GetPath(), self) {
			continue // loop: this hub is already in the path
		}
		snap := env.GetSnapshot()
		if snap == nil {
			continue
		}
		hostID, ms := transport.FromSnapshot(snap)
		id := domain.HostID(hostID)
		h.reg.Enroll(domain.Host{ID: id, Hostname: hostID, DisplayName: hostID}, time.Now())
		h.reg.Observe(id, ms, time.Now())
		origin := env.GetOriginHubId()
		if origin == "" {
			origin = self
		}
		h.recordOrigin(id, origin, env.GetPath())
		h.publish(snap)
		_ = stream.Send(&v1.RelayControl{AckedSeq: snap.GetSeq()})
	}
}

// RelayEnvelopes builds one upstream relay envelope per known host, appending
// this hub's id to each host's traversed path. A child hub sends these to its
// parent.
func (h *Hub) RelayEnvelopes() []*v1.RelayEnvelope {
	self := h.idOrDefault()
	views := h.reg.Hosts()
	out := make([]*v1.RelayEnvelope, 0, len(views))
	for _, v := range views {
		h.fmu.Lock()
		meta := h.fed[v.Host.ID]
		h.fmu.Unlock()
		origin := meta.origin
		if origin == "" {
			origin = self
		}
		path := append(append([]string{}, meta.path...), self)
		out = append(out, &v1.RelayEnvelope{
			OriginHubId: origin,
			Path:        path,
			Snapshot:    transport.ToSnapshot(string(v.Host.ID), v.LastSnapshot, 0, v.LastSeen),
		})
	}
	return out
}

// RunRelay relays this hub's hosts to an upstream parent hub at the given
// interval, reconnecting with backoff after a drop. The cross-hub link reuses
// the same token/TLS dial options as any client, so the parent re-authenticates
// on reconnect.
func RunRelay(ctx context.Context, h *Hub, upstream string, dialOpts []grpc.DialOption, interval time.Duration) {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		if err := relayOnce(ctx, h, upstream, dialOpts, interval); err == nil {
			return
		}
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func relayOnce(ctx context.Context, h *Hub, upstream string, dialOpts []grpc.DialOption, interval time.Duration) error {
	conn, err := grpc.NewClient(upstream, dialOpts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	stream, err := v1.NewFederationServiceClient(conn).Relay(ctx)
	if err != nil {
		return err
	}
	go func() {
		for {
			if _, err := stream.Recv(); err != nil {
				return
			}
		}
	}()

	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		for _, env := range h.RelayEnvelopes() {
			if err := stream.Send(env); err != nil {
				return err
			}
		}
		select {
		case <-t.C:
		case <-ctx.Done():
			_ = stream.CloseSend()
			return nil
		}
	}
}

func containsString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
