// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"fmt"
	"strings"
)

// Core-type ids are the standardized, platform-neutral core taxonomy. The
// daemon's adapters translate each platform's own notion — Apple perf-levels,
// Linux cpu_core/cpu_atom, Windows EfficiencyClass — into these ids before any
// data leaves the host, so every consumer (TUI, CLI) reads the same scheme.
// Lower id = higher-performance tier, so CorePerf is always "P".
const (
	CorePerf = 0 // performance (P) core
	CoreEff  = 1 // efficiency (E) core
	CoreLP   = 2 // low-power efficiency core
)

var coreTypeLabels = []string{"P", "E", "LP"}

// CoreTypeLabel is the display label for a core-type id ("P", "E", "LP"), or ""
// for an id outside the known taxonomy.
func CoreTypeLabel(id int) string {
	if id < 0 || id >= len(coreTypeLabels) {
		return ""
	}
	return coreTypeLabels[id]
}

// CoreGroup is one block of same-type logical cores, already standardized for a
// consumer to render verbatim: a label, the real logical core ids it owns, and
// their utilisations. No consumer re-derives grouping from raw type ids.
type CoreGroup struct {
	Label   string    // "P", "E", "LP"
	Type    int       // CorePerf / CoreEff / CoreLP
	Indices []int     // real logical core ids in this group, ascending
	Util    []float64 // per-core utilisation aligned to Indices; nil when unavailable
}

// CoreGroups turns the standardized per-core type layout into display-ready
// groups, ordered by first appearance (logical core order — whichever type owns
// core 0 comes first). ok is false — meaning "there is nothing to group, render
// the plain per-core list" — when the layout is empty, uniform (a single type),
// carries an unknown type id, or (when coreUtil is supplied) disagrees in length
// with it (a torn snapshot). coreUtil may be nil; when it matches coreType in
// length, each group's Util is filled from it.
//
// This is the one place the P/E grouping is computed. It lives in the domain
// core (no framework imports) so the TUI and CLI share it and stay dumb
// renderers — the grouping is not view logic.
func CoreGroups(coreUtil, coreType []float64) ([]CoreGroup, bool) {
	if len(coreType) == 0 {
		return nil, false
	}
	withUtil := len(coreUtil) == len(coreType)
	if coreUtil != nil && !withUtil {
		return nil, false // torn snapshot — don't mislabel
	}
	byType := map[int]*CoreGroup{}
	var order []*CoreGroup
	for i, tv := range coreType {
		t := int(tv)
		if float64(t) != tv || CoreTypeLabel(t) == "" {
			return nil, false
		}
		g, seen := byType[t]
		if !seen {
			g = &CoreGroup{Label: CoreTypeLabel(t), Type: t}
			byType[t] = g
			order = append(order, g)
		}
		g.Indices = append(g.Indices, i)
		if withUtil {
			g.Util = append(g.Util, coreUtil[i])
		}
	}
	if len(order) < 2 {
		return nil, false
	}
	out := make([]CoreGroup, len(order))
	for i, g := range order {
		out[i] = *g
	}
	return out, true
}

// CoreTypeSummary renders the human mix of a standardized type layout, e.g.
// "12P + 4E" for hybrid silicon or "16 cores (uniform)" for a single-type CPU.
func CoreTypeSummary(coreType []float64) string {
	if len(coreType) == 0 {
		return ""
	}
	counts := map[int]int{}
	var order []int
	for _, tv := range coreType {
		t := int(tv)
		if _, seen := counts[t]; !seen {
			order = append(order, t)
		}
		counts[t]++
	}
	if len(counts) <= 1 {
		return fmt.Sprintf("%d cores (uniform)", len(coreType))
	}
	// Summary reads by tier (P before E before LP), independent of core ordering.
	var parts []string
	for id := range coreTypeLabels {
		if counts[id] > 0 {
			parts = append(parts, fmt.Sprintf("%d%s", counts[id], coreTypeLabels[id]))
		}
	}
	return strings.Join(parts, " + ")
}
