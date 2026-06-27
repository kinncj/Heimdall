// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"testing"

	"heimdall/app/internal/domain"
)

func TestUptime(t *testing.T) {
	up := first(t, Uptime{})
	if up.Name != "host.uptime" || up.Status != domain.StatusOK || up.Gauge <= 0 {
		t.Errorf("uptime = %+v", up)
	}
}
