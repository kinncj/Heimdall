// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package topview

import (
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

func freqHost(g float64) domain.HostView {
	return domain.HostView{
		Host:  domain.Host{ID: "h"},
		State: domain.StateOnline,
		LastSnapshot: []domain.Metric{
			{Name: "cpu.freq", Status: domain.StatusOK, Gauge: g},
			{Name: "cpu.util", Status: domain.StatusOK, Gauge: 50},
		},
		Processes: []domain.ProcessRow{{PID: 1, CPUPct: 9, MemPct: 2, Command: "x"}},
	}
}

func TestFreqRendersGHzNotMHz(t *testing.T) {
	out := strip(New(freqHost(4200), nil, darkMode(t), 120, 40).View())
	if !strings.Contains(out, "4.20 GHz") {
		t.Errorf("expected 4.20 GHz, got:\n%s", out)
	}
	if strings.Contains(out, "4200.00") {
		t.Errorf("freq should not render raw MHz as GHz")
	}
}

func TestProcessTableHasNoUserColumn(t *testing.T) {
	out := strip(New(freqHost(3000), nil, darkMode(t), 120, 40).View())
	if strings.Contains(out, "USER") {
		t.Errorf("process table should not have a USER column:\n%s", out)
	}
}
