// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func TestDiskIODelta(t *testing.T) {
	d := &DiskIO{}
	if _, err := d.Collect(context.Background()); err != nil {
		t.Fatalf("diskio collect 1: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	ms, err := d.Collect(context.Background())
	if err != nil {
		t.Fatalf("diskio collect 2: %v", err)
	}
	if len(ms) != 2 || ms[0].Name != "disk.read" || ms[1].Name != "disk.write" {
		t.Fatalf("diskio metrics = %+v", ms)
	}
	for _, m := range ms {
		if m.Status == domain.StatusOK && m.Gauge < 0 {
			t.Errorf("%s rate negative: %v", m.Name, m.Gauge)
		}
	}
}
