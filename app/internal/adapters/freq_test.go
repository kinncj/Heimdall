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

func TestFreqFromKHzMeanAndPerCore(t *testing.T) {
	// sysfs scaling_cur_freq is kHz; 2_800_000 kHz -> 2800 MHz.
	m := freqFromKHz([]string{"2800000", "2400000", "bad", ""})
	if m.Status != domain.StatusOK || m.Gauge != 2600 {
		t.Fatalf("mean MHz wrong: %+v", m)
	}
	if len(m.PerCore) != 2 || m.PerCore[0] != 2800 {
		t.Errorf("PerCore should be MHz per readable core, got %v", m.PerCore)
	}
}

func TestFreqFromKHzDegradesWhenEmpty(t *testing.T) {
	if m := freqFromKHz([]string{"0", "bad"}); m.Status != domain.StatusUnavailable {
		t.Errorf("no readable clock should be Unavailable, got %+v", m)
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
