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
