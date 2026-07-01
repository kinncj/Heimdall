// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

//go:build linux

package helper

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"heimdall/app/internal/domain"
)

// amdGPUSysfs reads GPU metrics straight from the in-tree amdgpu driver's sysfs
// nodes — no amd-smi / ROCm install required, and readable without root. It is
// the fallback for any name amd-smi did not provide. When no amdgpu card is
// present it returns nil, so a non-AMD host contributes nothing.
//
//   - device/gpu_busy_percent      -> gpu.util  (percent)
//   - device/mem_info_vram_used    -> gpu.vram  (percent of total, bytes)
//   - device/mem_info_vram_total
//   - hwmon/power1_average         -> power.gpu (microwatts)
//   - hwmon/temp1_input            -> gpu.temp  (millidegrees C)
//
// The XDNA NPU (amdxdna) has no stable utilisation counter yet; when an amdgpu
// card is found we still surface npu.util as Unavailable so the host advertises
// the accelerator without faking a number.
func amdGPUSysfs() []domain.Metric {
	dev := firstAmdgpuDevice()
	if dev == "" {
		return nil
	}
	var out []domain.Metric

	if v, ok := readSysfsFloat(filepath.Join(dev, "gpu_busy_percent")); ok {
		out = append(out, domain.Metric{Name: "gpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: v})
	}
	used, uok := readSysfsFloat(filepath.Join(dev, "mem_info_vram_used"))
	total, tok := readSysfsFloat(filepath.Join(dev, "mem_info_vram_total"))
	if uok && tok {
		// sysfs reports bytes; amdVramMetric wants MiB.
		if m, ok := amdVramMetric(used/(1024*1024), total/(1024*1024)); ok {
			out = append(out, m)
		}
	}
	if hwmon := firstDir(filepath.Join(dev, "hwmon", "hwmon*")); hwmon != "" {
		if b, err := os.ReadFile(filepath.Join(hwmon, "power1_average")); err == nil {
			if w, ok := parseMicrowatts(string(b)); ok {
				out = append(out, powerMetric("power.gpu", w))
			}
		}
		if b, err := os.ReadFile(filepath.Join(hwmon, "temp1_input")); err == nil {
			if c, err := parseMilliCelsius(string(b)); err == nil && c > 0 {
				out = append(out, domain.Metric{Name: "gpu.temp", Unit: "celsius", Status: domain.StatusOK, Gauge: c})
			}
		}
		// freq1_input is the gfx (sclk) clock in Hz.
		if b, err := os.ReadFile(filepath.Join(hwmon, "freq1_input")); err == nil {
			if mhz, ok := hzToMHz(string(b)); ok {
				out = append(out, domain.Metric{Name: "gpu.clock", Unit: "mhz", Status: domain.StatusOK, Gauge: mhz})
			}
		}
		// pwm1 is the fan duty cycle (0–255); report it as a percentage. Absent on
		// APUs with no dedicated GPU fan (shared cooling).
		if v, ok := readSysfsFloat(filepath.Join(hwmon, "pwm1")); ok {
			if pct, ok := pwmToFanPercent(v); ok {
				out = append(out, domain.Metric{Name: "gpu.fan", Unit: "percent", Status: domain.StatusOK, Gauge: pct})
			}
		}
	}

	// Best-effort NPU: no stable amdxdna util counter — advertise as Unavailable.
	out = append(out, domain.Metric{Name: "npu.util", Status: domain.StatusUnavailable, Detail: "amdxdna exposes no util counter"})
	return out
}

// firstAmdgpuDevice returns the device sysfs dir of the first DRM card whose
// driver is amdgpu, or "" if none. Connector entries (card0-DP-1, with a hyphen)
// are skipped; only the bare cardN render nodes are considered.
func firstAmdgpuDevice() string {
	cards, _ := filepath.Glob("/sys/class/drm/card[0-9]")
	for _, card := range cards {
		dev := filepath.Join(card, "device")
		b, err := os.ReadFile(filepath.Join(dev, "uevent"))
		if err == nil && strings.Contains(string(b), "DRIVER=amdgpu") {
			return dev
		}
	}
	return ""
}

func readSysfsFloat(path string) (float64, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func firstDir(glob string) string {
	matches, _ := filepath.Glob(glob)
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}
