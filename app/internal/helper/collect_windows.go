// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build windows

package helper

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"heimdall/app/internal/domain"
)

// scaphandreDefaultURL is Scaphandre's Prometheus exporter default. Override with
// HEIMDALL_SCAPHANDRE_URL when it runs elsewhere.
const scaphandreDefaultURL = "http://127.0.0.1:8080/metrics"

// windowsPrivileged reports CPU power and package temperature on Windows.
//
// Windows exposes no RAPL to user space — reading the CPU energy MSRs needs a
// ring-0 driver, which Heimdall does not ship. Scaphandre installs the signed
// Hubblo RAPL driver and runs as a service exposing power over a Prometheus
// endpoint; when it's up we scrape it and report CPU-package power as power.cpu
// (Intel + AMD, pure Go, no cgo). When it isn't, power.cpu is Unavailable with a
// reason pointing at it. Package temperature still comes from the ACPI thermal
// zone via WMI.
func windowsPrivileged(ctx context.Context) []domain.Metric {
	var out []domain.Metric
	if w, ok := scaphandreCPUWatts(ctx); ok {
		out = append(out, domain.Metric{Name: "power.cpu", Unit: "watts", Status: domain.StatusOK, Gauge: w, Detail: "CPU package (Scaphandre)"})
	} else {
		out = append(out, domain.Metric{Name: "power.cpu", Status: domain.StatusUnavailable, Detail: "no RAPL on Windows — run Scaphandre for CPU power"})
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

// scaphandreCPUWatts scrapes a running Scaphandre Prometheus exporter for CPU
// package power. Absent (0, false) when Scaphandre isn't reachable.
func scaphandreCPUWatts(ctx context.Context) (float64, bool) {
	url := os.Getenv("HEIMDALL_SCAPHANDRE_URL")
	if url == "" {
		url = scaphandreDefaultURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, false
	}
	resp, err := (&http.Client{Timeout: 2 * time.Second}).Do(req)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, false
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, false
	}
	return parseScaphandreCPUWatts(string(b))
}
