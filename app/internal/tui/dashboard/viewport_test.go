// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/tui/theme"
)

func colTitles(lay gridLayout) []string {
	out := make([]string, len(lay.columns))
	for i, c := range lay.columns {
		out[i] = c.title
	}
	return out
}

func TestLayoutRespondsToWidth(t *testing.T) {
	var m Model // layout is pure arithmetic — no theme needed
	cases := []struct {
		width int
		nameW int
		badge bool
		cols  []string
	}{
		{110, 16, true, []string{"CPU", "MEM", "DISK", "TEMP", "GPU", "PWR"}},
		{64, 16, true, []string{"CPU", "MEM"}},
		{46, 16, false, []string{"CPU"}},
		{28, 8, false, []string{"CPU"}},
		{20, 8, false, []string{}},
	}
	for _, tc := range cases {
		lay := m.layout(tc.width)
		if lay.nameW != tc.nameW || lay.badge != tc.badge || !equalStrings(colTitles(lay), tc.cols) {
			t.Errorf("layout(%d) = name=%d badge=%v cols=%v, want name=%d badge=%v cols=%v",
				tc.width, lay.nameW, lay.badge, colTitles(lay), tc.nameW, tc.badge, tc.cols)
		}
	}
}

// The horizontal half of the small-screen fix: nothing the grid renders may be
// wider than the terminal (else it clips off the right edge / wraps).
func TestGridViewFitsTerminalWidth(t *testing.T) {
	for _, w := range []int{50, 64, 88} {
		m := smallModel(t, bigFleet(t, 8), 30)
		m.width = w
		for i, line := range strings.Split(m.GridView(), "\n") {
			if lw := lipgloss.Width(line); lw > w {
				t.Errorf("width %d: line %d is %d cells wide:\n%q", w, i, lw, line)
			}
		}
	}
}

// Regression: a detail-view modal's header must show the real online/total count,
// not 0/0 (the modal header was hardcoded to 0,0).
func TestModalHeaderShowsHostCount(t *testing.T) {
	m := smallModel(t, obsReg(t), 30) // obsReg has one online host
	m.detail = true
	for _, modal := range []modalKind{modalTop, modalLogList, modalCmdList} {
		m.modal = modal
		out := m.ModalView()
		if !strings.Contains(out, "1/1") {
			t.Fatalf("modal %d header should show 1/1 online, got header:\n%s",
				modal, strings.SplitN(out, "\n", 4)[1])
		}
	}
}

// Regression: the detail view must fit the terminal height and keep its header on
// screen on a short terminal (the body used to scroll the header off).
func TestDetailViewFitsTerminalHeight(t *testing.T) {
	m := smallModel(t, bigFleet(t, 1), 14)
	m.detail = true
	out := m.DetailView()
	if lines := strings.Count(out, "\n") + 1; lines > 14 {
		t.Fatalf("detail rendered %d lines on a height-14 terminal, want <= 14", lines)
	}
	if !strings.Contains(out, "HEIMDALL") {
		t.Fatalf("detail header must stay visible:\n%s", out)
	}
	if !strings.Contains(out, "esc") {
		t.Fatalf("detail footer must stay visible:\n%s", out)
	}
}

// The detail body scrolls (mouse wheel / shift+arrows) within bounds while the
// header and footer stay fixed.
func TestDetailScrollMovesAndClamps(t *testing.T) {
	m := smallModel(t, bigFleet(t, 1), 12) // short terminal → body overflows
	m.detail = true
	max := m.detailMaxScroll()
	if max == 0 {
		t.Fatal("a height-12 detail view should be scrollable")
	}

	m = m.scroll(1) // wheel down
	if m.detailScroll == 0 {
		t.Fatal("scrolling down should move the detail body")
	}
	for i := 0; i < 50; i++ { // can't scroll past the end
		m = m.scroll(1)
	}
	if m.detailScroll != max {
		t.Fatalf("scroll should clamp at max %d, got %d", max, m.detailScroll)
	}
	for i := 0; i < 50; i++ { // can't scroll above the top
		m = m.scroll(-1)
	}
	if m.detailScroll != 0 {
		t.Fatalf("scroll should clamp at 0, got %d", m.detailScroll)
	}
	// The frame still fits the terminal height while scrolled.
	m.detailScroll = max
	if lines := strings.Count(m.DetailView(), "\n") + 1; lines > 12 {
		t.Fatalf("scrolled detail rendered %d lines, want <= 12", lines)
	}
}

func TestGridViewWideShowsEveryColumn(t *testing.T) {
	m := smallModel(t, bigFleet(t, 4), 30)
	m.width = 120
	out := m.GridView()
	for _, col := range []string{"HOST", "STATE", "CPU", "MEM", "DISK", "TEMP", "GPU", "PWR"} {
		if !strings.Contains(out, col) {
			t.Errorf("wide grid missing column %q", col)
		}
	}
}

func bigFleet(t *testing.T, n int) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	for i := 0; i < n; i++ {
		id := domain.HostID(fmt.Sprintf("host-%02d", i))
		reg.Enroll(domain.Host{ID: id, DisplayName: string(id)}, now)
		reg.Observe(id, []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 10}}, nil, now)
	}
	reg.Evaluate(now)
	return reg
}

func smallModel(t *testing.T, reg *domain.HostRegistry, h int) Model {
	t.Helper()
	th, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	md, ok := th.Mode("")
	if !ok {
		t.Fatal("default theme mode missing")
	}
	m := New(md, reg, time.Unix(1_700_000_000, 0))
	m.width, m.height = 100, h
	return m
}

// The reported bug: on a screen shorter than the fleet, GridView must not emit a
// frame taller than the terminal (which scrolls the header off-screen and makes
// filtering look inert).
func TestGridViewFitsTerminalHeight(t *testing.T) {
	m := smallModel(t, bigFleet(t, 25), 14)
	lines := strings.Count(m.GridView(), "\n") + 1
	if lines > 14 {
		t.Fatalf("GridView rendered %d lines on a height-14 terminal, want <= 14", lines)
	}
}

func TestGridViewKeepsSelectedHostVisible(t *testing.T) {
	m := smallModel(t, bigFleet(t, 25), 14)
	m.cursor = 24 // last host, well below the fold
	out := m.GridView()
	if !strings.Contains(out, "host-24") {
		t.Fatalf("selected host host-24 not visible in overflowed grid:\n%s", out)
	}
	if lines := strings.Count(out, "\n") + 1; lines > 14 {
		t.Fatalf("grid still overflows: %d lines", lines)
	}
}

func TestGridViewShowsHiddenRowIndicator(t *testing.T) {
	m := smallModel(t, bigFleet(t, 25), 14)
	if out := m.GridView(); !strings.Contains(out, "more") {
		t.Fatalf("expected a hidden-rows indicator when the list overflows:\n%s", out)
	}
}
