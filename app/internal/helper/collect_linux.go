// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build linux

package helper

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"heimdall/app/internal/domain"
)

// linuxPrivileged reads CPU power/temperature from RAPL and hwmon. power.pkg is
// the whole CPU package (cores + uncore) and power.cpu is the cores alone — both
// are the CPU socket only; a discrete GPU is a separate rail (power.gpu), which
// is why on a workstation power.pkg can read far below power.gpu. Any source may
// be absent (no powercap, no core subdomain, no trusted sensor); whatever is
// found is returned and the rest is omitted, so an unsupported host degrades to
// the Unavailable fallback rather than failing.
func linuxPrivileged(ctx context.Context) []domain.Metric {
	var out []domain.Metric
	if w, ok := raplPackageWatts(ctx); ok {
		out = append(out, domain.Metric{Name: "power.pkg", Unit: "watts", Status: domain.StatusOK, Gauge: w, Detail: "CPU package"})
	}
	if w, ok := raplCoreWatts(ctx); ok {
		out = append(out, domain.Metric{Name: "power.cpu", Unit: "watts", Status: domain.StatusOK, Gauge: w, Detail: "CPU cores"})
	}
	if c, ok := hwmonPackageTemp(); ok {
		out = append(out, domain.Metric{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: c})
	}
	return out
}

// isCoreDomain reports whether a RAPL domain name is the CPU-core (pp0) subdomain
// that maps to power.cpu — as opposed to the package, uncore, dram, or psys.
func isCoreDomain(name string) bool {
	return strings.TrimSpace(name) == "core"
}

// raplCoreWatts samples the RAPL "core" subdomain to report CPU-core power as
// power.cpu, distinct from the package power.pkg. It is absent on CPUs that do
// not expose the core subdomain, in which case power.cpu simply stays unset.
func raplCoreWatts(ctx context.Context) (float64, bool) {
	path, ok := raplCoreEnergyPath()
	if !ok {
		return 0, false
	}
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

// raplCoreEnergyPath locates the energy counter of the RAPL "core" subdomain,
// matching by the domain's name file rather than by index.
func raplCoreEnergyPath() (string, bool) {
	for _, d := range raplDomainDirs() {
		name, err := os.ReadFile(filepath.Join(d, "name"))
		if err != nil || !isCoreDomain(string(name)) {
			continue
		}
		ep := filepath.Join(d, "energy_uj")
		if _, err := os.Stat(ep); err == nil {
			return ep, true
		}
	}
	return "", false
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
