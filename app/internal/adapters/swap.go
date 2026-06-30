// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"github.com/shirou/gopsutil/v4/mem"

	"heimdall/app/internal/domain"
)

// Swap reports swap usage as a percentage. Hosts without swap report 0%.
type Swap struct{}

func (Swap) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "swap", Metrics: []string{"mem.swap"}}
}

func (Swap) Collect(ctx context.Context) ([]domain.Metric, error) {
	sm, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return []domain.Metric{{Name: "mem.swap", Status: domain.StatusUnavailable, Detail: "no swap source"}}, nil
	}
	return []domain.Metric{{
		Name: "mem.swap", Unit: "percent", Status: domain.StatusOK,
		Gauge: sm.UsedPercent, Detail: usedTotal(sm.Used, sm.Total),
	}}, nil
}
