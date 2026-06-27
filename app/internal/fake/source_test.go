// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package fake

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

// The demo fleet must stay live across many ticks (regression: it used to be
// observed once and decay to all-offline as wall-clock advanced).
func TestFleetStaysLiveAcrossTicks(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	s := New(now)
	for i := 0; i < 120; i++ {
		now = now.Add(time.Second)
		s.Tick(now)
	}

	var online, stale, offline int
	for _, h := range s.Registry().Hosts() {
		switch h.State {
		case domain.StateOnline:
			online++
		case domain.StateStale:
			stale++
		case domain.StateOffline:
			offline++
		}
	}
	if online < 4 {
		t.Fatalf("after 120 ticks only %d hosts online; fleet decayed", online)
	}
	if stale < 1 {
		t.Errorf("expected at least one stale host, got %d", stale)
	}
	if offline < 1 {
		t.Errorf("expected at least one offline host, got %d", offline)
	}
}
