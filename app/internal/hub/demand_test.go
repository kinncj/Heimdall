// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package hub

import (
	"context"
	"testing"
	"time"

	v1 "heimdall/common/proto/monitoring/v1"
)

// A subscribing dashboard opens the demand window; the last one leaving closes it
// (ADR 0018, v2 demand-driven push).
func TestDemandWindowTracksSubscribers(t *testing.T) {
	h := New(10*time.Second, 30*time.Second)
	ctrl := make(chan *v1.StreamControl, 8)
	h.addDaemon(ctrl)

	recvWindow := func(what string) *v1.ObservabilityWindow {
		select {
		case c := <-ctrl:
			w := c.GetObservability()
			if w == nil {
				t.Fatalf("%s: expected an observability directive, got %T", what, c.GetControl())
			}
			return w
		default:
			t.Fatalf("%s: no directive sent", what)
			return nil
		}
	}

	sub := make(chan *v1.Snapshot, 1)
	h.addSub(sub)
	if w := recvWindow("subscribe"); !w.GetLogs() || !w.GetProcesses() {
		t.Fatalf("subscribe should open the window, got logs=%v procs=%v", w.GetLogs(), w.GetProcesses())
	}

	h.removeSub(sub)
	if w := recvWindow("unsubscribe"); w.GetLogs() || w.GetProcesses() {
		t.Fatalf("last unsubscribe should close the window, got logs=%v procs=%v", w.GetLogs(), w.GetProcesses())
	}
}

func TestObservabilityWindowReflectsDemand(t *testing.T) {
	h := New(10*time.Second, 30*time.Second)
	if w := h.observabilityWindow().GetObservability(); w.GetLogs() {
		t.Fatal("no subscribers should mean a closed window")
	}
	sub := make(chan *v1.Snapshot, 1)
	h.addSub(sub)
	if w := h.observabilityWindow().GetObservability(); !w.GetLogs() {
		t.Fatal("a subscriber should mean an open window")
	}
}

// RunCommand routes a directive to the owning daemon's sink and refuses an
// unknown host (v2 Phase 2, ADR 0018).
func TestRunCommandRouting(t *testing.T) {
	h := New(10*time.Second, 30*time.Second)
	ctrl := make(chan *v1.StreamControl, 4)
	h.bindDaemon("web-01", ctrl)

	ack, err := h.RunCommand(context.Background(), &v1.ControlRequest{
		RequestId: "r1", HostId: "web-01", AllowlistedCmd: "process.list",
	})
	if err != nil || !ack.GetAccepted() {
		t.Fatalf("connected host should accept, got accepted=%v err=%v", ack.GetAccepted(), err)
	}
	select {
	case c := <-ctrl:
		run := c.GetRun()
		if run == nil || run.GetRequestId() != "r1" || run.GetAllowlistedCmd() != "process.list" {
			t.Fatalf("directive not routed correctly: %+v", c.GetControl())
		}
	default:
		t.Fatal("no directive was routed to the daemon sink")
	}

	ack2, _ := h.RunCommand(context.Background(), &v1.ControlRequest{
		RequestId: "r2", HostId: "ghost", AllowlistedCmd: "process.list",
	})
	if ack2.GetAccepted() {
		t.Fatal("an unconnected host must not be accepted")
	}
	if ack2.GetError() == "" {
		t.Fatal("an unconnected host should report an error")
	}
}
