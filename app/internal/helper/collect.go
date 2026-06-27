// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"

	"heimdall/app/internal/domain"
)

// PrivilegedMetrics runs the platform's privileged collectors. On Apple Silicon
// it reads power from the IOReport energy counters first — these need no root —
// and uses powermetrics only to fill GPU utilisation (and any power a given SoC
// does not expose via IOReport). Anywhere nvidia-smi is present it queries the
// NVIDIA GPU. When no source yields a value it returns Unavailable readings
// rather than failing, so an unsupported host degrades gracefully.
func PrivilegedMetrics(ctx context.Context) []domain.Metric {
	var out []domain.Metric
	if runtime.GOOS == "darwin" {
		if cpu, gpu, ane, gpuUtil, ok := ioReportPower(200); ok {
			total := 0.0
			if cpu > 0 {
				out = append(out, powerMetric("power.cpu", cpu))
				total += cpu
			}
			if gpu > 0 {
				out = append(out, powerMetric("power.gpu", gpu))
				total += gpu
			}
			if ane > 0 {
				out = append(out, powerMetric("power.ane", ane))
				total += ane
			}
			if total > 0 {
				out = append(out, powerMetric("power.pkg", total))
			}
			if gpuUtil >= 0 {
				out = append(out, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: gpuUtil})
			}
		}
		// powermetrics (root only) adds GPU utilisation and any power the energy
		// counters did not expose; existing IOReport readings win.
		if text, err := runPowermetrics(ctx); err == nil {
			out = mergeByName(out, parsePowermetrics(text))
		}
	}
	// Linux privileged sources (RAPL power, hwmon temps); no-op off Linux.
	out = mergeByName(out, linuxPrivileged(ctx))
	if text, err := runNvidiaSMI(ctx); err == nil {
		out = mergeByName(out, parseNvidiaSMI(text))
	}
	if !hasOK(out) {
		return []domain.Metric{
			{Name: "power.pkg", Status: domain.StatusUnavailable, Detail: "no power source"},
			{Name: "gpu.util", Status: domain.StatusUnavailable, Detail: "no gpu source"},
		}
	}
	return out
}

func powerMetric(name string, w float64) domain.Metric {
	return domain.Metric{Name: name, Unit: "watts", Status: domain.StatusOK, Gauge: w}
}

// mergeByName appends secondary metrics whose names are not already present.
func mergeByName(primary, secondary []domain.Metric) []domain.Metric {
	seen := make(map[string]bool, len(primary))
	for _, m := range primary {
		seen[m.Name] = true
	}
	for _, m := range secondary {
		if !seen[m.Name] {
			primary = append(primary, m)
			seen[m.Name] = true
		}
	}
	return primary
}

func hasOK(ms []domain.Metric) bool {
	for _, m := range ms {
		if m.Status == domain.StatusOK {
			return true
		}
	}
	return false
}

func runPowermetrics(ctx context.Context) (string, error) {
	if os.Geteuid() != 0 {
		return "", errors.New("powermetrics requires root")
	}
	path, err := exec.LookPath("powermetrics")
	if err != nil {
		return "", err
	}
	out, err := exec.CommandContext(ctx, path,
		"--samplers", "cpu_power,gpu_power", "-n", "1", "-i", "200").Output()
	return string(out), err
}

func runNvidiaSMI(ctx context.Context) (string, error) {
	path, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return "", err
	}
	out, err := exec.CommandContext(ctx, path,
		"--query-gpu=utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw",
		"--format=csv,noheader,nounits").Output()
	return string(out), err
}
