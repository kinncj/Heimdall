// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"testing"
	"time"
)

func TestRecordPushStoresProcessesAndRingsLogs(t *testing.T) {
	reg := NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(Host{ID: "h"}, now)

	reg.RecordPush("h", []ProcessRow{{PID: 1, Command: "init"}}, now, []LogLine{{Source: "app", Line: "a"}})
	reg.RecordPush("h", nil, time.Time{}, []LogLine{{Source: "app", Line: "b"}})

	v, ok := reg.Host("h")
	if !ok {
		t.Fatal("host missing")
	}
	if len(v.Processes) != 1 || v.Processes[0].Command != "init" {
		t.Fatalf("processes = %+v", v.Processes)
	}
	if !v.ProcessesAt.Equal(now) {
		t.Fatalf("processesAt = %v", v.ProcessesAt)
	}
	// nil processes on the second push must not wipe the table.
	if len(v.Processes) != 1 {
		t.Fatal("nil processes wiped the table")
	}
	if len(v.Logs) != 2 || v.Logs[0].Line != "a" || v.Logs[1].Line != "b" {
		t.Fatalf("logs = %+v", v.Logs)
	}
}

func TestRecordPushBoundsLogRing(t *testing.T) {
	reg := NewHostRegistry(10*time.Second, 30*time.Second)
	reg.Enroll(Host{ID: "h"}, time.Unix(0, 0))
	for i := 0; i < LogRingCap+50; i++ {
		reg.RecordPush("h", nil, time.Time{}, []LogLine{{Source: "app", Line: "x"}})
	}
	v, _ := reg.Host("h")
	if len(v.Logs) != LogRingCap {
		t.Fatalf("log ring = %d, want capped at %d", len(v.Logs), LogRingCap)
	}
}
