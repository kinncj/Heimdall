// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"strconv"
	"strings"

	"heimdall/app/internal/domain"
)

// parseNvidiaSMI parses one CSV row produced by:
//
//	nvidia-smi --query-gpu=utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw \
//	  --format=csv,noheader,nounits
//
// Fields, in order: utilization %, memory used (MiB), memory total (MiB),
// temperature (C), power draw (W). VRAM is reported as a percentage of total.
func parseNvidiaSMI(text string) []domain.Metric {
	line := firstNonEmptyLine(text)
	if line == "" {
		return nil
	}
	f := strings.Split(line, ",")
	for i := range f {
		f[i] = strings.TrimSpace(f[i])
	}
	var out []domain.Metric
	if v, ok := parseField(f, 0); ok {
		out = append(out, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: v})
	}
	used, uok := parseField(f, 1)
	total, tok := parseField(f, 2)
	if uok && tok && total > 0 {
		out = append(out, domain.Metric{Name: "gpu.vram", Unit: "percent", Status: domain.StatusOK, Gauge: used / total * 100})
	}
	if v, ok := parseField(f, 3); ok {
		out = append(out, domain.Metric{Name: "gpu.temp", Unit: "celsius", Status: domain.StatusOK, Gauge: v})
	}
	if v, ok := parseField(f, 4); ok {
		out = append(out, domain.Metric{Name: "power.gpu", Unit: "watts", Status: domain.StatusOK, Gauge: v})
	}
	return out
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
