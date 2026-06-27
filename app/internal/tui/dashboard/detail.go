// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
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
// ring buffer for the detail-view sparklines (no TSDB; bounded RAM).
func (m Model) recordHistory() {
	const capN = 60
	for _, h := range m.reg.Hosts() {
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
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// DetailView renders the per-host drilldown: identity, context, and each metric
// as a wide gauge + value + trend sparkline.
func (m Model) DetailView() string {
	hosts := m.reg.Hosts()
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

	heading, _ := m.mode.Role("heading")
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")
	muted, _ := m.mode.Role("text_muted")
	keys, _ := m.mode.Role("keybinding")

	dn := h.Host.DisplayName
	if dn == "" {
		dn = string(h.Host.ID)
	}
	st, _ := m.mode.State(stateName(h.State))
	ctx := h.Host.Context
	osArch := orDash(strings.Trim(ctx.OS+"/"+ctx.Arch, "/"))

	var b strings.Builder
	b.WriteString(header + "\n\n")
	b.WriteString("  " + heading.Style().Render("HOST DETAIL — "+dn) + "\n")
	b.WriteString("  " + st.Badge() + "   " +
		label.Style().Render("os ") + val.Style().Render(osArch) + "   " +
		label.Style().Render("class ") + val.Style().Render(orDash(ctx.Labels["class"])) + "   " +
		label.Style().Render("seen ") + val.Style().Render(h.LastSeen.Format("15:04:05")) + "\n\n")

	byName := make(map[string]domain.Metric, len(h.LastSnapshot))
	for _, mm := range h.LastSnapshot {
		byName[mm.Name] = mm
	}
	hist := m.history[h.Host.ID]
	// Cap the trend to the space between the value column and the right edge so
	// it scrolls in place instead of growing past the frame as history fills.
	sparkW := w - 50
	if sparkW < 12 {
		sparkW = 12
	}

	for _, d := range []struct{ key, lab, unit string }{
		{"cpu.util", "CPU", "%"}, {"mem.used", "MEM", "%"}, {"disk.used", "DISK", "%"},
		{"gpu.util", "GPU", "%"}, {"power.pkg", "POWER", "W"}, {"temp.pkg", "TEMP", "°C"},
	} {
		lab := label.Style().Render(fmt.Sprintf("  %-6s", d.lab))
		mm, ok := byName[d.key]
		if !ok || mm.Status != domain.StatusOK {
			b.WriteString(lab + "  " + m.nonOK(mm) + "\n")
			continue
		}
		gauge := render.Gauge(m.mode, mm.Gauge, 28)
		value := val.Style().Render(fmt.Sprintf("%5.0f%s", mm.Gauge, d.unit))
		spark := ""
		if hist != nil {
			spark = render.Sparkline(m.mode, hist[d.key], sparkW)
		}
		b.WriteString(lab + "  " + gauge + " " + value + "   " + spark + "\n")
	}

	if cores := byName["cpu.cores"]; cores.Status == domain.StatusOK && len(cores.PerCore) > 0 {
		b.WriteString("\n  " + heading.Style().Render("CPU CORES") + "\n")
		b.WriteString("  " + render.Sparkline(m.mode, cores.PerCore, w-4) + "  " +
			muted.Style().Render(fmt.Sprintf("%d cores", len(cores.PerCore))) + "\n")
	}

	b.WriteString("\n  " + heading.Style().Render("NETWORK & SYSTEM") + "\n")
	b.WriteString("  " +
		label.Style().Render("net ↓ ") + m.netVal(byName, "net.rx") + "   " +
		label.Style().Render("net ↑ ") + m.netVal(byName, "net.tx") + "   " +
		label.Style().Render("ping ") + m.pingVal(byName) + "   " +
		label.Style().Render("gw ") + m.gatewayVal(byName) + "   " +
		label.Style().Render("uptime ") + m.uptimeVal(byName) + "\n")

	b.WriteString("\n  " + keys.Style().Render("esc") + muted.Style().Render(" back   ") +
		keys.Style().Render("↑/↓") + muted.Style().Render(" host   ") +
		keys.Style().Render("q") + muted.Style().Render(" quit"))
	return b.String()
}

func (m Model) netVal(byName map[string]domain.Metric, key string) string {
	mm, ok := byName[key]
	if !ok || mm.Status != domain.StatusOK {
		return m.nonOK(mm)
	}
	val, _ := m.mode.Role("value")
	return val.Style().Render(fmt.Sprintf("%.2f MB/s", mm.Gauge))
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
