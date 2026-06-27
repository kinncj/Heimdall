// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import "testing"

func TestParseThermalZoneCelsius(t *testing.T) {
	// tenths of a Kelvin: 3232 -> 50.05 °C (the hotter of two zones)
	if c, ok := parseThermalZoneCelsius("3012\n3232\n"); !ok || c < 49 || c > 51 {
		t.Fatalf("c=%v ok=%v, want ~50", c, ok)
	}
	// non-numeric output -> not ok
	if _, ok := parseThermalZoneCelsius("\nN/A\n"); ok {
		t.Error("non-numeric output should yield ok=false")
	}
	// 0 tenths-K is absolute zero -> implausible, filtered
	if _, ok := parseThermalZoneCelsius("0\n"); ok {
		t.Error("implausible reading should be filtered")
	}
}
