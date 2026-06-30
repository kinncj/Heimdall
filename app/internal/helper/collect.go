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
		cpu, gpu, ane, gpuUtil, ioOK := ioReportPower(200)
		smcPkg, smcOK := smcSystemPower()
		// powermetrics (root only) fills GPU utilisation and any power the energy
		// counters did not expose.
		var pm []domain.Metric
		if text, err := runPowermetrics(ctx); err == nil {
			pm = parsePowermetrics(text)
		}
		out = append(out, assembleApplePower(cpu, gpu, ane, gpuUtil, ioOK, smcPkg, smcOK, pm)...)
	}
	// Linux privileged sources (RAPL power, hwmon temps); no-op off Linux.
	out = mergeByName(out, linuxPrivileged(ctx))
	// Windows privileged sources (WMI thermal zones); no-op off Windows.
	out = mergeByName(out, windowsPrivileged(ctx))
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

// assembleApplePower builds the macOS power/util metrics from the available
// sources. Per-domain CPU/GPU/ANE power and GPU utilisation come from the
// IOReport energy counters when present. Package power has a strict precedence:
// SMC PSTR ("System Total Power") first, then powermetrics, then the IOReport
// per-domain sum as a last resort. The ordering matters: on Apple Silicon
// Pro/Max chips IOReport reports 0 for CPU/ANE and only a sub-watt GPU figure,
// so the energy-sum is a phantom — it must never shadow a real SMC or
// powermetrics reading. pm is the parsed powermetrics output (may be nil).
func assembleApplePower(cpu, gpu, ane, gpuUtil float64, ioOK bool, smcPkg float64, smcOK bool, pm []domain.Metric) []domain.Metric {
	var out []domain.Metric
	if ioOK {
		if cpu > 0 {
			out = append(out, powerMetric("power.cpu", cpu))
		}
		if gpu > 0 {
			out = append(out, powerMetric("power.gpu", gpu))
		}
		if ane > 0 {
			out = append(out, powerMetric("power.npu", ane))
		}
		if gpuUtil >= 0 {
			out = append(out, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: gpuUtil})
		}
	}
	// SMC PSTR is the authoritative whole-system figure and needs no root.
	if smcOK && smcPkg > 0 {
		out = append(out, powerMetric("power.pkg", smcPkg))
	}
	// powermetrics fills any name not already set (package when SMC is absent,
	// plus CPU/GPU/util gaps).
	out = mergeByName(out, pm)
	// Last resort: the IOReport per-domain sum, only when nothing better gave a
	// package figure.
	if !hasName(out, "power.pkg") {
		sum := 0.0
		if ioOK {
			for _, w := range []float64{cpu, gpu, ane} {
				if w > 0 {
					sum += w
				}
			}
		}
		if sum > 0 {
			out = append(out, powerMetric("power.pkg", sum))
		}
	}
	return out
}

func hasName(ms []domain.Metric, name string) bool {
	for _, m := range ms {
		if m.Name == name {
			return true
		}
	}
	return false
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
