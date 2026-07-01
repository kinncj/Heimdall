// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"testing"

	"heimdall/app/internal/domain"
)

// On Apple Silicon Pro/Max the IOReport energy channels report 0 for CPU/ANE and
// only a sub-watt GPU figure. SMC PSTR ("System Total Power") is the real number.
// power.total must come from SMC, not the ~0.1W IOReport sum.
func TestAssembleApplePower_SMCWinsOverPhantomIOReport(t *testing.T) {
	got := byName(assembleApplePower(0, 0.1, 0, 5, true /*ioOK*/, 42.0, true /*smcOK*/, nil))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 42.0 {
		t.Errorf("power.total = %+v, want 42W from SMC", m)
	}
	if m := got["power.gpu"]; m.Status != domain.StatusOK || m.Gauge != 0.1 {
		t.Errorf("power.gpu = %+v, want 0.1W from IOReport", m)
	}
	if _, ok := got["power.cpu"]; ok {
		t.Errorf("power.cpu should be absent when IOReport CPU energy is 0")
	}
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 5 {
		t.Errorf("gpu.util = %+v, want 5%%", m)
	}
}

// Without SMC, a sub-watt IOReport sum must not shadow a real powermetrics package
// reading (the old mergeByName poisoning).
func TestAssembleApplePower_PowermetricsBeatsPhantomIOReport(t *testing.T) {
	pm := []domain.Metric{{Name: "power.total", Unit: "watts", Status: domain.StatusOK, Gauge: 12.7}}
	got := byName(assembleApplePower(0, 0.1, 0, -1, true, 0, false /*no smc*/, pm))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 12.7 {
		t.Errorf("power.total = %+v, want 12.7W from powermetrics", m)
	}
}

// With neither SMC nor powermetrics, a meaningful IOReport sum is still used as a
// last resort (back-compat for chips that do expose per-domain energy).
func TestAssembleApplePower_IOReportSumLastResort(t *testing.T) {
	got := byName(assembleApplePower(8, 3, 0, -1, true, 0, false, nil))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 11 {
		t.Errorf("power.total = %+v, want 11W from IOReport sum", m)
	}
}

// No sources at all: no power.total emitted (caller renders Unavailable).
func TestAssembleApplePower_NoSources(t *testing.T) {
	got := byName(assembleApplePower(0, 0, 0, -1, false, 0, false, nil))
	if _, ok := got["power.total"]; ok {
		t.Errorf("power.total should be absent when no source yields a value")
	}
}
