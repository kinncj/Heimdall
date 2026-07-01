// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"fmt"
	"strconv"
	"strings"

	"heimdall/app/internal/domain"
)

// parseNvidiaSMI parses one CSV row produced by:
//
//	nvidia-smi --query-gpu=utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw,\
//	  clocks.current.graphics,utilization.memory,fan.speed --format=csv,noheader,nounits
//
// Fields, in order: utilization %, memory used (MiB), memory total (MiB),
// temperature (C), power draw (W), graphics clock (MHz), memory-controller
// utilisation %, fan speed %. VRAM is reported as a percentage of total. The
// last three are newer and may read "[N/A]" / "[Not Supported]" on a given SoC
// (e.g. fan on a passively cooled DGX) — those fields fail to parse and are
// simply omitted, so a five-field row from an older query still works.
func parseNvidiaSMI(text string) []domain.Metric {
	var rows [][]string
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		f := strings.Split(ln, ",")
		for i := range f {
			f[i] = strings.TrimSpace(f[i])
		}
		rows = append(rows, f)
	}
	if len(rows) == 0 {
		return nil
	}

	// Aggregate across every GPU (one row each): power sums so power.total is
	// right, util/clock/mem-util average, temp/fan take the hottest, and VRAM is
	// the pooled used/total. A single-GPU host is just the one-row case.
	var utilSum, utilN, usedSum, totalSum, clkSum, clkN, memSum, memN float64
	var powerSum float64
	var tempMax, fanMax float64
	var powerOK, tempOK, fanOK bool
	for _, f := range rows {
		if v, ok := parseField(f, 0); ok {
			utilSum, utilN = utilSum+v, utilN+1
		}
		used, uok := parseField(f, 1)
		total, tok := parseField(f, 2)
		if uok && tok && total > 0 {
			usedSum, totalSum = usedSum+used, totalSum+total
		}
		if v, ok := parseField(f, 3); ok && (!tempOK || v > tempMax) {
			tempMax, tempOK = v, true
		}
		if v, ok := parseField(f, 4); ok {
			powerSum, powerOK = powerSum+v, true
		}
		if v, ok := parseField(f, 5); ok {
			clkSum, clkN = clkSum+v, clkN+1
		}
		if v, ok := parseField(f, 6); ok {
			memSum, memN = memSum+v, memN+1
		}
		if v, ok := parseField(f, 7); ok && (!fanOK || v > fanMax) {
			fanMax, fanOK = v, true
		}
	}

	suffix := ""
	if len(rows) > 1 {
		suffix = fmt.Sprintf(" (%d GPUs)", len(rows))
	}

	var out []domain.Metric
	if utilN > 0 {
		out = append(out, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: utilSum / utilN})
	}
	if totalSum > 0 {
		out = append(out, domain.Metric{
			Name: "gpu.vram", Unit: "percent", Status: domain.StatusOK,
			Gauge:  usedSum / totalSum * 100,
			Detail: fmt.Sprintf("%.1f / %.1f GB%s", usedSum/1024, totalSum/1024, suffix),
		})
	}
	if tempOK {
		out = append(out, domain.Metric{Name: "gpu.temp", Unit: "celsius", Status: domain.StatusOK, Gauge: tempMax})
	}
	if powerOK {
		out = append(out, domain.Metric{Name: "power.gpu", Unit: "watts", Status: domain.StatusOK, Gauge: powerSum})
	}
	if clkN > 0 {
		out = append(out, domain.Metric{Name: "gpu.clock", Unit: "mhz", Status: domain.StatusOK, Gauge: clkSum / clkN})
	}
	if memN > 0 {
		out = append(out, domain.Metric{Name: "gpu.mem.util", Unit: "percent", Status: domain.StatusOK, Gauge: memSum / memN})
	}
	if fanOK {
		out = append(out, domain.Metric{Name: "gpu.fan", Unit: "percent", Status: domain.StatusOK, Gauge: fanMax})
	}
	return out
}

// nvidiaVRAMFromComputeApps derives gpu.vram on unified-memory NVIDIA hosts
// (GB10 Grace-Blackwell), where nvidia-smi reports memory.used/total as [N/A]
// because there is no discrete VRAM. It sums the per-process GPU memory from
// `nvidia-smi --query-compute-apps=used_memory` and expresses it as a percent of
// the system RAM total — the physical ceiling of the shared LPDDR5X pool. Extra
// CSV columns (e.g. a pid,used_memory query) are tolerated; the first field is
// used_memory in MiB.
// A successful-but-empty query means the GPU is idle (no resident contexts); that
// is a real 0%, not a missing reading — reporting it keeps gpu.vram stable rather
// than flickering in and out as workloads come and go. Only a missing system
// total (the denominator) makes the metric unavailable.
func nvidiaVRAMFromComputeApps(text string, totalMiB float64) (domain.Metric, bool) {
	if totalMiB <= 0 {
		return domain.Metric{}, false
	}
	var usedMiB float64
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(strings.Split(ln, ",")[0]), 64)
		if err != nil {
			continue
		}
		usedMiB += v
	}
	return domain.Metric{
		Name: "gpu.vram", Unit: "percent", Status: domain.StatusOK,
		Gauge:  usedMiB / totalMiB * 100,
		Detail: fmt.Sprintf("%.0f/%.0f GB shared", usedMiB/1024, totalMiB/1024),
	}, true
}

// nvidiaErrorMetrics turns a failed nvidia-smi invocation into a visible reason
// rather than a silent gap. nvidia-smi is present but exited non-zero — a common
// cause is a driver/library version mismatch after upgrading the NVIDIA driver
// without rebooting — so gpu.util and gpu.vram are reported Unavailable carrying
// the reason instead of disappearing.
func nvidiaErrorMetrics(reason string) []domain.Metric {
	detail := "nvidia-smi: " + firstNonEmptyLine(reason)
	return []domain.Metric{
		{Name: "gpu.util", Status: domain.StatusUnavailable, Detail: detail},
		{Name: "gpu.vram", Status: domain.StatusUnavailable, Detail: detail},
	}
}

func parseField(f []string, i int) (float64, bool) {
	if i >= len(f) {
		return 0, false
	}
	v, err := strconv.ParseFloat(f[i], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func firstNonEmptyLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			return strings.TrimSpace(ln)
		}
	}
	return ""
}
