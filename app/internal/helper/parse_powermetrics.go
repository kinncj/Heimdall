// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"regexp"
	"strconv"

	"heimdall/app/internal/domain"
)

var (
	reCPUPower = regexp.MustCompile(`(?mi)^\s*CPU Power:\s*([\d.]+)\s*mW`)
	reGPUPower = regexp.MustCompile(`(?mi)^\s*GPU Power:\s*([\d.]+)\s*mW`)
	reCombined = regexp.MustCompile(`(?mi)^\s*Combined Power[^:]*:\s*([\d.]+)\s*mW`)
	rePackage  = regexp.MustCompile(`(?mi)^\s*Package Power:\s*([\d.]+)\s*mW`)
	reECluster = regexp.MustCompile(`(?mi)^\s*E-Cluster Power:\s*([\d.]+)\s*mW`)
	rePCluster = regexp.MustCompile(`(?mi)^\s*P-Cluster Power:\s*([\d.]+)\s*mW`)
	reGPUResid = regexp.MustCompile(`(?mi)^\s*GPU (?:active|HW active) residency:\s*([\d.]+)\s*%`)
)

// parsePowermetrics extracts power and GPU metrics from `powermetrics` text.
// Powers convert mW -> W. On Apple Silicon CPU power is sometimes reported only
// per cluster, so the E-Cluster + P-Cluster sum is used when it exceeds (or
// replaces) the explicit CPU Power line. When the combined/package line is
// absent it derives package power from CPU+GPU. Absent fields are omitted.
func parsePowermetrics(text string) []domain.Metric {
	var out []domain.Metric
	cpu, cpuOK := matchFloat(reCPUPower, text)
	// Some Apple Silicon SoCs report CPU power only per cluster; sum E + P.
	e, eOK := matchFloat(reECluster, text)
	p, pOK := matchFloat(rePCluster, text)
	if eOK || pOK {
		if sum := e + p; sum > cpu {
			cpu, cpuOK = sum, true
		}
	}
	// On several SoCs powermetrics prints "CPU Power: 0 mW" despite an active CPU
	// (the figure simply isn't exposed). Treat a hard zero with no cluster data
	// as unavailable rather than a misleading 0 W.
	cpuGap := cpuOK && cpu == 0

	gpu, gpuOK := matchFloat(reGPUPower, text)

	pkg, pkgOK := matchFloat(reCombined, text)
	if !pkgOK {
		pkg, pkgOK = matchFloat(rePackage, text)
	}
	if !pkgOK {
		// Derive package power from the components actually measured, skipping a
		// CPU-power telemetry gap so it never contributes a phantom zero.
		sum, have := 0.0, false
		if cpuOK && !cpuGap {
			sum, have = sum+cpu, true
		}
		if gpuOK {
			sum, have = sum+gpu, true
		}
		if have {
			pkg, pkgOK = sum, true
		}
	}

	switch {
	case cpuGap:
		out = append(out, domain.Metric{Name: "power.cpu", Status: domain.StatusUnavailable, Detail: "not exposed by SoC"})
	case cpuOK:
		out = append(out, watts("power.cpu", cpu))
	}
	if gpuOK {
		out = append(out, watts("power.gpu", gpu))
	}
	if pkgOK {
		out = append(out, watts("power.total", pkg))
	}
	if util, ok := matchFloat(reGPUResid, text); ok {
		out = append(out, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: util})
	}
	return out
}

func watts(name string, mW float64) domain.Metric {
	return domain.Metric{Name: name, Unit: "watts", Status: domain.StatusOK, Gauge: mW / 1000}
}

func matchFloat(re *regexp.Regexp, s string) (float64, bool) {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
