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
		add(m.row2(m.cpuPanel(t, col), m.memPanel(t, col), col))
		add(m.row2(m.powerPanel(t, col), m.gpuPanel(t, col), col))
		add(m.row2(m.netDiskPanel(t, col), m.loadUptimePanel(), col))
		add(m.renderPanel(m.processPanel(t, m.procRows(out)), full))
		return out
	}

	inner := m.width - 4
	if inner < 10 {
		inner = 10
	}
	for _, p := range []panelSpec{
		m.cpuPanel(t, inner), m.memPanel(t, inner), m.powerPanel(t, inner),
		m.gpuPanel(t, inner), m.netDiskPanel(t, inner),
	} {
		add(m.renderPanel(p, inner))
	}
	add(m.renderPanel(m.processPanel(t, m.procRows(out)), inner))
	return out
}

// procRows is how many content lines the PROCESSES box should hold so it grows
// to fill the height left below the other panels (btop-style), instead of
// leaving a void. out is the body built so far; add() will insert one blank
// separator, and the box itself costs a title row + two borders.
func (m Model) procRows(out []string) int {
	n := m.bodyHeight() - (len(out) + 1) - 3
	if n < 4 {
		n = 4
	}
	return n
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

func (m Model) cpuPanel(t tier, w int) panelSpec {
	label, _ := m.mode.Role("label")
	lab := func(s string) string { return label.Style().Render(s) }

	util := m.pctVal("cpu.util")
	freq := m.freqVal()
	load := m.numVal("cpu.load", "", "%.2f")

	if t == tierNarrow {
		return panelSpec{title: "CPU", lines: []string{
			lab("util ") + util + "   " + lab("freq ") + freq,
			lab("load ") + load,
			lab("util ") + m.sparkW("cpu.util", w) + " " + util,
			m.coresAggregate(),
		}}
	}

	cols := 3
	if t == tierWide {
		cols = 4 // a wider grid -> a shorter, denser per-core block
	}
	lines := []string{
		lab("util ") + util + "   " + lab("freq ") + freq + "   " + lab("load ") + load,
		lab("util ") + m.sparkW("cpu.util", w) + "  " + util,
	}
	if cores, ok := m.ok("cpu.cores"); ok && len(cores.PerCore) > 0 {
		lines = append(lines, lab(fmt.Sprintf("per-core (%d):", len(cores.PerCore))))
		lines = append(lines, m.coreMatrix(cores.PerCore, cols)...)
	}
	return panelSpec{title: "CPU", lines: lines}
}

func (m Model) memPanel(t tier, w int) panelSpec {
	label, _ := m.mode.Role("label")
	muted, _ := m.mode.Role("text_muted")
	lab := func(s string) string { return label.Style().Render(s) }

	used := m.pctVal("mem.used")
	usedDet := muted.Style().Render(m.detailOr("mem.used", ""))
	swap := m.pctVal("mem.swap")
	bw := m.numVal("mem.bw", " GB/s", "%.0f")

	if t == tierNarrow {
		return panelSpec{title: "MEMORY", lines: []string{
			lab("used ") + used + " " + usedDet + "   " + lab("swap ") + swap,
			lab("bw ") + m.sparkW("mem.bw", w) + " " + bw,
		}}
	}
	// Fill the panel with gauge bars + a usage trend so it isn't a near-empty box
	// next to the tall CPU panel.
	return panelSpec{title: "MEMORY", lines: []string{
		lab("used ") + m.barVal("mem.used", 10) + "   " + usedDet,
		lab("swap ") + m.barVal("mem.swap", 10),
		lab("mem  ") + m.sparkW("mem.used", w) + "  " + used,
		lab("bw   ") + m.sparkW("mem.bw", w) + "  " + bw,
	}}
}

func (m Model) powerPanel(t tier, w int) panelSpec {
	label, _ := m.mode.Role("label")
	lab := func(s string) string { return label.Style().Render(s) }

	cpu := m.numVal("power.cpu", " W", "%.1f")
	gpu := m.numVal("power.gpu", " W", "%.1f")
	npu := m.numVal("power.npu", " W", "%.1f")
	// Headline the whole-machine total (CPU + GPU + NPU on non-Apple, or the SMC
	// whole-system figure on Apple) so a busy GPU isn't hidden behind the much
	// smaller CPU figure.
	totalVal := m.numVal("power.total", " W", "%.0f")

	if t == tierNarrow {
		return panelSpec{title: "POWER", lines: []string{
			lab("cpu ") + cpu + "  " + lab("gpu ") + gpu + "  " + lab("npu ") + npu,
			lab("total ") + m.sparkW("power.total", w) + "  " + totalVal,
		}}
	}
	if t == tierMedium {
		return panelSpec{title: "POWER", lines: []string{
			lab("cpu ") + cpu + "   " + lab("gpu ") + gpu + "   " + lab("npu ") + npu,
			lab("total ") + m.sparkW("power.total", w) + "  " + totalVal,
		}}
	}
	return panelSpec{title: "POWER", lines: []string{
		lab("total ") + m.sparkW("power.total", w) + "  " + totalVal,
		lab("cpu ") + cpu + "   " + lab("gpu ") + gpu + "   " + lab("npu ") + npu,
	}}
}

func (m Model) gpuPanel(t tier, w int) panelSpec {
	label, _ := m.mode.Role("label")
	lab := func(s string) string { return label.Style().Render(s) }

	gUtil := m.pctVal("gpu.util")
	vram := m.pctVal("gpu.vram")
	vramDet := m.detailOr("gpu.vram", "")
	temp := m.numVal("gpu.temp", "°C", "%.0f")
	clk := m.numVal("gpu.clock", " MHz", "%.0f")
	memU := m.pctVal("gpu.mem.util")
	fan := m.pctVal("gpu.fan")
	nUtil := m.pctVal("npu.util")

	if t == tierNarrow {
		return panelSpec{title: "GPU / NPU", lines: []string{
			lab("gpu ") + gUtil + "  " + lab("vram ") + vram + "  " + lab("temp ") + temp,
			lab("clk ") + clk + "  " + lab("mem ") + memU + "  " + lab("fan ") + fan,
			lab("npu ") + nUtil + "  " + lab("(NPU)"),
		}}
	}
	if t == tierMedium {
		return panelSpec{title: "GPU / NPU", lines: []string{
			lab("gpu  ") + m.barVal("gpu.util", 10) + "   " + lab("temp ") + temp,
			lab("vram ") + m.barVal("gpu.vram", 10),
			lab("clk ") + clk + "  " + lab("mem ") + memU + "  " + lab("fan ") + fan,
			lab("npu  ") + lab("util ") + nUtil,
		}}
	}
	// WIDE: gauge bars for util and vram (no duplicated vram line), plus a clock /
	// mem / fan line and a util trend, so the panel reads at a glance and fills
	// its height.
	muted, _ := m.mode.Role("text_muted")
	// Clip the vram detail to the space left on its line so a long value (e.g. the
	// GB10 "42/122 GB shared") can't wrap and break the panel box.
	vbar := m.barVal("gpu.vram", 12)
	vramDet = clip(vramDet, w-2-lipgloss.Width(lab("vram ")+vbar+"  "))
	return panelSpec{title: "GPU / NPU", lines: []string{
		lab("gpu  ") + m.barVal("gpu.util", 12) + "   " + lab("temp ") + temp,
		lab("vram ") + vbar + "  " + muted.Style().Render(vramDet),
		lab("clk  ") + clk + "   " + lab("mem ") + memU + "   " + lab("fan ") + fan,
		lab("util ") + m.sparkW("gpu.util", w) + "  " + gUtil,
		lab("npu  ") + lab("util ") + nUtil,
	}}
}

func (m Model) netDiskPanel(t tier, w int) panelSpec {
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
		lab("net ↓  ") + m.sparkW("net.rx", w) + "  " + m.numVal("net.rx", " MB/s", "%.2f"),
		lab("net ↑  ") + m.sparkW("net.tx", w) + "  " + m.numVal("net.tx", " MB/s", "%.2f"),
		lab("disk r ") + m.sparkW("disk.read", w) + "  " + m.numVal("disk.read", " MB/s", "%.2f"),
		lab("disk w ") + m.sparkW("disk.write", w) + "  " + m.numVal("disk.write", " MB/s", "%.2f"),
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

func (m Model) processPanel(t tier, rows int) panelSpec {
	label, _ := m.mode.Role("label")
	val, _ := m.mode.Role("value")
	muted, _ := m.mode.Role("text_muted")

	title := "PROCESSES"
	if t == tierWide || t == tierMedium {
		title = "PROCESSES (top by cpu)"
	}
	if rows < 2 {
		rows = 2
	}

	procs := m.sortedProcs()
	if len(procs) == 0 {
		lines := []string{muted.Style().Render("waiting for a process table…")}
		return panelSpec{title: title, lines: padTo(lines, rows)}
	}

	narrow := t == tierNarrow
	var lines []string
	if narrow {
		lines = append(lines, label.Style().Render(fmt.Sprintf("%-7s %5s  %s", "PID", "CPU%", "COMMAND")))
	} else {
		// No USER column: the pushed process table carries no username, and a column
		// of dashes reads as broken. PID / CPU% / MEM% / COMMAND are what we have.
		lines = append(lines, label.Style().Render(fmt.Sprintf("%-7s %6s %6s  %s", "PID", "CPU%", "MEM%", "COMMAND")))
	}
	for i, p := range procs {
		if i >= rows-1 { // leave room for the header within the budget
			break
		}
		if narrow {
			lines = append(lines, val.Style().Render(fmt.Sprintf("%-7d %5.1f  %s", p.PID, p.CPUPct, clip(p.Command, 24))))
		} else {
			lines = append(lines, val.Style().Render(fmt.Sprintf("%-7d %6.1f %6.1f  %s",
				p.PID, p.CPUPct, p.MemPct, clip(p.Command, 48))))
		}
	}
	return panelSpec{title: title, lines: padTo(lines, rows)}
}

// padTo extends lines with blank rows up to n so the boxed panel grows to fill
// the height budgeted for it.
func padTo(lines []string, n int) []string {
	for len(lines) < n {
		lines = append(lines, "")
	}
	return lines
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
		lab("pwr") + m.numVal("power.total", " W", "%.0f"),
		lab("temp") + m.numVal("temp.pkg", "°C", "%.0f"),
		lab("gpu") + m.pctVal("gpu.util"),
		lab("npu") + m.pctVal("npu.util"),
		lab("load") + m.numVal("cpu.load", "", "%.2f"),
		lab("freq") + m.freqVal(),
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

// pctBar renders a severity-coloured gauge bar for a 0–100 metric, or the dash.
func (m Model) pctBar(name string, cells int) string {
	mm, okv := m.ok(name)
	if !okv {
		return m.dash()
	}
	return render.Gauge(m.mode, mm.Gauge, cells)
}

// barVal renders "<bar> NN%" for an OK percentage metric, or a single dash when
// unavailable (so it never reads as a doubled "— —").
func (m Model) barVal(name string, cells int) string {
	if _, okv := m.ok(name); !okv {
		return m.dash()
	}
	return m.pctBar(name, cells) + " " + m.pctVal(name)
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

// freqVal renders cpu.freq (collected in MHz) as GHz, or the dash. Showing
// "4200.00 GHz" was wrong — the counter is MHz, so 4200 MHz is 4.20 GHz.
func (m Model) freqVal() string {
	mm, okv := m.ok("cpu.freq")
	if !okv {
		return m.dash()
	}
	val, _ := m.mode.Role("value")
	uRole, _ := m.mode.Role("unit")
	return val.Style().Render(fmt.Sprintf("%.2f", mm.Gauge/1000)) + uRole.Style().Render(" GHz")
}

// sparkW renders a braille sparkline sized to fill a panel of inner width w,
// leaving room for the line's label and trailing value.
func (m Model) sparkW(name string, w int) string {
	// Reserve room for the box padding (2) plus the longest label + trailing value
	// on a sparkline row (e.g. "disk r " + "12.40 MB/s") so the value never wraps.
	sw := w - 22
	if sw < 8 {
		sw = 8
	}
	if sw > 80 {
		sw = 80
	}
	return render.BrailleSparkline(m.mode, m.history[name], sw)
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
