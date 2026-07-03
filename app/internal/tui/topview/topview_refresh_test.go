// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package topview

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/domain"
)

func TestRefreshPreservesFocusAndScroll(t *testing.T) {
	th := darkMode(t)
	h := domain.HostView{Host: domain.Host{ID: "h", DisplayName: "h"},
		State:        domain.StateOnline,
		LastSnapshot: []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 50}}}
	// A short terminal so the panel stack overflows and Tab page-scrolls.
	m := New(h, map[string][]float64{}, th, 120, 10)
	for i := 0; i < 3; i++ { // Tab focus down a few panels
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}
	focus := m.group.Focus()
	page := m.group.PageScroll()
	if focus == 0 {
		t.Fatal("Tab did not move focus")
	}
	r := m.Refresh(h, map[string][]float64{})
	if r.group.Focus() != focus {
		t.Errorf("Refresh dropped focus: got %d want %d", r.group.Focus(), focus)
	}
	if r.group.PageScroll() != page {
		t.Errorf("Refresh dropped page scroll: got %d want %d", r.group.PageScroll(), page)
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
