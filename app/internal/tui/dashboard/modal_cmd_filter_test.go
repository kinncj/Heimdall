// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"strings"
	"testing"
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
