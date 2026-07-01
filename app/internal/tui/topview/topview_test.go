// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package topview

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/theme"
)

var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

func strip(s string) string { return ansiRE.ReplaceAllString(s, "") }

func darkMode(t *testing.T) theme.Mode {
	t.Helper()
	th, err := theme.Load()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := th.Mode("dark")
	if !ok {
		t.Fatal("dark mode not found")
	}
	return m
}

// sampleHost builds a host with the full v2.2 metric set. npu.util is deliberately
// Unavailable so the dash affordance can be asserted.
func sampleHost() domain.HostView {
	ok := func(name string, g float64, unit, detail string) domain.Metric {
		return domain.Metric{Name: name, Unit: unit, Status: domain.StatusOK, Kind: domain.KindGauge, Gauge: g, Detail: detail}
	}
	return domain.HostView{
		Host: domain.Host{
			ID:          "workstation",
			DisplayName: "workstation",
			Context:     domain.HostContext{OS: "macOS 14.4", Arch: "arm64"},
		},
		State:    domain.StateOnline,
		LastSeen: time.Date(2026, 6, 30, 14, 22, 8, 0, time.UTC),
		LastSnapshot: []domain.Metric{
			ok("cpu.util", 72, "%", ""),
			ok("cpu.freq", 3.20, "GHz", "3.20 GHz"),
			ok("cpu.load", 2.41, "", ""),
			{Name: "cpu.cores", Status: domain.StatusOK, Kind: domain.KindPerCore,
				PerCore: []float64{71, 52, 80, 41, 63, 33, 70, 55, 88, 44}},
			ok("mem.used", 61, "%", "18.4 / 32 GB"),
			ok("mem.swap", 2, "%", "0.6 / 8 GB"),
			ok("mem.bw", 41, "GB/s", "41 GB/s"),
			ok("power.total", 20.1, "W", ""),
			ok("power.cpu", 14.1, "W", ""),
			ok("power.gpu", 6.0, "W", ""),
			{Name: "power.npu", Status: domain.StatusUnavailable, Kind: domain.KindGauge,
				Detail: "npu power residency unavailable"},
			ok("gpu.util", 64, "%", ""),
			ok("gpu.vram", 47, "%", "7.5 / 16 GB"),
			ok("gpu.temp", 58, "°C", ""),
			{Name: "npu.util", Status: domain.StatusUnavailable, Kind: domain.KindGauge},
			ok("temp.pkg", 58, "°C", ""),
			ok("net.rx", 3.20, "MB/s", ""),
			ok("net.tx", 0.80, "MB/s", ""),
			ok("disk.read", 12.4, "MB/s", ""),
			ok("disk.write", 2.10, "MB/s", ""),
			ok("host.uptime", 540000, "s", ""),
		},
		Processes: []domain.ProcessRow{
			{PID: 1234, CPUPct: 42.1, MemPct: 3.2, Command: "heimdall-dashboard"},
			{PID: 880, CPUPct: 11.7, MemPct: 1.1, Command: "heimdall-daemon"},
			{PID: 2051, CPUPct: 8.4, MemPct: 6.0, Command: "firefox"},
			{PID: 3120, CPUPct: 3.2, MemPct: 2.4, Command: "ghostty"},
		},
	}
}

func sampleHistory() map[string][]float64 {
	ramp := []float64{10, 20, 35, 50, 60, 72, 65, 58, 70, 72}
	return map[string][]float64{
		"cpu.util":    ramp,
		"mem.bw":      ramp,
		"power.total": ramp,
		"gpu.vram":    ramp,
		"net.rx":      ramp,
		"net.tx":      ramp,
		"disk.read":   ramp,
		"disk.write":  ramp,
	}
}

func newModel(t *testing.T, w, h int) Model {
	t.Helper()
	return New(sampleHost(), sampleHistory(), darkMode(t), w, h)
}

func TestLayoutTiers(t *testing.T) {
	cases := []struct {
		width int
		want  tier
	}{
		{120, tierWide},
		{100, tierWide},
		{80, tierMedium},
		{60, tierMedium},
		{50, tierNarrow},
		{40, tierNarrow},
		{30, tierTiny},
		{20, tierTiny},
	}
	for _, c := range cases {
		if got := layout(c.width); got != c.want {
			t.Errorf("layout(%d) = %v, want %v", c.width, got, c.want)
		}
	}
}

func TestNoLineExceedsWidth(t *testing.T) {
	for _, w := range []int{120, 80, 50, 30} {
		m := newModel(t, w, 40)
		for i, line := range strings.Split(m.View(), "\n") {
			// lipgloss.Width is the ANSI-aware display width; for this width-1
			// glyph set it equals runewidth(strip(line)).
			if got := lipgloss.Width(line); got > w {
				t.Errorf("width %d: line %d display width %d > %d: %q", w, i, got, w, strip(line))
			}
		}
	}
}

func TestTinyShowsKeyNumbersOnly(t *testing.T) {
	m := newModel(t, 30, 40)
	s := strip(m.View())
	for _, want := range []string{"72", "61", "20", "58"} { // cpu, mem, power, temp
		if !strings.Contains(s, want) {
			t.Errorf("tiny view missing key number %q\n%s", want, s)
		}
	}
	if strings.Contains(s, "c0") {
		t.Errorf("tiny view must not show per-core labels (c0)\n%s", s)
	}
	if strings.ContainsAny(s, "⣀⣤⣶⣷⣿") {
		t.Errorf("tiny view must not show braille sparklines\n%s", s)
	}
}

func TestUnavailableRendersDash(t *testing.T) {
	m := newModel(t, 30, 40) // tiny shows one metric per line, easy to target
	s := strip(m.View())
	var npuLine string
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(strings.TrimSpace(line), "npu") {
			npuLine = line
		}
	}
	if npuLine == "" {
		t.Fatalf("no npu line in tiny view\n%s", s)
	}
	if !strings.Contains(npuLine, "—") {
		t.Errorf("unavailable npu must render —, got %q", npuLine)
	}
	if strings.Contains(npuLine, "0") {
		t.Errorf("unavailable npu must not fabricate a 0, got %q", npuLine)
	}
}

func TestExitOnEscQ(t *testing.T) {
	m := newModel(t, 120, 40)
	if _, act := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); act != ActBack {
		t.Errorf("esc should go back, got %v", act)
	}
	if _, act := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}); act != ActQuit {
		t.Errorf("q should quit the app, got %v", act)
	}
	if _, act := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); act != ActQuit {
		t.Errorf("ctrl+c should quit the app, got %v", act)
	}
	if _, act := m.Update(tea.KeyMsg{Type: tea.KeyDown}); act != ActNone {
		t.Errorf("down should stay in the view, got %v", act)
	}
}

func TestNpuLabel(t *testing.T) {
	for _, w := range []int{120, 80, 50, 30} {
		m := newModel(t, w, 40)
		s := strip(m.View())
		if !strings.Contains(s, "NPU") && !strings.Contains(s, "npu") {
			t.Errorf("width %d: view should reference the NPU\n%s", w, s)
		}
		if strings.Contains(s, "ANE") {
			t.Errorf("width %d: view must never show the legacy ANE label\n%s", w, s)
		}
	}
}
