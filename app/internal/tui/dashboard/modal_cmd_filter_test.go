// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func TestFilteredCmdKeysNarrowsByName(t *testing.T) {
	var m Model // filtering is pure string work — no theme needed

	all := m.filteredCmdKeys()
	if len(all) == 0 {
		t.Fatal("expected some commands with an empty filter")
	}
	if contains(all, "dir.list") {
		t.Error("dir.list needs a path arg and must stay out of the picker")
	}

	m.cmdQuery = "os"
	got := m.filteredCmdKeys()
	if !contains(got, "os.info") {
		t.Errorf("filter %q should keep os.info, got %v", m.cmdQuery, got)
	}
	for _, k := range got {
		if !strings.Contains(k, "os") {
			t.Errorf("filter %q returned non-matching %q", m.cmdQuery, k)
		}
	}

	m.cmdQuery = "zzz-nothing"
	if n := len(m.filteredCmdKeys()); n != 0 {
		t.Errorf("a non-matching filter should return no commands, got %d", n)
	}
}

func TestUpdateCmdSearchTypingAndClear(t *testing.T) {
	var m Model
	m.cmdSel = 3

	m = m.updateCmdSearch("o", []rune("o"))
	m = m.updateCmdSearch("s", []rune("s"))
	if m.cmdQuery != "os" {
		t.Fatalf("typing built query %q, want os", m.cmdQuery)
	}
	// Narrowing to one result must pull the (stale) selection back in range.
	if m.cmdSel >= len(m.filteredCmdKeys()) {
		t.Fatalf("cmdSel %d out of range after filter (%d rows)", m.cmdSel, len(m.filteredCmdKeys()))
	}

	m = m.updateCmdSearch("backspace", nil)
	if m.cmdQuery != "o" {
		t.Fatalf("backspace left %q, want o", m.cmdQuery)
	}
	m = m.updateCmdSearch("esc", nil)
	if m.cmdQuery != "" || m.cmdSearching {
		t.Fatalf("esc should clear and close the filter, got q=%q searching=%v", m.cmdQuery, m.cmdSearching)
	}
}

func TestMatchesCmdOut(t *testing.T) {
	var m Model
	if !m.matchesCmdOut("anything") {
		t.Error("empty filter should match every line")
	}
	m.cmdOutQuery = "ERROR"
	if !m.matchesCmdOut("some error here") { // case-insensitive
		t.Error("filter should match case-insensitively")
	}
	if m.matchesCmdOut("all good") {
		t.Error("non-matching line should be filtered out")
	}
}

func TestCmdResultBodyFiltersOutput(t *testing.T) {
	h := domain.HostView{
		Host: domain.Host{ID: "h"},
		LastCommand: &domain.CommandResult{
			RequestID: "r1", Status: domain.StatusOK,
			Stdout: "alpha line\nbeta line\ngamma line",
		},
	}
	m := Model{mode: darkMode(t), width: 80, runCmd: func(string, string, []string, string) {}, cmdReqID: "r1"}

	m.cmdOutQuery = "beta"
	body := strings.Join(m.cmdResultBody(h, 80), "\n")
	if !strings.Contains(body, "beta") {
		t.Errorf("filtered body should keep the matching line:\n%s", body)
	}
	if strings.Contains(body, "alpha") || strings.Contains(body, "gamma") {
		t.Errorf("filtered body should drop non-matching lines:\n%s", body)
	}

	m.cmdOutQuery = "nomatch"
	empty := strings.Join(m.cmdResultBody(h, 80), "\n")
	if !strings.Contains(empty, "no output lines match") {
		t.Errorf("a filter matching nothing should show an empty state:\n%s", empty)
	}
}
