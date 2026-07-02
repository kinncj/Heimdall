// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"testing"

	"heimdall/app/internal/domain"
)

// On Apple Pro/Max the SMC per-cluster rails are read anyway to build power.cpu;
// the P/E split must be surfaced as its own metrics instead of being thrown away.
func TestAssembleApplePower_EmitsClusterSplit(t *testing.T) {
	smc := smcCPUReading{Total: 18.1, P: 15.0, E: 3.1, OK: true, POK: true, EOK: true}
	got := byName(assembleApplePower(0 /*ioreport cpu 0*/, 0.1, 0, 5, true, 42, true, smc, nil))
	if m := got["power.cpu"]; m.Status != domain.StatusOK || m.Gauge != 18.1 {
		t.Fatalf("power.cpu = %+v, want 18.1W total", m)
	}
	if m := got["power.cpu.pcluster"]; m.Status != domain.StatusOK || m.Gauge != 15.0 {
		t.Errorf("power.cpu.pcluster = %+v, want 15.0W", m)
	}
	if m := got["power.cpu.ecluster"]; m.Status != domain.StatusOK || m.Gauge != 3.1 {
		t.Errorf("power.cpu.ecluster = %+v, want 3.1W", m)
	}
}

// A cluster rail that did not read must not fabricate a 0 — only the rails that
// answered are emitted.
func TestAssembleApplePower_PartialClusterSplit(t *testing.T) {
	smc := smcCPUReading{Total: 15.0, P: 15.0, OK: true, POK: true}
	got := byName(assembleApplePower(0, 0.1, 0, 5, true, 42, true, smc, nil))
	if m := got["power.cpu.pcluster"]; m.Status != domain.StatusOK || m.Gauge != 15.0 {
		t.Errorf("power.cpu.pcluster = %+v, want 15.0W", m)
	}
	if _, present := got["power.cpu.ecluster"]; present {
		t.Error("power.cpu.ecluster must be absent when the E rail did not read")
	}
}

// Base dies get power.cpu from IOReport and have no mapped cluster keys — no
// cluster metrics may appear.
func TestAssembleApplePower_NoClusterMetricsWithoutSMCKeys(t *testing.T) {
	got := byName(assembleApplePower(0.5 /*ioreport cpu*/, 0.1, 0, 5, true, 20, true, smcCPUReading{}, nil))
	if _, present := got["power.cpu.pcluster"]; present {
		t.Error("power.cpu.pcluster must be absent without SMC cluster keys")
	}
	if _, present := got["power.cpu.ecluster"]; present {
		t.Error("power.cpu.ecluster must be absent without SMC cluster keys")
	}
}

// The cluster split is real even when IOReport already supplied power.cpu (both
// sources healthy): the split decomposes the CPU complex either way.
func TestAssembleApplePower_ClusterSplitAlongsideIOReportCPU(t *testing.T) {
	smc := smcCPUReading{Total: 18.1, P: 15.0, E: 3.1, OK: true, POK: true, EOK: true}
	got := byName(assembleApplePower(17.9 /*ioreport cpu*/, 0.1, 0, 5, true, 42, true, smc, nil))
	if m := got["power.cpu"]; m.Status != domain.StatusOK || m.Gauge != 17.9 {
		t.Fatalf("power.cpu = %+v, want the IOReport 17.9W", m)
	}
	if m := got["power.cpu.pcluster"]; m.Status != domain.StatusOK || m.Gauge != 15.0 {
		t.Errorf("power.cpu.pcluster = %+v, want 15.0W", m)
	}
}
