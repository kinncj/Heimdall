// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package topview is the full-screen single-host "top" view (codename
// Hliðskjálf). It is a pure, render-only Bubble Tea sub-model: it takes a host
// snapshot and a per-metric history slice in and never collects anything. The
// dashboard switches into it on `t` and leaves on `esc`/`q`.
//
// Responsive behaviour mirrors the established dashboard pattern — layout(width)
// picks the densest plan that fits, and a self-contained scrollWindow clamps the
// body between a fixed header and footer so the frame never exceeds the terminal.
package topview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/domain"
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

// Model is the top view's immutable render state plus the local scroll offset.
type Model struct {
	host    domain.HostView
	history map[string][]float64
	mode    theme.Mode
	width   int
	height  int
	scroll  int

	byName map[string]domain.Metric // host.LastSnapshot indexed by name
}

// New builds a top view for one host. history is the per-metric recent-value
// buffer (each value 0–100 for percentages); the view reads it for sparklines
// and never appends to it.
func New(host domain.HostView, history map[string][]float64, mode theme.Mode, width, height int) Model {
	bn := make(map[string]domain.Metric, len(host.LastSnapshot))
	for _, mm := range host.LastSnapshot {
		bn[mm.Name] = mm
	}
	return Model{host: host, history: history, mode: mode, width: width, height: height, byName: bn}
}

// Refresh returns a copy bound to a newer host snapshot and history while keeping
// the current scroll offset, so a live tick updates the numbers without jumping
// the view. The next key press re-clamps scroll against the new content height.
func (m Model) Refresh(host domain.HostView, history map[string][]float64) Model {
	n := New(host, history, m.mode, m.width, m.height)
	n.scroll = m.scroll
	return n
}

// Resize returns a copy at a new terminal size, preserving scroll.
func (m Model) Resize(width, height int) Model {
	m.width, m.height = width, height
	return m
}

// Action is what the dashboard should do after a key press in the top view.
type Action int

const (
	ActNone Action = iota // stayed in the view (e.g. scrolled)
	ActBack               // esc: return to the dashboard
	ActQuit               // q / ctrl+c: quit the whole app
)

// Update handles a key press. esc returns ActBack (leave the view); q and ctrl+c
// return ActQuit (quit the app), matching the rest of the TUI. All other keys
// scroll the body in place and return ActNone.
func (m Model) Update(msg tea.KeyMsg) (Model, Action) {
	switch msg.String() {
	case "esc":
		return m, ActBack
	case "q", "ctrl+c":
		return m, ActQuit
	case "up", "k":
		m.scroll = clampScroll(m.scroll-1, m.maxScroll())
	case "down", "j":
		m.scroll = clampScroll(m.scroll+1, m.maxScroll())
	case "pgup":
		m.scroll = clampScroll(m.scroll-m.pageStep(), m.maxScroll())
	case "pgdown", "pgdn":
		m.scroll = clampScroll(m.scroll+m.pageStep(), m.maxScroll())
	case "home":
		m.scroll = 0
	case "end":
		m.scroll = m.maxScroll()
	}
	return m, ActNone
}

// View renders the fixed header, the scrollable panel body, and the fixed footer,
// clamped to the terminal height. Every line is finally bounded to the terminal
// width so nothing clips past the frame.
func (m Model) View() string {
	t := layout(m.width)
	header := m.header(t)
	footer := m.footer(t)
	body := m.body(t)

	windowed, off := scrollWindow(m, body, m.scroll, m.bodyHeight())
	m.scroll = off

	out := header + "\n\n" + strings.Join(windowed, "\n") + "\n\n" + footer

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

// maxScroll is the largest valid body scroll offset (0 when everything fits).
func (m Model) maxScroll() int {
	body := m.body(layout(m.width))
	if vis := m.bodyHeight(); len(body) > vis {
		return len(body) - vis
	}
	return 0
}

// pageStep is one page of body scroll (a near-full screen, minus one row of
// overlap so context is kept).
func (m Model) pageStep() int {
	if s := m.bodyHeight() - 1; s > 1 {
		return s
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
		return k("↑/↓") + x(" scroll · ") + k("pgup/pgdn") + x(" page · ") + k("esc") + x(" back · ") + k("q") + x(" quit")
	case tierMedium:
		return k("↑/↓") + x(" scroll · ") + k("esc") + x(" back · ") + k("q") + x(" quit")
	case tierNarrow:
		return k("↑/↓") + x(" scroll · ") + k("esc") + x(" back")
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

func clampScroll(v, max int) int {
	if v < 0 {
		return 0
	}
	if v > max {
		return max
	}
	return v
}

// scrollWindow clamps lines to a vis-row window around offset, replacing the edge
// rows with text scroll affordances ("▲ more above" / "▼ more below" + position)
// so the body stays height-bounded. It returns the clamped offset. Kept local to
// this package so the view is self-contained.
func scrollWindow(m Model, lines []string, offset, vis int) ([]string, int) {
	if vis < 1 {
		vis = 1
	}
	if len(lines) <= vis {
		return lines, 0
	}
	if offset > len(lines)-vis {
		offset = len(lines) - vis
	}
	if offset < 0 {
		offset = 0
	}
	out := append([]string(nil), lines[offset:offset+vis]...)
	caption, _ := m.mode.Role("caption")
	pos := func(above bool) string {
		glyph := "▼ more below"
		if above {
			glyph = "▲ more above"
		}
		return caption.Style().Render(fmt.Sprintf("%s   scroll %d/%d", glyph, offset+1, len(lines)))
	}
	if offset > 0 {
		out[0] = pos(true)
	}
	if offset+vis < len(lines) {
		out[len(out)-1] = pos(false)
	}
	return out, offset
}
