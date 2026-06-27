// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build linux

package helper

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"heimdall/app/internal/domain"
)

// linuxPrivileged reads CPU package power from RAPL and package temperature from
// hwmon. Either may be absent (no powercap, no trusted sensor); whatever is found
// is returned and the rest is simply omitted, so an unsupported host degrades to
// the Unavailable fallback rather than failing.
func linuxPrivileged(ctx context.Context) []domain.Metric {
	var out []domain.Metric
	if w, ok := raplPackageWatts(ctx); ok {
		out = append(out, powerMetric("power.pkg", w))
	}
	if c, ok := hwmonPackageTemp(); ok {
		out = append(out, domain.Metric{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: c})
	}
	return out
}

// raplPackageWatts samples a RAPL energy counter twice and derives average power.
func raplPackageWatts(ctx context.Context) (float64, bool) {
	matches, _ := filepath.Glob("/sys/class/powercap/intel-rapl:*/energy_uj")
	if len(matches) == 0 {
		matches, _ = filepath.Glob("/sys/class/powercap/*/energy_uj")
	}
	if len(matches) == 0 {
		return 0, false
	}
	path := matches[0]
	e0, err := readMicrojoules(path)
	if err != nil {
		return 0, false
	}
	const window = 200 * time.Millisecond
	select {
	case <-time.After(window):
	case <-ctx.Done():
		return 0, false
	}
	e1, err := readMicrojoules(path)
	if err != nil {
		return 0, false
	}
	w := raplWatts(e0, e1, window)
	return w, w > 0
}

func readMicrojoules(path string) (uint64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return parseMicrojoules(string(b))
}

// hwmonPackageTemp finds a trusted CPU package sensor and returns its first
// temperature input in degrees Celsius.
func hwmonPackageTemp() (float64, bool) {
	chips, _ := filepath.Glob("/sys/class/hwmon/hwmon*")
	for _, chip := range chips {
		name, err := os.ReadFile(filepath.Join(chip, "name"))
		if err != nil || !isPackageSensor(string(name)) {
			continue
		}
		inputs, _ := filepath.Glob(filepath.Join(chip, "temp*_input"))
		for _, in := range inputs {
			b, err := os.ReadFile(in)
			if err != nil {
				continue
			}
			if c, err := parseMilliCelsius(string(b)); err == nil && c > 0 {
				return c, true
			}
		}
	}
	return 0, false
}
