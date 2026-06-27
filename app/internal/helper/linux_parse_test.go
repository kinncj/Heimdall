// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"testing"
	"time"
)

func TestRaplWatts(t *testing.T) {
	// 10 J consumed over 1s = 10 W. (10 J = 10_000_000 µJ)
	if w := raplWatts(0, 10_000_000, time.Second); w != 10 {
		t.Errorf("raplWatts = %v, want 10", w)
	}
	// counter wrap -> 0 (treated as no reading)
	if w := raplWatts(100, 50, time.Second); w != 0 {
		t.Errorf("wrap should yield 0, got %v", w)
	}
	// zero interval -> 0
	if w := raplWatts(0, 10_000_000, 0); w != 0 {
		t.Errorf("zero dt should yield 0, got %v", w)
	}
}

func TestParseMicrojoulesAndMilliCelsius(t *testing.T) {
	if v, err := parseMicrojoules("  12345\n"); err != nil || v != 12345 {
		t.Errorf("parseMicrojoules = %v, %v; want 12345", v, err)
	}
	if c, err := parseMilliCelsius("52000\n"); err != nil || c != 52 {
		t.Errorf("parseMilliCelsius = %v, %v; want 52", c, err)
	}
}

func TestIsPackageSensor(t *testing.T) {
	for _, ok := range []string{"coretemp", "k10temp", "zenpower"} {
		if !isPackageSensor(ok) {
			t.Errorf("%q should be a package sensor", ok)
		}
	}
	for _, no := range []string{"acpitz", "nvme", "iwlwifi_1"} {
		if isPackageSensor(no) {
			t.Errorf("%q should not be a package sensor", no)
		}
	}
}
