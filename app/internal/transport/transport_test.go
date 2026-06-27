// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package transport

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func TestSnapshotRoundTrip(t *testing.T) {
	ts := time.Unix(1_000_000, 0)
	in := []domain.Metric{
		{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: 42.5},
		{Name: "gpu.util", Status: domain.StatusUnavailable, Detail: "no GPU"},
		{Name: "power.pkg", Status: domain.StatusInsufficientPermission, Detail: "needs helper"},
		{Name: "temp.pkg", Status: domain.StatusError, Detail: "sensor offline"},
	}

	snap := ToSnapshot("dgx-spark", in, 7, ts)
	if snap.GetHostId() != "dgx-spark" || snap.GetSeq() != 7 || snap.GetTsUnixMillis() != ts.UnixMilli() {
		t.Fatalf("snapshot header wrong: %+v", snap)
	}

	host, out := FromSnapshot(snap)
	if host != "dgx-spark" {
		t.Errorf("host = %q", host)
	}
	if len(out) != len(in) {
		t.Fatalf("got %d metrics, want %d", len(out), len(in))
	}
	for i, m := range out {
		if m.Name != in[i].Name || m.Status != in[i].Status || m.Detail != in[i].Detail {
			t.Errorf("metric %d = %+v, want %+v", i, m, in[i])
		}
	}
	if out[0].Gauge != 42.5 {
		t.Errorf("cpu.util gauge = %v, want 42.5", out[0].Gauge)
	}
	// non-OK metrics carry no value
	if out[1].Gauge != 0 {
		t.Errorf("unavailable metric should carry no value, got %v", out[1].Gauge)
	}
}

func TestPerCoreRoundTrip(t *testing.T) {
	in := domain.Metric{
		Name: "cpu.cores", Unit: "percent", Status: domain.StatusOK,
		Kind: domain.KindPerCore, PerCore: []float64{10, 20, 30, 40},
	}
	got := MetricFromProto(MetricToProto(in))
	if got.Kind != domain.KindPerCore {
		t.Fatalf("kind = %v, want per-core", got.Kind)
	}
	if len(got.PerCore) != 4 || got.PerCore[2] != 30 {
		t.Fatalf("per-core round trip = %+v", got.PerCore)
	}
}
