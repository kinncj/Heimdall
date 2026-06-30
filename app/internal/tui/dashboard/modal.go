// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/brand"
)

// modalKind is the host-detail overlay currently open (Heimdallr's sight).
type modalKind int

const (
	modalNone      modalKind = iota
	modalLogList             // pick a log source
	modalLogView             // stream the chosen source
	modalTop                 // live process table
	modalTopSort             // pick the top sort column (v2, ADR 0019)
	modalCmdList             // pick an on-demand command (v2 Phase 2)
	modalCmdResult           // show an on-demand command's result
)

// Reserved capability labels the hub/daemon set; never shown as user tags.
const (
	labelLogs = "_logs" // comma-separated log source aliases the host pushes
	labelProc = "_proc" // host pushes a process table
)

// reservedLabel reports whether a label key is hub/daemon-managed metadata rather
// than a user Realms tag. Reserved keys are underscore-prefixed.
func reservedLabel(k string) bool { return strings.HasPrefix(k, "_") }

// selectedHost returns the host the detail view is focused on.
func (m Model) selectedHost() (domain.HostView, bool) {
	hosts := m.orderedList()
	if len(hosts) == 0 {
		return domain.HostView{}, false
	}
	i := m.cursor
	if i >= len(hosts) {
		i = len(hosts) - 1
	}
	return hosts[i], true
}

// logSourcesOf returns the host's advertised log source aliases (from _logs).
func logSourcesOf(h domain.HostView) []string {
	v := h.Host.Context.Labels[labelLogs]
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// hasProc reports whether the host exposes a process table (top).
func hasProc(h domain.HostView) bool {
	return h.Host.Context.Labels[labelProc] != "" || len(h.Processes) > 0
}

// updateDetail handles keys while the detail view (and its modals) is open. esc is
// the universal back button, unwinding one level at a time.
func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	h, ok := m.selectedHost()
	// The log search input owns every key while open (so typing "q" searches).
	if m.modal == modalLogView && m.logSearching {
		return m.updateLogSearch(msg.String(), msg.Runes), nil
	}
	// Clamp a scroll offset that may be a "pin to tail" sentinel or stale after the
	// buffer shrank, so up/down respond immediately.
	if m.modal == modalLogView || m.modal == modalTop || m.modal == modalCmdResult {
		if mx := m.modalMaxScroll(); m.modalScroll > mx {
			m.modalScroll = mx
		}
	}
	switch m.modal {
	case modalNone:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.detail = false
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.detailScroll = 0 // a new host starts at the top
			}
		case "down", "j":
			if m.cursor < len(m.orderedList())-1 {
				m.cursor++
				m.detailScroll = 0
			}
		case "shift+up", "pgup":
			m.detailScroll = clampScroll(m.detailScroll-3, m.detailMaxScroll())
		case "shift+down", "pgdown":
			m.detailScroll = clampScroll(m.detailScroll+3, m.detailMaxScroll())
		case "home":
			m.detailScroll = 0
		case "end":
			m.detailScroll = m.detailMaxScroll()
		case "l":
			if ok && len(logSourcesOf(h)) > 0 {
				m.modal, m.modalSel = modalLogList, 0
			}
		case "p":
			if ok && hasProc(h) {
				m.modal, m.modalScroll = modalTop, 0
			}
		case "c":
			if ok && hasCmd(h) && m.runCmd != nil {
				m.modal, m.cmdSel = modalCmdList, 0
			}
		}
	case modalLogList:
		srcs := logSourcesOf(h)
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.modal = modalNone
		case "up", "k":
			if m.modalSel > 0 {
				m.modalSel--
			}
		case "down", "j":
			if m.modalSel < len(srcs)-1 {
				m.modalSel++
			}
		case "enter":
			if m.modalSel < len(srcs) {
				m.logSource = srcs[m.modalSel]
				m.modal, m.modalScroll = modalLogView, 1<<30 // start pinned to the tail
			}
		}
	case modalLogView:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.logSearching = true
		case "esc":
			if m.logQuery != "" {
				m.logQuery = "" // first esc clears an active search
			} else {
				m.modal = modalLogList // then steps back to the source list
			}
		case "up", "k":
			if m.modalScroll > 0 {
				m.modalScroll--
			}
		case "down", "j":
			if m.modalScroll < m.modalMaxScroll() {
				m.modalScroll++
			}
		}
	case modalTop:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.modal = modalNone
		case "s":
			m.modal, m.topSortSel = modalTopSort, m.activeTopSortIndex()
		case "up", "k":
			if m.modalScroll > 0 {
				m.modalScroll--
			}
		case "down", "j":
			if m.modalScroll < m.modalMaxScroll() {
				m.modalScroll++
			}
		}
	case modalTopSort:
		opts := topSortOptions()
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.modal = modalTop
		case "up", "k":
			if m.topSortSel > 0 {
				m.topSortSel--
			}
		case "down", "j":
			if m.topSortSel < len(opts)-1 {
				m.topSortSel++
			}
		case "enter":
			if m.topSortSel < len(opts) {
				m.topSort = opts[m.topSortSel].key
				if m.persistSort != nil {
					m.persistSort(m.topSort)
				}
			}
			m.modal = modalTop
		}
	case modalCmdList:
		keys := cmdModalKeys()
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.modal = modalNone
		case "up", "k":
			if m.cmdSel > 0 {
				m.cmdSel--
			}
		case "down", "j":
			if m.cmdSel < len(keys)-1 {
				m.cmdSel++
			}
		case "enter":
			if m.cmdSel < len(keys) && m.runCmd != nil && ok {
				m.cmdReqID = fmt.Sprintf("dash-%d", m.now.UnixNano())
				m.runCmd(string(h.Host.ID), keys[m.cmdSel], nil, m.cmdReqID)
				m.modal, m.modalScroll = modalCmdResult, 0
			}
		}
	case modalCmdResult:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.modal = modalCmdList // back to the command list
		case "up", "k":
			if m.modalScroll > 0 {
				m.modalScroll--
			}
		case "down", "j":
			if m.modalScroll < m.modalMaxScroll() {
				m.modalScroll++
			}
		}
	}
	return m, nil
}

// ModalView renders the active detail-view overlay, height-bounded like the grid.
func (m Model) ModalView() string {
	h, ok := m.selectedHost()
	if !ok {
		return m.DetailView()
	}
	w := m.width
	if w < 24 {
		w = 24
	}
	online, total := m.fleetCounts()
	header := brand.SkinnyHeader(m.mode, w, online, total, m.now.Format("15:04:05"))
	heading, _ := m.mode.Role("heading")
	muted, _ := m.mode.Role("text_muted")
	keys, _ := m.mode.Role("keybinding")
	val, _ := m.mode.Role("value")

	dn := h.Host.DisplayName
	if dn == "" {
		dn = string(h.Host.ID)
	}

	var title, footer string
	var body []string
	switch m.modal {
	case modalLogList:
		title = heading.Style().Render("  LOGS — " + dn)
		body = m.logListBody(h)
		footer = "  " + keys.Style().Render("↑/↓") + muted.Style().Render(" pick  ") +
			keys.Style().Render("⏎") + muted.Style().Render(" open  ") +
			keys.Style().Render("esc") + muted.Style().Render(" back")
	case modalLogView:
		title = heading.Style().Render("  LOG — "+dn+" / ") + keys.Style().Render(m.logSource)
		if q := m.logQuery; q != "" || m.logSearching {
			if m.logSearching {
				q += "▏"
			}
			title += muted.Style().Render("   search: ") + val.Style().Render(q)
		}
		body = m.logViewBody(h, w)
		footer = "  " + keys.Style().Render("↑/↓") + muted.Style().Render(" scroll  ") +
			keys.Style().Render("/") + muted.Style().Render(" search  ") +
			keys.Style().Render("esc") + muted.Style().Render(" back")
	case modalTop:
		when := "—"
		if !h.ProcessesAt.IsZero() {
			when = h.ProcessesAt.Format("15:04:05")
		}
		title = heading.Style().Render("  PROCESSES — "+dn) +
			muted.Style().Render("   sort ") + val.Style().Render(m.activeTopSort().key) +
			muted.Style().Render("   updated "+when)
		body = m.topBody(h, w)
		footer = "  " + keys.Style().Render("↑/↓") + muted.Style().Render(" scroll  ") +
			keys.Style().Render("s") + muted.Style().Render(" sort  ") +
			keys.Style().Render("esc") + muted.Style().Render(" back")
	case modalTopSort:
		title = heading.Style().Render("  SORT — top processes")
		body = m.topSortBody()
		footer = "  " + keys.Style().Render("↑/↓") + muted.Style().Render(" pick  ") +
			keys.Style().Render("⏎") + muted.Style().Render(" apply  ") +
			keys.Style().Render("esc") + muted.Style().Render(" cancel")
	case modalCmdList:
		title = heading.Style().Render("  COMMAND — " + dn)
		body = m.cmdListBody()
		footer = "  " + keys.Style().Render("↑/↓") + muted.Style().Render(" pick  ") +
			keys.Style().Render("⏎") + muted.Style().Render(" run  ") +
			keys.Style().Render("esc") + muted.Style().Render(" back")
	case modalCmdResult:
		title = heading.Style().Render("  COMMAND — "+dn+" / ") + keys.Style().Render(m.cmdResultName())
		body = m.cmdResultBody(h, w)
		footer = "  " + keys.Style().Render("↑/↓") + muted.Style().Render(" scroll  ") +
			keys.Style().Render("esc") + muted.Style().Render(" commands")
	default:
		return m.DetailView()
	}

	// Bound the body to the terminal height: header(3) + blank + title + blank +
	// blank + footer ≈ 7 lines of chrome.
	maxBody := m.height - (lineCount(header) + 5)
	windowed, off := scrollWindow(body, m.modalScroll, maxBody)
	m.modalScroll = off
	return strings.Join([]string{header, "", title, "", strings.Join(windowed, "\n"), "", footer}, "\n")
}

func (m Model) logListBody(h domain.HostView) []string {
	val, _ := m.mode.Role("value")
	focus, _ := m.mode.Role("focus")
	muted, _ := m.mode.Role("text_muted")
	srcs := logSourcesOf(h)
	if len(srcs) == 0 {
		return []string{muted.Style().Render("  no log sources")}
	}
	out := make([]string, len(srcs))
	for i, s := range srcs {
		if i == m.modalSel {
			out[i] = focus.Style().Render("  ▸ " + s)
		} else {
			out[i] = val.Style().Render("    " + s)
		}
	}
	return out
}

func (m Model) logViewBody(h domain.HostView, w int) []string {
	muted, _ := m.mode.Role("text_muted")
	val, _ := m.mode.Role("value")
	al, _ := m.mode.State("error")
	var out []string
	for _, l := range h.Logs {
		if l.Source != m.logSource || !m.matchesLogQuery(l) {
			continue
		}
		ts := muted.Style().Render(l.At.Format("15:04:05"))
		line := l.Line
		if l.RateLimited {
			line = al.Style().Render("[rate-limited] ") + line
		}
		out = append(out, "  "+ts+"  "+val.Style().Render(clip(line, w-13)))
	}
	if len(out) == 0 {
		msg := "  waiting for lines…"
		if strings.TrimSpace(m.logQuery) != "" {
			msg = "  no lines match the search"
		}
		out = append(out, muted.Style().Render(msg))
	}
	return out
}

func (m Model) topBody(h domain.HostView, w int) []string {
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")
	muted, _ := m.mode.Role("text_muted")
	out := []string{label.Style().Render(fmt.Sprintf("  %7s %7s %6s %6s  %s", "PID", "PPID", "CPU%", "MEM%", "COMMAND"))}
	if len(h.Processes) == 0 {
		return append(out, muted.Style().Render("  waiting for a process table…"))
	}
	for _, p := range m.sortedProcesses(h) {
		out = append(out, val.Style().Render(fmt.Sprintf("  %7d %7d %5.1f%% %5.1f%%  %s",
			p.PID, p.PPID, p.CPUPct, p.MemPct, clip(p.Command, w-38))))
	}
	return out
}

func clampScroll(v, max int) int {
	if v < 0 {
		return 0
	}
	if v > max {
		return max
	}
	return v
}

// scroll handles a mouse-wheel tick (dir -1 up, +1 down) over whatever is
// scrollable: the grid cursor, the detail body, or a scrollable modal.
func (m Model) scroll(dir int) Model {
	if !m.detail {
		n := len(m.orderedList())
		m.cursor += dir
		switch {
		case m.cursor < 0 || n == 0:
			m.cursor = 0
		case m.cursor > n-1:
			m.cursor = n - 1
		}
		return m
	}
	step := dir * 3
	switch m.modal {
	case modalNone:
		m.detailScroll = clampScroll(m.detailScroll+step, m.detailMaxScroll())
	case modalLogView, modalTop, modalCmdResult:
		m.modalScroll = clampScroll(m.modalScroll+step, m.modalMaxScroll())
	}
	return m
}

// modalMaxScroll is the largest valid scroll offset for the active modal body,
// given the terminal height. Chrome around the body is header(3) + 5 lines.
func (m Model) modalMaxScroll() int {
	h, ok := m.selectedHost()
	if !ok {
		return 0
	}
	var bodyLen int
	switch m.modal {
	case modalLogView:
		for _, l := range h.Logs {
			if l.Source == m.logSource && m.matchesLogQuery(l) {
				bodyLen++
			}
		}
	case modalTop:
		bodyLen = len(h.Processes) + 1 // header row
	case modalCmdResult:
		bodyLen = len(m.cmdResultBody(h, m.width))
	default:
		return 0
	}
	maxBody := m.height - 8
	if maxBody < 1 {
		maxBody = 1
	}
	if bodyLen <= maxBody {
		return 0
	}
	return bodyLen - maxBody
}

// scrollWindow clamps lines to max around a scroll offset, replacing edge lines
// with "↑/↓ N more" indicators so the frame height stays bounded. Returns the
// clamped offset so the caller can pin to the tail.
func scrollWindow(lines []string, offset, max int) ([]string, int) {
	if max < 1 {
		max = 1
	}
	if len(lines) <= max {
		return lines, 0
	}
	if offset > len(lines)-max {
		offset = len(lines) - max
	}
	if offset < 0 {
		offset = 0
	}
	out := append([]string(nil), lines[offset:offset+max]...)
	if offset > 0 {
		out[0] = moreIndicator(offset, true)
	}
	if offset+max < len(lines) {
		out[len(out)-1] = moreIndicator(len(lines)-offset-max, false)
	}
	return out, offset
}

func clip(s string, n int) string {
	if n < 1 {
		n = 1
	}
	if r := []rune(s); len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
}
