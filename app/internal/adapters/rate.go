// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

// deltaRateMB converts the growth of a monotonic byte counter over dt seconds
// into MB/s. A non-positive interval or a counter reset (cur < prev) yields 0,
// so a wrapped or restarted counter never produces a bogus spike. Shared by the
// Network and DiskIO adapters.
func deltaRateMB(prev, cur uint64, dt float64) float64 {
	if dt <= 0 || cur < prev {
		return 0
	}
	return float64(cur-prev) / dt / 1e6
}
