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
// the helper's existing shell-out pattern (no new dependency). CPU power is not
// available on Windows without a kernel driver (RAPL is inaccessible), so
// power.cpu is reported Unavailable-with-reason; the GPU rail (nvidia-smi) still
// comes through and power.total upstream is then the GPU alone.
func windowsPrivileged(ctx context.Context) []domain.Metric {
	out := []domain.Metric{{Name: "power.cpu", Status: domain.StatusUnavailable, Detail: "no CPU power counter on Windows (RAPL inaccessible)"}}
	b, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command",
		"Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature "+
			"| Select-Object -ExpandProperty CurrentTemperature").Output()
	if err != nil {
		return out
	}
	if c, ok := parseThermalZoneCelsius(string(b)); ok {
		out = append(out, domain.Metric{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusOK, Gauge: c})
	}
	return out
}
