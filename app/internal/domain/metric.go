// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package domain holds Heimdall's core monitoring types and rules. It has zero
// dependencies on transport, persistence, or UI frameworks (Clean Architecture):
// metrics, hosts, and the metric-adapter contract live here and are imported
// inward by every outer layer.
package domain

import "time"

// MetricStatus is an adapter's self-reported health for a single reading.
//
// StatusUnspecified is the zero value and is deliberately NOT "healthy": a
// monitoring system must never read a missing or default status as OK.
type MetricStatus int

const (
	StatusUnspecified MetricStatus = iota
	StatusOK
	StatusUnavailable            // adapter not applicable on this host (e.g. no NVIDIA GPU)
	StatusInsufficientPermission // needs the privileged helper / denied
	StatusError                  // adapter errored, timed out, or panicked
)

func (s MetricStatus) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusUnavailable:
		return "unavailable"
	case StatusInsufficientPermission:
		return "insufficient_permission"
	case StatusError:
		return "error"
	default:
		return "unspecified"
	}
}

// MetricKind tags which value field of a Metric is populated.
type MetricKind int

const (
	KindGauge   MetricKind = iota // instantaneous value in Gauge
	KindPerCore                   // per-core values in PerCore
	KindCounter                   // monotonic counter + rate in Counter
)

// Counter is a monotonic total plus its derived per-second rate.
type Counter struct {
	Total uint64
	Rate  float64
}

// Metric is one reading from one adapter. Status is always meaningful; a non-OK
// status (Unavailable, InsufficientPermission, Error) carries no value and is
// explained by Detail. This is the failure-isolation guarantee in the domain.
type Metric struct {
	Name    string
	Unit    string
	Status  MetricStatus
	Kind    MetricKind
	Gauge   float64
	PerCore []float64
	Counter Counter
	Detail  string
	At      time.Time
}
