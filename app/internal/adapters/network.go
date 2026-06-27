// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"sort"
	"sync"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"

	"heimdall/app/internal/domain"
)

// Network reports receive/transmit throughput in MB/s: an aggregate across all
// non-loopback interfaces (net.rx / net.tx) plus a per-NIC breakdown
// (net.rx.<iface> / net.tx.<iface>), derived from the delta between collects.
type Network struct {
	mu   sync.Mutex
	last map[string]nicCounters
	at   time.Time
}

type nicCounters struct{ rx, tx uint64 }

func (n *Network) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "net", Metrics: []string{"net.rx", "net.tx"}}
}

func (n *Network) Collect(ctx context.Context) ([]domain.Metric, error) {
	counters, err := gnet.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	n.mu.Lock()
	defer n.mu.Unlock()
	dt := 0.0
	if !n.at.IsZero() {
		dt = now.Sub(n.at).Seconds()
	}

	cur := make(map[string]nicCounters, len(counters))
	names := make([]string, 0, len(counters))
	for _, c := range counters {
		if isLoopback(c.Name) {
			continue
		}
		cur[c.Name] = nicCounters{rx: c.BytesRecv, tx: c.BytesSent}
		names = append(names, c.Name)
	}
	sort.Strings(names)

	perNIC := make([]domain.Metric, 0, len(names)*2)
	var aggRx, aggTx float64
	for _, name := range names {
		prev, ok := n.last[name]
		if dt <= 0 || !ok {
			continue
		}
		rxRate := deltaRateMB(prev.rx, cur[name].rx, dt)
		txRate := deltaRateMB(prev.tx, cur[name].tx, dt)
		perNIC = append(perNIC,
			domain.Metric{Name: "net.rx." + name, Unit: "MB/s", Status: domain.StatusOK, Gauge: rxRate},
			domain.Metric{Name: "net.tx." + name, Unit: "MB/s", Status: domain.StatusOK, Gauge: txRate},
		)
		aggRx += rxRate
		aggTx += txRate
	}
	n.last = cur
	n.at = now

	out := []domain.Metric{
		{Name: "net.rx", Unit: "MB/s", Status: domain.StatusOK, Gauge: aggRx},
		{Name: "net.tx", Unit: "MB/s", Status: domain.StatusOK, Gauge: aggTx},
	}
	return append(out, perNIC...), nil
}

// isLoopback reports whether name is the loopback interface, which carries no
// meaningful throughput and is excluded from the breakdown.
func isLoopback(name string) bool {
	return name == "lo" || name == "lo0"
}
