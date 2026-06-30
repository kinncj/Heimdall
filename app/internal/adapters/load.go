// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v4/load"

	"heimdall/app/internal/domain"
)

// Load reports the 1/5/15-minute load average. Windows has no real load average,
// so the source errors there and the metric degrades to Unavailable rather than
// reporting a fabricated 0.
type Load struct{}

func (Load) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "load", Metrics: []string{"cpu.load"}}
}

func (Load) Collect(ctx context.Context) ([]domain.Metric, error) {
	avg, err := load.AvgWithContext(ctx)
	return []domain.Metric{buildLoad(avg, err)}, nil
}

// buildLoad turns a gopsutil load reading into a metric. Gauge carries the
// 1-minute figure; PerCore carries [1m,5m,15m] so the top view can show all
// three. On error (e.g. Windows) it returns Unavailable.
func buildLoad(avg *load.AvgStat, err error) domain.Metric {
	if err != nil || avg == nil {
		return domain.Metric{Name: "cpu.load", Status: domain.StatusUnavailable, Detail: "no load average on this platform"}
	}
	return domain.Metric{
		Name: "cpu.load", Status: domain.StatusOK,
		Gauge:   avg.Load1,
		PerCore: []float64{avg.Load1, avg.Load5, avg.Load15},
		Detail:  fmt.Sprintf("%.2f %.2f %.2f (1/5/15m)", avg.Load1, avg.Load5, avg.Load15),
	}
}
