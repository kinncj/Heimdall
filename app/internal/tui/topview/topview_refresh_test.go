// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package topview

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/domain"
)

func TestRefreshPreservesScroll(t *testing.T) {
	th := darkMode(t)
	h := domain.HostView{Host: domain.Host{ID: "h", DisplayName: "h"},
		State:        domain.StateOnline,
		LastSnapshot: []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 50}}}
	m := New(h, map[string][]float64{}, th, 120, 8)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // scroll to 1
	if m.scroll == 0 {
		t.Skip("content fit on screen; nothing to scroll")
	}
	before := m.scroll
	r := m.Refresh(h, map[string][]float64{})
	if r.scroll != before {
		t.Errorf("Refresh dropped scroll: got %d want %d", r.scroll, before)
	}
}

func TestResizeChangesWidth(t *testing.T) {
	th := darkMode(t)
	h := domain.HostView{Host: domain.Host{ID: "h"}, State: domain.StateOnline}
	m := New(h, nil, th, 120, 40).Resize(30, 20)
	if m.width != 30 || m.height != 20 {
		t.Fatalf("resize failed: %dx%d", m.width, m.height)
	}
	if layout(m.width) != tierTiny {
		t.Errorf("width 30 should be tiny tier")
	}
}
