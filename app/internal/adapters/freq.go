// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"github.com/shirou/gopsutil/v4/cpu"

	"heimdall/app/internal/domain"
)

// Freq reports CPU clock frequency (MHz). It is best-effort: platforms that do
// not expose a clock through gopsutil (notably Apple Silicon) report Unavailable
// rather than a fabricated 0.
type Freq struct{}

func (Freq) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "freq", Metrics: []string{"cpu.freq"}}
}

func (Freq) Collect(ctx context.Context) ([]domain.Metric, error) {
	info, err := cpu.InfoWithContext(ctx)
	return []domain.Metric{buildFreq(info, err)}, nil
}

// buildFreq derives a cpu.freq metric from gopsutil's per-CPU info. Gauge is the
// mean MHz across entries; PerCore carries each entry's MHz. When no entry
// reports a positive clock it degrades to Unavailable.
func buildFreq(info []cpu.InfoStat, err error) domain.Metric {
	if err != nil || len(info) == 0 {
		return domain.Metric{Name: "cpu.freq", Status: domain.StatusUnavailable, Detail: "no cpu clock source"}
	}
	per := make([]float64, 0, len(info))
	sum := 0.0
	for _, c := range info {
		if c.Mhz > 0 {
			per = append(per, c.Mhz)
			sum += c.Mhz
		}
	}
	if len(per) == 0 {
		return domain.Metric{Name: "cpu.freq", Status: domain.StatusUnavailable, Detail: "cpu clock not exposed"}
	}
	return domain.Metric{
		Name: "cpu.freq", Unit: "MHz", Status: domain.StatusOK,
		Gauge: sum / float64(len(per)), PerCore: per,
	}
}
