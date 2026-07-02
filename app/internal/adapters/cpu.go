// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"github.com/shirou/gopsutil/v4/cpu"

	"heimdall/app/internal/domain"
)

// CPU reports total CPU utilisation plus per-core utilisation.
type CPU struct{}

func (CPU) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "cpu", Metrics: []string{"cpu.util", "cpu.cores", "cpu.topology"}}
}

func (CPU) Collect(ctx context.Context) ([]domain.Metric, error) {
	per, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return nil, err
	}
	avg := 0.0
	for _, v := range per {
		avg += v
	}
	if len(per) > 0 {
		avg /= float64(len(per))
	}
	out := []domain.Metric{
		{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: avg},
	}
	if len(per) > 0 {
		out = append(out, domain.Metric{
			Name: "cpu.cores", Unit: "percent", Status: domain.StatusOK,
			Kind: domain.KindPerCore, Gauge: avg, PerCore: per,
		})
		// Core-type layout (P/E on Apple Silicon and Intel hybrid; uniform
		// elsewhere) so the top view can label the per-core grid. Omitted when
		// no probe answers — the renderer falls back to the unlabelled grid.
		if topo, ok := coreTopologyMetric(len(per)); ok {
			out = append(out, topo)
		}
	}
	return out, nil
}
