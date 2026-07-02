// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build windows

package adapters

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetLogicalProcInfoEx = kernel32.NewProc("GetLogicalProcessorInformationEx")
)

// coreTypes probes Windows hybrid topology (Intel P/E, no privileges needed):
// GetLogicalProcessorInformationEx(RelationProcessorCore) yields one record per
// physical core carrying its EfficiencyClass and logical-CPU affinity mask.
// Multi-group machines (>64 logical CPUs) are refused — they don't ship hybrid
// silicon, and an unlabelled grid beats a mislabelled one.
func coreTypes(total int) ([]int, bool) {
	buf, ok := logicalProcInfoEx()
	if !ok {
		return uniformTypes(total), true
	}
	classes, ok := efficiencyClassesFromBuffer(buf, total)
	if !ok {
		return nil, false
	}
	return typesFromEfficiencyClasses(classes)
}

func logicalProcInfoEx() ([]byte, bool) {
	var size uint32
	r, _, _ := procGetLogicalProcInfoEx.Call(relationProcessorCore, 0, uintptr(unsafe.Pointer(&size)))
	if r != 0 || size == 0 {
		return nil, false
	}
	buf := make([]byte, size)
	r, _, _ = procGetLogicalProcInfoEx.Call(relationProcessorCore,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if r == 0 {
		return nil, false
	}
	return buf[:size], true
}
