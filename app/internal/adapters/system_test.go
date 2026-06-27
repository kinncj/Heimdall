// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func first(t *testing.T, a domain.Adapter) domain.Metric {
	t.Helper()
	ms, err := a.Collect(context.Background())
	if err != nil {
		t.Fatalf("%s collect: %v", a.Describe().ID, err)
	}
	if len(ms) == 0 {
		t.Fatalf("%s: no metrics", a.Describe().ID)
	}
	return ms[0]
}

func TestCPUReportsRealUtil(t *testing.T) {
	m := first(t, CPU{})
	if m.Name != "cpu.util" || m.Status != domain.StatusOK {
		t.Fatalf("got %+v", m)
	}
	if m.Gauge < 0 || m.Gauge > 100 {
		t.Errorf("cpu.util %.1f out of [0,100]", m.Gauge)
	}
}

func TestMemReportsRealUsed(t *testing.T) {
	m := first(t, Mem{})
	if m.Name != "mem.used" || m.Status != domain.StatusOK {
		t.Fatalf("got %+v", m)
	}
	if m.Gauge <= 0 || m.Gauge > 100 {
		t.Errorf("mem.used %.1f not in (0,100]", m.Gauge)
	}
}

func TestDiskReportsRealUsed(t *testing.T) {
	m := first(t, Disk{Path: "/"})
	if m.Name != "disk.used" || m.Status != domain.StatusOK {
		t.Fatalf("got %+v", m)
	}
	if m.Gauge <= 0 || m.Gauge > 100 {
		t.Errorf("disk.used %.1f not in (0,100]", m.Gauge)
	}
}

func TestDefaultAdapters(t *testing.T) {
	if n := len(Default()); n != 10 {
		t.Fatalf("Default() = %d adapters, want 10", n)
	}
}

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

func TestNetworkDeltaAndUptime(t *testing.T) {
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
	up := first(t, Uptime{})
	if up.Name != "host.uptime" || up.Status != domain.StatusOK || up.Gauge <= 0 {
		t.Errorf("uptime = %+v", up)
	}
}

func TestCPUPerCore(t *testing.T) {
	ms, err := CPU{}.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var cores domain.Metric
	for _, m := range ms {
		if m.Name == "cpu.cores" {
			cores = m
		}
	}
	if cores.Kind != domain.KindPerCore || len(cores.PerCore) == 0 {
		t.Fatalf("cpu.cores = %+v, want per-core values", cores)
	}
}
