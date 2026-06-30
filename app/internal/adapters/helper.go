// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/helper"
)

// metricCollector is the helper-socket capability the adapter depends on (DIP):
// just the metric read. helper.Client satisfies it; tests inject a fake.
type metricCollector interface {
	Collect(ctx context.Context) ([]domain.Metric, error)
}

// Helper bridges privileged metrics into the adapter framework with a two-step
// strategy. It first tries the out-of-process heimdall-helper over its socket,
// so the daemon can stay unprivileged. If the helper is absent OR returns no ok
// metrics, it falls back to collecting in-process, which succeeds when the
// daemon itself is privileged (e.g. started with sudo) or where the platform
// tool is readable unprivileged (nvidia-smi, and IOReport/SMC on Apple Silicon).
// Only when neither path yields a value does it report insufficient-permission —
// the dashboard's needs-helper affordance.
type Helper struct {
	// Client reads from the helper socket. A nil Client uses a default
	// helper.Client (the daemon's normal path); injectable for tests.
	Client metricCollector
	// Direct collects privileged metrics in-process when the helper socket is
	// unavailable or empty. Defaults to helper.PrivilegedMetrics.
	Direct func(context.Context) []domain.Metric
}

func (Helper) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{
		ID:                "privileged",
		Metrics:           []string{"power.pkg", "power.cpu", "power.gpu", "power.npu", "gpu.util", "gpu.vram", "gpu.temp", "gpu.clock", "gpu.mem.util", "gpu.fan", "npu.util"},
		RequiresPrivilege: true,
	}
}

func (h Helper) Collect(ctx context.Context) ([]domain.Metric, error) {
	var client metricCollector = h.Client
	if client == nil {
		client = helper.Client{}
	}
	// Use the helper only when it actually produced an ok reading. A reachable
	// helper that returns nothing ok (e.g. a non-cgo helper, or one that can't
	// read IOReport) must NOT shadow a working in-process source — otherwise
	// running the helper on a Mac, where IOReport is unprivileged, blanks out the
	// power/gpu the daemon was reading itself.
	if ms, err := client.Collect(ctx); err == nil && anyOK(ms) {
		return ms, nil
	}
	direct := h.Direct
	if direct == nil {
		direct = helper.PrivilegedMetrics
	}
	if ms := direct(ctx); anyOK(ms) {
		return ms, nil
	}
	return []domain.Metric{
		{Name: "power.pkg", Status: domain.StatusInsufficientPermission, Detail: "needs helper"},
		{Name: "gpu.util", Status: domain.StatusInsufficientPermission, Detail: "needs helper"},
	}, nil
}

func anyOK(ms []domain.Metric) bool {
	for _, m := range ms {
		if m.Status == domain.StatusOK {
			return true
		}
	}
	return false
}
