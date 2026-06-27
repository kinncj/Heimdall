// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package adapters provides concrete domain.Adapter implementations that read
// real metrics from the host via gopsutil. They are unprivileged; signals that
// need elevation report INSUFFICIENT_PERMISSION rather than failing.
//
// Each adapter lives in its own file (cpu.go, mem.go, disk.go, diskio.go,
// network.go, …) so adding a metric is a matter of dropping in one new file and
// registering its type in Build — no existing adapter is touched.
package adapters

import "heimdall/app/internal/domain"

// Options configures the adapter set built by Build.
type Options struct {
	// PingTarget is the internet host whose round-trip latency the Reachability
	// adapter measures. Empty means 1.1.1.1.
	PingTarget string
	// Version is the Heimdall build version reported by the Inventory adapter.
	// Empty means "dev".
	Version string
}

// Build returns the unprivileged adapters plus the helper bridge, configured by
// o. It is the single place the daemon assembles its adapter set; register a new
// adapter here.
func Build(o Options) []domain.Adapter {
	return []domain.Adapter{
		&Inventory{Version: o.Version},
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
