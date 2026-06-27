// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

func TestResolveOutputDefaultsToTTY(t *testing.T) {
	lg, mw, mjson, _, err := resolveOutput("", true)
	if err != nil || mw != os.Stdout || lg == nil {
		t.Fatalf("default: mw=%v err=%v; want stdout", mw, err)
	}
	if !mjson {
		t.Error("default should honour the --json flag")
	}
}

func TestResolveOutputFalseDiscardsEverything(t *testing.T) {
	for _, v := range []string{"false", "FALSE", " off ", "none", "0"} {
		_, mw, mjson, _, err := resolveOutput(v, true)
		if err != nil || mw != io.Discard || mjson {
			t.Fatalf("resolveOutput(%q): mw=%v json=%v err=%v; want discard, json=false", v, mw, mjson, err)
		}
	}
}

func TestFormatMetricInfoUsesDetail(t *testing.T) {
	// Info metrics (host.os, host.cpu, …) carry their value as a Detail string
	// with no gauge. They must render the string, not the zero gauge.
	m := domain.Metric{Name: "host.os", Unit: "info", Status: domain.StatusOK, Detail: "macOS 15.0"}
	got := formatMetric(m)
	want := "host.os=macOS 15.0"
	if got != want {
		t.Fatalf("formatMetric(info) = %q, want %q", got, want)
	}
}

func TestEmitJSONInfoCarriesDetail(t *testing.T) {
	var b strings.Builder
	emit(&b, "host-a", []domain.Metric{{Name: "host.os", Unit: "info", Status: domain.StatusOK, Detail: "macOS 15.0"}}, true)
	out := b.String()
	if !strings.Contains(out, `"detail":"macOS 15.0"`) {
		t.Fatalf("emit JSON info dropped the detail string: %s", out)
	}
}

func TestResolveOutputFileWritesJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d.log")
	lg, mw, mjson, closeFn, err := resolveOutput(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if !mjson {
		t.Error("a file destination should force JSON metric output")
	}
	lg.Info("hello", "k", "v")
	emit(mw, "host-a", []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 5, Unit: "percent"}}, mjson)
	if err := closeFn(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{`"msg":"hello"`, `"component":"heimdall-daemon"`, `"metric":"cpu.util"`} {
		if !strings.Contains(s, want) {
			t.Errorf("log file missing %s in:\n%s", want, s)
		}
	}
}
