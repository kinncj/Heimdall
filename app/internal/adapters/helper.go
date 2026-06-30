// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"

	"heimdall/app/internal/domain"
	"heimdall/app/internal/helper"
)

// Helper bridges privileged metrics into the adapter framework with a two-step
// strategy. It first tries the out-of-process heimdall-helper over its socket,
// so the daemon can stay unprivileged. If no helper is running it falls back to
// collecting in-process, which succeeds when the daemon itself is privileged
// (e.g. started with sudo) or where the platform tool is readable unprivileged
// (nvidia-smi). Only when neither path yields a value does it report
// insufficient-permission — the dashboard's needs-helper affordance.
type Helper struct {
	Client helper.Client
	// Direct collects privileged metrics in-process when the helper socket is
	// unavailable. Defaults to helper.PrivilegedMetrics.
	Direct func(context.Context) []domain.Metric
}

func (Helper) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{
		ID:                "privileged",
		Metrics:           []string{"power.pkg", "power.cpu", "power.gpu", "power.npu", "gpu.util", "gpu.vram", "gpu.temp"},
		RequiresPrivilege: true,
	}
}

func (h Helper) Collect(ctx context.Context) ([]domain.Metric, error) {
	if ms, err := h.Client.Collect(ctx); err == nil {
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
