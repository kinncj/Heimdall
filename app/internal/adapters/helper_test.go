// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"path/filepath"
	"testing"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/helper"
)

func absentClient(t *testing.T) helper.Client {
	return helper.Client{SockPath: filepath.Join(t.TempDir(), "absent.sock")}
}

// With no helper socket and no in-process privilege, the adapter degrades to
// needs-helper rather than erroring.
func TestHelperAdapterNeedsHelperWhenAbsent(t *testing.T) {
	a := Helper{
		Client: absentClient(t),
		Direct: func(context.Context) []domain.Metric {
			return []domain.Metric{{Name: "power.cpu", Status: domain.StatusUnavailable}}
		},
	}
	ms, err := a.Collect(context.Background())
	if err != nil {
		t.Fatalf("collect returned error, want graceful degradation: %v", err)
	}
	got := make(map[string]domain.MetricStatus, len(ms))
	for _, m := range ms {
		got[m.Name] = m.Status
	}
	for _, name := range []string{"power.cpu", "gpu.util"} {
		if got[name] != domain.StatusInsufficientPermission {
			t.Errorf("%s status = %v, want insufficient_permission", name, got[name])
		}
	}
}

// With no helper socket but in-process privilege (e.g. sudo), the adapter
// returns the directly collected metrics instead of needs-helper.
func TestHelperAdapterUsesDirectWhenPrivileged(t *testing.T) {
	a := Helper{
		Client: absentClient(t),
		Direct: func(context.Context) []domain.Metric {
			return []domain.Metric{
				{Name: "power.cpu", Unit: "W", Status: domain.StatusOK, Gauge: 12.5},
				{Name: "gpu.util", Unit: "%", Status: domain.StatusOK, Gauge: 40},
			}
		},
	}
	ms, err := a.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]domain.Metric, len(ms))
	for _, m := range ms {
		got[m.Name] = m
	}
	if got["power.cpu"].Status != domain.StatusOK || got["power.cpu"].Gauge != 12.5 {
		t.Errorf("power.cpu = %+v, want 12.5W ok", got["power.cpu"])
	}
	if got["gpu.util"].Status != domain.StatusOK || got["gpu.util"].Gauge != 40 {
		t.Errorf("gpu.util = %+v, want 40%% ok", got["gpu.util"])
	}
}

// fakeMetricClient is a reachable helper that returns canned metrics (no socket).
type fakeMetricClient struct {
	ms  []domain.Metric
	err error
}

func (f fakeMetricClient) Collect(context.Context) ([]domain.Metric, error) { return f.ms, f.err }

// A reachable helper that returns NO ok metrics must not shadow a working
// in-process source. On Apple Silicon IOReport is unprivileged, so a daemon that
// can read power/gpu itself must not be blanked out by an empty helper reply.
func TestHelperEmptyReplyFallsBackToDirect(t *testing.T) {
	a := Helper{
		Client: fakeMetricClient{ms: []domain.Metric{
			{Name: "power.cpu", Status: domain.StatusInsufficientPermission, Detail: "needs helper"},
			{Name: "gpu.util", Status: domain.StatusInsufficientPermission, Detail: "needs helper"},
		}},
		Direct: func(context.Context) []domain.Metric {
			return []domain.Metric{
				{Name: "power.gpu", Unit: "W", Status: domain.StatusOK, Gauge: 8},
				{Name: "gpu.util", Unit: "%", Status: domain.StatusOK, Gauge: 42},
			}
		},
	}
	ms, _ := a.Collect(context.Background())
	got := make(map[string]domain.Metric, len(ms))
	for _, m := range ms {
		got[m.Name] = m
	}
	if got["gpu.util"].Status != domain.StatusOK || got["gpu.util"].Gauge != 42 {
		t.Errorf("gpu.util = %+v, want in-process 42 (not the empty helper reply)", got["gpu.util"])
	}
	if got["power.gpu"].Status != domain.StatusOK {
		t.Errorf("power.gpu should come from Direct, got %+v", got["power.gpu"])
	}
}

// A reachable helper WITH ok metrics is authoritative — Direct must not run.
func TestHelperOKReplyIsUsedWithoutDirect(t *testing.T) {
	directRan := false
	a := Helper{
		Client: fakeMetricClient{ms: []domain.Metric{
			{Name: "power.cpu", Unit: "W", Status: domain.StatusOK, Gauge: 15},
		}},
		Direct: func(context.Context) []domain.Metric { directRan = true; return nil },
	}
	ms, _ := a.Collect(context.Background())
	if directRan {
		t.Fatal("Direct must not run when the helper already has ok metrics")
	}
	if len(ms) != 1 || ms[0].Name != "power.cpu" || ms[0].Status != domain.StatusOK {
		t.Fatalf("want helper power.cpu ok, got %+v", ms)
	}
}
