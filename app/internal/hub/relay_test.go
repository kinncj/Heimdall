// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package hub

import (
	"context"
	"testing"
	"time"

	"heimdall/app/internal/domain"
	v1 "heimdall/common/proto/monitoring/v1"
)

// A relayed envelope whose path already contains this hub's id is a loop and is
// dropped; an envelope from a child not yet in the path is ingested.
func TestRelayLoopPreventionAndIngest(t *testing.T) {
	lis, parent, stop := serve(t, "")
	defer stop()
	parent.SetID("P")

	conn := dial(t, lis, "")
	defer conn.Close()
	stream, err := v1.NewFederationServiceClient(conn).Relay(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := stream.Send(&v1.RelayEnvelope{
		OriginHubId: "C", Path: []string{"P"},
		Snapshot: &v1.Snapshot{HostId: "loopy"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := stream.Send(&v1.RelayEnvelope{
		OriginHubId: "C", Path: []string{"C"},
		Snapshot: &v1.Snapshot{HostId: "alpha"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := stream.Recv(); err != nil { // ack confirms alpha was processed
		t.Fatalf("expected ack: %v", err)
	}

	if _, ok := parent.Registry().Host("loopy"); ok {
		t.Error("looped host was ingested")
	}
	if _, ok := parent.Registry().Host("alpha"); !ok {
		t.Error("relayed host alpha not ingested")
	}
}

// A child hub's relay envelopes are ingested by the parent (scenario 1), and the
// parent's own upstream path appends its id so a return relay would be dropped
// by the child (scenario 3, loop-free).
func TestTwoHubRelay(t *testing.T) {
	lis, parent, stop := serve(t, "")
	defer stop()
	parent.SetID("P")

	child := New(2*time.Second, 5*time.Second)
	child.SetID("C")
	child.reg.Enroll(domain.Host{ID: "alpha", Hostname: "alpha", DisplayName: "alpha"}, time.Now())
	child.reg.Observe("alpha", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 10}}, time.Now())
	child.recordOrigin("alpha", "C", nil)

	envs := child.RelayEnvelopes()
	if len(envs) != 1 || envs[0].OriginHubId != "C" || len(envs[0].Path) != 1 || envs[0].Path[0] != "C" {
		t.Fatalf("child envelopes = %+v", envs)
	}

	conn := dial(t, lis, "")
	defer conn.Close()
	stream, err := v1.NewFederationServiceClient(conn).Relay(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range envs {
		if err := stream.Send(e); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := stream.Recv(); err != nil {
		t.Fatalf("ack: %v", err)
	}

	hv, ok := parent.Registry().Host("alpha")
	if !ok {
		t.Fatal("parent missing relayed host")
	}
	if hv.State != domain.StateOnline {
		t.Errorf("relayed host state = %v, want online", hv.State)
	}

	penvs := parent.RelayEnvelopes()
	if len(penvs) != 1 || len(penvs[0].Path) != 2 || penvs[0].Path[1] != "P" {
		t.Fatalf("parent relay path = %+v, want [C P]", penvs)
	}
}

// Several dashboards subscribing to one hub all receive its host state.
func TestMultipleSubscribers(t *testing.T) {
	lis, h, stop := serve(t, "")
	defer stop()
	h.reg.Enroll(domain.Host{ID: "alpha", Hostname: "alpha"}, time.Now())
	h.reg.Observe("alpha", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 5}}, time.Now())

	subscribeOnce := func() {
		conn := dial(t, lis, "")
		t.Cleanup(func() { conn.Close() })
		st, err := v1.NewFederationServiceClient(conn).Subscribe(
			context.Background(), &v1.SubscribeRequest{SubscriberId: "d"})
		if err != nil {
			t.Fatal(err)
		}
		snap, err := st.Recv()
		if err != nil {
			t.Fatalf("recv: %v", err)
		}
		if snap.GetHostId() != "alpha" {
			t.Errorf("host id = %q, want alpha", snap.GetHostId())
		}
	}
	subscribeOnce()
	subscribeOnce()
}
