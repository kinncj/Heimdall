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

// linuxPrivileged reads CPU power from RAPL and package temperature from hwmon.
// power.cpu is the RAPL package (the whole CPU socket — cores + uncore), which is
// what btop/top report as "CPU power". The GPU is a separate rail (power.gpu) and
// the two are summed into power.total upstream. When the SoC exposes no RAPL at
// all (e.g. GB10 Grace/ARM), power.cpu is reported Unavailable-with-reason rather
// than left blank; a trusted temp sensor is added when present.
func linuxPrivileged(ctx context.Context) []domain.Metric {
	var out []domain.Metric
	if w, ok := raplPackageWatts(ctx); ok {
		out = append(out, domain.Metric{Name: "power.cpu", Unit: "watts", Status: domain.StatusOK, Gauge: w, Detail: "CPU package"})
	} else if len(raplDomainDirs()) == 0 {
		// No powercap tree at all — the SoC exposes no RAPL (e.g. GB10 Grace/ARM).
		out = append(out, domain.Metric{Name: "power.cpu", Status: domain.StatusUnavailable, Detail: "no RAPL power sensor (SoC/ARM)"})
	}
	if c, ok := hwmonPackageTemp(); ok {
		out = append(out, domain.Metric{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: c})
	}
	return out
}

// raplDomainDirs returns the RAPL domain directories, preferring the intel-rapl
// tree (present on both Intel and most AMD kernels) and falling back to the
// generic powercap tree.
func raplDomainDirs() []string {
	dirs, _ := filepath.Glob("/sys/class/powercap/intel-rapl:*")
	if len(dirs) == 0 {
		dirs, _ = filepath.Glob("/sys/class/powercap/*")
	}
	return dirs
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
