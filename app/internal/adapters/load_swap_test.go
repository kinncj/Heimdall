// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v4/load"

	"heimdall/app/internal/domain"
)

func TestSwapReportsPercent(t *testing.T) {
	m := first(t, Swap{})
	if m.Name != "mem.swap" {
		t.Fatalf("got %+v", m)
	}
	// macOS/Linux test hosts have a swap source; status must be OK with a sane %.
	if m.Status == domain.StatusOK && (m.Gauge < 0 || m.Gauge > 100) {
		t.Errorf("mem.swap %.1f out of [0,100]", m.Gauge)
	}
}

func TestBuildLoadOK(t *testing.T) {
	m := buildLoad(&load.AvgStat{Load1: 2.4, Load5: 1.9, Load15: 1.5}, nil)
	if m.Status != domain.StatusOK || m.Gauge != 2.4 {
		t.Fatalf("got %+v", m)
	}
	if len(m.PerCore) != 3 || m.PerCore[2] != 1.5 {
		t.Errorf("PerCore should carry [1m,5m,15m], got %v", m.PerCore)
	}
}

func TestBuildLoadDegradesOnError(t *testing.T) {
	m := buildLoad(nil, errors.New("not supported"))
	if m.Status != domain.StatusUnavailable {
		t.Errorf("load on an unsupported platform must be Unavailable, got %+v", m)
	}
}
