// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package cli

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func cliSeed(t *testing.T) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "web-01", DisplayName: "web-01",
		Context: domain.HostContext{Labels: map[string]string{"env": "prod", "_logs": "app,sys", "_proc": "1"}}}, now)
	reg.Observe("web-01", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 42}}, nil, now)
	reg.RecordPush("web-01",
		[]domain.ProcessRow{{PID: 1, Command: "init"}, {PID: 9, CPUPct: 80, Command: "node"}},
		now, []domain.LogLine{{Source: "app", At: now, Line: "up"}, {Source: "sys", At: now, Line: "boot"}})
	reg.Evaluate(now)
	return reg
}

func TestCLIHostShapeStripsReservedLabels(t *testing.T) {
	h, _ := cliSeed(t).Host("web-01")
	j := newJHost(h)
	if j.ID != "web-01" || !j.HasLogs || !j.HasProcesses {
		t.Fatalf("capabilities wrong: %+v", j)
	}
	if _, leaked := j.Labels["_logs"]; leaked {
		t.Fatal("reserved label leaked into CLI output")
	}
	if j.Labels["env"] != "prod" {
		t.Fatalf("user label missing: %+v", j.Labels)
	}
	if j.Metrics["cpu.util"] != 42 {
		t.Fatalf("metric missing: %+v", j.Metrics)
	}
	if len(j.LogSources) != 2 {
		t.Fatalf("log sources = %v", j.LogSources)
	}
}

func TestCLISurfacesUnavailableMetrics(t *testing.T) {
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "mac-01", DisplayName: "mac-01"}, now)
	reg.Observe("mac-01", []domain.Metric{
		{Name: "gpu.util", Status: domain.StatusOK, Gauge: 18},
		{Name: "gpu.vram", Status: domain.StatusUnavailable, Detail: "unified memory (no discrete VRAM)"},
	}, nil, now)
	reg.Evaluate(now)
	h, _ := reg.Host("mac-01")
	j := newJHost(h)

	if _, ok := j.Metrics["gpu.vram"]; ok {
		t.Error("an unavailable metric must not leak into .metrics")
	}
	if j.Metrics["gpu.util"] != 18 {
		t.Errorf("ok metric still expected in .metrics: %+v", j.Metrics)
	}
	u, ok := j.Unavailable["gpu.vram"]
	if !ok || u.Status != "unavailable" || u.Detail != "unified memory (no discrete VRAM)" {
		t.Fatalf("gpu.vram unavailable entry = %+v (present=%v)", u, ok)
	}
}

func TestCLISurfacesOKMetricDetails(t *testing.T) {
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "gb10", DisplayName: "gb10"}, now)
	reg.Observe("gb10", []domain.Metric{
		{Name: "host.version", Unit: "info", Status: domain.StatusOK, Detail: "v2.2.8"},
		{Name: "host.gpu", Unit: "info", Status: domain.StatusOK, Detail: "NVIDIA GB10"},
		{Name: "gpu.vram", Status: domain.StatusOK, Gauge: 34.2, Detail: "43/122 GB shared"},
		{Name: "gpu.util", Status: domain.StatusOK, Gauge: 95}, // no detail
	}, nil, now)
	reg.Evaluate(now)
	h, _ := reg.Host("gb10")
	j := newJHost(h)

	if j.Details["host.version"] != "v2.2.8" || j.Details["host.gpu"] != "NVIDIA GB10" {
		t.Errorf("info-metric details missing: %+v", j.Details)
	}
	if j.Details["gpu.vram"] != "43/122 GB shared" {
		t.Errorf("gpu.vram detail missing: %+v", j.Details)
	}
	if _, ok := j.Details["gpu.util"]; ok {
		t.Error("a metric with no detail must not appear in details")
	}
	// The value map is unchanged (backward compatible).
	if j.Metrics["gpu.vram"] != 34.2 || j.Metrics["gpu.util"] != 95 {
		t.Errorf("metrics values changed: %+v", j.Metrics)
	}
}

// The CLI surfaces the standardized P/E core grouping (domain.CoreGroups), so an
// agent reads structured groups — not just the "12P + 4E" detail string.
func TestCLISurfacesCoreGroups(t *testing.T) {
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "hybrid", DisplayName: "hybrid"}, now)
	reg.Observe("hybrid", []domain.Metric{
		{Name: "cpu.cores", Status: domain.StatusOK, Kind: domain.KindPerCore,
			PerCore: []float64{5, 6, 80, 70}},
		{Name: "cpu.topology", Status: domain.StatusOK, Kind: domain.KindPerCore,
			PerCore: []float64{domain.CoreEff, domain.CoreEff, domain.CorePerf, domain.CorePerf},
			Detail:  "2P + 2E"},
	}, nil, now)
	reg.Evaluate(now)
	h, _ := reg.Host("hybrid")
	j := newJHost(h)

	if len(j.CoreGroups) != 2 {
		t.Fatalf("core_groups = %+v, want 2 groups", j.CoreGroups)
	}
	if j.CoreGroups[0].Type != "E" || len(j.CoreGroups[0].Cores) != 2 {
		t.Errorf("group 0 = %+v, want E with 2 cores (logical order)", j.CoreGroups[0])
	}
	if j.CoreGroups[1].Type != "P" {
		t.Errorf("group 1 = %+v, want P", j.CoreGroups[1])
	}
}

// A uniform CPU emits no core_groups — nothing to group.
func TestCLIUniformCPUHasNoCoreGroups(t *testing.T) {
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "amd", DisplayName: "amd"}, now)
	reg.Observe("amd", []domain.Metric{
		{Name: "cpu.cores", Status: domain.StatusOK, Kind: domain.KindPerCore,
			PerCore: []float64{10, 20, 30, 40}},
		{Name: "cpu.topology", Status: domain.StatusOK, Kind: domain.KindPerCore,
			PerCore: []float64{0, 0, 0, 0}, Detail: "4 cores (uniform)"},
	}, nil, now)
	reg.Evaluate(now)
	h, _ := reg.Host("amd")
	if j := newJHost(h); j.CoreGroups != nil {
		t.Errorf("uniform CPU must emit no core_groups, got %+v", j.CoreGroups)
	}
}

func TestCLITopAndLogs(t *testing.T) {
	h, _ := cliSeed(t).Host("web-01")
	top := newJTop(h)
	if len(top.Processes) != 2 || top.Processes[1].Command != "node" {
		t.Fatalf("top = %+v", top.Processes)
	}
	all := newJLogs(h, "")
	if len(all.Lines) != 2 {
		t.Fatalf("all logs = %d", len(all.Lines))
	}
	appOnly := newJLogs(h, "app")
	if len(appOnly.Lines) != 1 || appOnly.Lines[0].Line != "up" {
		t.Fatalf("source-filtered logs = %+v", appOnly.Lines)
	}
}

func TestCLIFleetSummary(t *testing.T) {
	s := fleetSummary(cliSeed(t))
	if s["total"].(int) != 1 {
		t.Fatalf("total = %v", s["total"])
	}
	if by := s["by_state"].(map[string]int); by["online"] != 1 {
		t.Fatalf("by_state = %v", by)
	}
}
