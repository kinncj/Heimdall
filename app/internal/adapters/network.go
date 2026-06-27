// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"sync"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"

	"heimdall/app/internal/domain"
)

// Network reports aggregate receive/transmit throughput in MB/s, derived from
// the delta between consecutive collects.
type Network struct {
	mu     sync.Mutex
	lastRx uint64
	lastTx uint64
	last   time.Time
}

func (n *Network) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "net", Metrics: []string{"net.rx", "net.tx"}}
}

func (n *Network) Collect(ctx context.Context) ([]domain.Metric, error) {
	counters, err := gnet.IOCountersWithContext(ctx, false)
	if err != nil {
		return nil, err
	}
	if len(counters) == 0 {
		return []domain.Metric{
			{Name: "net.rx", Status: domain.StatusUnavailable},
			{Name: "net.tx", Status: domain.StatusUnavailable},
		}, nil
	}
	rx, tx, now := counters[0].BytesRecv, counters[0].BytesSent, time.Now()

	n.mu.Lock()
	defer n.mu.Unlock()
	var rxRate, txRate float64
	if !n.last.IsZero() {
		dt := now.Sub(n.last).Seconds()
		rxRate = deltaRateMB(n.lastRx, rx, dt)
		txRate = deltaRateMB(n.lastTx, tx, dt)
	}
	n.lastRx, n.lastTx, n.last = rx, tx, now
	return []domain.Metric{
		{Name: "net.rx", Unit: "MB/s", Status: domain.StatusOK, Gauge: rxRate},
		{Name: "net.tx", Unit: "MB/s", Status: domain.StatusOK, Gauge: txRate},
	}, nil
}
