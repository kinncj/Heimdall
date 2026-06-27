// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package domain

import "context"

// AdapterInfo describes an Adapter without collecting from it: a stable id, the
// metric names it produces, and whether it needs the privileged helper.
type AdapterInfo struct {
	ID                string
	Metrics           []string
	RequiresPrivilege bool
}

// Adapter is the metric-collection contract shared by the daemon and the hub.
//
// It is intentionally small (Interface Segregation) so a new signal is added by
// writing a new Adapter and registering it — never by modifying existing
// adapters (Open/Closed). The Registry depends on this abstraction, not on any
// concrete collector (Dependency Inversion).
type Adapter interface {
	Describe() AdapterInfo
	Collect(ctx context.Context) ([]Metric, error)
}
