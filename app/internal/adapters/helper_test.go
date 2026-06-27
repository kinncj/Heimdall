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
			return []domain.Metric{{Name: "power.pkg", Status: domain.StatusUnavailable}}
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
	for _, name := range []string{"power.pkg", "gpu.util"} {
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
				{Name: "power.pkg", Unit: "W", Status: domain.StatusOK, Gauge: 12.5},
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
	if got["power.pkg"].Status != domain.StatusOK || got["power.pkg"].Gauge != 12.5 {
		t.Errorf("power.pkg = %+v, want 12.5W ok", got["power.pkg"])
	}
	if got["gpu.util"].Status != domain.StatusOK || got["gpu.util"].Gauge != 40 {
		t.Errorf("gpu.util = %+v, want 40%% ok", got["gpu.util"])
	}
}
