// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"sort"
	"strings"

	"heimdall/app/internal/domain"
)

// topSortOption is one process-table sort column (v2, ADR 0019): a persisted key,
// a menu label, and a less function. Registered in priority order, like the
// grouping/column/matcher registries — adding a sort is registering one.
type topSortOption struct {
	key   string
	label string
	less  func(a, b domain.ProcessRow) bool
}

func topSortOptions() []topSortOption {
	return []topSortOption{
		{"cpu", "CPU%  (high → low)", func(a, b domain.ProcessRow) bool { return a.CPUPct > b.CPUPct }},
		{"mem", "MEM%  (high → low)", func(a, b domain.ProcessRow) bool { return a.MemPct > b.MemPct }},
		{"pid", "PID   (low → high)", func(a, b domain.ProcessRow) bool { return a.PID < b.PID }},
		{"command", "COMMAND  (A → Z)", func(a, b domain.ProcessRow) bool { return a.Command < b.Command }},
	}
}

// activeTopSort returns the option for the model's sort key, defaulting to the
// first (cpu) when unset or unknown.
func (m Model) activeTopSort() topSortOption {
	for _, o := range topSortOptions() {
		if o.key == m.topSort {
			return o
		}
	}
	return topSortOptions()[0]
}

// activeTopSortIndex is the index of the active sort in the option registry, for
// seeding the picker's selection.
func (m Model) activeTopSortIndex() int {
	want := m.activeTopSort().key
	for i, o := range topSortOptions() {
		if o.key == want {
			return i
		}
	}
	return 0
}

// sortedProcesses returns a copy of the host's process table sorted by the active
// sort, with the command as a stable tiebreaker.
func (m Model) sortedProcesses(h domain.HostView) []domain.ProcessRow {
	rows := append([]domain.ProcessRow(nil), h.Processes...)
	less := m.activeTopSort().less
	sort.SliceStable(rows, func(i, j int) bool {
		if less(rows[i], rows[j]) {
			return true
		}
		if less(rows[j], rows[i]) {
			return false
		}
		return rows[i].Command < rows[j].Command
	})
	return rows
}

// matchesLogQuery reports whether a line satisfies the active log search (empty
// query matches everything), case-insensitive substring over source + text.
func (m Model) matchesLogQuery(l domain.LogLine) bool {
	q := strings.ToLower(strings.TrimSpace(m.logQuery))
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(l.Source+" "+l.Line), q)
}

// updateLogSearch handles keystrokes while the log search input is open: type to
// narrow, backspace to edit, enter to keep, esc to clear.
func (m Model) updateLogSearch(s string, runes []rune) Model {
	switch s {
	case "enter":
		m.logSearching = false
	case "esc":
		m.logSearching = false
		m.logQuery = ""
	case "backspace":
		if r := []rune(m.logQuery); len(r) > 0 {
			m.logQuery = string(r[:len(r)-1])
		}
	default:
		if len(runes) > 0 {
			m.logQuery += string(runes)
		}
	}
	m.modalScroll = 1 << 30 // keep pinned to the tail as results change
	return m
}

// topSortBody renders the sort picker entries.
func (m Model) topSortBody() []string {
	val, _ := m.mode.Role("value")
	focus, _ := m.mode.Role("focus")
	opts := topSortOptions()
	out := make([]string, len(opts))
	for i, o := range opts {
		mark := "    "
		style := val.Style()
		if i == m.topSortSel {
			mark, style = "  ▸ ", focus.Style()
		}
		active := ""
		if o.key == m.activeTopSort().key {
			active = "  ●"
		}
		out[i] = style.Render(mark+o.label) + active
	}
	return out
}
