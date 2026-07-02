// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build !darwin && !linux && !windows

package adapters

// coreTypes has no hybrid probe on this platform: uniform topology.
func coreTypes(total int) ([]int, bool) {
	return uniformTypes(total), true
}
