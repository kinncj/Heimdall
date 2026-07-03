// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
	"strings"

	"heimdall/app/internal/command"
	"heimdall/app/internal/domain"
)

// labelCmd marks a host that exposes on-demand commands (daemon --allow-commands).
const labelCmd = "_cmd"

// hasCmd reports whether the host exposes on-demand commands (gates the modal).
func hasCmd(h domain.HostView) bool { return h.Host.Context.Labels[labelCmd] != "" }

// cmdModalKeys are the no-argument allow-listed commands offered in the modal.
// dir.list needs a path, so it stays CLI-only for now.
func cmdModalKeys() []string {
	out := make([]string, 0, len(command.Keys()))
	for _, k := range command.Keys() {
		if k != "dir.list" {
			out = append(out, k)
		}
	}
	return out
}

// filteredCmdKeys narrows the command list by the picker's name filter
// (case-insensitive substring). An empty query returns every command.
func (m Model) filteredCmdKeys() []string {
	all := cmdModalKeys()
	q := strings.ToLower(strings.TrimSpace(m.cmdQuery))
	if q == "" {
		return all
	}
	out := make([]string, 0, len(all))
	for _, k := range all {
		if strings.Contains(strings.ToLower(k), q) {
			out = append(out, k)
		}
	}
	return out
}

func (m Model) cmdResultName() string {
	keys := m.filteredCmdKeys()
	if m.cmdSel < len(keys) {
		return keys[m.cmdSel]
	}
	return ""
}

func (m Model) cmdListBody() []string {
	val, _ := m.mode.Role("value")
	focus, _ := m.mode.Role("focus")
	muted, _ := m.mode.Role("text_muted")
	keys := m.filteredCmdKeys()
	if len(keys) == 0 {
		return []string{muted.Style().Render("  no commands match the filter")}
	}
	out := make([]string, len(keys))
	for i, k := range keys {
		if i == m.cmdSel {
			out[i] = focus.Style().Render("  ▸ " + k)
		} else {
			out[i] = val.Style().Render("    " + k)
		}
	}
	return out
}

// updateCmdSearch handles keystrokes while the command filter input is open: type
// to narrow by name, backspace to edit, enter to keep, esc to clear. Mirrors the
// log-view search so the two pickers behave the same.
func (m Model) updateCmdSearch(s string, runes []rune) Model {
	switch s {
	case "enter":
		m.cmdSearching = false
	case "esc":
		m.cmdSearching = false
		m.cmdQuery = ""
	case "backspace":
		if r := []rune(m.cmdQuery); len(r) > 0 {
			m.cmdQuery = string(r[:len(r)-1])
		}
	default:
		if len(runes) > 0 {
			m.cmdQuery += string(runes)
		}
	}
	// A narrowed list resets the cursor and scroll so the selection stays valid.
	if m.cmdSel >= len(m.filteredCmdKeys()) {
		m.cmdSel = 0
	}
	m.modalScroll = 0
	return m
}

func (m Model) cmdResultBody(h domain.HostView, w int) []string {
	muted, _ := m.mode.Role("text_muted")
	val, _ := m.mode.Role("value")
	if m.runCmd == nil {
		return []string{muted.Style().Render("  commands are unavailable (no hub connection)")}
	}
	cr := h.LastCommand
	if cr == nil || cr.RequestID != m.cmdReqID {
		return []string{muted.Style().Render("  running…")}
	}

	head := fmt.Sprintf("  status: %s   exit: %d", statusWord(cr.Status), cr.ExitCode)
	if cr.Status != domain.StatusOK {
		st, _ := m.mode.State("error")
		head = st.Style().Render(head)
	} else {
		head = muted.Style().Render(head)
	}
	out := []string{head, ""}

	if s := strings.TrimRight(cr.Stdout, "\n"); s != "" {
		for _, line := range strings.Split(s, "\n") {
			out = append(out, "  "+val.Style().Render(clip(line, w-4)))
		}
	}
	// Keep stderr visually distinct and never merged onto a stdout line.
	if s := strings.TrimRight(cr.Stderr, "\n"); s != "" {
		al, _ := m.mode.State("error")
		out = append(out, "", muted.Style().Render("  stderr:"))
		for _, line := range strings.Split(s, "\n") {
			out = append(out, "  "+al.Style().Render(clip(line, w-4)))
		}
	}
	if cr.Truncated {
		out = append(out, muted.Style().Render("  [output truncated]"))
	}
	return out
}

func statusWord(s domain.MetricStatus) string {
	switch s {
	case domain.StatusOK:
		return "ok"
	case domain.StatusInsufficientPermission:
		return "insufficient_permission"
	case domain.StatusError:
		return "error"
	default:
		return "unspecified"
	}
}
