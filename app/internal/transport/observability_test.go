// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package transport

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func TestObservabilityRoundTrip(t *testing.T) {
	at := time.UnixMilli(1_700_000_000_000)
	procs := []domain.ProcessRow{
		{PID: 1, PPID: 0, CPUPct: 2.5, MemPct: 1.2, Command: "init"},
		{PID: 42, PPID: 1, CPUPct: 88, MemPct: 30, Command: "heimdall-daemon"},
	}
	logs := []domain.LogLine{
		{Source: "app", At: at, Line: "boot", Level: "info"},
		{Source: "app", At: at, Line: "dropped", RateLimited: true},
	}

	snap := ToSnapshot("h1", nil, nil, 7, at)
	AttachObservability(snap, procs, at, "h1", logs)

	gotProcs, gotAt, gotLogs := ObservabilityFromSnapshot(snap)
	if len(gotProcs) != 2 || gotProcs[1].PID != 42 || gotProcs[1].Command != "heimdall-daemon" {
		t.Fatalf("processes round-trip = %+v", gotProcs)
	}
	if gotProcs[1].CPUPct < 87 || gotProcs[1].CPUPct > 89 {
		t.Fatalf("cpu%% lost precision: %v", gotProcs[1].CPUPct)
	}
	if !gotAt.Equal(at) {
		t.Fatalf("processesAt = %v, want %v", gotAt, at)
	}
	if len(gotLogs) != 2 || gotLogs[0].Line != "boot" || !gotLogs[1].RateLimited {
		t.Fatalf("logs round-trip = %+v", gotLogs)
	}
}

func TestAttachObservabilityEmptyStaysAbsent(t *testing.T) {
	snap := ToSnapshot("h1", nil, nil, 1, time.UnixMilli(0))
	AttachObservability(snap, nil, time.Time{}, "h1", nil)
	if len(snap.GetProcesses()) != 0 || len(snap.GetLogLines()) != 0 {
		t.Fatal("empty push must leave snapshot fields absent")
	}
}
