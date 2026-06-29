// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func sortReg(t *testing.T) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "h", DisplayName: "h",
		Context: domain.HostContext{Labels: map[string]string{"_proc": "1", "_logs": "app"}}}, now)
	reg.Observe("h", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 1}}, nil, now)
	reg.RecordPush("h", []domain.ProcessRow{
		{PID: 1, CPUPct: 10, MemPct: 50, Command: "a"},
		{PID: 2, CPUPct: 90, MemPct: 5, Command: "b"},
		{PID: 3, CPUPct: 50, MemPct: 30, Command: "c"},
	}, now, []domain.LogLine{
		{Source: "app", Line: "boot sequence started"},
		{Source: "app", Line: "request handled"},
		{Source: "app", Line: "boot complete"},
	})
	reg.Evaluate(now)
	return reg
}

func commands(rows []domain.ProcessRow) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Command
	}
	return out
}

func TestTopSortDefaultsToCPU(t *testing.T) {
	m := Model{reg: sortReg(t)}
	h, _ := m.selectedHost()
	if got := commands(m.sortedProcesses(h)); !equalStrings(got, []string{"b", "c", "a"}) {
		t.Fatalf("default sort = %v, want cpu-desc [b c a]", got)
	}
}

func TestTopSortPickerAppliesAndPersists(t *testing.T) {
	var persisted string
	m := Model{reg: sortReg(t), detail: true, height: 40, width: 100, modal: modalTop}
	m.persistSort = func(k string) { persisted = k }

	m = press(m, "s")
	if m.modal != modalTopSort || m.topSortSel != 0 {
		t.Fatalf("s should open the picker at cpu, got modal=%d sel=%d", m.modal, m.topSortSel)
	}
	m = press(m, "down") // -> mem
	m = press(m, "enter")
	if m.modal != modalTop || m.topSort != "mem" {
		t.Fatalf("enter should apply mem and return to top, got modal=%d sort=%q", m.modal, m.topSort)
	}
	if persisted != "mem" {
		t.Fatalf("sort should persist, got %q", persisted)
	}
	h, _ := m.selectedHost()
	if got := commands(m.sortedProcesses(h)); !equalStrings(got, []string{"a", "c", "b"}) {
		t.Fatalf("mem sort = %v, want mem-desc [a c b]", got)
	}
}

func TestTopSortPickerEscCancels(t *testing.T) {
	m := Model{reg: sortReg(t), detail: true, height: 40, width: 100, modal: modalTop}
	m = press(m, "s")
	m = press(m, "down")
	m = press(m, "esc")
	if m.modal != modalTop || m.topSort != "" {
		t.Fatalf("esc should cancel without changing the sort, got modal=%d sort=%q", m.modal, m.topSort)
	}
}

func TestLogSearchFiltersByQuery(t *testing.T) {
	m := Model{reg: sortReg(t), detail: true, height: 40, width: 100, modal: modalLogView, logSource: "app"}

	// "/" opens the search; typing narrows to matching lines.
	m = press(m, "/")
	if !m.logSearching {
		t.Fatal("/ should open the log search input")
	}
	for _, r := range "boot" {
		m = press(m, string(r))
	}
	if m.logQuery != "boot" {
		t.Fatalf("query = %q, want boot", m.logQuery)
	}
	h, _ := m.selectedHost()
	body := m.logViewBody(h, 100)
	if len(body) != 2 { // "boot sequence started" + "boot complete"
		t.Fatalf("search should match 2 lines, got %d: %v", len(body), body)
	}

	// enter keeps the query; esc clears it, then esc steps back to the list.
	m = press(m, "enter")
	if m.logSearching || m.logQuery != "boot" {
		t.Fatalf("enter should keep the query and close the input")
	}
	m = press(m, "esc")
	if m.logQuery != "" || m.modal != modalLogView {
		t.Fatalf("first esc should clear the query, staying in the view")
	}
	m = press(m, "esc")
	if m.modal != modalLogList {
		t.Fatalf("second esc should return to the source list, got %d", m.modal)
	}
}
