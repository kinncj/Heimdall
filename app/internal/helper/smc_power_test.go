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
	got := byName(assembleApplePower(0, 0.1, 0, 5, true /*ioOK*/, 42.0, true /*smcOK*/, smcCPUReading{}, nil))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 42.0 {
		t.Errorf("power.total = %+v, want 42W from SMC", m)
	}
	if m := got["power.gpu"]; m.Status != domain.StatusOK || m.Gauge != 0.1 {
		t.Errorf("power.gpu = %+v, want 0.1W from IOReport", m)
	}
	// Pro/Max exposes no per-domain CPU power (IOReport and powermetrics both read
	// 0); say why rather than leaving a silent blank.
	if m := got["power.cpu"]; m.Status != domain.StatusUnavailable || m.Detail == "" {
		t.Errorf("power.cpu = %+v, want unavailable-with-reason on Pro/Max", m)
	}
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 5 {
		t.Errorf("gpu.util = %+v, want 5%%", m)
	}
}

// On Pro/Max IOReport gives no CPU power (0), but the raw SMC P-core cluster keys
// do — power.cpu is filled from SMC then, tagged so its source is clear.
func TestAssembleApplePower_SMCCPUFillsProMax(t *testing.T) {
	got := byName(assembleApplePower(0 /*ioreport cpu 0*/, 0.1, 0, 5, true, 42, true, smcCPUReading{Total: 10.6, P: 8.0, E: 2.6, OK: true, POK: true, EOK: true}, nil))
	if m := got["power.cpu"]; m.Status != domain.StatusOK || m.Gauge != 10.6 {
		t.Fatalf("power.cpu = %+v, want 10.6W from SMC", m)
	}
	if got["power.cpu"].Detail == "" {
		t.Error("SMC-sourced power.cpu should be labelled")
	}
}

// GPU power is always a valid IOReport channel on Apple Silicon, so it must be
// reported even at idle (~0 W) — dropping it made a base M4 look like it "has no
// GPU" while a Pro/Max (whose CPU/ANE channels genuinely read 0) showed the GPU.
func TestAssembleApplePower_ShowsIdleGPU(t *testing.T) {
	got := byName(assembleApplePower(0.5, 0 /*gpu idle*/, 0, 3, true /*ioOK*/, 20, true, smcCPUReading{}, nil))
	if m, ok := got["power.gpu"]; !ok || m.Status != domain.StatusOK || m.Gauge != 0 {
		t.Fatalf("power.gpu = %+v (present=%v), want a 0 W reading at idle", m, ok)
	}
}

// Without SMC, a sub-watt IOReport sum must not shadow a real powermetrics package
// reading (the old mergeByName poisoning).
func TestAssembleApplePower_PowermetricsBeatsPhantomIOReport(t *testing.T) {
	pm := []domain.Metric{{Name: "power.total", Unit: "watts", Status: domain.StatusOK, Gauge: 12.7}}
	got := byName(assembleApplePower(0, 0.1, 0, -1, true, 0, false /*no smc*/, smcCPUReading{}, pm))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 12.7 {
		t.Errorf("power.total = %+v, want 12.7W from powermetrics", m)
	}
}

// With neither SMC nor powermetrics, a meaningful IOReport sum is still used as a
// last resort (back-compat for chips that do expose per-domain energy).
func TestAssembleApplePower_IOReportSumLastResort(t *testing.T) {
	got := byName(assembleApplePower(8, 3, 0, -1, true, 0, false, smcCPUReading{}, nil))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 11 {
		t.Errorf("power.total = %+v, want 11W from IOReport sum", m)
	}
}

// No sources at all: no power.total emitted (caller renders Unavailable).
func TestAssembleApplePower_NoSources(t *testing.T) {
	got := byName(assembleApplePower(0, 0, 0, -1, false, 0, false, smcCPUReading{}, nil))
	if _, ok := got["power.total"]; ok {
		t.Errorf("power.total should be absent when no source yields a value")
	}
}
