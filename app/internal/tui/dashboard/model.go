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
			return m, tea.Quit
		case "r":
			m.refresh()
			return m, nil
		case "enter":
			if !m.detail && m.reg.Count() > 0 {
				m.detail = true
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.reg.Count()-1 {
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
	hosts := m.reg.Hosts()
	online := 0
	for _, h := range hosts {
		if h.State == domain.StateOnline {
			online++
		}
	}
	w := m.width
	if w < 88 {
		w = 88
	}
	clock := m.now.Format("15:04:05")

	header := brand.SkinnyHeader(m.mode, w, online, len(hosts), clock)
	legend, _ := m.mode.Role("label")
	cols := legend.Style().Render(fmt.Sprintf("  %-16s %-13s %-14s %-14s %-14s %-7s %-6s %-6s",
		"HOST", "STATE", "CPU", "MEM", "DISK", "TEMP", "GPU", "PWR"))

	rows := make([]string, 0, len(hosts))
	for i, h := range hosts {
		rows = append(rows, m.row(h, i == m.cursor))
	}

	status := brand.StatusBar(m.mode, w, true, "2s", "low-bw gRPC", "edge relay", true, clock)
	foot, _ := m.mode.Role("text_muted")
	keys, _ := m.mode.Role("keybinding")
	footer := foot.Style().Render("  ") + keys.Style().Render("↑/↓") + foot.Style().Render(" nav  ") +
		keys.Style().Render("⏎") + foot.Style().Render(" detail  ") +
		keys.Style().Render("r") + foot.Style().Render(" refresh  ") +
		keys.Style().Render("q") + foot.Style().Render(" quit  ") +
		keys.Style().Render("?") + foot.Style().Render(" help")

	return strings.Join([]string{header, "", cols, strings.Join(rows, "\n"), "", status, footer}, "\n")
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

	marker := "  "
	if selected {
		f, _ := m.mode.Role("focus")
		marker = f.Style().Render("▸ ")
	}

	st, _ := m.mode.State(stateName(h.State))
	badge := lipgloss.NewStyle().Width(13).Render(st.Badge())

	name, _ := m.mode.Role("value")
	dn := h.Host.DisplayName
	if dn == "" {
		dn = string(h.Host.ID)
	}
	nameCell := name.Style().Render(fmt.Sprintf("%-16s", clipName(dn, 16)))

	line := marker + nameCell + " " + badge +
		" " + m.pct(byName["cpu.util"]) +
		" " + m.pct(byName["mem.used"]) +
		" " + m.pct(byName["disk.used"]) +
		" " + m.temp(byName["temp.pkg"]) +
		" " + m.plain(byName["gpu.util"], "%") +
		" " + m.plain(byName["power.pkg"], "W")

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
