// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"testing"

	"heimdall/app/internal/domain"
)

func TestInventoryReportsHostDescriptors(t *testing.T) {
	inv := &Inventory{Version: "v9.9.9"}
	ms, err := inv.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	by := make(map[string]domain.Metric, len(ms))
	for _, m := range ms {
		by[m.Name] = m
	}

	// CPU model and version are always reported; OS comes from gopsutil where
	// available. All descriptors are OK info metrics carrying a Detail string.
	cpu, ok := by["host.cpu"]
	if !ok || cpu.Status != domain.StatusOK || cpu.Detail == "" {
		t.Fatalf("host.cpu = %+v, want OK with detail", cpu)
	}
	if v := by["host.version"]; v.Detail != "v9.9.9" {
		t.Fatalf("host.version detail = %q, want v9.9.9", v.Detail)
	}
}

func TestInventoryCachesFirstGather(t *testing.T) {
	inv := &Inventory{Version: "v1"}
	a, _ := inv.Collect(context.Background())
	b, _ := inv.Collect(context.Background())
	if len(a) != len(b) || len(a) == 0 {
		t.Fatalf("inventory not stable/cached: %d then %d", len(a), len(b))
	}
}
