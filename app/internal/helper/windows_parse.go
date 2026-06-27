// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"strconv"
	"strings"
)

// parseThermalZoneCelsius parses WMI MSAcpi_ThermalZoneTemperature
// CurrentTemperature values — tenths of a Kelvin, one per thermal zone — and
// returns the hottest as degrees Celsius. Implausible readings (outside 0–150 °C)
// are ignored; ok is false when nothing usable is found.
func parseThermalZoneCelsius(out string) (celsius float64, ok bool) {
	for _, f := range strings.Fields(out) {
		n, err := strconv.Atoi(f)
		if err != nil {
			continue
		}
		c := float64(n)/10.0 - 273.15
		if c < 0 || c > 150 {
			continue
		}
		if !ok || c > celsius {
			celsius, ok = c, true
		}
	}
	return celsius, ok
}
