// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import (
	"testing"
	"time"
)

func TestCanonicalMetricNameAliasesAneToNpu(t *testing.T) {
	if got := canonicalMetricName("power.ane"); got != "power.npu" {
		t.Errorf("power.ane should canonicalise to power.npu, got %q", got)
	}
	// Non-aliased names pass through unchanged.
	for _, n := range []string{"power.npu", "power.cpu", "gpu.util", "cpu.load"} {
		if got := canonicalMetricName(n); got != n {
			t.Errorf("%q should pass through unchanged, got %q", n, got)
		}
	}
}

// An older daemon emits the legacy power.ane key; a v2.2 registry must store it
// under the canonical power.npu so the UI renders it under the NPU label.
func TestObserveNormalizesLegacyAneToNpu(t *testing.T) {
	r := NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	r.Enroll(Host{ID: "h", DisplayName: "h"}, now)
	r.Observe("h", []Metric{{Name: "power.ane", Status: StatusOK, Gauge: 3.5}}, nil, now)

	v, ok := r.Host("h")
	if !ok || len(v.LastSnapshot) != 1 {
		t.Fatalf("expected one stored metric, got %+v", v.LastSnapshot)
	}
	if v.LastSnapshot[0].Name != "power.npu" {
		t.Errorf("legacy power.ane should be stored as power.npu, got %q", v.LastSnapshot[0].Name)
	}
}
