// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build linux

package adapters

import "os"

// coreTypes probes Intel hybrid's sysfs core lists (unprivileged):
// /sys/devices/cpu_core/cpus are the P cores, /sys/devices/cpu_atom/cpus the
// E cores, both as index ranges ("0-15", "16-23"). Hosts without the hybrid
// interface (AMD, older Intel, ARM) are uniform.
func coreTypes(total int) ([]int, bool) {
	perf, errP := os.ReadFile("/sys/devices/cpu_core/cpus")
	eff, errE := os.ReadFile("/sys/devices/cpu_atom/cpus")
	if errP != nil || errE != nil {
		return uniformTypes(total), true
	}
	if types, ok := typesFromRanges(string(perf), string(eff), total); ok {
		return types, true
	}
	// hybrid interface present but inconsistent with the visible core count
	// (offline cores, torn read) — an unlabelled grid beats a mislabelled one.
	return nil, false
}
