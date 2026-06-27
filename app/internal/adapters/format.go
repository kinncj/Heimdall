// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import "fmt"

// humanBytes formats a byte count as a compact binary-scaled string (KB, MB, GB,
// TB, …), e.g. 137438953472 -> "128 GB". It is used to annotate percentage
// metrics with their absolute used/total figures.
func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	exp := 0
	for value >= unit && exp < 4 {
		value /= unit
		exp++
	}
	suffix := []string{"KB", "MB", "GB", "TB", "PB"}[exp-1]
	if value >= 100 {
		return fmt.Sprintf("%.0f %s", value, suffix)
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}

// usedTotal renders an absolute "used / total" annotation for a percentage gauge.
func usedTotal(used, total uint64) string {
	return humanBytes(used) + " / " + humanBytes(total)
}
