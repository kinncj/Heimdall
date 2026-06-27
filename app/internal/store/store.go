// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package store is the optional durable sink (Mímir, ADR 0016). Heimdall never
// embeds a TSDB; instead a Store writes the live fleet to the operator's own
// Prometheus-compatible backend and restores the hub's last-known state from it
// on restart. It is off unless a backend is configured. See ADR 0016 / 0008.
package store

import (
	"context"
	"time"

	"heimdall/app/internal/domain"
)

// Store is the durable backend port: write the current fleet, and restore the
// last-known fleet on startup. Implementations talk to an external TSDB.
type Store interface {
	// Write persists the current host views as samples timestamped at now.
	Write(ctx context.Context, views []domain.HostView, now time.Time) error
	// Restore returns the last-known host views reconstructed from the backend,
	// best-effort: scalar gauges + labels + last-seen, no info strings or alerts.
	Restore(ctx context.Context) ([]domain.HostView, error)
}
