// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"testing"

	"heimdall/app/internal/domain"
)

func TestParseNvidiaSMI_ClockMemUtilFan(t *testing.T) {
	// Extended row: util, mem.used, mem.total, temp, power, clock(MHz),
	// mem.util(%), fan(%).
	got := byName(parseNvidiaSMI("42, 2048, 8192, 63, 88.5, 1410, 55, 37\n"))
	if m := got["gpu.clock"]; m.Status != domain.StatusOK || m.Gauge != 1410 {
		t.Errorf("gpu.clock = %+v, want 1410 MHz", m)
	}
	if m := got["gpu.mem.util"]; m.Status != domain.StatusOK || m.Gauge != 55 {
		t.Errorf("gpu.mem.util = %+v, want 55%%", m)
	}
	if m := got["gpu.fan"]; m.Status != domain.StatusOK || m.Gauge != 37 {
		t.Errorf("gpu.fan = %+v, want 37%%", m)
	}
}

func TestParseNvidiaSMI_NAFieldsSkipped(t *testing.T) {
	// Fan reads [N/A] (passively cooled / no sensor); it must be skipped, the rest kept.
	got := byName(parseNvidiaSMI("10, 1024, 8192, 50, 30.0, 900, 20, [N/A]\n"))
	if _, ok := got["gpu.fan"]; ok {
		t.Error("gpu.fan should be absent when nvidia-smi reports [N/A]")
	}
	if m := got["gpu.clock"]; m.Status != domain.StatusOK || m.Gauge != 900 {
		t.Errorf("gpu.clock = %+v, want 900", m)
	}
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 10 {
		t.Errorf("gpu.util = %+v, want 10", m)
	}
}

func TestParseNvidiaSMI_OldFiveFieldRowStillWorks(t *testing.T) {
	// Backward compatibility: a 5-field row (no clock/mem.util/fan) must still parse.
	got := byName(parseNvidiaSMI("42, 2048, 8192, 63, 88.5\n"))
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 42 {
		t.Errorf("gpu.util = %+v, want 42", m)
	}
	if _, ok := got["gpu.clock"]; ok {
		t.Error("gpu.clock should be absent for a 5-field row")
	}
}

func TestParseAmdSMICSV_ClockAndMemUtil(t *testing.T) {
	text := "gpu,gfx_activity,socket_power,edge_temperature,used_vram,total_vram,gfx_clk,umc_activity\n" +
		"0,37,55.0,48,4096,16384,2100,61\n"
	got := byName(parseAmdSMICSV(text))
	if m := got["gpu.clock"]; m.Status != domain.StatusOK || m.Gauge != 2100 {
		t.Errorf("gpu.clock = %+v, want 2100 MHz", m)
	}
	if m := got["gpu.mem.util"]; m.Status != domain.StatusOK || m.Gauge != 61 {
		t.Errorf("gpu.mem.util = %+v, want 61%%", m)
	}
	// gfx_activity must still map to util, not be mistaken for the clock.
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 37 {
		t.Errorf("gpu.util = %+v, want 37", m)
	}
}

func TestParseHz_ToMHz(t *testing.T) {
	if mhz, ok := hzToMHz("2100000000"); !ok || mhz != 2100 {
		t.Errorf("hzToMHz(2.1GHz) = %v %v, want 2100", mhz, ok)
	}
	if _, ok := hzToMHz("0"); ok {
		t.Error("zero Hz should be no reading")
	}
	if _, ok := hzToMHz("nope"); ok {
		t.Error("garbage should not parse")
	}
}
