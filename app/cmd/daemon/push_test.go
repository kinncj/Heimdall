// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
	v1 "heimdall/common/proto/monitoring/v1"
)

// Multiple command results that finish before the next send must all be
// delivered — one per snapshot — not collapsed onto a single slot (the wire
// carries one result per snapshot).
func TestPusherQueuesMultipleResults(t *testing.T) {
	p := &pusher{}
	p.setResult(&v1.ControlResponse{RequestId: "a"})
	p.setResult(&v1.ControlResponse{RequestId: "b"})

	first := p.drainResult()
	if first == nil || first.GetRequestId() != "a" {
		t.Fatalf("want FIFO order a first, got %v", first)
	}
	if !p.hasResults() {
		t.Fatal("the second result must still be queued")
	}
	second := p.drainResult()
	if second == nil || second.GetRequestId() != "b" {
		t.Fatalf("want b second, got %v", second)
	}
	if p.hasResults() || p.drainResult() != nil {
		t.Fatal("queue must be empty after draining both")
	}
}

func TestPusherGatesOnDemandWindow(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	p := &pusher{wantLogs: true, wantProcs: true}

	// Open window: drain returns both the table and the lines.
	p.setProcs([]domain.ProcessRow{{PID: 1, Command: "init"}}, now)
	p.addLine("app", "hello")
	procs, _, logs := p.drain()
	if len(procs) != 1 || len(logs) != 1 {
		t.Fatalf("open window should drain both, got procs=%d logs=%d", len(procs), len(logs))
	}

	// Closed window: nothing is sent, and buffered lines are dropped (not retained).
	p.setProcs([]domain.ProcessRow{{PID: 2, Command: "node"}}, now)
	p.addLine("app", "dropped")
	p.setWindow(false, false)
	procs, _, logs = p.drain()
	if procs != nil || len(logs) != 0 {
		t.Fatalf("closed window should drain nothing, got procs=%v logs=%v", procs, logs)
	}
	p.addLine("app", "still dropped")
	if _, _, logs = p.drain(); len(logs) != 0 {
		t.Fatal("closed window must not accumulate lines")
	}

	// Re-open: pushing resumes.
	p.setWindow(true, true)
	p.setProcs([]domain.ProcessRow{{PID: 3}}, now)
	p.addLine("app", "back")
	if procs, _, logs = p.drain(); len(procs) != 1 || len(logs) != 1 {
		t.Fatalf("re-opened window should drain again, got procs=%d logs=%d", len(procs), len(logs))
	}
}
