// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"github.com/shirou/gopsutil/v4/disk"

	"heimdall/app/internal/domain"
)

// Disk reports used percentage of a mount point.
type Disk struct{ Path string }

func (d Disk) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "disk", Metrics: []string{"disk.used"}}
}

func (d Disk) Collect(ctx context.Context) ([]domain.Metric, error) {
	path := d.Path
	if path == "" {
		path = "/"
	}
	u, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return nil, err
	}
	return []domain.Metric{{Name: "disk.used", Unit: "percent", Status: domain.StatusOK, Gauge: u.UsedPercent}}, nil
}
