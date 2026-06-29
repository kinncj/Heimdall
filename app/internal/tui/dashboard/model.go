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
	// Heimdallr's sight (ADR 0017): host-detail observability overlays.
	modal       modalKind // active overlay (none / log list / log view / top / sort)
	modalSel    int       // selection index in the log-source list
	modalScroll int       // scroll offset in the log/top view
	logSource   string    // chosen log source in the log view
	// v2 (ADR 0019): log search + top sorting.
	topSort      string       // active top sort key ("" = cpu default)
	topSortSel   int          // selection index in the sort picker
	logQuery     string       // log-view search query
	logSearching bool         // true while the log search input is open
	persistSort  func(string) // persist the chosen top sort to config (injected)
}

// WithTopSort sets the initial top-modal sort key (the persisted default).
func (m Model) WithTopSort(key string) Model {
	m.topSort = key
	return m
}

// WithPersistSort injects the callback that persists a chosen top sort to the
// dashboard config, so the choice becomes the default on next launch.
func (m Model) WithPersistSort(fn func(string)) Model {
	m.persistSort = fn
	return m
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
		if m.detail {
			return m.updateDetail(msg)
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
		if m.modal != modalNone {
			return m.ModalView()
		}
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
	if w < 24 {
		w = 24
	}
	clock := m.now.Format("15:04:05")

	header := brand.SkinnyHeader(m.mode, w, online, len(all), clock)
	lay := m.layout(w)
	cols := m.colsHeader(lay)

	status := brand.StatusBar(m.mode, w, m.live != nil && m.live(), m.source, clock)
	footer := m.footerBar(w)

	hosts, groups := m.orderedHosts()
	body := make([]string, 0, len(hosts)+4)
	cursorLine := 0
	lastGroup := "\x00"
	for i, h := range hosts {
		if groups != nil && groups[i] != lastGroup {
			lastGroup = groups[i]
			body = append(body, m.sectionHeader(groups[i], countGroup(groups, groups[i])))
		}
		if i == m.cursor {
			cursorLine = len(body)
		}
		body = append(body, m.row(h, i == m.cursor, lay))
	}
	if len(hosts) == 0 {
		muted, _ := m.mode.Role("text_muted")
		msg := "no hosts yet"
		if m.filter != "" {
			msg = fmt.Sprintf("no hosts match %q", m.filter)
		}
		body = append(body, muted.Style().Render("  "+msg))
	}

	// Clamp the host rows so the whole frame fits the terminal height. Without
	// this the frame overruns short screens (SSH from a tablet/phone), scrolling
	// the header off and making ungrouped filtering look inert. The fixed chrome
	// is header + meta + cols + blank + status + footer.
	chrome := lineCount(header) + 3 + lineCount(status) + 1
	body = m.windowBody(window(body, cursorLine, m.height-chrome))

	return strings.Join([]string{header, m.metaLine(all), cols, strings.Join(body, "\n"), "", status, footer}, "\n")
}

// lineCount returns how many terminal lines a rendered block occupies.
func lineCount(s string) int { return strings.Count(s, "\n") + 1 }

// Identity-block widths (marker + name + state), shared by header and rows.
const (
	markerW = 2  // selection/alert marker
	badgeW  = 13 // full state badge ("● ONLINE")
	glyphW  = 2  // compact state glyph
)

// gridColumn is one metric column of the fleet grid — a title, a fixed cell
// width, and a renderer over a host's metrics. Columns are an ordered registry
// (most → least essential); the grid drops them right-to-left as the terminal
// narrows. Adding a column means registering one here, mirroring groupDim and
// fieldMatcher — no width switch.
type gridColumn struct {
	title string
	width int
	of    func(m Model, byName map[string]domain.Metric) string
}

func gridColumns() []gridColumn {
	return []gridColumn{
		{"CPU", 14, func(m Model, b map[string]domain.Metric) string { return m.pct(b["cpu.util"]) }},
		{"MEM", 14, func(m Model, b map[string]domain.Metric) string { return m.pct(b["mem.used"]) }},
		{"DISK", 14, func(m Model, b map[string]domain.Metric) string { return m.pct(b["disk.used"]) }},
		{"TEMP", 7, func(m Model, b map[string]domain.Metric) string { return m.temp(pickMetric(b, "temp.pkg", "gpu.temp")) }},
		{"GPU", 6, func(m Model, b map[string]domain.Metric) string { return m.plain(b["gpu.util"], "%") }},
		{"PWR", 6, func(m Model, b map[string]domain.Metric) string {
			return m.plain(pickMetric(b, "power.pkg", "power.gpu"), "W")
		}},
	}
}

// gridLayout is the responsive column plan for a terminal width: the host-name
// cell width, whether state shows the full badge or a compact glyph, and which
// metric columns fit. Computed once per frame and shared by the column header and
// every row so they stay aligned.
type gridLayout struct {
	nameW   int
	badge   bool
	columns []gridColumn
}

// layout chooses the densest column plan that fits width. It keeps the full state
// badge while name + badge + CPU fit; below that it switches to a compact glyph
// and shrinks the name as needed so at least CPU stays visible.
func (m Model) layout(width int) gridLayout {
	first := gridColumns()[0].width
	full := markerW + 16 + 1 + badgeW
	if full+1+first <= width {
		return gridLayout{16, true, fitColumns(full, width)}
	}
	nameW := 16
	for nameW > 8 && markerW+nameW+1+glyphW+1+first > width {
		nameW -= 2
	}
	return gridLayout{nameW, false, fitColumns(markerW+nameW+1+glyphW, width)}
}

// fitColumns returns the leading run of columns that fit in width after the
// identity block (used cells). Stopping at the first column that overflows drops
// it and every lower-priority column after it.
func fitColumns(used, width int) []gridColumn {
	out := []gridColumn{}
	for _, c := range gridColumns() {
		if used+1+c.width > width {
			break
		}
		used += 1 + c.width
		out = append(out, c)
	}
	return out
}

// colsHeader renders the column-title row for the active layout.
func (m Model) colsHeader(lay gridLayout) string {
	legend, _ := m.mode.Role("label")
	stateW, stateTitle := badgeW, "STATE"
	if !lay.badge {
		stateW, stateTitle = glyphW, "ST"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "  %-*s %-*s", lay.nameW, "HOST", stateW, stateTitle)
	for _, c := range lay.columns {
		fmt.Fprintf(&b, " %-*s", c.width, c.title)
	}
	return legend.Style().Render(b.String())
}

// footerBar renders the keybinding footer, falling back to a glyph-only form when
// the labelled form would not fit the terminal width.
func (m Model) footerBar(width int) string {
	foot, _ := m.mode.Role("text_muted")
	keys, _ := m.mode.Role("keybinding")
	full := foot.Style().Render("  ") + keys.Style().Render("↑/↓") + foot.Style().Render(" nav  ") +
		keys.Style().Render("⏎") + foot.Style().Render(" detail  ") +
		keys.Style().Render("g") + foot.Style().Render(" group  ") +
		keys.Style().Render("/") + foot.Style().Render(" filter  ") +
		keys.Style().Render("r") + foot.Style().Render(" refresh  ") +
		keys.Style().Render("q") + foot.Style().Render(" quit  ") +
		keys.Style().Render("?") + foot.Style().Render(" help")
	if lipgloss.Width(full) <= width {
		return full
	}
	sep := foot.Style().Render(" ")
	parts := []string{"↑↓", "⏎", "g", "/", "r", "q", "?"}
	out := foot.Style().Render("  ")
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += keys.Style().Render(p)
	}
	return out
}

// window clamps body lines to at most max, keeping the cursor line in view. When
// the list overflows, the edge lines become "↑/↓ N more" indicators so the
// rendered height stays exactly max. max < 1 is treated as 1.
func window(lines []string, cursor, max int) []string {
	if max < 1 {
		max = 1
	}
	if len(lines) <= max {
		return lines
	}
	start := cursor - max/2
	if start < 0 {
		start = 0
	}
	if start > len(lines)-max {
		start = len(lines) - max
	}
	end := start + max
	out := append([]string(nil), lines[start:end]...)
	if start > 0 {
		out[0] = moreIndicator(start, true)
	}
	if end < len(lines) {
		out[len(out)-1] = moreIndicator(len(lines)-end, false)
	}
	return out
}

func moreIndicator(n int, up bool) string {
	arrow := "↓"
	if up {
		arrow = "↑"
	}
	return fmt.Sprintf("  %s %d more", arrow, n)
}

// windowBody themes the "N more" overflow indicators produced by window. Kept
// separate so window stays a pure, testable function with no theme dependency.
func (m Model) windowBody(lines []string) []string {
	muted, ok := m.mode.Role("text_muted")
	if !ok {
		return lines
	}
	for i, l := range lines {
		if strings.Contains(l, " more") && (strings.Contains(l, "↑") || strings.Contains(l, "↓")) {
			lines[i] = muted.Style().Render(l)
		}
	}
	return lines
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

func (m Model) row(h domain.HostView, selected bool, lay gridLayout) string {
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
	stateCell := lipgloss.NewStyle().Width(badgeW).Render(st.Badge())
	if !lay.badge {
		stateCell = cell(st.Style().Render(st.Glyph), glyphW)
	}

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
	nameCell := nameStyle.Render(fmt.Sprintf("%-*s", lay.nameW, clipName(dn, lay.nameW)))

	line := marker + nameCell + " " + stateCell
	for _, c := range lay.columns {
		line += " " + c.of(m, byName)
	}

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
