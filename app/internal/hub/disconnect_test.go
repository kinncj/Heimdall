// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package hub

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"heimdall/app/internal/domain"
	v1 "heimdall/common/proto/monitoring/v1"
)

// When a daemon's stream ends, the hub must flip the host Offline at once and push
// a disconnected snapshot to subscribers — real-time, not on the freshness window.
// The window here is 30s, so anything that turns Offline within the test proves it
// came from the stream-end signal, not a timeout.
func TestDisconnectIsRealTimeNotTimeout(t *testing.T) {
	h := New(10*time.Second, 30*time.Second)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	v1.RegisterMetricStreamServiceServer(srv, h)
	v1.RegisterFederationServiceServer(srv, h)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()
	addr := lis.Addr().String()

	dial := func() *grpc.ClientConn {
		c, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatal(err)
		}
		return c
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Subscriber first, so it sees the disconnected snapshot the hub publishes.
	cConn := dial()
	defer cConn.Close()
	sub, err := v1.NewFederationServiceClient(cConn).Subscribe(ctx, &v1.SubscribeRequest{SubscriberId: "test"})
	if err != nil {
		t.Fatal(err)
	}
	gotDisconnect := make(chan string, 4)
	go func() {
		for {
			snap, err := sub.Recv()
			if err != nil {
				return
			}
			if snap.GetDisconnected() {
				gotDisconnect <- snap.GetHostId()
			}
		}
	}()

	// Fake daemon: register host "d1", then end the stream (clean CloseSend, as the
	// real daemon does on SIGTERM).
	dConn := dial()
	stream, err := v1.NewMetricStreamServiceClient(dConn).Stream(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Send(&v1.Snapshot{HostId: "d1"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "host registered online", func() bool {
		hv, ok := h.reg.Host("d1")
		return ok && hv.State == domain.StateOnline
	})

	if err := stream.CloseSend(); err != nil { // graceful disconnect
		t.Fatal(err)
	}

	// The hub's own registry flips Offline immediately — well within the 30s window.
	waitFor(t, "host offline right after stream end", func() bool {
		hv, ok := h.reg.Host("d1")
		return ok && hv.State == domain.StateOffline
	})

	// And the subscriber is told, so a dashboard flips in real time.
	select {
	case id := <-gotDisconnect:
		if id != "d1" {
			t.Fatalf("disconnected snapshot for %q, want d1", id)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("subscriber never received a disconnected snapshot")
	}
}
