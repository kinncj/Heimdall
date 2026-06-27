// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package observe

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func view(id string, gauge float64) domain.HostView {
	return domain.HostView{
		Host:         domain.Host{ID: domain.HostID(id)},
		State:        domain.StateOnline,
		LastSnapshot: []domain.Metric{{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: gauge}},
	}
}

func TestHistoryRetainsRecentSamples(t *testing.T) {
	h := NewHistory(3)
	base := time.Unix(1_700_000_000, 0)
	h.Record([]domain.HostView{view("a", 10)}, base)
	h.Record([]domain.HostView{view("a", 20)}, base.Add(time.Second))
	s := h.Series("a", "cpu.util")
	if len(s) != 2 || s[0].Value != 10 || s[1].Value != 20 {
		t.Fatalf("series = %+v, want two samples 10 then 20", s)
	}
}

func TestHistoryIsBounded(t *testing.T) {
	h := NewHistory(2)
	base := time.Unix(1_700_000_000, 0)
	for i, v := range []float64{1, 2, 3} {
		h.Record([]domain.HostView{view("a", v)}, base.Add(time.Duration(i)*time.Second))
	}
	s := h.Series("a", "cpu.util")
	if len(s) != 2 || s[0].Value != 2 || s[1].Value != 3 {
		t.Fatalf("series = %+v, want last two samples 2 then 3 (oldest dropped)", s)
	}
}
