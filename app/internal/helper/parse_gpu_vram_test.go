// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

// On unified-memory NVIDIA (GB10 Grace-Blackwell), nvidia-smi reports
// memory.used/total as [N/A], so gpu.vram must be derived from the sum of
// per-process compute-apps memory over the system RAM total.
func TestNvidiaVRAMFromComputeApps(t *testing.T) {
	// One compute app (ollama) holding 42625 MiB; host has 124546 MiB of RAM.
	m, ok := nvidiaVRAMFromComputeApps("42625\n", 124546)
	if !ok {
		t.Fatal("expected a gpu.vram reading from compute-apps")
	}
	if m.Name != "gpu.vram" || m.Unit != "percent" || m.Status != domain.StatusOK {
		t.Fatalf("metric = %+v, want ok gpu.vram in percent", m)
	}
	if m.Gauge < 34 || m.Gauge > 35 {
		t.Errorf("gpu.vram = %.2f%%, want ~34%%", m.Gauge)
	}
	if m.Detail != "42/122 GB shared" {
		t.Errorf("detail = %q, want \"42/122 GB shared\"", m.Detail)
	}
}

// Multiple compute apps sum; extra CSV columns (pid,used_memory) are tolerated.
func TestNvidiaVRAMFromComputeApps_SumsAndToleratesColumns(t *testing.T) {
	m, ok := nvidiaVRAMFromComputeApps("40000\n2625\n", 124546)
	if !ok || m.Gauge < 34 || m.Gauge > 35 {
		t.Fatalf("summed gpu.vram = %+v, want ~34%%", m)
	}
}

// Without a system RAM total there is no denominator, so no reading.
func TestNvidiaVRAMFromComputeApps_NoTotalMeansNoReading(t *testing.T) {
	if _, ok := nvidiaVRAMFromComputeApps("42625\n", 0); ok {
		t.Error("no system total should mean no gpu.vram")
	}
}

// A successful-but-empty query = idle GPU = a stable 0%, not a vanished metric.
func TestNvidiaVRAMFromComputeApps_IdleReportsZero(t *testing.T) {
	for _, in := range []string{"\n", ""} {
		m, ok := nvidiaVRAMFromComputeApps(in, 124546)
		if !ok || m.Status != domain.StatusOK || m.Gauge != 0 {
			t.Fatalf("idle gpu.vram for %q = %+v, want ok 0%%", in, m)
		}
	}
}

// When nvidia-smi is present but exits with an error (e.g. a driver/library
// version mismatch after a driver upgrade without a reboot), the GPU must not
// silently vanish — gpu.util is reported Unavailable carrying the reason.
func TestNvidiaErrorMetrics_SurfacesReason(t *testing.T) {
	got := byName(nvidiaErrorMetrics("Failed to initialize NVML: Driver/library version mismatch\nNVML library version: 580.159"))
	m := got["gpu.util"]
	if m.Status != domain.StatusUnavailable {
		t.Fatalf("gpu.util status = %v, want unavailable", m.Status)
	}
	if !strings.Contains(strings.ToLower(m.Detail), "driver/library version mismatch") {
		t.Fatalf("detail = %q, want the nvidia-smi reason", m.Detail)
	}
}

// NPUs (Apple ANE, Intel AI Boost, AMD XDNA) expose no utilisation counter — so
// npu.util must be reported Unavailable-with-reason, not left as a bare dash. An
// existing reading (e.g. the AMD path's own reason) must be preserved.
func TestEnsureNPUUtil(t *testing.T) {
	got := byName(ensureNPUUtil(nil))
	if m := got["npu.util"]; m.Status != domain.StatusUnavailable || m.Detail == "" {
		t.Fatalf("npu.util = %+v, want unavailable with a reason", m)
	}
	amd := []domain.Metric{{Name: "npu.util", Status: domain.StatusUnavailable, Detail: "amdxdna exposes no util counter"}}
	if d := byName(ensureNPUUtil(amd))["npu.util"].Detail; d != "amdxdna exposes no util counter" {
		t.Errorf("existing npu.util reason must be preserved, got %q", d)
	}
}

// Apple Silicon has unified memory and no discrete VRAM to report; gpu.vram must
// be Unavailable with a reason rather than absent, so the panel explains the dash.
func TestAssembleApplePower_VRAMUnavailableWithReason(t *testing.T) {
	got := byName(assembleApplePower(5, 2, 0, 30, true, 22, true, nil))
	m, ok := got["gpu.vram"]
	if !ok || m.Status != domain.StatusUnavailable {
		t.Fatalf("gpu.vram = %+v, want unavailable", m)
	}
	if !strings.Contains(m.Detail, "unified memory") {
		t.Errorf("detail = %q, want a unified-memory note", m.Detail)
	}
	// The real Apple power/util readings must still be present.
	if got["power.gpu"].Status != domain.StatusOK || got["gpu.util"].Status != domain.StatusOK {
		t.Error("apple power.gpu / gpu.util must remain OK alongside the vram note")
	}
}
