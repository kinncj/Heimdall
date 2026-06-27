// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"github.com/shirou/gopsutil/v4/sensors"

	"heimdall/app/internal/domain"
)

// Temperature reports the hottest reported sensor. Many platforms (notably
// macOS) expose no sensors without elevation — reported as unavailable / needs
// the privileged helper, never an error.
type Temperature struct{}

func (Temperature) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "temp", Metrics: []string{"temp.pkg"}}
}

func (Temperature) Collect(ctx context.Context) ([]domain.Metric, error) {
	temps, err := sensors.TemperaturesWithContext(ctx)
	if err != nil || len(temps) == 0 {
		return []domain.Metric{{Name: "temp.pkg", Status: domain.StatusInsufficientPermission, Detail: "needs helper"}}, nil
	}
	var max float64
	for _, s := range temps {
		if s.Temperature > max {
			max = s.Temperature
		}
	}
	return []domain.Metric{{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: max}}, nil
}
