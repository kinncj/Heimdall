// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/brand"
	"heimdall/app/internal/tui/render"
)

func onlineCount(hosts []domain.HostView) int {
	n := 0
	for _, h := range hosts {
		if h.State == domain.StateOnline {
			n++
		}
	}
	return n
}

// recordHistory appends each host's OK metric values into a capped in-memory
// ring buffer for the detail-view sparklines (no TSDB; bounded RAM), and drops
// history for hosts the registry has purged so the trend store never outgrows
// the live fleet.
func (m Model) recordHistory() {
	const capN = 60
	live := make(map[domain.HostID]struct{})
	for _, h := range m.reg.Hosts() {
		live[h.Host.ID] = struct{}{}
		hm := m.history[h.Host.ID]
		if hm == nil {
			hm = make(map[string][]float64)
			m.history[h.Host.ID] = hm
		}
		for _, mm := range h.LastSnapshot {
			if mm.Status != domain.StatusOK {
				continue
			}
			s := append(hm[mm.Name], mm.Gauge)
			if len(s) > capN {
				s = s[len(s)-capN:]
			}
			hm[mm.Name] = s
		}
	}
	for id := range m.history {
		if _, ok := live[id]; !ok {
			delete(m.history, id)
		}
	}
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// pickMetric returns the first OK metric among keys; if none are OK it returns
// the first key's metric (possibly absent/non-OK) for the affordance. This lets
// a row fall back from a package-level signal to a device-level one — e.g. POWER
// from power.pkg to power.gpu, or TEMP from temp.pkg to gpu.temp — on a host
// that only exposes the GPU's figures.
func pickMetric(byName map[string]domain.Metric, keys ...string) domain.Metric {
	for _, k := range keys {
		if mm, ok := byName[k]; ok && mm.Status == domain.StatusOK {
			return mm
		}
	}
	return byName[keys[0]]
}

// DetailView renders the per-host drilldown: identity, context, and each metric
// as a wide gauge + value + trend sparkline.
func (m Model) DetailView() string {
	hosts := m.orderedList()
	if len(hosts) == 0 {
		return m.GridView()
	}
	i := m.cursor
	if i >= len(hosts) {
		i = len(hosts) - 1
	}
	h := hosts[i]

	w := m.width
	if w < 88 {
		w = 88
	}
	header := brand.SkinnyHeader(m.mode, w, onlineCount(hosts), len(hosts), m.now.Format("15:04:05"))

	// Fixed header + footer; the body sections scroll (detailScroll) so the view
	// fits short terminals (e.g. SSH from a phone). Chrome = header + 2 blanks +
	// footer.
	body, _ := scrollWindow(m.detailBody(h, w), m.detailScroll, m.height-(lineCount(header)+3))
	return strings.Join([]string{header, "", strings.Join(body, "\n"), "", m.detailFooter(h)}, "\n")
}

// detailMaxScroll is the largest valid detail-body scroll offset for the terminal
// height (0 when everything fits).
func (m Model) detailMaxScroll() int {
	h, ok := m.selectedHost()
	if !ok {
		return 0
	}
	w := m.width
	if w < 88 {
		w = 88
	}
	maxBody := m.height - 6 // header(3) + 2 blanks + footer(1)
	if maxBody < 1 {
		maxBody = 1
	}
	if body := m.detailBody(h, w); len(body) > maxBody {
		return len(body) - maxBody
	}
	return 0
}

// detailFooter is the fixed keybind line, with logs/top/cmd shown only when the
// host advertises them.
func (m Model) detailFooter(h domain.HostView) string {
	muted, _ := m.mode.Role("text_muted")
	keys, _ := m.mode.Role("keybinding")
	footer := "  " + keys.Style().Render("esc") + muted.Style().Render(" back   ") +
		keys.Style().Render("↑/↓") + muted.Style().Render(" host   ") +
		keys.Style().Render("⇧↑/↓") + muted.Style().Render(" scroll   ")
	if len(logSourcesOf(h)) > 0 {
		footer += keys.Style().Render("l") + muted.Style().Render(" logs   ")
	}
	if hasProc(h) {
		footer += keys.Style().Render("t") + muted.Style().Render(" top   ")
	}
	if hasCmd(h) && m.runCmd != nil {
		footer += keys.Style().Render("c") + muted.Style().Render(" cmd   ")
	}
	return footer + keys.Style().Render("q") + muted.Style().Render(" quit")
}

// detailBody renders the per-host detail sections (everything between the brand
// header and the footer) as lines, so the view can scroll them.
func (m Model) detailBody(h domain.HostView, w int) []string {
	heading, _ := m.mode.Role("heading")
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")
	muted, _ := m.mode.Role("text_muted")

	byName := make(map[string]domain.Metric, len(h.LastSnapshot))
	for _, mm := range h.LastSnapshot {
		byName[mm.Name] = mm
	}
	desc := func(key, fallback string) string {
		if d := byName[key].Detail; d != "" {
			return d
		}
		return fallback
	}

	dn := h.Host.DisplayName
	if dn == "" {
		dn = string(h.Host.ID)
	}
	st, _ := m.mode.State(stateName(h.State))
	ctx := h.Host.Context
	osArch := orDash(strings.TrimSpace(desc("host.os", ctx.OS) + " " + desc("host.arch", ctx.Arch)))

	var b strings.Builder
	b.WriteString("  " + heading.Style().Render("HOST DETAIL — "+dn) + "\n")
	b.WriteString("  " + st.Badge() + "   " +
		label.Style().Render("os ") + val.Style().Render(osArch) + "   " +
		label.Style().Render("kernel ") + val.Style().Render(orDash(desc("host.kernel", ""))) + "   " +
		label.Style().Render("seen ") + val.Style().Render(h.LastSeen.Format("15:04:05")) + "\n\n")

	hist := m.history[h.Host.ID]
	// Cap the trend to the space between the value column and the right edge so
	// it scrolls in place instead of growing past the frame as history fills.
	sparkW := w - 66
	if sparkW < 12 {
		sparkW = 12
	}

	for _, d := range []struct {
		keys      []string
		lab, unit string
	}{
		{[]string{"cpu.util"}, "CPU", "%"},
		{[]string{"mem.used"}, "MEM", "%"},
		{[]string{"disk.used"}, "DISK", "%"},
		{[]string{"gpu.util"}, "GPU", "%"},
		{[]string{"gpu.vram"}, "VRAM", "%"},
		{[]string{"power.pkg", "power.gpu"}, "POWER", "W"},
		{[]string{"temp.pkg", "gpu.temp"}, "TEMP", "°C"},
	} {
		lab := label.Style().Render(fmt.Sprintf("  %-6s", d.lab))
		mm := pickMetric(byName, d.keys...)
		if mm.Status != domain.StatusOK {
			b.WriteString(lab + "  " + m.nonOK(mm) + "\n")
			continue
		}
		gauge := render.Gauge(m.mode, mm.Gauge, 28)
		valAbs := val.Style().Render(fmt.Sprintf("%5.0f%s", mm.Gauge, d.unit))
		if mm.Detail != "" {
			valAbs += "  " + muted.Style().Render(mm.Detail)
		}
		spark := ""
		if hist != nil {
			spark = render.Sparkline(m.mode, hist[mm.Name], sparkW)
		}
		b.WriteString(lab + "  " + gauge + " " + cell(valAbs, 26) + " " + spark + "\n")
	}

	if cores := byName["cpu.cores"]; cores.Status == domain.StatusOK && len(cores.PerCore) > 0 {
		b.WriteString("\n  " + heading.Style().Render("CPU CORES") + "\n")
		b.WriteString("  " + render.Sparkline(m.mode, cores.PerCore, w-4) + "  " +
			muted.Style().Render(fmt.Sprintf("%d cores", len(cores.PerCore))) + "\n")
	}

	b.WriteString("\n  " + heading.Style().Render("HARDWARE") + "\n")
	b.WriteString("  " +
		label.Style().Render("cpu ") + val.Style().Render(orDash(desc("host.cpu", ""))) + "   " +
		label.Style().Render("gpu ") + val.Style().Render(orDash(desc("host.gpu", ""))) + "   " +
		label.Style().Render("heimdall ") + val.Style().Render(orDash(desc("host.version", ""))) + "\n")

	b.WriteString("\n  " + heading.Style().Render("NETWORK & SYSTEM") + "\n")
	b.WriteString("  " +
		label.Style().Render("net ↓ ") + m.throughputVal(byName, "net.rx") + "   " +
		label.Style().Render("net ↑ ") + m.throughputVal(byName, "net.tx") + "   " +
		label.Style().Render("ping ") + m.pingVal(byName) + "   " +
		label.Style().Render("gw ") + m.gatewayVal(byName) + "   " +
		label.Style().Render("uptime ") + m.uptimeVal(byName) + "\n")
	b.WriteString("  " +
		label.Style().Render("disk r ") + m.throughputVal(byName, "disk.read") + "   " +
		label.Style().Render("disk w ") + m.throughputVal(byName, "disk.write") + "\n")

	b.WriteString(m.nicsSection(byName))

	return strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
}

func (m Model) throughputVal(byName map[string]domain.Metric, key string) string {
	mm, ok := byName[key]
	if !ok || mm.Status != domain.StatusOK {
		return m.nonOK(mm)
	}
	val, _ := m.mode.Role("value")
	return val.Style().Render(fmt.Sprintf("%.2f MB/s", mm.Gauge))
}

// nicsSection lists each non-loopback NIC with its rx/tx throughput and, when
// known, its gateway latency and IP. NIC names come from the net.rx.<iface>
// metrics the daemon emits per interface.
func (m Model) nicsSection(byName map[string]domain.Metric) string {
	ifaces := make([]string, 0, 4)
	for name := range byName {
		if strings.HasPrefix(name, "net.rx.") {
			ifaces = append(ifaces, strings.TrimPrefix(name, "net.rx."))
		}
	}
	if len(ifaces) == 0 {
		return ""
	}
	sort.Strings(ifaces)

	heading, _ := m.mode.Role("heading")
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")

	var b strings.Builder
	rows := 0
	for _, iface := range ifaces {
		rx := byName["net.rx."+iface]
		tx := byName["net.tx."+iface]
		gw, hasGW := byName["net.gateway."+iface]
		// Skip the OS's many idle virtual interfaces: show a NIC only if it has a
		// default gateway or is currently moving traffic.
		if !hasGW && rx.Gauge == 0 && tx.Gauge == 0 {
			continue
		}
		if rows == 0 {
			b.WriteString("\n  " + heading.Style().Render("NICS") + "\n")
		}
		rows++
		line := "  " + val.Style().Render(fmt.Sprintf("%-12s", clipName(iface, 12))) +
			label.Style().Render("↓ ") + val.Style().Render(fmt.Sprintf("%6.2f MB/s", rx.Gauge)) +
			label.Style().Render("  ↑ ") + val.Style().Render(fmt.Sprintf("%6.2f MB/s", tx.Gauge))
		if hasGW && gw.Status == domain.StatusOK {
			line += label.Style().Render("   gw ") + val.Style().Render(fmt.Sprintf("%.0f ms", gw.Gauge))
			if gw.Detail != "" {
				line += " " + label.Style().Render(gw.Detail)
			}
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) pingVal(byName map[string]domain.Metric) string {
	mm, ok := byName["net.latency"]
	if !ok || mm.Status != domain.StatusOK {
		return m.nonOK(mm)
	}
	val, _ := m.mode.Role("value")
	return val.Style().Render(fmt.Sprintf("%.0f ms", mm.Gauge))
}

func (m Model) gatewayVal(byName map[string]domain.Metric) string {
	mm, ok := byName["net.gateway"]
	if !ok || mm.Status != domain.StatusOK {
		return m.nonOK(mm)
	}
	val, _ := m.mode.Role("value")
	return val.Style().Render(fmt.Sprintf("%.0f ms", mm.Gauge))
}

func (m Model) uptimeVal(byName map[string]domain.Metric) string {
	mm, ok := byName["host.uptime"]
	if !ok || mm.Status != domain.StatusOK {
		return m.nonOK(mm)
	}
	val, _ := m.mode.Role("value")
	return val.Style().Render(uptimeStr(mm.Gauge))
}

func uptimeStr(secs float64) string {
	d := time.Duration(secs) * time.Second
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	mn := int(d.Minutes()) % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, h)
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, mn)
	default:
		return fmt.Sprintf("%dm", mn)
	}
}
