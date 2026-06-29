// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/theme"
)

func bigFleet(t *testing.T, n int) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	for i := 0; i < n; i++ {
		id := domain.HostID(fmt.Sprintf("host-%02d", i))
		reg.Enroll(domain.Host{ID: id, DisplayName: string(id)}, now)
		reg.Observe(id, []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 10}}, nil, now)
	}
	reg.Evaluate(now)
	return reg
}

func smallModel(t *testing.T, reg *domain.HostRegistry, h int) Model {
	t.Helper()
	th, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	md, ok := th.Mode("")
	if !ok {
		t.Fatal("default theme mode missing")
	}
	m := New(md, reg, time.Unix(1_700_000_000, 0))
	m.width, m.height = 100, h
	return m
}

// The reported bug: on a screen shorter than the fleet, GridView must not emit a
// frame taller than the terminal (which scrolls the header off-screen and makes
// filtering look inert).
func TestGridViewFitsTerminalHeight(t *testing.T) {
	m := smallModel(t, bigFleet(t, 25), 14)
	lines := strings.Count(m.GridView(), "\n") + 1
	if lines > 14 {
		t.Fatalf("GridView rendered %d lines on a height-14 terminal, want <= 14", lines)
	}
}

func TestGridViewKeepsSelectedHostVisible(t *testing.T) {
	m := smallModel(t, bigFleet(t, 25), 14)
	m.cursor = 24 // last host, well below the fold
	out := m.GridView()
	if !strings.Contains(out, "host-24") {
		t.Fatalf("selected host host-24 not visible in overflowed grid:\n%s", out)
	}
	if lines := strings.Count(out, "\n") + 1; lines > 14 {
		t.Fatalf("grid still overflows: %d lines", lines)
	}
}

func TestGridViewShowsHiddenRowIndicator(t *testing.T) {
	m := smallModel(t, bigFleet(t, 25), 14)
	if out := m.GridView(); !strings.Contains(out, "more") {
		t.Fatalf("expected a hidden-rows indicator when the list overflows:\n%s", out)
	}
}
