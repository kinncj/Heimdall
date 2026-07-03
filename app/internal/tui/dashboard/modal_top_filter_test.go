// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

func TestMatchesTopRow(t *testing.T) {
	var m Model
	p := domain.ProcessRow{Command: "/usr/bin/coreaudiod"}
	if !m.matchesTopRow(p) {
		t.Error("empty filter should match every process")
	}
	m.topQuery = "AUDIO" // case-insensitive
	if !m.matchesTopRow(p) {
		t.Error("filter should match the command case-insensitively")
	}
	if m.matchesTopRow(domain.ProcessRow{Command: "launchd"}) {
		t.Error("non-matching command should be filtered out")
	}
}

func TestTopBodyFiltersByCommand(t *testing.T) {
	h := domain.HostView{
		Host: domain.Host{ID: "h"},
		Processes: []domain.ProcessRow{
			{PID: 1, CPUPct: 9, Command: "launchd"},
			{PID: 2, CPUPct: 8, Command: "WindowServer"},
			{PID: 3, CPUPct: 7, Command: "coreaudiod"},
		},
	}
	m := Model{mode: darkMode(t), width: 80}

	m.topQuery = "window"
	body := strings.Join(m.topBody(h, 80), "\n")
	if !strings.Contains(body, "WindowServer") {
		t.Errorf("filtered body should keep the matching process:\n%s", body)
	}
	if strings.Contains(body, "launchd") || strings.Contains(body, "coreaudiod") {
		t.Errorf("filtered body should drop non-matching processes:\n%s", body)
	}

	m.topQuery = "nomatch"
	empty := strings.Join(m.topBody(h, 80), "\n")
	if !strings.Contains(empty, "no processes match") {
		t.Errorf("a filter matching nothing should show an empty state:\n%s", empty)
	}
}

func TestUpdateTopSearchTypingAndClear(t *testing.T) {
	var m Model
	m = m.updateTopSearch("w", []rune("w"))
	m = m.updateTopSearch("m", []rune("m")) // "wm"
	if m.topQuery != "wm" {
		t.Fatalf("typing built %q, want wm", m.topQuery)
	}
	m = m.updateTopSearch("backspace", nil)
	if m.topQuery != "w" {
		t.Fatalf("backspace left %q, want w", m.topQuery)
	}
	m = m.updateTopSearch("esc", nil)
	if m.topQuery != "" || m.topSearching {
		t.Fatalf("esc should clear and close, got q=%q searching=%v", m.topQuery, m.topSearching)
	}
}
