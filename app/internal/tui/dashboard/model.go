// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package dashboard is the Bubble Tea model for the Heimdall TUI: brand header,
// live host grid, and status bar, themed from terminal-theme.json.
package dashboard

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/brand"
	"heimdall/app/internal/tui/render"
	"heimdall/app/internal/tui/theme"
)

// Model is the dashboard state.
type Model struct {
	mode          theme.Mode
	reg           *domain.HostRegistry
	width, height int
	now           time.Time
	cursor        int
	onTick        func(time.Time)
	detail        bool
	help          bool
	history       map[domain.HostID]map[string][]float64
	source        string
	live          func() bool
	groupBy       int    // 0 = no grouping; 1..N index dimensions()+1 (Yggdrasil)
	filter        string // active filter query
	filtering     bool   // true while the filter input is open
}

type tickMsg time.Time

// New builds a dashboard model over a host registry.
func New(mode theme.Mode, reg *domain.HostRegistry, now time.Time) Model {
	return Model{mode: mode, reg: reg, width: 104, height: 30, now: now,
		history: make(map[domain.HostID]map[string][]float64)}
}

// WithTick sets a per-tick callback (e.g. a live metric collector).
func (m Model) WithTick(fn func(time.Time)) Model {
	m.onTick = fn
	return m
}

// WithStatus sets what the footer reports: the data source (hub address or
// "demo") and a predicate telling whether the dashboard is currently receiving.
func (m Model) WithStatus(source string, live func() bool) Model {
	m.source = source
	m.live = live
	return m
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Init starts the clock/liveness tick.
func (m Model) Init() tea.Cmd { return tick() }

// Update handles input, resize, and the periodic tick.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.filtering {
			return m.updateFilter(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.help = !m.help
			return m, nil
		case "esc":
			if m.help {
				m.help = false
				return m, nil
			}
			if m.detail {
				m.detail = false
				return m, nil
			}
			if m.filter != "" {
				m.filter = ""
				m.cursor = 0
				return m, nil
			}
			return m, tea.Quit
		case "r":
			m.refresh()
			return m, nil
		case "g":
			// cycle the grouping dimension: none -> hub -> os -> tag keys -> none
			m.groupBy = (m.groupBy + 1) % (len(m.dimensions(m.reg.Hosts())) + 1)
			m.cursor = 0
			return m, nil
		case "/":
			m.filtering = true
			return m, nil
		case "enter":
			if !m.detail && len(m.orderedList()) > 0 {
				m.detail = true
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.orderedList())-1 {
				m.cursor++
			}
		}
	case tickMsg:
		m.now = time.Time(msg)
		if m.onTick != nil {
			m.onTick(m.now)
		}
		m.reg.Evaluate(m.now)
		m.recordHistory()
		return m, tick()
	}
	return m, nil
}

// updateFilter handles keystrokes while the filter input is open: type to
// narrow, backspace to edit, enter to keep, esc to clear.
func (m Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filtering = false
	case "esc":
		m.filtering = false
		m.filter = ""
	case "backspace":
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
		}
	default:
		if len(msg.Runes) > 0 {
			m.filter += string(msg.Runes)
		}
	}
	m.cursor = 0
	return m, nil
}

// View renders one frame.
func (m Model) View() string {
	if m.help {
		return m.HelpView()
	}
	if m.detail {
		return m.DetailView()
	}
	return m.GridView()
}

// GridView renders the dashboard grid frame.
func (m Model) GridView() string {
	all := m.reg.Hosts()
	online := 0
	for _, h := range all {
		if h.State == domain.StateOnline {
			online++
		}
	}
	w := m.width
	if w < 88 {
		w = 88
	}
	clock := m.now.Format("15:04:05")

	header := brand.SkinnyHeader(m.mode, w, online, len(all), clock)
	legend, _ := m.mode.Role("label")
	cols := legend.Style().Render(fmt.Sprintf("  %-16s %-13s %-14s %-14s %-14s %-7s %-6s %-6s",
		"HOST", "STATE", "CPU", "MEM", "DISK", "TEMP", "GPU", "PWR"))

	hosts, groups := m.orderedHosts()
	rows := make([]string, 0, len(hosts)+4)
	lastGroup := "\x00"
	for i, h := range hosts {
		if groups != nil && groups[i] != lastGroup {
			lastGroup = groups[i]
			rows = append(rows, m.sectionHeader(groups[i], countGroup(groups, groups[i])))
		}
		rows = append(rows, m.row(h, i == m.cursor))
	}
	if len(hosts) == 0 {
		muted, _ := m.mode.Role("text_muted")
		msg := "no hosts yet"
		if m.filter != "" {
			msg = fmt.Sprintf("no hosts match %q", m.filter)
		}
		rows = append(rows, muted.Style().Render("  "+msg))
	}

	status := brand.StatusBar(m.mode, w, m.live != nil && m.live(), m.source, clock)
	foot, _ := m.mode.Role("text_muted")
	keys, _ := m.mode.Role("keybinding")
	footer := foot.Style().Render("  ") + keys.Style().Render("↑/↓") + foot.Style().Render(" nav  ") +
		keys.Style().Render("⏎") + foot.Style().Render(" detail  ") +
		keys.Style().Render("g") + foot.Style().Render(" group  ") +
		keys.Style().Render("/") + foot.Style().Render(" filter  ") +
		keys.Style().Render("r") + foot.Style().Render(" refresh  ") +
		keys.Style().Render("q") + foot.Style().Render(" quit  ") +
		keys.Style().Render("?") + foot.Style().Render(" help")

	return strings.Join([]string{header, m.metaLine(all), cols, strings.Join(rows, "\n"), "", status, footer}, "\n")
}

// metaLine shows the active grouping dimension, the filter query (with a cursor
// while editing), and a fleet alert count when any host is firing.
func (m Model) metaLine(all []domain.HostView) string {
	muted, _ := m.mode.Role("text_muted")
	val, _ := m.mode.Role("value")
	groupName := "off"
	if dim, ok := m.activeDim(all); ok {
		groupName = dim.name
	}
	parts := []string{muted.Style().Render("  group: ") + val.Style().Render(groupName)}
	if m.filtering || m.filter != "" {
		q := m.filter
		if m.filtering {
			q += "▏"
		}
		parts = append(parts, muted.Style().Render("filter: ")+val.Style().Render(q))
	}
	if n := alertCount(all); n > 0 {
		al, _ := m.mode.State("error")
		suffix := "s"
		if n == 1 {
			suffix = ""
		}
		parts = append(parts, al.Style().Render(fmt.Sprintf("⚠ %d alert%s", n, suffix)))
	}
	return strings.Join(parts, muted.Style().Render("    "))
}

// sectionHeader renders a group divider like "── home (3) ──".
func (m Model) sectionHeader(name string, n int) string {
	lbl, _ := m.mode.Role("label")
	return lbl.Style().Render(fmt.Sprintf("  ── %s (%d) ──", name, n))
}

func countGroup(groups []string, name string) int {
	n := 0
	for _, g := range groups {
		if g == name {
			n++
		}
	}
	return n
}

func clipName(s string, n int) string {
	s = strings.TrimSuffix(s, ".local")
	if r := []rune(s); len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
}

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

func (m Model) row(h domain.HostView, selected bool) string {
	byName := make(map[string]domain.Metric, len(h.LastSnapshot))
	for _, mm := range h.LastSnapshot {
		byName[mm.Name] = mm
	}

	alerting := len(h.Alerts) > 0
	marker := "  "
	if selected {
		f, _ := m.mode.Role("focus")
		marker = f.Style().Render("▸ ")
	} else if alerting {
		al, _ := m.mode.State("error")
		marker = al.Style().Render("⚠ ")
	}

	st, _ := m.mode.State(stateName(h.State))
	badge := lipgloss.NewStyle().Width(13).Render(st.Badge())

	name, _ := m.mode.Role("value")
	nameStyle := name.Style()
	if alerting {
		al, _ := m.mode.State("error")
		nameStyle = al.Style()
	}
	dn := h.Host.DisplayName
	if dn == "" {
		dn = string(h.Host.ID)
	}
	nameCell := nameStyle.Render(fmt.Sprintf("%-16s", clipName(dn, 16)))

	line := marker + nameCell + " " + badge +
		" " + m.pct(byName["cpu.util"]) +
		" " + m.pct(byName["mem.used"]) +
		" " + m.pct(byName["disk.used"]) +
		" " + m.temp(pickMetric(byName, "temp.pkg", "gpu.temp")) +
		" " + m.plain(byName["gpu.util"], "%") +
		" " + m.plain(pickMetric(byName, "power.pkg", "power.gpu"), "W")

	if selected {
		sel, _ := m.mode.Role("selection")
		if sel.BG != "" {
			return lipgloss.NewStyle().Background(lipgloss.Color(sel.BG)).Render(line)
		}
	}
	return line
}

// cell pads/truncates styled content to a fixed display width for column alignment.
func cell(s string, w int) string { return lipgloss.NewStyle().Width(w).Render(s) }

// pct renders a metric as gauge + value, or the non-OK affordance, in a fixed cell.
func (m Model) pct(metric domain.Metric) string {
	if metric.Status != domain.StatusOK {
		return cell(m.nonOK(metric), 14)
	}
	val, _ := m.mode.Role("value")
	return cell(render.Gauge(m.mode, metric.Gauge, 8)+" "+val.Style().Render(fmt.Sprintf("%3.0f%%", metric.Gauge)), 14)
}

func (m Model) temp(metric domain.Metric) string {
	if metric.Status != domain.StatusOK {
		return cell(m.nonOK(metric), 7)
	}
	val, _ := m.mode.Role("value")
	return cell(val.Style().Render(fmt.Sprintf("%3.0f°C", metric.Gauge)), 7)
}

func (m Model) plain(metric domain.Metric, unit string) string {
	if metric.Status != domain.StatusOK {
		return cell(m.nonOK(metric), 6)
	}
	val, _ := m.mode.Role("value")
	return cell(val.Style().Render(fmt.Sprintf("%3.0f%s", metric.Gauge, unit)), 6)
}

// nonOK renders the compact grid affordance (— unavailable, ⚿ needs-helper, ⚠ error);
// the full detail belongs in the per-host detail view.
func (m Model) nonOK(metric domain.Metric) string {
	var role, sym string
	switch metric.Status {
	case domain.StatusUnavailable, domain.StatusUnspecified:
		role, sym = "unavailable", "—"
	case domain.StatusInsufficientPermission:
		role, sym = "needs_helper", "⚿"
	default:
		role, sym = "error", "⚠"
	}
	st, _ := m.mode.State(role)
	return st.Style().Render(sym)
}
