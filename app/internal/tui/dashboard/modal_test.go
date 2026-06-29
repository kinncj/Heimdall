// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/domain"
)

func obsReg(t *testing.T) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "h", DisplayName: "h",
		Context: domain.HostContext{Labels: map[string]string{"_logs": "app,sys", "_proc": "1"}}}, now)
	reg.Observe("h", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 1}}, nil, now)
	reg.RecordPush("h",
		[]domain.ProcessRow{{PID: 1, Command: "init"}},
		now,
		[]domain.LogLine{{Source: "app", Line: "hello"}, {Source: "sys", Line: "boot"}})
	reg.Evaluate(now)
	return reg
}

func mkey(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func press(m Model, s string) Model {
	next, _ := m.updateDetail(mkey(s))
	return next.(Model)
}

func TestModalUnwindsWithEscape(t *testing.T) {
	m := Model{reg: obsReg(t), detail: true, height: 40, width: 100}

	if m = press(m, "l"); m.modal != modalLogList {
		t.Fatalf("l should open the log list, got %d", m.modal)
	}
	if m = press(m, "down"); m.modalSel != 1 {
		t.Fatalf("down should move selection to 1, got %d", m.modalSel)
	}
	if m = press(m, "enter"); m.modal != modalLogView || m.logSource != "sys" {
		t.Fatalf("enter should open the sys log view, got modal=%d src=%q", m.modal, m.logSource)
	}
	// esc unwinds: view -> list -> detail.
	if m = press(m, "esc"); m.modal != modalLogList {
		t.Fatalf("esc from log view should return to the list, got %d", m.modal)
	}
	if m = press(m, "esc"); m.modal != modalNone {
		t.Fatalf("esc from the list should close the modal, got %d", m.modal)
	}
	if m = press(m, "esc"); m.detail {
		t.Fatal("esc with no modal should leave the detail view")
	}
}

func TestTopModalOpensAndCloses(t *testing.T) {
	m := Model{reg: obsReg(t), detail: true, height: 40, width: 100}
	if m = press(m, "t"); m.modal != modalTop {
		t.Fatalf("t should open top, got %d", m.modal)
	}
	if m = press(m, "esc"); m.modal != modalNone {
		t.Fatalf("esc should close top, got %d", m.modal)
	}
}

func TestModalAffordancesGatedByCapability(t *testing.T) {
	// A host advertising nothing must not open the modals.
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "bare", DisplayName: "bare"}, now)
	reg.Observe("bare", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 1}}, nil, now)
	reg.Evaluate(now)
	m := Model{reg: reg, detail: true, height: 40, width: 100}
	if m = press(m, "l"); m.modal != modalNone {
		t.Fatal("l must do nothing without log sources")
	}
	if m = press(m, "t"); m.modal != modalNone {
		t.Fatal("t must do nothing without a process table")
	}
}
