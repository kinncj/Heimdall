// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"reflect"
	"testing"
)

func TestCoreGroupsHybrid(t *testing.T) {
	// E owns the low ids (Apple Silicon shape): E first in logical order.
	coreType := []float64{CoreEff, CoreEff, CorePerf, CorePerf, CorePerf, CorePerf}
	coreUtil := []float64{5, 6, 80, 70, 60, 90}
	groups, ok := CoreGroups(coreUtil, coreType)
	if !ok {
		t.Fatal("expected grouping for a hybrid layout")
	}
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(groups))
	}
	if groups[0].Label != "E" || !reflect.DeepEqual(groups[0].Indices, []int{0, 1}) {
		t.Errorf("group 0 = %+v, want E {0,1} first (logical order)", groups[0])
	}
	if !reflect.DeepEqual(groups[0].Util, []float64{5, 6}) {
		t.Errorf("E util = %v, want [5 6]", groups[0].Util)
	}
	if groups[1].Label != "P" || !reflect.DeepEqual(groups[1].Indices, []int{2, 3, 4, 5}) {
		t.Errorf("group 1 = %+v, want P {2,3,4,5}", groups[1])
	}
}

// The Gauge-on-the-wire trap: consumers get PerCore but no Gauge. Grouping must
// work from coreType alone, and with a nil coreUtil (a consumer that only has
// the topology).
func TestCoreGroupsWithoutUtil(t *testing.T) {
	coreType := []float64{CorePerf, CorePerf, CoreEff, CoreEff}
	groups, ok := CoreGroups(nil, coreType)
	if !ok || len(groups) != 2 {
		t.Fatalf("groups=%v ok=%v, want 2 groups", groups, ok)
	}
	if groups[0].Label != "P" || groups[0].Util != nil {
		t.Errorf("group 0 = %+v, want P with nil util", groups[0])
	}
}

func TestCoreGroupsUniformIsNotGrouped(t *testing.T) {
	coreType := []float64{CorePerf, CorePerf, CorePerf, CorePerf}
	if _, ok := CoreGroups(nil, coreType); ok {
		t.Error("a uniform CPU must not group — render the plain grid")
	}
}

func TestCoreGroupsRejectsBadInput(t *testing.T) {
	cases := []struct {
		name              string
		coreUtil, coreType []float64
	}{
		{"empty", nil, nil},
		{"length mismatch with util", []float64{1, 2, 3}, []float64{CorePerf, CoreEff}},
		{"unknown type id", nil, []float64{CorePerf, 9}},
		{"non-integer type", nil, []float64{CorePerf, 1.5}},
	}
	for _, c := range cases {
		if _, ok := CoreGroups(c.coreUtil, c.coreType); ok {
			t.Errorf("%s: expected ok=false", c.name)
		}
	}
}

func TestCoreTypeSummary(t *testing.T) {
	cases := []struct {
		name     string
		coreType []float64
		want     string
	}{
		{"m3 max", rep(CorePerf, 12, CoreEff, 4), "12P + 4E"},
		{"uniform", rep(CorePerf, 16), "16 cores (uniform)"},
		{"three tiers", rep(CorePerf, 2, CoreEff, 4, CoreLP, 2), "2P + 4E + 2LP"},
		{"empty", nil, ""},
	}
	for _, c := range cases {
		if got := CoreTypeSummary(c.coreType); got != c.want {
			t.Errorf("%s: summary = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestCoreTypeLabel(t *testing.T) {
	if CoreTypeLabel(CorePerf) != "P" || CoreTypeLabel(CoreEff) != "E" || CoreTypeLabel(CoreLP) != "LP" {
		t.Error("standard labels wrong")
	}
	if CoreTypeLabel(-1) != "" || CoreTypeLabel(99) != "" {
		t.Error("out-of-range id must yield empty label")
	}
}

// rep builds a type slice from (id, count) pairs.
func rep(pairs ...int) []float64 {
	var out []float64
	for i := 0; i+1 < len(pairs); i += 2 {
		for n := 0; n < pairs[i+1]; n++ {
			out = append(out, float64(pairs[i]))
		}
	}
	return out
}
