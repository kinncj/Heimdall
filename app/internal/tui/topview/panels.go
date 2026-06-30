// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package topview

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/render"
)

// panelSpec is one titled panel's content before it is boxed.
type panelSpec struct {
	title string
	lines []string
}

// body builds the scrollable panel region for the tier as a flat list of lines.
func (m Model) body(t tier) []string {
	if t == tierTiny {
		return m.tinyBody()
	}

	var out []string
	add := func(block string) {
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, strings.Split(block, "\n")...)
	}

	if t == tierWide {
		full := m.width - 4
		col := (m.width-2)/2 - 4
		if col < 10 {
			col = 10
		}
		add(m.row2(m.cpuPanel(t), m.memPanel(t), col))
		add(m.row2(m.powerPanel(t), m.gpuPanel(t), col))
		add(m.row2(m.netDiskPanel(t), m.loadUptimePanel(), col))
		add(m.renderPanel(m.processPanel(t), full))
		return out
	}

	inner := m.width - 4
	if inner < 10 {
		inner = 10
	}
	for _, p := range []panelSpec{
		m.cpuPanel(t), m.memPanel(t), m.powerPanel(t),
		m.gpuPanel(t), m.netDiskPanel(t), m.processPanel(t),
	} {
		add(m.renderPanel(p, inner))
	}
	return out
}

// row2 renders two panels side by side and returns the joined block.
func (m Model) row2(a, b panelSpec, col int) string {
	left := m.renderPanel(a, col)
	right := m.renderPanel(b, col)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

// renderPanel boxes a panel's title + content with a single border, fixed to the
// given inner content width. lipgloss wraps any overlong content and guarantees
// the box width, so no line escapes the frame.
func (m Model) renderPanel(p panelSpec, inner int) string {
	heading, _ := m.mode.Role("heading")
	border, _ := m.mode.Role("border")
	content := heading.Style().Render(p.title)
	if len(p.lines) > 0 {
		content += "\n" + strings.Join(p.lines, "\n")
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(border.FG)).
		Padding(0, 1).
		Width(inner).
		Render(content)
}

// --- panels -----------------------------------------------------------------

func (m Model) cpuPanel(t tier) panelSpec {
	label, _ := m.mode.Role("label")
	lab := func(s string) string { return label.Style().Render(s) }

	util := m.pctVal("cpu.util")
	freq := m.numVal("cpu.freq", " GHz", "%.2f")
	load := m.numVal("cpu.load", "", "%.2f")
	spark := m.spark("cpu.util", t)

	if t == tierNarrow {
		return panelSpec{title: "CPU", lines: []string{
			lab("util ") + util + "   " + lab("freq ") + freq,
			lab("load ") + load,
			lab("util ") + spark + " " + util,
			m.coresAggregate(),
		}}
	}

	lines := []string{
		lab("util ") + util + "   " + lab("freq ") + freq + "   " + lab("load ") + load,
		lab("util ") + spark + "  " + util,
	}
	if cores, ok := m.ok("cpu.cores"); ok && len(cores.PerCore) > 0 {
		lines = append(lines, lab(fmt.Sprintf("per-core (%d):", len(cores.PerCore))))
		lines = append(lines, m.coreMatrix(cores.PerCore, 3)...)
	}
	return panelSpec{title: "CPU", lines: lines}
}

func (m Model) memPanel(t tier) panelSpec {
	label, _ := m.mode.Role("label")
	muted, _ := m.mode.Role("text_muted")
	lab := func(s string) string { return label.Style().Render(s) }

	used := m.pctVal("mem.used")
	usedDet := muted.Style().Render(m.detailOr("mem.used", ""))
	swap := m.pctVal("mem.swap")
	swapDet := muted.Style().Render(m.detailOr("mem.swap", ""))
	bw := m.numVal("mem.bw", " GB/s", "%.0f")
	bwSpark := m.spark("mem.bw", t)

	if t == tierNarrow {
		return panelSpec{title: "MEMORY", lines: []string{
			lab("used ") + used + " " + usedDet + "   " + lab("swap ") + swap,
			lab("bw ") + bwSpark + " " + bw,
		}}
	}
	return panelSpec{title: "MEMORY", lines: []string{
		lab("used ") + used + "   " + usedDet,
		lab("swap ") + swap + "   " + swapDet,
		lab("bw   ") + bwSpark + "  " + bw,
	}}
}

func (m Model) powerPanel(t tier) panelSpec {
	label, _ := m.mode.Role("label")
	muted, _ := m.mode.Role("text_muted")
	lab := func(s string) string { return label.Style().Render(s) }

	pkg := m.numVal("power.pkg", " W", "%.1f")
	cpu := m.numVal("power.cpu", " W", "%.1f")
	gpu := m.numVal("power.gpu", " W", "%.1f")
	npu := m.numVal("power.npu", " W", "%.1f")
	pwrSpark := m.spark("power.pkg", t)
	pwrVal := m.numVal("power.pkg", " W", "%.0f")

	if t == tierNarrow {
		return panelSpec{title: "POWER", lines: []string{
			lab("pkg ") + pkg + "  " + lab("cpu ") + cpu + "  " + lab("gpu ") + gpu + "  " + lab("npu ") + npu,
			lab("pwr ") + pwrSpark + "  " + pwrVal,
		}}
	}
	if t == tierMedium {
		return panelSpec{title: "POWER", lines: []string{
			lab("pkg ") + pkg + "   " + lab("cpu ") + cpu + "   " + lab("gpu ") + gpu + "   " + lab("npu ") + npu,
			lab("pwr ") + pwrSpark + "  " + pwrVal,
		}}
	}
	lines := []string{
		lab("pkg ") + pkg,
		lab("cpu ") + cpu + "   " + lab("gpu ") + gpu + "   " + lab("npu ") + npu,
		lab("pwr ") + pwrSpark + "  " + pwrVal,
	}
	if d := m.detailOr("power.npu", ""); d != "" {
		lines = append(lines, muted.Style().Render("note: "+d+" → —"))
	}
	return panelSpec{title: "POWER", lines: lines}
}

func (m Model) gpuPanel(t tier) panelSpec {
	label, _ := m.mode.Role("label")
	lab := func(s string) string { return label.Style().Render(s) }

	gUtil := m.pctVal("gpu.util")
	vram := m.pctVal("gpu.vram")
	temp := m.numVal("gpu.temp", "°C", "%.0f")
	nUtil := m.pctVal("npu.util")

	if t == tierNarrow {
		return panelSpec{title: "GPU / NPU", lines: []string{
			lab("gpu ") + gUtil + "  " + lab("vram ") + vram + "  " + lab("temp ") + temp,
			lab("npu ") + nUtil + "  " + lab("(NPU)"),
		}}
	}
	if t == tierMedium {
		return panelSpec{title: "GPU / NPU", lines: []string{
			lab("gpu ") + lab("util ") + gUtil + "  " + lab("vram ") + vram + "  " + lab("temp ") + temp,
			lab("npu ") + lab("util ") + nUtil,
		}}
	}
	return panelSpec{title: "GPU / NPU", lines: []string{
		lab("gpu  ") + lab("util ") + gUtil + "   " + lab("vram ") + vram + "   " + lab("temp ") + temp,
		lab("npu  ") + lab("util ") + nUtil,
		lab("vram ") + m.spark("gpu.vram", t) + "  " + m.numVal("gpu.vram", "%", "%.0f"),
	}}
}

func (m Model) netDiskPanel(t tier) panelSpec {
	label, _ := m.mode.Role("label")
	lab := func(s string) string { return label.Style().Render(s) }

	if t == tierNarrow {
		// Two sparklines share each line here, so they are kept short to leave room
		// for the numeric values without wrapping.
		sp := func(name string) string { return render.BrailleSparkline(m.mode, m.history[name], 3) }
		return panelSpec{title: "NET & DISK", lines: []string{
			lab("net ↓ ") + sp("net.rx") + " " + m.numVal("net.rx", "", "%.1f") + "  " +
				lab("↑ ") + sp("net.tx") + " " + m.numVal("net.tx", " MB/s", "%.1f"),
			lab("disk r ") + sp("disk.read") + " " + m.numVal("disk.read", "", "%.1f") + "  " +
				lab("w ") + sp("disk.write") + " " + m.numVal("disk.write", " MB/s", "%.1f"),
		}}
	}
	return panelSpec{title: "NET & DISK", lines: []string{
		lab("net ↓  ") + m.spark("net.rx", t) + "  " + m.numVal("net.rx", " MB/s", "%.2f"),
		lab("net ↑  ") + m.spark("net.tx", t) + "  " + m.numVal("net.tx", " MB/s", "%.2f"),
		lab("disk r ") + m.spark("disk.read", t) + "  " + m.numVal("disk.read", " MB/s", "%.2f"),
		lab("disk w ") + m.spark("disk.write", t) + "  " + m.numVal("disk.write", " MB/s", "%.2f"),
	}}
}

func (m Model) loadUptimePanel() panelSpec {
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")
	lab := func(s string) string { return label.Style().Render(s) }

	load := m.numVal("cpu.load", "", "%.2f")
	seen := val.Style().Render(m.host.LastSeen.Format("15:04:05"))
	return panelSpec{title: "LOAD / UPTIME", lines: []string{
		lab("load   ") + load + "  " + lab("(1/5/15m)"),
		lab("uptime ") + val.Style().Render(m.uptime()) + "   " + lab("seen ") + seen,
	}}
}

func (m Model) processPanel(t tier) panelSpec {
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")
	muted, _ := m.mode.Role("text_muted")

	title := "PROCESSES"
	if t == tierWide || t == tierMedium {
		title = "PROCESSES (top by cpu)"
	}

	procs := m.sortedProcs()
	if len(procs) == 0 {
		return panelSpec{title: title, lines: []string{muted.Style().Render("waiting for a process table…")}}
	}

	var lines []string
	if t == tierNarrow {
		lines = append(lines, label.Style().Render(fmt.Sprintf("%-7s %5s  %s", "PID", "CPU%", "COMMAND")))
		for _, p := range procs {
			lines = append(lines, val.Style().Render(fmt.Sprintf("%-7d %5.1f  %s", p.PID, p.CPUPct, clip(p.Command, 24))))
		}
		return panelSpec{title: title, lines: lines}
	}

	lines = append(lines, label.Style().Render(fmt.Sprintf("%-7s %-8s %6s %6s  %s", "PID", "USER", "CPU%", "MEM%", "COMMAND")))
	for _, p := range procs {
		lines = append(lines, val.Style().Render(fmt.Sprintf("%-7d %-8s %6.1f %6.1f  %s",
			p.PID, "—", p.CPUPct, p.MemPct, clip(p.Command, 40))))
	}
	return panelSpec{title: title, lines: lines}
}

// tinyBody renders the key-numbers-only layout: one metric per line, no graphs,
// no per-core bars.
func (m Model) tinyBody() []string {
	label, _ := m.mode.Role("label")
	lab := func(s string) string { return label.Style().Render(fmt.Sprintf("%-5s ", s)) }
	return []string{
		lab("cpu") + m.pctVal("cpu.util"),
		lab("mem") + m.pctVal("mem.used"),
		lab("swap") + m.pctVal("mem.swap"),
		lab("pwr") + m.numVal("power.pkg", " W", "%.0f"),
		lab("temp") + m.numVal("temp.pkg", "°C", "%.0f"),
		lab("gpu") + m.pctVal("gpu.util"),
		lab("npu") + m.pctVal("npu.util"),
		lab("load") + m.numVal("cpu.load", "", "%.2f"),
		lab("freq") + m.numVal("cpu.freq", " GHz", "%.2f"),
	}
}

// --- per-core ----------------------------------------------------------------

// coreMatrix renders per-core bars in rows of `cols`: "c0 ███▌71  c1 ...".
func (m Model) coreMatrix(cores []float64, cols int) []string {
	muted, _ := m.mode.Role("text_muted")
	val, _ := m.mode.Role("value")
	cell := func(i int, v float64) string {
		return muted.Style().Render(fmt.Sprintf("c%d ", i)) +
			render.Gauge(m.mode, v, 4) +
			val.Style().Render(fmt.Sprintf("%2.0f", v))
	}
	var lines []string
	for i := 0; i < len(cores); i += cols {
		var row []string
		for j := i; j < i+cols && j < len(cores); j++ {
			row = append(row, cell(j, cores[j]))
		}
		lines = append(lines, strings.Join(row, "  "))
	}
	return lines
}

// coresAggregate collapses per-core into a single bar plus a count/avg/max
// summary (NARROW tier).
func (m Model) coresAggregate() string {
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")
	cores, ok := m.ok("cpu.cores")
	if !ok || len(cores.PerCore) == 0 {
		return label.Style().Render("cores ") + m.dash()
	}
	var sum, max float64
	for _, v := range cores.PerCore {
		sum += v
		if v > max {
			max = v
		}
	}
	avg := sum / float64(len(cores.PerCore))
	return label.Style().Render("cores ") + render.Gauge(m.mode, avg, 10) + "  " +
		val.Style().Render(fmt.Sprintf("%d cores", len(cores.PerCore))) + " " +
		label.Style().Render("avg ") + val.Style().Render(fmt.Sprintf("%.0f", avg)) + " " +
		label.Style().Render("max ") + val.Style().Render(fmt.Sprintf("%.0f", max))
}

func (m Model) sortedProcs() []domain.ProcessRow {
	ps := append([]domain.ProcessRow(nil), m.host.Processes...)
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].CPUPct > ps[j].CPUPct })
	if len(ps) > 6 {
		ps = ps[:6]
	}
	return ps
}

// --- value helpers -----------------------------------------------------------

// ok returns a metric only when it is present and StatusOK.
func (m Model) ok(name string) (domain.Metric, bool) {
	mm, present := m.byName[name]
	return mm, present && mm.Status == domain.StatusOK
}

// detailOr returns a metric's Detail string (only when the metric is OK) or the
// fallback.
func (m Model) detailOr(name, fallback string) string {
	if mm, present := m.byName[name]; present && mm.Detail != "" {
		return mm.Detail
	}
	return fallback
}

// dash is the unavailable affordance: a faint "—", never a fabricated 0.
func (m Model) dash() string {
	un, _ := m.mode.State("unavailable")
	return un.Style().Render("—")
}

// pctVal renders a 0–100 percentage value or the dash when unavailable.
func (m Model) pctVal(name string) string {
	mm, okv := m.ok(name)
	if !okv {
		return m.dash()
	}
	val, _ := m.mode.Role("value")
	unit, _ := m.mode.Role("unit")
	return val.Style().Render(fmt.Sprintf("%.0f", mm.Gauge)) + unit.Style().Render("%")
}

// numVal renders a gauge value with the given printf format and unit suffix, or
// the dash when unavailable. A leading space in unit is treated as a separator.
func (m Model) numVal(name, unit, format string) string {
	mm, okv := m.ok(name)
	if !okv {
		return m.dash()
	}
	val, _ := m.mode.Role("value")
	uRole, _ := m.mode.Role("unit")
	return val.Style().Render(fmt.Sprintf(format, mm.Gauge)) + uRole.Style().Render(unit)
}

// spark renders a braille sparkline from the history buffer, sized for the tier.
func (m Model) spark(name string, t tier) string {
	w := 16
	if t == tierNarrow {
		w = 10
	}
	return render.BrailleSparkline(m.mode, m.history[name], w)
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
