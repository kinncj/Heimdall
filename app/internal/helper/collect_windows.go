// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build windows

package helper

import (
	"context"
	"os/exec"
	"strconv"

	"heimdall/app/internal/domain"
)

// windowsPrivileged reports CPU power and package temperature on Windows.
//
// Windows exposes no RAPL to user space — reading the CPU energy MSRs needs a
// ring-0 driver, which Heimdall does not ship. So CPU power comes from a
// driver-backed monitor the operator already runs: LibreHardwareMonitor publishes
// CPU package power over WMI (root/LibreHardwareMonitor), which we read here (no
// cgo, Intel + AMD). When it isn't running, power.cpu is Unavailable with a reason
// pointing at how to enable it. Package temperature still comes from the ACPI
// thermal zone via WMI.
func windowsPrivileged(ctx context.Context) []domain.Metric {
	var out []domain.Metric
	if w, ok := lhmCPUPackageWatts(ctx); ok {
		out = append(out, domain.Metric{Name: "power.cpu", Unit: "watts", Status: domain.StatusOK, Gauge: w, Detail: "CPU package (LibreHardwareMonitor)"})
	} else {
		out = append(out, domain.Metric{Name: "power.cpu", Status: domain.StatusUnavailable, Detail: "no RAPL on Windows — run LibreHardwareMonitor for CPU power"})
	}
	if b, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command",
		"Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature "+
			"| Select-Object -ExpandProperty CurrentTemperature").Output(); err == nil {
		if c, ok := parseThermalZoneCelsius(string(b)); ok {
			out = append(out, domain.Metric{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: c})
		}
	}
	return out
}

// lhmCPUPackageWatts reads CPU package power from a running LibreHardwareMonitor
// via its WMI provider. LHM ships a signed driver that reads the RAPL MSRs Windows
// won't expose otherwise, and surfaces them as Sensor instances. Absent (0, false)
// when LHM isn't running or exposes no CPU package power sensor.
func lhmCPUPackageWatts(ctx context.Context) (float64, bool) {
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command",
		"$s = Get-CimInstance -Namespace root/LibreHardwareMonitor -ClassName Sensor -ErrorAction SilentlyContinue "+
			"| Where-Object { $_.SensorType -eq 'Power' -and $_.Name -like '*Package*' } "+
			"| Select-Object -First 1 -ExpandProperty Value; if ($s) { $s }").Output()
	if err != nil {
		return 0, false
	}
	line := firstNonEmptyLine(string(out))
	if line == "" {
		return 0, false
	}
	w, err := strconv.ParseFloat(line, 64)
	if err != nil || w <= 0 {
		return 0, false
	}
	return w, true
}
