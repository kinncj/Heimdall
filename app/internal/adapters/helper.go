// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"time"

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
		Metrics:           []string{"power.total", "power.cpu", "power.gpu", "power.npu", "gpu.util", "gpu.vram", "gpu.temp", "gpu.clock", "gpu.mem.util", "gpu.fan", "npu.util"},
		RequiresPrivilege: true,
	}
}

func (h Helper) Collect(ctx context.Context) ([]domain.Metric, error) {
	direct := h.Direct
	if direct == nil {
		direct = helper.PrivilegedMetrics
	}

	// In-process first. Where the daemon can already read a privileged CPU power
	// rail itself — macOS (SMC/IOReport are unprivileged) or a root daemon reading
	// RAPL — the helper can only duplicate it, so we return the in-process read and
	// never call the helper. That's what makes running the helper on Apple Silicon
	// harmless: the daemon has power without it, so a slow helper (macOS
	// powermetrics can take ~1s) is never on the critical path.
	inproc := direct(ctx)
	if hasOKPower(inproc) {
		return inproc, nil
	}

	// Otherwise the daemon is unprivileged and can't read the CPU power rail
	// (typically a Linux daemon that needs the root helper for RAPL). Consult the
	// helper, bounded so a slow or hung one can't stall the collection cycle. A
	// reachable-but-empty helper must never shadow the in-process read, so we only
	// take the helper's result when it has an ok metric; else we fall back to
	// whatever in-process produced.
	var client metricCollector = h.Client
	if client == nil {
		client = helper.Client{}
	}
	hctx, cancel := helperCallContext(ctx)
	defer cancel()
	if ms, err := client.Collect(hctx); err == nil && anyOK(ms) {
		return ms, nil
	}
	if anyOK(inproc) {
		return inproc, nil
	}
	return []domain.Metric{
		{Name: "power.cpu", Status: domain.StatusInsufficientPermission, Detail: "needs helper"},
		{Name: "gpu.util", Status: domain.StatusInsufficientPermission, Detail: "needs helper"},
	}, nil
}

// hasOKPower reports whether the metrics already include a privileged CPU power
// rail the daemon read itself — the signal that the helper would only duplicate.
func hasOKPower(ms []domain.Metric) bool {
	for _, m := range ms {
		if m.Status != domain.StatusOK {
			continue
		}
		switch m.Name {
		case "power.total", "power.cpu", "power.pkg":
			return true
		}
	}
	return false
}

// helperCallContext caps the helper socket call so it leaves budget for the
// in-process fallback within the adapter deadline. Without a deadline it applies
// a sane default; when the deadline is already near, the helper still gets a
// token slice so a healthy helper isn't starved on a tight cycle.
func helperCallContext(ctx context.Context) (context.Context, context.CancelFunc) {
	const reserve = 350 * time.Millisecond
	dl, ok := ctx.Deadline()
	if !ok {
		return context.WithTimeout(ctx, 650*time.Millisecond)
	}
	remaining := time.Until(dl)
	if remaining <= reserve {
		return context.WithTimeout(ctx, remaining/2)
	}
	return context.WithTimeout(ctx, remaining-reserve)
}

func anyOK(ms []domain.Metric) bool {
	for _, m := range ms {
		if m.Status == domain.StatusOK {
			return true
		}
	}
	return false
}
