// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"

	"heimdall/app/internal/domain"
)

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
