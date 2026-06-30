// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"strings"
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

func TestProcessesModalOpensAndCloses(t *testing.T) {
	// v2.2: "p" opens the process table (was "t"); "t" is now the full-screen top view.
	m := Model{reg: obsReg(t), detail: true, height: 40, width: 100}
	if m = press(m, "p"); m.modal != modalTop {
		t.Fatalf("p should open the process table, got %d", m.modal)
	}
	if m = press(m, "esc"); m.modal != modalNone {
		t.Fatalf("esc should close the process table, got %d", m.modal)
	}
}

func TestProcessesModalTitleSaysProcesses(t *testing.T) {
	m := Model{reg: obsReg(t), detail: true, height: 40, width: 100}
	m = press(m, "p")
	if got := m.ModalView(); !strings.Contains(strings.ToLower(ansiRe.ReplaceAllString(got, "")), "processes") {
		t.Fatalf("process modal title should say 'processes', got:\n%s", got)
	}
}

func TestTopViewEntersAndExits(t *testing.T) {
	// From the grid, "t" enters the full-screen top view for the focused host;
	// "esc" and "q" each leave it.
	base := Model{reg: obsReg(t), detail: false, height: 40, width: 100}
	for _, exitKey := range []string{"esc", "q"} {
		next, _ := base.Update(mkey("t"))
		m := next.(Model)
		if m.top == nil {
			t.Fatalf("t should enter the top view")
		}
		out, _ := m.Update(mkey(exitKey))
		if out.(Model).top != nil {
			t.Fatalf("%q should exit the top view", exitKey)
		}
	}
}

func TestDetailFooterShowsTopAndProc(t *testing.T) {
	m := Model{reg: obsReg(t), detail: true, height: 40, width: 120}
	h, _ := m.selectedHost()
	f := ansiRe.ReplaceAllString(m.detailFooter(h), "")
	for _, want := range []string{"top", "proc"} {
		if !strings.Contains(f, want) {
			t.Errorf("detail footer missing %q: %q", want, f)
		}
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
	if m = press(m, "p"); m.modal != modalNone {
		t.Fatal("p must do nothing without a process table")
	}
}
