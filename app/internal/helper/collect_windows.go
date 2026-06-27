// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build windows

package helper

import (
	"context"
	"os/exec"

	"heimdall/app/internal/domain"
)

// windowsPrivileged reads CPU/zone temperature from WMI via PowerShell, matching
// the helper's existing shell-out pattern (no new dependency). CPU package power
// is not available on Windows without a kernel driver (RAPL is inaccessible), so
// power.pkg stays Unavailable through the daemon's normal adapter path.
func windowsPrivileged(ctx context.Context) []domain.Metric {
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command",
		"Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature "+
			"| Select-Object -ExpandProperty CurrentTemperature").Output()
	if err != nil {
		return nil
	}
	if c, ok := parseThermalZoneCelsius(string(out)); ok {
		return []domain.Metric{{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: c}}
	}
	return nil
}
