// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"testing"

	"heimdall/app/internal/domain"
)

func TestCPUReportsRealUtil(t *testing.T) {
	m := first(t, CPU{})
	if m.Name != "cpu.util" || m.Status != domain.StatusOK {
		t.Fatalf("got %+v", m)
	}
	if m.Gauge < 0 || m.Gauge > 100 {
		t.Errorf("cpu.util %.1f out of [0,100]", m.Gauge)
	}
}

func TestCPUPerCore(t *testing.T) {
	ms, err := CPU{}.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var cores domain.Metric
	for _, m := range ms {
		if m.Name == "cpu.cores" {
			cores = m
		}
	}
	if cores.Kind != domain.KindPerCore || len(cores.PerCore) == 0 {
		t.Fatalf("cpu.cores = %+v, want per-core values", cores)
	}
}
