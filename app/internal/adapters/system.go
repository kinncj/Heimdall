// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package adapters provides concrete domain.Adapter implementations that read
// real metrics from the host via gopsutil. They are unprivileged; signals that
// need elevation report INSUFFICIENT_PERMISSION rather than failing.
package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"

	"heimdall/app/internal/domain"
)

// CPU reports total CPU utilisation plus per-core utilisation.
type CPU struct{}

func (CPU) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "cpu", Metrics: []string{"cpu.util", "cpu.cores"}}
}

func (CPU) Collect(ctx context.Context) ([]domain.Metric, error) {
	per, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return nil, err
	}
	avg := 0.0
	for _, v := range per {
		avg += v
	}
	if len(per) > 0 {
		avg /= float64(len(per))
	}
	out := []domain.Metric{
		{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: avg},
	}
	if len(per) > 0 {
		out = append(out, domain.Metric{
			Name: "cpu.cores", Unit: "percent", Status: domain.StatusOK,
			Kind: domain.KindPerCore, PerCore: per,
		})
	}
	return out, nil
}

// Mem reports virtual memory used percentage.
type Mem struct{}

func (Mem) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "mem", Metrics: []string{"mem.used"}}
}

func (Mem) Collect(ctx context.Context) ([]domain.Metric, error) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return []domain.Metric{{Name: "mem.used", Unit: "percent", Status: domain.StatusOK, Gauge: vm.UsedPercent}}, nil
}

// Disk reports used percentage of a mount point.
type Disk struct{ Path string }

func (d Disk) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "disk", Metrics: []string{"disk.used"}}
}

func (d Disk) Collect(ctx context.Context) ([]domain.Metric, error) {
	path := d.Path
	if path == "" {
		path = "/"
	}
	u, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return nil, err
	}
	return []domain.Metric{{Name: "disk.used", Unit: "percent", Status: domain.StatusOK, Gauge: u.UsedPercent}}, nil
}

// DiskIO reports aggregate read/write throughput in MB/s, summed across every
// physical device and derived from the delta between consecutive collects.
type DiskIO struct {
	mu        sync.Mutex
	lastRead  uint64
	lastWrite uint64
	last      time.Time
}

func (d *DiskIO) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "diskio", Metrics: []string{"disk.read", "disk.write"}}
}

func (d *DiskIO) Collect(ctx context.Context) ([]domain.Metric, error) {
	counters, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if len(counters) == 0 {
		return []domain.Metric{
			{Name: "disk.read", Status: domain.StatusUnavailable},
			{Name: "disk.write", Status: domain.StatusUnavailable},
		}, nil
	}
	var read, write uint64
	for _, c := range counters {
		read += c.ReadBytes
		write += c.WriteBytes
	}
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()
	var readRate, writeRate float64
	if !d.last.IsZero() {
		dt := now.Sub(d.last).Seconds()
		readRate = deltaRateMB(d.lastRead, read, dt)
		writeRate = deltaRateMB(d.lastWrite, write, dt)
	}
	d.lastRead, d.lastWrite, d.last = read, write, now
	return []domain.Metric{
		{Name: "disk.read", Unit: "MB/s", Status: domain.StatusOK, Gauge: readRate},
		{Name: "disk.write", Unit: "MB/s", Status: domain.StatusOK, Gauge: writeRate},
	}, nil
}

// Options configures the adapter set built by Build.
type Options struct {
	// PingTarget is the internet host whose round-trip latency the Reachability
	// adapter measures. Empty means 1.1.1.1.
	PingTarget string
}

// Build returns the unprivileged adapters plus the helper bridge, configured by
// o. It is the single place the daemon assembles its adapter set.
func Build(o Options) []domain.Adapter {
	return []domain.Adapter{
		CPU{}, Mem{}, Disk{Path: "/"}, &DiskIO{},
		Temperature{}, &Network{}, Reachability{Target: o.PingTarget}, Uptime{},
		&Gateway{}, Helper{},
	}
}

// Default returns the always-available unprivileged adapters plus the helper
// bridge (which self-reports needs-helper until heimdall-helper is installed),
// with default options.
func Default() []domain.Adapter {
	return Build(Options{})
}
