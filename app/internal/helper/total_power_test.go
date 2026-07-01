// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"testing"

	"heimdall/app/internal/domain"
)

// power.total must be source-aware so it never double-counts an integrated GPU
// (already inside the CPU package) nor Apple's SMC PSTR (already whole-system).
func TestWithTotalPower(t *testing.T) {
	watts := func(name string, w float64) domain.Metric {
		return domain.Metric{Name: name, Unit: "watts", Status: domain.StatusOK, Gauge: w}
	}

	// Linux + discrete/superchip NVIDIA: CPU package + a separate GPU rail.
	got := byName(withTotalPower([]domain.Metric{watts("power.pkg", 32), watts("power.gpu", 506)}, "linux", true))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 538 {
		t.Errorf("nvidia total = %+v, want 538", m)
	}

	// GB10 (ARM, no RAPL package): only the GPU rail is measurable.
	got = byName(withTotalPower([]domain.Metric{watts("power.gpu", 50)}, "linux", true))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 50 {
		t.Errorf("gb10 total = %+v, want 50", m)
	}

	// Apple: power.pkg is SMC PSTR (already whole-system) — do not add the GPU.
	got = byName(withTotalPower([]domain.Metric{watts("power.pkg", 20), watts("power.gpu", 5)}, "darwin", false))
	if m := got["power.total"]; m.Gauge != 20 {
		t.Errorf("apple total = %+v, want 20 (pkg only)", m)
	}

	// AMD APU (integrated GPU inside the package): total is the package alone.
	got = byName(withTotalPower([]domain.Metric{watts("power.pkg", 40), watts("power.gpu", 15)}, "linux", false))
	if m := got["power.total"]; m.Gauge != 40 {
		t.Errorf("apu total = %+v, want 40 (no double-count)", m)
	}

	// Nothing to total → no metric rather than a phantom 0.
	if _, ok := byName(withTotalPower(nil, "linux", true))["power.total"]; ok {
		t.Error("no power sources should yield no power.total")
	}
}
