// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"strconv"
	"strings"
	"time"
)

// parseMicrojoules parses a Linux RAPL energy_uj counter value (microjoules).
func parseMicrojoules(s string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(s), 10, 64)
}

// raplWatts derives average power from two RAPL energy readings (microjoules)
// taken dt apart. A counter that went backwards (wrap-around) or a non-positive
// interval yields 0 — the caller treats that as "no reading" rather than a spike.
func raplWatts(prev, cur uint64, dt time.Duration) float64 {
	if dt <= 0 || cur < prev {
		return 0
	}
	joules := float64(cur-prev) / 1e6
	return joules / dt.Seconds()
}

// parseMilliCelsius parses a hwmon tempN_input value (thousandths of a degree)
// into degrees Celsius.
func parseMilliCelsius(s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	return v / 1000.0, nil
}

// isPackageSensor reports whether a hwmon chip name is a CPU package thermal
// source we trust for temp.pkg (Intel coretemp, AMD k10temp/zenpower).
func isPackageSensor(name string) bool {
	switch strings.TrimSpace(name) {
	case "coretemp", "k10temp", "zenpower":
		return true
	}
	return false
}
