// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/mem"

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
	if nv, present, err := runNvidiaSMI(ctx); present {
		if err != nil {
			// nvidia-smi is installed but broke (e.g. driver/library mismatch after
			// an un-rebooted driver upgrade) — surface why instead of blanking GPU.
			out = mergeByName(out, nvidiaErrorMetrics(nv))
		} else {
			ms := parseNvidiaSMI(nv)
			// Unified-memory NVIDIA (GB10 Grace-Blackwell) has no aggregate VRAM
			// counter — nvidia-smi memory.* reads [N/A]. Derive gpu.vram from the
			// per-process compute-apps memory over the system RAM total.
			if !hasName(ms, "gpu.vram") {
				if apps, ok := runNvidiaComputeApps(ctx); ok {
					if total, ok := systemMemoryTotalMiB(ctx); ok {
						if m, ok := nvidiaVRAMFromComputeApps(apps, total); ok {
							ms = append(ms, m)
						}
					}
				}
			}
			out = mergeByName(out, ms)
		}
	}
	// AMD: amd-smi when present (richer), then amdgpu sysfs fills any gaps. Both
	// only add names not already set, so an NVIDIA or Apple reading is never
	// shadowed, and a non-AMD host contributes nothing.
	out = mergeByName(out, amdGPU(ctx))
	out = withTotalPower(out)
	out = ensureNPUUtil(out)
	if !hasOK(out) {
		return []domain.Metric{
			{Name: "power.cpu", Status: domain.StatusUnavailable, Detail: "no power source"},
			{Name: "gpu.util", Status: domain.StatusUnavailable, Detail: "no gpu source"},
		}
	}
	return out
}

// withTotalPower appends the whole-machine power.total so the panel can headline
// real draw. On Apple it is already provided directly (SMC PSTR is whole-system),
// so this leaves it untouched. Everywhere else power.cpu (the RAPL package) and
// power.gpu are physically separate rails — the same split btop/top show — so the
// total is their sum plus any NPU rail.
func withTotalPower(ms []domain.Metric) []domain.Metric {
	if hasName(ms, "power.total") {
		return ms
	}
	cpu, _ := okGauge(ms, "power.cpu")
	gpu, _ := okGauge(ms, "power.gpu")
	npu, _ := okGauge(ms, "power.npu")
	total := cpu + gpu + npu
	if total <= 0 {
		return ms
	}
	return append(ms, domain.Metric{Name: "power.total", Unit: "watts", Status: domain.StatusOK, Gauge: total, Detail: "system total"})
}

// okGauge returns the gauge of an OK metric by name.
func okGauge(ms []domain.Metric, name string) (float64, bool) {
	for _, m := range ms {
		if m.Name == name && m.Status == domain.StatusOK {
			return m.Gauge, true
		}
	}
	return 0, false
}

// ensureNPUUtil guarantees an npu.util reading. NPUs (Apple ANE, Intel AI Boost,
// AMD XDNA) expose no utilisation counter, so if no source set it, report it
// Unavailable-with-reason rather than leaving a bare dash. A reading already set
// by a platform collector (e.g. the AMD path's own reason) is preserved.
func ensureNPUUtil(ms []domain.Metric) []domain.Metric {
	if hasName(ms, "npu.util") {
		return ms
	}
	return append(ms, domain.Metric{Name: "npu.util", Status: domain.StatusUnavailable, Detail: "no NPU utilisation counter"})
}

func powerMetric(name string, w float64) domain.Metric {
	return domain.Metric{Name: name, Unit: "watts", Status: domain.StatusOK, Gauge: w}
}

// assembleApplePower builds the macOS power/util metrics from the available
// sources. Per-domain CPU/GPU/ANE power and GPU utilisation come from the
// IOReport energy counters when present. The whole-system total has a strict
// precedence: SMC PSTR ("System Total Power") first, then powermetrics, then the
// IOReport per-domain sum as a last resort. The ordering matters: on Apple
// Silicon Pro/Max chips IOReport reports 0 for CPU/ANE and only a sub-watt GPU
// figure, so the energy-sum is a phantom — it must never shadow a real SMC or
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
	// SMC PSTR is the authoritative whole-system total and needs no root.
	if smcOK && smcPkg > 0 {
		out = append(out, powerMetric("power.total", smcPkg))
	}
	// powermetrics fills any name not already set (the total when SMC is absent,
	// plus CPU/GPU/util gaps).
	out = mergeByName(out, pm)
	// Last resort: the IOReport per-domain sum, only when nothing better gave a
	// whole-system figure.
	if !hasName(out, "power.total") {
		sum := 0.0
		if ioOK {
			for _, w := range []float64{cpu, gpu, ane} {
				if w > 0 {
					sum += w
				}
			}
		}
		if sum > 0 {
			out = append(out, powerMetric("power.total", sum))
		}
	}
	// Apple Pro/Max SoCs expose no per-domain CPU power — IOReport and powermetrics
	// both report 0 for the CPU/ANE channels (only the SMC whole-system total is
	// real). Base M-series populate it. Say why rather than leaving power.cpu a
	// silent blank; power.total still carries the real whole-machine figure.
	if ioOK && !hasName(out, "power.cpu") {
		out = append(out, domain.Metric{
			Name: "power.cpu", Status: domain.StatusUnavailable,
			Detail: "Pro/Max: no per-domain CPU power",
		})
	}
	// Apple Silicon is unified memory — there is no discrete VRAM to read. Report
	// gpu.vram as Unavailable-with-reason so the panel explains the dash instead
	// of silently omitting the metric.
	out = append(out, domain.Metric{
		Name: "gpu.vram", Status: domain.StatusUnavailable,
		Detail: "unified memory (no discrete VRAM)",
	})
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

// runNvidiaSMI queries the NVIDIA GPU. It distinguishes "not present" (nvidia-smi
// not on PATH → present=false, contribute nothing) from "present but failed"
// (present=true, err!=nil, out carries the reason) so the caller can surface a
// broken driver instead of silently dropping every gpu.* key.
func runNvidiaSMI(ctx context.Context) (out string, present bool, err error) {
	path, lookErr := exec.LookPath("nvidia-smi")
	if lookErr != nil {
		return "", false, lookErr
	}
	b, runErr := exec.CommandContext(ctx, path,
		"--query-gpu=utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw,"+
			"clocks.current.graphics,utilization.memory,fan.speed",
		"--format=csv,noheader,nounits").Output()
	if runErr != nil {
		reason := strings.TrimSpace(string(b))
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			if s := strings.TrimSpace(string(ee.Stderr)); s != "" {
				reason = s
			}
		}
		if reason == "" {
			reason = runErr.Error()
		}
		return reason, true, runErr
	}
	return string(b), true, nil
}

// runNvidiaComputeApps returns per-process GPU memory (MiB, one per line) for the
// unified-memory VRAM fallback. Absent or failing nvidia-smi yields ok=false.
func runNvidiaComputeApps(ctx context.Context) (string, bool) {
	path, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return "", false
	}
	out, err := exec.CommandContext(ctx, path,
		"--query-compute-apps=used_memory", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

// systemMemoryTotalMiB reports total system RAM, the ceiling of the shared pool
// on a unified-memory host. Used only as the denominator for the NVIDIA
// compute-apps VRAM fallback.
func systemMemoryTotalMiB(ctx context.Context) (float64, bool) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil || vm == nil || vm.Total == 0 {
		return 0, false
	}
	return float64(vm.Total) / (1024 * 1024), true
}

// amdGPU collects AMD GPU metrics, preferring amd-smi and falling back to amdgpu
// sysfs for any name amd-smi did not provide (or when amd-smi is absent).
func amdGPU(ctx context.Context) []domain.Metric {
	var out []domain.Metric
	if text, err := runAmdSMI(ctx); err == nil {
		out = parseAmdSMICSV(text)
	}
	return mergeByName(out, amdGPUSysfs())
}

func runAmdSMI(ctx context.Context) (string, error) {
	path, err := exec.LookPath("amd-smi")
	if err != nil {
		return "", err
	}
	out, err := exec.CommandContext(ctx, path, "metric",
		"--usage", "--power", "--temperature", "--mem-usage", "--clock", "--csv").Output()
	return string(out), err
}
