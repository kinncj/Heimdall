// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/transport"
	v1 "heimdall/common/proto/monitoring/v1"
)

// A retained snapshot for a long-dead daemon carries an old timestamp. The
// dashboard must age it from that timestamp — not from wall-clock now — so a
// reconnect or --snapshot frame does not resurrect the host as ONLINE.
func TestFoldSnapshotHonorsTimestamp(t *testing.T) {
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	stale := transport.ToSnapshot(
		"host-a",
		[]domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 12}},
		nil,
		0,
		time.Now().Add(-60*time.Second),
	)

	foldSnapshot(reg, stale)
	reg.Evaluate(time.Now())

	hv, ok := reg.Host("host-a")
	if !ok {
		t.Fatal("host-a not registered after foldSnapshot")
	}
	if hv.State != domain.StateOffline {
		t.Fatalf("60s-old snapshot should be OFFLINE, got %v", hv.State)
	}
}

func TestFoldSnapshotFreshIsOnline(t *testing.T) {
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	fresh := transport.ToSnapshot(
		"host-b",
		[]domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 5}},
		nil,
		0,
		time.Now(),
	)

	foldSnapshot(reg, fresh)
	reg.Evaluate(time.Now())

	hv, _ := reg.Host("host-b")
	if hv.State != domain.StateOnline {
		t.Fatalf("fresh snapshot should be ONLINE, got %v", hv.State)
	}
}

// Defensive: a snapshot with no timestamp must fall back to now rather than
// 1970 (which would make every host instantly OFFLINE).
func TestFoldSnapshotZeroTimestampFallsBackToNow(t *testing.T) {
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	snap := &v1.Snapshot{HostId: "host-c", TsUnixMillis: 0}

	foldSnapshot(reg, snap)
	reg.Evaluate(time.Now())

	hv, _ := reg.Host("host-c")
	if hv.State != domain.StateOnline {
		t.Fatalf("zero-ts snapshot should fall back to now (ONLINE), got %v", hv.State)
	}
}
