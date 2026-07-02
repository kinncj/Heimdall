// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build darwin

package adapters

import "golang.org/x/sys/unix"

// coreTypes probes Apple Silicon's perf-level sysctls (no root, no cgo):
// hw.perflevel0 is the Performance tier, hw.perflevel1 the Efficiency tier.
// The E-cluster owns the low logical core ids on Apple Silicon (cluster 0 is
// the E-cluster in XNU; verified live on an M3 Max — background-QoS load lands
// on c0–c3), so eFirst is true. Intel Macs have no perflevel sysctls and fall
// back to a uniform topology.
func coreTypes(total int) ([]int, bool) {
	p, errP := unix.SysctlUint32("hw.perflevel0.logicalcpu")
	e, errE := unix.SysctlUint32("hw.perflevel1.logicalcpu")
	if errP != nil || errE != nil || e == 0 {
		return uniformTypes(total), true
	}
	return typesFromCounts(int(p), int(e), total, true)
}
