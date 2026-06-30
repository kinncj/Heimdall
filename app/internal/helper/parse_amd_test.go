// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

func TestParseAmdSMICSV_HappyPath(t *testing.T) {
	// amd-smi metric --csv with usage, power, temperature, mem-usage. Header
	// names are matched by token so exact wording can drift across versions.
	text := "gpu,gfx_activity,socket_power,edge_temperature,used_vram,total_vram\n" +
		"0,37,55.0,48,4096,16384\n"
	got := byName(parseAmdSMICSV(text))

	if m, ok := got["gpu.util"]; !ok || m.Status != domain.StatusOK || m.Gauge != 37 {
		t.Errorf("gpu.util = %+v, want 37 OK", m)
	}
	if m, ok := got["power.gpu"]; !ok || m.Status != domain.StatusOK || m.Gauge != 55.0 {
		t.Errorf("power.gpu = %+v, want 55 OK", m)
	}
	if m, ok := got["gpu.temp"]; !ok || m.Status != domain.StatusOK || m.Gauge != 48 {
		t.Errorf("gpu.temp = %+v, want 48 OK", m)
	}
	m, ok := got["gpu.vram"]
	if !ok || m.Status != domain.StatusOK || m.Gauge != 25 { // 4096/16384 = 25%
		t.Errorf("gpu.vram = %+v, want 25%% OK", m)
	}
	if !strings.Contains(m.Detail, "GB") {
		t.Errorf("gpu.vram detail = %q, want a GB figure", m.Detail)
	}
}

func TestParseAmdSMICSV_ReorderedAndExtraColumns(t *testing.T) {
	// Columns in a different order, with unrelated extra columns present.
	text := "gpu,total_vram,fan,edge_temperature,used_vram,socket_power,gfx_activity\n" +
		"0,16384,30,50,8192,12.5,90\n"
	got := byName(parseAmdSMICSV(text))
	if m := got["gpu.util"]; m.Gauge != 90 {
		t.Errorf("gpu.util = %v, want 90", m.Gauge)
	}
	if m := got["power.gpu"]; m.Gauge != 12.5 {
		t.Errorf("power.gpu = %v, want 12.5", m.Gauge)
	}
	if m := got["gpu.vram"]; m.Gauge != 50 { // 8192/16384
		t.Errorf("gpu.vram = %v, want 50", m.Gauge)
	}
}

func TestParseAmdSMICSV_EmptyOrGarbage(t *testing.T) {
	for _, in := range []string{"", "   \n", "no,header,row\n", "gpu,gfx_activity\n"} {
		if ms := parseAmdSMICSV(in); len(ms) != 0 {
			t.Errorf("parseAmdSMICSV(%q) = %v, want none", in, ms)
		}
	}
}

func TestAmdVramMetric(t *testing.T) {
	m, ok := amdVramMetric(4096, 16384)
	if !ok || m.Gauge != 25 {
		t.Errorf("amdVramMetric(4096,16384) = %+v %v, want 25%%", m, ok)
	}
	if !strings.Contains(m.Detail, "4.0") || !strings.Contains(m.Detail, "16.0") {
		t.Errorf("detail = %q, want 4.0 / 16.0 GB", m.Detail)
	}
	if _, ok := amdVramMetric(100, 0); ok {
		t.Error("zero total should not produce a vram metric")
	}
}

func TestAmdDoesNotShadowExistingSource(t *testing.T) {
	// An earlier source (e.g. Apple IOReport) already set power.gpu; the AMD
	// reading must not overwrite it.
	primary := []domain.Metric{{Name: "power.gpu", Unit: "watts", Status: domain.StatusOK, Gauge: 99}}
	amd := parseAmdSMICSV("gpu,socket_power\n0,12.5\n")
	got := byName(mergeByName(primary, amd))
	if got["power.gpu"].Gauge != 99 {
		t.Errorf("power.gpu = %v, want the earlier 99 kept", got["power.gpu"].Gauge)
	}
}

func TestParseMicrowatts(t *testing.T) {
	if w, ok := parseMicrowatts(" 15000000\n"); !ok || w != 15 {
		t.Errorf("parseMicrowatts(15000000) = %v %v, want 15W", w, ok)
	}
	if _, ok := parseMicrowatts("nope"); ok {
		t.Error("garbage should not parse")
	}
	if _, ok := parseMicrowatts("0"); ok {
		t.Error("zero microwatts should be treated as no reading")
	}
}
