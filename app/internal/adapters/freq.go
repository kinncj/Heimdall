// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"

	"heimdall/app/internal/domain"
)

// Freq reports CPU clock frequency (MHz). It is best-effort: platforms that do
// not expose a clock report Unavailable rather than a fabricated 0. On Linux it
// reads per-core clocks from /sys cpufreq (which works on ARM, where gopsutil
// reports nothing); elsewhere it falls back to gopsutil's per-CPU info.
type Freq struct{}

func (Freq) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "freq", Metrics: []string{"cpu.freq"}}
}

func (Freq) Collect(ctx context.Context) ([]domain.Metric, error) {
	// Prefer /sys cpufreq — it gives real per-core MHz on Linux/ARM where
	// gopsutil's Mhz is 0. Empty off Linux or when the path is absent.
	if khz := readSysCPUFreqKHz(); len(khz) > 0 {
		return []domain.Metric{freqFromKHz(khz)}, nil
	}
	info, err := cpu.InfoWithContext(ctx)
	return []domain.Metric{buildFreq(info, err)}, nil
}

// readSysCPUFreqKHz reads scaling_cur_freq (kHz) for each CPU from sysfs. Returns
// nil when the path doesn't exist (non-Linux, or a kernel without cpufreq).
func readSysCPUFreqKHz() []string {
	paths, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/scaling_cur_freq")
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		out = append(out, strings.TrimSpace(string(b)))
	}
	return out
}

// freqFromKHz builds a cpu.freq metric from sysfs kHz strings: mean MHz in Gauge,
// per-core MHz in PerCore. Unparseable/empty input degrades to Unavailable.
func freqFromKHz(khz []string) domain.Metric {
	per := make([]float64, 0, len(khz))
	sum := 0.0
	for _, s := range khz {
		k, err := strconv.ParseFloat(s, 64)
		if err != nil || k <= 0 {
			continue
		}
		mhz := k / 1000
		per = append(per, mhz)
		sum += mhz
	}
	if len(per) == 0 {
		return domain.Metric{Name: "cpu.freq", Status: domain.StatusUnavailable, Detail: "cpu clock not exposed"}
	}
	return domain.Metric{
		Name: "cpu.freq", Unit: "MHz", Status: domain.StatusOK,
		Gauge: sum / float64(len(per)), PerCore: per,
	}
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
