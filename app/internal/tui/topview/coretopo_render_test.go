// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package topview

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"heimdall/app/internal/domain"
)

// hybridHost is sampleHost plus a 4E+6P cpu.topology (E-cluster owns the low
// core ids, like Apple Silicon) and the SMC per-cluster power rails.
func hybridHost() domain.HostView {
	h := sampleHost()
	h.LastSnapshot = append(h.LastSnapshot,
		// Gauge is deliberately 0: per-core metrics ride the proto `per_core` oneof
		// arm, so the Gauge (distinct-type count) is dropped on the wire and the
		// dashboard receives it as 0. The grouping must derive the count from the
		// PerCore slice, not the Gauge.
		domain.Metric{Name: "cpu.topology", Status: domain.StatusOK, Kind: domain.KindPerCore,
			Gauge:   0,
			PerCore: []float64{1, 1, 1, 1, 0, 0, 0, 0, 0, 0}, // 0=P, 1=E
			Detail:  "6P + 4E"},
		domain.Metric{Name: "power.cpu.pcluster", Unit: "watts", Status: domain.StatusOK, Kind: domain.KindGauge, Gauge: 15.0},
		domain.Metric{Name: "power.cpu.ecluster", Unit: "watts", Status: domain.StatusOK, Kind: domain.KindGauge, Gauge: 3.1},
	)
	return h
}

func hybridModel(t *testing.T, w, h int) Model {
	t.Helper()
	return New(hybridHost(), sampleHistory(), darkMode(t), w, h)
}

// With topology, the per-core grid is grouped under text headers in logical core
// order (E owns c0 here, so E renders first), and the anonymous "per-core (N):"
// header disappears.
func TestCPUGridGroupedByCoreType(t *testing.T) {
	for _, w := range []int{120, 80} {
		s := strip(hybridModel(t, w, 50).View())
		if !strings.Contains(s, "E-cores (4):") {
			t.Errorf("width %d: missing E group header\n%s", w, s)
		}
		if !strings.Contains(s, "P-cores (6):") {
			t.Errorf("width %d: missing P group header\n%s", w, s)
		}
		if strings.Contains(s, "per-core (10):") {
			t.Errorf("width %d: anonymous per-core header must be replaced by groups\n%s", w, s)
		}
		if strings.Index(s, "E-cores (4):") > strings.Index(s, "P-cores (6):") {
			t.Errorf("width %d: E owns c0 and must render first (logical core order)\n%s", w, s)
		}
	}
}

// Core ids must stay the real logical ids — grouping never renumbers. c4 is the
// first P core, so "c4" must appear after the P header; "c0" after the E header.
func TestGroupedGridKeepsLogicalIDs(t *testing.T) {
	s := strip(hybridModel(t, 120, 50).View())
	pIdx := strings.Index(s, "P-cores (6):")
	eIdx := strings.Index(s, "E-cores (4):")
	c4 := strings.Index(s, "c4 ")
	if c4 < pIdx {
		t.Errorf("c4 (first P core) must render inside the P block (pIdx=%d c4=%d)", pIdx, c4)
	}
	c0 := strings.Index(s, "c0 ")
	if !(c0 > eIdx && c0 < pIdx) {
		t.Errorf("c0 (first E core) must render inside the E block (eIdx=%d c0=%d pIdx=%d)", eIdx, c0, pIdx)
	}
}

// The POWER panel gains the cluster line when the rails exist.
func TestPowerPanelShowsClusterSplit(t *testing.T) {
	for _, w := range []int{120, 80} {
		s := strip(hybridModel(t, w, 50).View())
		if !strings.Contains(s, "clusters") {
			t.Errorf("width %d: missing cpu clusters line\n%s", w, s)
		}
		if !strings.Contains(s, "15.0") || !strings.Contains(s, "3.1") {
			t.Errorf("width %d: cluster watts missing\n%s", w, s)
		}
	}
}

// NARROW collapses to one aggregate per type instead of one anonymous bar.
func TestNarrowAggregatePerType(t *testing.T) {
	s := strip(hybridModel(t, 50, 50).View())
	if !strings.Contains(s, "P (6)") || !strings.Contains(s, "E (4)") {
		t.Errorf("narrow view should aggregate per core type\n%s", s)
	}
}

// The grouped grid and cluster line must respect every tier's width budget —
// the a11y min-width-resize check (nothing clips, nothing wraps).
func TestNoLineExceedsWidthHybrid(t *testing.T) {
	for _, w := range []int{120, 100, 80, 60, 50, 40, 30} {
		m := hybridModel(t, w, 50)
		for i, line := range strings.Split(m.View(), "\n") {
			if got := lipgloss.Width(line); got > w {
				t.Errorf("width %d: line %d display width %d > %d: %q", w, i, got, w, strip(line))
			}
		}
	}
}

// Mixed-version fleet: no cpu.topology → today's unlabelled grid, no group
// headers, no cluster line. This is the BC guarantee.
func TestFallbackWithoutTopologyUnchanged(t *testing.T) {
	s := strip(newModel(t, 120, 50).View())
	if !strings.Contains(s, "per-core (10):") {
		t.Errorf("fallback must keep the anonymous per-core header\n%s", s)
	}
	for _, banned := range []string{"P-cores", "E-cores", "clusters"} {
		if strings.Contains(s, banned) {
			t.Errorf("fallback must not render %q\n%s", banned, s)
		}
	}
}

// A uniform topology (single type) adds nothing over the plain grid — render it
// exactly like the fallback.
func TestUniformTopologyRendersPlainGrid(t *testing.T) {
	h := sampleHost()
	h.LastSnapshot = append(h.LastSnapshot, domain.Metric{
		Name: "cpu.topology", Status: domain.StatusOK, Kind: domain.KindPerCore,
		Gauge:   0, // wire drops Gauge for per-core metrics
		PerCore: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Detail:  "10 cores (uniform)",
	})
	m := New(h, sampleHistory(), darkMode(t), 120, 50)
	s := strip(m.View())
	if !strings.Contains(s, "per-core (10):") {
		t.Errorf("uniform topology must render the plain grid\n%s", s)
	}
	if strings.Contains(s, "P-cores") {
		t.Errorf("uniform topology must not invent a P group\n%s", s)
	}
}

// A topology whose length disagrees with cpu.cores (torn snapshot, hotplug) must
// be ignored rather than mis-labelling bars.
func TestMismatchedTopologyIgnored(t *testing.T) {
	h := sampleHost()
	h.LastSnapshot = append(h.LastSnapshot, domain.Metric{
		Name: "cpu.topology", Status: domain.StatusOK, Kind: domain.KindPerCore,
		Gauge:   2,
		PerCore: []float64{1, 1, 0, 0}, // 4 entries vs 10 cores
		Detail:  "2P + 2E",
	})
	m := New(h, sampleHistory(), darkMode(t), 120, 50)
	s := strip(m.View())
	if !strings.Contains(s, "per-core (10):") {
		t.Errorf("mismatched topology must fall back to the plain grid\n%s", s)
	}
}
