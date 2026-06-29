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

	v1 "heimdall/common/proto/monitoring/v1"
)

// TestSocketRoundTrip exercises the whole v2 socket transport at the hub boundary:
// a daemon registers over the bidi metric stream, a subscriber opens the demand
// window, a RunCommand directive is routed down to the daemon, and the result is
// fanned back up to the subscriber — all over real gRPC (ADR 0018).
func TestSocketRoundTrip(t *testing.T) {
	h := New(10*time.Second, 30*time.Second)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	v1.RegisterMetricStreamServiceServer(srv, h)
	v1.RegisterFederationServiceServer(srv, h)
	v1.RegisterEnrollmentServiceServer(srv, h)
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

	// Fake daemon: register host "d1", capture demand windows, answer run directives.
	dConn := dial()
	defer dConn.Close()
	stream, err := v1.NewMetricStreamServiceClient(dConn).Stream(ctx)
	if err != nil {
		t.Fatal(err)
	}
	windows := make(chan *v1.ObservabilityWindow, 16)
	go func() {
		for {
			ctrl, err := stream.Recv()
			if err != nil {
				return
			}
			if w := ctrl.GetObservability(); w != nil {
				windows <- w
			}
			if req := ctrl.GetRun(); req != nil {
				_ = stream.Send(&v1.Snapshot{
					HostId: "d1",
					CommandResult: &v1.ControlResponse{
						RequestId: req.GetRequestId(),
						Status:    v1.MetricStatus_METRIC_STATUS_OK,
						Stdout:    "ran " + req.GetAllowlistedCmd(),
					},
				})
			}
		}
	}()
	if err := stream.Send(&v1.Snapshot{HostId: "d1"}); err != nil {
		t.Fatal(err)
	}

	waitFor(t, "host registered", func() bool { _, ok := h.reg.Host("d1"); return ok })

	// Subscriber: opens the demand window and reads command results.
	cConn := dial()
	defer cConn.Close()
	fed := v1.NewFederationServiceClient(cConn)
	sub, err := fed.Subscribe(ctx, &v1.SubscribeRequest{SubscriberId: "test"})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan *v1.ControlResponse, 16)
	go func() {
		for {
			snap, err := sub.Recv()
			if err != nil {
				return
			}
			if cr := snap.GetCommandResult(); cr != nil {
				results <- cr
			}
		}
	}()

	// Subscribing opens the observability window; the daemon must learn of it.
	gotOpen := false
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline) && !gotOpen; {
		select {
		case w := <-windows:
			if w.GetLogs() && w.GetProcesses() {
				gotOpen = true
			}
		case <-time.After(200 * time.Millisecond):
		}
	}
	if !gotOpen {
		t.Fatal("daemon never received an open demand window after a subscriber connected")
	}

	// Run a command; the directive routes to the daemon and the result comes back.
	ack, err := fed.RunCommand(ctx, &v1.ControlRequest{RequestId: "r1", HostId: "d1", AllowlistedCmd: "process.list"})
	if err != nil {
		t.Fatal(err)
	}
	if !ack.GetAccepted() {
		t.Fatalf("RunCommand not accepted: %s", ack.GetError())
	}
	select {
	case cr := <-results:
		if cr.GetRequestId() != "r1" || cr.GetStdout() != "ran process.list" {
			t.Fatalf("unexpected result: %+v", cr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no command result returned through the hub")
	}

	// An unknown host is refused without routing anywhere.
	ack2, _ := fed.RunCommand(ctx, &v1.ControlRequest{RequestId: "r2", HostId: "ghost", AllowlistedCmd: "process.list"})
	if ack2.GetAccepted() {
		t.Fatal("an unknown host must be refused")
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting: %s", what)
}
