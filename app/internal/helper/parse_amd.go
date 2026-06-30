// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"fmt"
	"strconv"
	"strings"

	"heimdall/app/internal/domain"
)

// parseAmdSMICSV parses one row of CSV produced by amd-smi, e.g.
//
//	amd-smi metric --usage --power --temperature --mem-usage --csv
//
// Header names differ across amd-smi releases (gfx_activity vs gpu_use,
// socket_power vs power, edge_temperature vs temp_edge), so columns are matched
// by token rather than by a fixed position. Anything we cannot map is ignored,
// and a missing column simply yields no metric — the caller fills the gap from
// amdgpu sysfs. Only the first GPU row is read (index 0).
func parseAmdSMICSV(text string) []domain.Metric {
	lines := nonEmptyLines(text)
	if len(lines) < 2 {
		return nil
	}
	head := splitCSV(lines[0])
	row := splitCSV(lines[1])
	col := func(match func(string) bool) (float64, bool) {
		for i, h := range head {
			if i < len(row) && match(h) {
				if v, err := strconv.ParseFloat(strings.TrimSpace(row[i]), 64); err == nil {
					return v, true
				}
			}
		}
		return 0, false
	}

	var out []domain.Metric
	if v, ok := col(isAmdUtilHeader); ok {
		out = append(out, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: v})
	}
	if v, ok := col(isAmdPowerHeader); ok && v > 0 {
		out = append(out, powerMetric("power.gpu", v))
	}
	if v, ok := col(isAmdTempHeader); ok && v > 0 {
		out = append(out, domain.Metric{Name: "gpu.temp", Unit: "celsius", Status: domain.StatusOK, Gauge: v})
	}
	used, uok := col(isAmdVramUsedHeader)
	total, tok := col(isAmdVramTotalHeader)
	if uok && tok {
		if m, ok := amdVramMetric(used, total); ok {
			out = append(out, m)
		}
	}
	if v, ok := col(isAmdClockHeader); ok && v > 0 {
		out = append(out, domain.Metric{Name: "gpu.clock", Unit: "mhz", Status: domain.StatusOK, Gauge: v})
	}
	if v, ok := col(isAmdMemUtilHeader); ok {
		out = append(out, domain.Metric{Name: "gpu.mem.util", Unit: "percent", Status: domain.StatusOK, Gauge: v})
	}
	return out
}

// hzToMHz converts a sysfs clock value in hertz to megahertz. Zero or
// unparseable input is reported as no reading.
func hzToMHz(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || v <= 0 {
		return 0, false
	}
	return v / 1e6, true
}

// amdVramMetric builds the gpu.vram percentage metric from used/total VRAM in
// MiB. A non-positive total yields no metric (we cannot compute a percentage).
func amdVramMetric(usedMiB, totalMiB float64) (domain.Metric, bool) {
	if totalMiB <= 0 {
		return domain.Metric{}, false
	}
	return domain.Metric{
		Name: "gpu.vram", Unit: "percent", Status: domain.StatusOK,
		Gauge:  usedMiB / totalMiB * 100,
		Detail: fmt.Sprintf("%.1f / %.1f GB", usedMiB/1024, totalMiB/1024),
	}, true
}

// parseMicrowatts parses an amdgpu hwmon power1_average value (microwatts) into
// watts. Zero or unparseable input is reported as no reading.
func parseMicrowatts(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || v <= 0 {
		return 0, false
	}
	return v / 1e6, true
}

func isAmdUtilHeader(h string) bool {
	h = strings.ToLower(h)
	return strings.Contains(h, "gfx_activity") ||
		(strings.Contains(h, "gpu") && (strings.Contains(h, "use") || strings.Contains(h, "util"))) ||
		strings.Contains(h, "gfx_util")
}

func isAmdPowerHeader(h string) bool {
	h = strings.ToLower(h)
	return strings.Contains(h, "socket_power") || h == "power" ||
		strings.Contains(h, "power_w") || strings.Contains(h, "gpu_power")
}

func isAmdTempHeader(h string) bool {
	h = strings.ToLower(h)
	return strings.Contains(h, "edge") || strings.Contains(h, "temp_") ||
		strings.Contains(h, "gpu_temp") || strings.Contains(h, "_temperature")
}

func isAmdVramUsedHeader(h string) bool {
	h = strings.ToLower(h)
	return strings.Contains(h, "used_vram") || strings.Contains(h, "vram_used")
}

func isAmdVramTotalHeader(h string) bool {
	h = strings.ToLower(h)
	return strings.Contains(h, "total_vram") || strings.Contains(h, "vram_total")
}

func isAmdClockHeader(h string) bool {
	h = strings.ToLower(h)
	// Graphics/shader clock: gfx_clk, gfx_0_clk, sclk. "clk"/"clock" never appears
	// in the activity, power, temp, or vram columns, so this can't collide.
	return strings.Contains(h, "sclk") ||
		(strings.Contains(h, "gfx") && (strings.Contains(h, "clk") || strings.Contains(h, "clock")))
}

func isAmdMemUtilHeader(h string) bool {
	h = strings.ToLower(h)
	// Memory-controller activity: umc_activity (or mem_activity). Distinct from
	// gfx_activity (GPU util) and from the *_vram size columns.
	return strings.Contains(h, "umc_activity") || strings.Contains(h, "mem_activity")
}

func splitCSV(line string) []string {
	f := strings.Split(line, ",")
	for i := range f {
		f[i] = strings.TrimSpace(f[i])
	}
	return f
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			out = append(out, strings.TrimSpace(ln))
		}
	}
	return out
}
