// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build !darwin || !cgo

package helper

// ioReportPower is unavailable off Apple Silicon or in CGO-free builds; callers
// fall back to powermetrics / nvidia-smi.
func ioReportPower(int) (cpu, gpu, ane, gpuUtil float64, ok bool) { return 0, 0, 0, -1, false }
