// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build !darwin || !cgo

package helper

// smcSystemPower is macOS-only; off Apple Silicon or in CGO-free builds it
// reports no value and callers fall back to IOReport / powermetrics.
func smcSystemPower() (watts float64, ok bool) { return 0, false }
