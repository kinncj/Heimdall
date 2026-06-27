// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"github.com/shirou/gopsutil/v4/mem"

	"heimdall/app/internal/domain"
)

// Mem reports virtual memory used percentage.
type Mem struct{}

func (Mem) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "mem", Metrics: []string{"mem.used"}}
}

func (Mem) Collect(ctx context.Context) ([]domain.Metric, error) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return []domain.Metric{{Name: "mem.used", Unit: "percent", Status: domain.StatusOK, Gauge: vm.UsedPercent}}, nil
}
