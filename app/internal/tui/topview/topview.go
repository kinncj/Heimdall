// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package topview is the full-screen single-host "top" view (codename
// Hliðskjálf). It is a pure, render-only Bubble Tea sub-model: it takes a host
// snapshot and a per-metric history slice in and never collects anything. The
// dashboard switches into it on `t` and leaves on `esc`/`q`.
//
// Responsive behaviour mirrors the established dashboard pattern — layout(width)
// picks the densest plan that fits. The panels are a pane.Group (Himinbjörg): Tab
// moves the focus ring between panels, each focused panel scrolls its own overflow,
// the whole screen page-scrolls when it is taller than the terminal, and the mouse
// wheel/click targets the panel under the pointer.
package topview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/pane"
	"heimdall/app/internal/tui/theme"
)

// tier is the chosen responsive plan for a given terminal width.
type tier int

const (
	tierTiny   tier = iota // <40: key numbers only, one per line
	tierNarrow             // 40–59: single column, aggregate per-core
	tierMedium             // 60–99: single column, sparklines kept
	tierWide               // >=100: two-column panel grid
)

// layout returns the densest plan that fits the given width.
func layout(width int) tier {
	switch {
	case width >= 100:
		return tierWide
	case width >= 60:
		return tierMedium
	case width >= 40:
		return tierNarrow
	default:
		return tierTiny
	}
}

// Model is the top view's render state. The panels live in a pane.Group, which
// holds the mutable focus/scroll state across ticks.
type Model struct {
	host    domain.HostView
	history map[string][]float64
	mode    theme.Mode
	width   int
	height  int

	byName map[string]domain.Metric // host.LastSnapshot indexed by name
	group  *pane.Group
	srcs   []*linesSource // panel content, in reading order, updated on refresh
}

// linesSource is a static pane source: the panel's pre-rendered content lines.
type linesSource struct{ lines []string }

func (s *linesSource) Rows() []string { return s.lines }

// New builds a top view for one host. history is the per-metric recent-value
// buffer (each value 0–100 for percentages); the view reads it for sparklines
// and never appends to it.
func New(host domain.HostView, history map[string][]float64, mode theme.Mode, width, height int) Model {
	bn := make(map[string]domain.Metric, len(host.LastSnapshot))
	for _, mm := range host.LastSnapshot {
		bn[mm.Name] = mm
	}
	m := Model{host: host, history: history, mode: mode, width: width, height: height, byName: bn}
	m.build()
	return m
}

// build assembles the panel panes into a Group for the current width tier.
func (m *Model) build() {
	rows := m.specRows(layout(m.width))
	grid := make([][]*pane.Pane, 0, len(rows))
	var srcs []*linesSource
	for _, row := range rows {
		panes := make([]*pane.Pane, 0, len(row))
		for _, ps := range row {
			s := &linesSource{lines: ps.lines}
			srcs = append(srcs, s)
			panes = append(panes, pane.New(ps.title, s))
		}
		grid = append(grid, panes)
	}
	m.srcs = srcs
	// The whole top view is the first focus: ↑/↓ and the wheel scroll the entire
	// screen, so it is always scrollable at any resolution. Tab then steps into
	// each panel and wraps back around to the whole-view stop (Himinbjörg).
	m.group = pane.NewGrid(grid).EnablePageFocus()
}

// Refresh returns a copy bound to a newer host snapshot and history while keeping
// focus and scroll where the user left them, so a live tick updates the numbers
// without jumping the view.
func (m Model) Refresh(host domain.HostView, history map[string][]float64) Model {
	n := New(host, history, m.mode, m.width, m.height)
	n.group.TransferStateFrom(m.group)
	return n
}

// Resize returns a copy at a new terminal size, preserving focus and scroll when
// the layout shape is unchanged.
func (m Model) Resize(width, height int) Model {
	n := New(m.host, m.history, m.mode, width, height)
	n.group.TransferStateFrom(m.group)
	return n
}

// Action is what the dashboard should do after a key press in the top view.
type Action int

const (
	ActNone Action = iota // stayed in the view (e.g. scrolled)
	ActBack               // esc: return to the dashboard
	ActQuit               // q / ctrl+c: quit the whole app
)

// Update handles a key press. esc returns ActBack (leave the view); q and ctrl+c
// return ActQuit (quit the app), matching the rest of the TUI. Everything else —
// Tab focus, arrow/page scroll — is routed to the focused panel and returns ActNone.
func (m Model) Update(msg tea.KeyMsg) (Model, Action) {
	switch msg.String() {
	case "esc":
		return m, ActBack
	case "q", "ctrl+c":
		return m, ActQuit
	}
	if m.group != nil {
		m.group.Update(msg, m.bodyHeight())
	}
	return m, ActNone
}

// Mouse routes a wheel/click to the panel under the pointer (Himinbjörg). The
// dashboard forwards mouse events here while the top view is active.
func (m Model) Mouse(msg tea.MouseMsg) Model {
	if m.group != nil {
		m.group.Mouse(msg, m.bodyHeight())
	}
	return m
}

// View renders the fixed header, the pane.Group panel body, and the fixed footer,
// clamped to the terminal height. Every line is finally bounded to the terminal
// width so nothing clips past the frame.
func (m Model) View() string {
	t := layout(m.width)
	header := m.header(t)
	footer := m.footer(t)

	body := ""
	if m.group != nil {
		body = m.group.View(m.mode, pane.Rect{X: 0, Y: lineCount(header) + 1, W: m.width, H: m.bodyHeight()})
	}

	out := header + "\n\n" + body + "\n\n" + footer

	lines := strings.Split(out, "\n")
	for i, l := range lines {
		lines[i] = lipgloss.NewStyle().MaxWidth(m.width).Render(l)
	}
	return strings.Join(lines, "\n")
}

// bodyHeight is the number of body rows that fit between the fixed header and
// footer (header + blank + body + blank + footer).
func (m Model) bodyHeight() int {
	t := layout(m.width)
	chrome := lineCount(m.header(t)) + lineCount(m.footer(t)) + 2
	if h := m.height - chrome; h >= 1 {
		return h
	}
	return 1
}

// header renders the fixed brand/host/state line(s) for the tier.
func (m Model) header(t tier) string {
	title, _ := m.mode.Role("title")
	muted, _ := m.mode.Role("text_muted")
	accent, _ := m.mode.Role("accent")

	sigil := accent.Style().Render("⬢")
	badge := m.badge(false)

	switch t {
	case tierWide:
		left := sigil + " " + title.Style().Render("HEIMDALL") +
			muted.Style().Render(" · top · "+m.hostName()+" · "+m.osArch()+" · up "+m.uptime())
		return joinEnds(left, badge, m.width)
	case tierMedium:
		line1 := joinEnds(sigil+" "+title.Style().Render("HEIMDALL")+
			muted.Style().Render(" · top · "+m.hostName()), badge, m.width)
		line2 := muted.Style().Render(m.osArch() + " · up " + m.uptime())
		return line1 + "\n" + line2
	case tierNarrow:
		left := sigil + muted.Style().Render(" top · "+m.hostName())
		return joinEnds(left, badge, m.width)
	default: // tierTiny
		left := sigil + muted.Style().Render(" top · "+m.shortName())
		return joinEnds(left, m.badge(true), m.width)
	}
}

// footer renders the fixed keybind legend; it shortens as the width drops.
func (m Model) footer(t tier) string {
	keys, _ := m.mode.Role("keybinding")
	muted, _ := m.mode.Role("text_muted")
	k := func(s string) string { return keys.Style().Render(s) }
	x := func(s string) string { return muted.Style().Render(s) }

	switch t {
	case tierWide:
		return k("↑/↓") + x(" scroll · ") + k("tab") + x(" panel · ") + k("pgup/pgdn") + x(" page · ") + k("esc") + x(" back · ") + k("q") + x(" quit")
	case tierMedium:
		return k("↑/↓") + x(" scroll · ") + k("tab") + x(" panel · ") + k("esc") + x(" back · ") + k("q") + x(" quit")
	case tierNarrow:
		return k("↑/↓") + x(" scroll · ") + k("tab") + x(" panel · ") + k("esc") + x(" back")
	default: // tierTiny
		return k("↑/↓") + x(" · ") + k("esc")
	}
}

// badge renders the host-state pill: glyph + word, never colour alone. The short
// form (TINY) abbreviates ONLINE to ON.
func (m Model) badge(short bool) string {
	st, ok := m.mode.State(stateName(m.host.State))
	if !ok {
		return ""
	}
	word := st.Label
	if short {
		word = shortState(m.host.State)
	}
	return st.Style().Render(st.Glyph + " " + word)
}

func (m Model) hostName() string {
	if m.host.Host.DisplayName != "" {
		return m.host.Host.DisplayName
	}
	return string(m.host.Host.ID)
}

func (m Model) shortName() string {
	n := m.hostName()
	if len(n) > 2 {
		return n[:2]
	}
	return n
}

func (m Model) osArch() string {
	os := m.detailOr("host.os", m.host.Host.Context.OS)
	arch := m.detailOr("host.arch", m.host.Host.Context.Arch)
	if s := strings.TrimSpace(os + " " + arch); s != "" {
		return s
	}
	return "—"
}

func (m Model) uptime() string {
	if mm, ok := m.ok("host.uptime"); ok {
		return uptimeStr(mm.Gauge)
	}
	return "—"
}

func uptimeStr(secs float64) string {
	mins := int(secs) / 60
	days := mins / (60 * 24)
	hrs := (mins / 60) % 24
	mn := mins % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hrs)
	case hrs > 0:
		return fmt.Sprintf("%dh %dm", hrs, mn)
	default:
		return fmt.Sprintf("%dm", mn)
	}
}

// stateName maps a host state to its theme state-role key.
func stateName(s domain.HostState) string {
	switch s {
	case domain.StateOnline:
		return "online"
	case domain.StateStale:
		return "stale"
	case domain.StateOffline:
		return "offline"
	default:
		return "enrolling"
	}
}

func shortState(s domain.HostState) string {
	switch s {
	case domain.StateOnline:
		return "ON"
	case domain.StateStale:
		return "STALE"
	case domain.StateOffline:
		return "OFF"
	default:
		return "…"
	}
}

// joinEnds places left and right on one line, filling the gap with spaces so
// right is right-aligned to width. left/right are already styled; widths are
// measured ANSI-aware.
func joinEnds(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func lineCount(s string) int { return strings.Count(s, "\n") + 1 }
