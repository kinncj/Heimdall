// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"

	"heimdall/app/internal/domain"
)

func TestBuildFreqMeanAndPerCore(t *testing.T) {
	m := buildFreq([]cpu.InfoStat{{Mhz: 3000}, {Mhz: 3200}}, nil)
	if m.Status != domain.StatusOK || m.Gauge != 3100 {
		t.Fatalf("mean MHz wrong: %+v", m)
	}
	if len(m.PerCore) != 2 {
		t.Errorf("PerCore should carry each clock, got %v", m.PerCore)
	}
}

func TestBuildFreqDegradesWhenNoClock(t *testing.T) {
	// Apple Silicon: entries exist but report Mhz 0.
	if m := buildFreq([]cpu.InfoStat{{Mhz: 0}}, nil); m.Status != domain.StatusUnavailable {
		t.Errorf("zero-clock should be Unavailable, got %+v", m)
	}
	if m := buildFreq(nil, errors.New("x")); m.Status != domain.StatusUnavailable {
		t.Errorf("error should be Unavailable, got %+v", m)
	}
}
