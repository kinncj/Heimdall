// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package adapters

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"

	"heimdall/app/internal/domain"
)

// Inventory reports slow-changing host descriptors — OS, architecture, kernel,
// CPU and GPU model, and the Heimdall build version — as info metrics whose
// value rides in Detail (Unit "info", no gauge). They are gathered once and
// cached, since they do not change while the daemon runs.
type Inventory struct {
	Version string

	once   sync.Once
	cached []domain.Metric
}

func (i *Inventory) Describe() domain.AdapterInfo {
	return domain.AdapterInfo{ID: "inventory", Metrics: []string{
		"host.os", "host.arch", "host.kernel", "host.cpu", "host.gpu", "host.version",
	}}
}

func (i *Inventory) Collect(ctx context.Context) ([]domain.Metric, error) {
	i.once.Do(func() { i.cached = i.gather(ctx) })
	return i.cached, nil
}

func (i *Inventory) gather(ctx context.Context) []domain.Metric {
	out := make([]domain.Metric, 0, 6)
	if h, err := host.InfoWithContext(ctx); err == nil {
		out = append(out,
			infoMetric("host.os", strings.TrimSpace(h.Platform+" "+h.PlatformVersion)),
			infoMetric("host.arch", h.KernelArch),
			infoMetric("host.kernel", h.KernelVersion),
		)
	}
	out = append(out, infoMetric("host.cpu", cpuModel(ctx)))
	gpu := gpuModel(ctx)
	if gpu == "" && runtime.GOOS == "darwin" {
		gpu = appleGPU(cpuModel(ctx))
	}
	if gpu != "" {
		out = append(out, infoMetric("host.gpu", gpu))
	}
	version := i.Version
	if version == "" {
		version = "dev"
	}
	out = append(out, infoMetric("host.version", version))
	return out
}

// infoMetric builds a string-valued descriptor metric: OK status, no gauge, the
// human-readable value in Detail.
func infoMetric(name, detail string) domain.Metric {
	return domain.Metric{Name: name, Unit: "info", Status: domain.StatusOK, Detail: detail}
}

func cpuModel(ctx context.Context) string {
	model := ""
	if infos, err := cpu.InfoWithContext(ctx); err == nil && len(infos) > 0 {
		model = strings.TrimSpace(infos[0].ModelName)
	}
	if model == "" {
		model = "CPU"
	}
	if threads, err := cpu.CountsWithContext(ctx, true); err == nil && threads > 0 {
		return fmt.Sprintf("%s (%d threads)", model, threads)
	}
	return model
}

// gpuModel returns the first GPU's marketing name via nvidia-smi, or "" when no
// NVIDIA GPU/driver is present. Best-effort and cached by the caller.
func gpuModel(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name", "--format=csv,noheader").Output()
	if err != nil {
		return ""
	}
	line := out
	if i := strings.IndexByte(string(out), '\n'); i >= 0 {
		line = out[:i]
	}
	return strings.TrimSpace(string(line))
}

// appleGPU derives the integrated GPU name from an Apple Silicon CPU model, e.g.
// "Apple M3 Max (16 threads)" -> "Apple M3 Max GPU". Returns "" for non-Apple.
func appleGPU(cpuModel string) string {
	soc := cpuModel
	if i := strings.Index(soc, " ("); i >= 0 {
		soc = soc[:i]
	}
	soc = strings.TrimSpace(soc)
	if strings.HasPrefix(soc, "Apple ") {
		return soc + " GPU"
	}
	return ""
}
