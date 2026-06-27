// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"time"

	"heimdall/app/internal/domain"
)

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
