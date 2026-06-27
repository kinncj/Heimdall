// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func TestNetworkDelta(t *testing.T) {
	n := &Network{}
	if _, err := n.Collect(context.Background()); err != nil {
		t.Fatalf("net collect 1: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	ms, err := n.Collect(context.Background())
	if err != nil {
		t.Fatalf("net collect 2: %v", err)
	}
	if len(ms) != 2 || ms[0].Name != "net.rx" || ms[1].Name != "net.tx" {
		t.Fatalf("net metrics = %+v", ms)
	}
	for _, m := range ms {
		if m.Status == domain.StatusOK && m.Gauge < 0 {
			t.Errorf("%s rate negative: %v", m.Name, m.Gauge)
		}
	}
}
