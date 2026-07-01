// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func byName(ms []domain.Metric) map[string]domain.Metric {
	m := make(map[string]domain.Metric, len(ms))
	for _, x := range ms {
		m[x.Name] = x
	}
	return m
}

func TestParsePowermetrics(t *testing.T) {
	sample := `
*** Sampled system activity ***

**** GPU usage ****
GPU HW active frequency: 444 MHz
GPU active residency:  37.50% (444 MHz: 37%)
GPU Power: 3100 mW

**** Processor usage ****
CPU Power: 8400 mW
GPU Power: 3100 mW
ANE Power: 0 mW
Combined Power (CPU + GPU + ANE): 12700 mW
`
	got := byName(parsePowermetrics(sample))
	if m := got["power.cpu"]; m.Status != domain.StatusOK || m.Gauge != 8.4 {
		t.Errorf("power.cpu = %+v, want 8.4W ok", m)
	}
	if m := got["power.gpu"]; m.Status != domain.StatusOK || m.Gauge != 3.1 {
		t.Errorf("power.gpu = %+v, want 3.1W ok", m)
	}
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 12.7 {
		t.Errorf("power.total = %+v, want 12.7W ok", m)
	}
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 37.5 {
		t.Errorf("gpu.util = %+v, want 37.5%% ok", m)
	}
}

func TestParsePowermetricsDerivesPackageFromParts(t *testing.T) {
	sample := "CPU Power: 5000 mW\nGPU Power: 2000 mW\n"
	got := byName(parsePowermetrics(sample))
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 7.0 {
		t.Errorf("derived power.total = %+v, want 7.0W", m)
	}
}

func TestParsePowermetricsCPUFromClusters(t *testing.T) {
	// Apple Silicon can report a low/zero explicit CPU Power while the per-cluster
	// lines carry the real figure; the parser should use the cluster sum.
	sample := "E-Cluster Power: 1200 mW\nP-Cluster Power: 3400 mW\nCPU Power: 0 mW\nGPU Power: 560 mW\n"
	got := byName(parsePowermetrics(sample))
	if m := got["power.cpu"]; m.Status != domain.StatusOK || m.Gauge != 4.6 {
		t.Errorf("power.cpu = %+v, want 4.6W (E+P clusters)", m)
	}
	if m := got["gpu.util"]; m.Status == domain.StatusOK && m.Unit != "percent" {
		t.Errorf("gpu.util unit = %q, want percent", m.Unit)
	}
}

func TestParsePowermetricsCPUPowerGap(t *testing.T) {
	// Real Apple Silicon shape: CPU Power reads 0 with no cluster power line and
	// no combined line; only GPU (and a zero ANE) are exposed.
	sample := "CPU Power: 0 mW\nGPU Power: 570 mW\nANE Power: 0 mW\nGPU HW active residency:  24.00%\n"
	got := byName(parsePowermetrics(sample))
	if m := got["power.cpu"]; m.Status != domain.StatusUnavailable {
		t.Errorf("power.cpu = %+v, want unavailable (telemetry gap)", m)
	}
	if m := got["power.gpu"]; m.Status != domain.StatusOK || m.Gauge != 0.57 {
		t.Errorf("power.gpu = %+v, want 0.57W", m)
	}
	if m := got["power.total"]; m.Status != domain.StatusOK || m.Gauge != 0.57 {
		t.Errorf("power.total = %+v, want 0.57W (measured components only)", m)
	}
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 24 {
		t.Errorf("gpu.util = %+v, want 24%%", m)
	}
}

func TestParseNvidiaSMI(t *testing.T) {
	got := byName(parseNvidiaSMI("42, 2048, 8192, 63, 88.5\n"))
	if m := got["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 42 {
		t.Errorf("gpu.util = %+v, want 42", m)
	}
	if m := got["gpu.vram"]; m.Status != domain.StatusOK || m.Gauge != 25 {
		t.Errorf("gpu.vram = %+v, want 25%%", m)
	}
	if m := got["gpu.temp"]; m.Status != domain.StatusOK || m.Gauge != 63 {
		t.Errorf("gpu.temp = %+v, want 63C", m)
	}
	if m := got["power.gpu"]; m.Status != domain.StatusOK || m.Gauge != 88.5 {
		t.Errorf("power.gpu = %+v, want 88.5W", m)
	}
}

func TestServerClientRoundTrip(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "h.sock")
	want := []domain.Metric{
		{Name: "power.total", Unit: "W", Status: domain.StatusOK, Gauge: 11.5},
		{Name: "gpu.util", Unit: "%", Status: domain.StatusOK, Gauge: 40},
	}
	srv := &Server{SockPath: sock, Collect: func(context.Context) []domain.Metric { return want }}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	var got []domain.Metric
	var err error
	for time.Now().Before(deadline) {
		got, err = (Client{SockPath: sock}).Collect(context.Background())
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	g := byName(got)
	if m := g["power.total"]; m.Status != domain.StatusOK || m.Gauge != 11.5 {
		t.Errorf("power.total = %+v", m)
	}
	if m := g["gpu.util"]; m.Status != domain.StatusOK || m.Gauge != 40 {
		t.Errorf("gpu.util = %+v", m)
	}
}

func TestClientUnavailableWhenAbsent(t *testing.T) {
	_, err := (Client{SockPath: filepath.Join(t.TempDir(), "missing.sock")}).Collect(context.Background())
	if err != ErrUnavailable {
		t.Fatalf("got %v, want ErrUnavailable", err)
	}
}
