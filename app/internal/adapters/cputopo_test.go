// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"reflect"
	"testing"

	"heimdall/app/internal/domain"
)

func TestTypesFromCounts(t *testing.T) {
	cases := []struct {
		name           string
		p, e, total    int
		eFirst         bool
		want           []int
		ok             bool
	}{
		{"m3max e-first", 12, 4, 16, true,
			[]int{coreEff, coreEff, coreEff, coreEff,
				corePerf, corePerf, corePerf, corePerf, corePerf, corePerf,
				corePerf, corePerf, corePerf, corePerf, corePerf, corePerf}, true},
		{"p-first", 2, 2, 4, false, []int{corePerf, corePerf, coreEff, coreEff}, true},
		{"count mismatch is refused", 12, 4, 20, true, nil, false},
		{"no efficiency cores is uniform", 8, 0, 8, true,
			[]int{corePerf, corePerf, corePerf, corePerf, corePerf, corePerf, corePerf, corePerf}, true},
		{"zero total refused", 0, 0, 0, true, nil, false},
	}
	for _, c := range cases {
		got, ok := typesFromCounts(c.p, c.e, c.total, c.eFirst)
		if ok != c.ok {
			t.Errorf("%s: ok = %v, want %v", c.name, ok, c.ok)
			continue
		}
		if ok && !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: types = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestTypesFromRanges(t *testing.T) {
	cases := []struct {
		name       string
		perf, eff  string
		total      int
		wantPerf   []int // indices expected to be performance
		wantEff    []int
		ok         bool
	}{
		{"alder lake", "0-15", "16-23", 24,
			seq(0, 15), seq(16, 23), true},
		{"comma list", "0-3,8-11", "4-7", 12,
			append(seq(0, 3), seq(8, 11)...), seq(4, 7), true},
		{"single index", "0", "1", 2, []int{0}, []int{1}, true},
		{"out of bounds refused", "0-15", "16-23", 16, nil, nil, false},
		{"garbage refused", "banana", "4-7", 8, nil, nil, false},
		{"empty perf refused", "", "0-7", 8, nil, nil, false},
	}
	for _, c := range cases {
		got, ok := typesFromRanges(c.perf, c.eff, c.total)
		if ok != c.ok {
			t.Errorf("%s: ok = %v, want %v", c.name, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		for _, i := range c.wantPerf {
			if got[i] != corePerf {
				t.Errorf("%s: index %d = %d, want performance", c.name, i, got[i])
			}
		}
		for _, i := range c.wantEff {
			if got[i] != coreEff {
				t.Errorf("%s: index %d = %d, want efficiency", c.name, i, got[i])
			}
		}
	}
}

func seq(a, b int) []int {
	var out []int
	for i := a; i <= b; i++ {
		out = append(out, i)
	}
	return out
}

func TestTopologyMetric(t *testing.T) {
	types := []int{coreEff, coreEff, corePerf, corePerf, corePerf, corePerf}
	m := topologyMetric(types)
	if m.Name != "cpu.topology" {
		t.Fatalf("name = %q", m.Name)
	}
	if m.Status != domain.StatusOK || m.Kind != domain.KindPerCore {
		t.Fatalf("status/kind = %v/%v", m.Status, m.Kind)
	}
	if len(m.PerCore) != 6 {
		t.Fatalf("percore len = %d", len(m.PerCore))
	}
	if m.PerCore[0] != float64(coreEff) || m.PerCore[2] != float64(corePerf) {
		t.Fatalf("percore values wrong: %v", m.PerCore)
	}
	// Gauge is deliberately unset: per-core metrics ride the proto per_core oneof,
	// so a Gauge would be dropped on the wire. The layout lives in PerCore + Detail.
	if m.Gauge != 0 {
		t.Errorf("gauge = %v, want 0 (unused for per-core metrics)", m.Gauge)
	}
	if m.Detail != "4P + 2E" {
		t.Errorf("detail = %q, want \"4P + 2E\"", m.Detail)
	}
}

func TestTopologyMetricUniform(t *testing.T) {
	types := []int{corePerf, corePerf, corePerf, corePerf}
	m := topologyMetric(types)
	if m.Detail != "4 cores (uniform)" {
		t.Errorf("detail = %q, want \"4 cores (uniform)\"", m.Detail)
	}
}

// TestUniformTopology covers the platforms with no hybrid probe: every core the
// same type, detail says so, and the metric is still OK (the renderer decides
// not to label a uniform grid).
func TestUniformTypes(t *testing.T) {
	got := uniformTypes(32)
	if len(got) != 32 {
		t.Fatalf("len = %d", len(got))
	}
	for i, v := range got {
		if v != corePerf {
			t.Fatalf("index %d = %d, want performance", i, v)
		}
	}
	if _, ok := typesFromCounts(0, 0, 0, true); ok {
		t.Error("zero cores must not build a topology")
	}
}
