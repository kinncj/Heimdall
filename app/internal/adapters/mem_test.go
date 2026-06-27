// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"testing"

	"heimdall/app/internal/domain"
)

func TestMemReportsRealUsed(t *testing.T) {
	m := first(t, Mem{})
	if m.Name != "mem.used" || m.Status != domain.StatusOK {
		t.Fatalf("got %+v", m)
	}
	if m.Gauge <= 0 || m.Gauge > 100 {
		t.Errorf("mem.used %.1f not in (0,100]", m.Gauge)
	}
}
