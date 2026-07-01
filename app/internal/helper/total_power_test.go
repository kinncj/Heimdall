// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"testing"

	"heimdall/app/internal/domain"
)

// power.total is CPU (RAPL package) + GPU + NPU — the same rails btop/top show —
// except on Apple, where the collector already provides the SMC whole-system
// total, which must be left untouched.
func TestWithTotalPower(t *testing.T) {
	watts := func(name string, w float64) domain.Metric {
		return domain.Metric{Name: name, Unit: "watts", Status: domain.StatusOK, Gauge: w}
	}

	// Discrete NVIDIA: CPU package + GPU rail.
	got := byName(withTotalPower([]domain.Metric{watts("power.cpu", 32), watts("power.gpu", 506)}))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 538 {
		t.Errorf("nvidia total = %+v, want 538", m)
	}

	// AMD APU (Strix Halo): RAPL package and amdgpu are separate rails — sum them.
	got = byName(withTotalPower([]domain.Metric{watts("power.cpu", 51.7), watts("power.gpu", 70)}))
	if m := got["power.total"]; m.Gauge < 121 || m.Gauge > 122 {
		t.Errorf("apu total = %+v, want ~121.7", m)
	}

	// GB10 (no CPU power sensor): GPU rail alone.
	got = byName(withTotalPower([]domain.Metric{watts("power.gpu", 50)}))
	if m := got["power.total"]; m.Gauge != 50 {
		t.Errorf("gb10 total = %+v, want 50", m)
	}

	// A separate NPU power rail is included in the sum (never silently dropped).
	got = byName(withTotalPower([]domain.Metric{watts("power.cpu", 30), watts("power.gpu", 60), watts("power.npu", 10)}))
	if m := got["power.total"]; m.Gauge != 100 {
		t.Errorf("cpu+gpu+npu total = %+v, want 100", m)
	}

	// Apple: a total is already present (SMC PSTR) — leave it, never recompute.
	got = byName(withTotalPower([]domain.Metric{watts("power.total", 26), watts("power.cpu", 3), watts("power.gpu", 1)}))
	if m := got["power.total"]; m.Gauge != 26 {
		t.Errorf("apple total = %+v, want 26 (untouched)", m)
	}

	// Nothing to total → no metric rather than a phantom 0.
	if _, ok := byName(withTotalPower(nil))["power.total"]; ok {
		t.Error("no power sources should yield no power.total")
	}
}
