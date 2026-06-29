// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/discovery"
	"heimdall/app/internal/tui/theme"
)

func TestHubPickerSelectsAndConfirms(t *testing.T) {
	th, err := theme.Load()
	if err != nil {
		t.Fatal(err)
	}
	md, _ := th.Mode("")
	hubs := []discovery.Found{
		{Name: "home", Addr: "10.0.0.1:9090"},
		{Name: "central", Addr: "10.0.0.2:9090"},
	}
	m := hubPicker{hubs: hubs, mode: md}

	step := func(k tea.KeyMsg) {
		next, _ := m.Update(k)
		m = next.(hubPicker)
	}
	step(tea.KeyMsg{Type: tea.KeyDown})
	if m.sel != 1 {
		t.Fatalf("down should move selection to 1, got %d", m.sel)
	}
	step(tea.KeyMsg{Type: tea.KeyUp})
	if m.sel != 0 {
		t.Fatalf("up should move selection to 0, got %d", m.sel)
	}
	step(tea.KeyMsg{Type: tea.KeyDown})
	step(tea.KeyMsg{Type: tea.KeyEnter})
	if m.chosen != "10.0.0.2:9090" {
		t.Fatalf("enter should choose the selected hub, got %q", m.chosen)
	}

	// View renders both hubs.
	if v := m.View(); !contains(v, "home") || !contains(v, "central") {
		t.Fatalf("view should list both hubs:\n%s", v)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
