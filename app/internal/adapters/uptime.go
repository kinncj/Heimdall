// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"github.com/shirou/gopsutil/v4/host"

	"heimdall/app/internal/domain"
)

// Uptime reports seconds since boot.
type Uptime struct{}

func (Uptime) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "uptime", Metrics: []string{"host.uptime"}}
}

func (Uptime) Collect(ctx context.Context) ([]domain.Metric, error) {
	up, err := host.UptimeWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return []domain.Metric{{Name: "host.uptime", Unit: "s", Status: domain.StatusOK, Gauge: float64(up)}}, nil
}
