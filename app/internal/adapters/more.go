// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"

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

// deltaRateMB converts the growth of a monotonic byte counter over dt seconds
// into MB/s. A non-positive interval or a counter reset (cur < prev) yields 0,
// so a wrapped or restarted counter never produces a bogus spike.
func deltaRateMB(prev, cur uint64, dt float64) float64 {
	if dt <= 0 || cur < prev {
		return 0
	}
	return float64(cur-prev) / dt / 1e6
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

// Reachability measures round-trip latency to an internet target via ICMP echo
// (the system `ping`). An unreachable target is reported as an error on that
// metric only (the host stays online), satisfying story 0007's isolation
// requirement. Target is configurable; it defaults to 1.1.1.1.
type Reachability struct{ Target string }

func (r Reachability) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "reach", Metrics: []string{"net.latency"}}
}

func (r Reachability) Collect(ctx context.Context) ([]domain.Metric, error) {
	target := r.Target
	if target == "" {
		target = "1.1.1.1"
	}
	ms, ok := pingFn(ctx, target, 1*time.Second)
	if !ok {
		return []domain.Metric{{Name: "net.latency", Status: domain.StatusError, Detail: "no reply from " + target}}, nil
	}
	return []domain.Metric{{Name: "net.latency", Unit: "ms", Status: domain.StatusOK, Gauge: ms}}, nil
}

// Uptime reports seconds since boot.
type Uptime struct{}

func (Uptime) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "uptime", Metrics: []string{"host.uptime"}}
}

func (Uptime) Collect(ctx context.Context) ([]domain.Metric, error) {
	up, err := host.UptimeWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return []domain.Metric{{Name: "host.uptime", Unit: "s", Status: domain.StatusOK, Gauge: float64(up)}}, nil
}

// Temperature reports the hottest reported sensor. Many platforms (notably
// macOS) expose no sensors without elevation — reported as unavailable / needs
// the privileged helper, never an error.
type Temperature struct{}

func (Temperature) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "temp", Metrics: []string{"temp.pkg"}}
}

func (Temperature) Collect(ctx context.Context) ([]domain.Metric, error) {
	temps, err := sensors.TemperaturesWithContext(ctx)
	if err != nil || len(temps) == 0 {
		return []domain.Metric{{Name: "temp.pkg", Status: domain.StatusInsufficientPermission, Detail: "needs helper"}}, nil
	}
	var max float64
	for _, s := range temps {
		if s.Temperature > max {
			max = s.Temperature
		}
	}
	return []domain.Metric{{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: max}}, nil
}
