// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package observe

import (
	"strings"
	"testing"

	"heimdall/app/internal/domain"
)

func TestRenderOpenMetricsGaugeLabelsAndLiveness(t *testing.T) {
	views := []domain.HostView{{
		Host:  domain.Host{ID: "alpha", Context: domain.HostContext{Labels: map[string]string{"env": "prod"}}},
		State: domain.StateOnline,
		LastSnapshot: []domain.Metric{
			{Name: "cpu.util", Unit: "percent", Status: domain.StatusOK, Gauge: 21},
			{Name: "host.os", Unit: "info", Status: domain.StatusOK, Detail: "darwin"},
			{Name: "temp.pkg", Unit: "celsius", Status: domain.StatusInsufficientPermission},
		},
	}}
	out := RenderOpenMetrics(views)
	for _, w := range []string{
		"# TYPE heimdall_cpu_util gauge",
		`heimdall_cpu_util{env="prod",host="alpha"} 21`,
		`heimdall_host_up{env="prod",host="alpha",state="online"} 1`,
	} {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q in:\n%s", w, out)
		}
	}
	if strings.Contains(out, "temp_pkg") {
		t.Errorf("non-OK metric must be skipped:\n%s", out)
	}
	if strings.Contains(out, "heimdall_host_os") {
		t.Errorf("info metric must not be a numeric series:\n%s", out)
	}
}

func TestRenderOpenMetricsPerCore(t *testing.T) {
	views := []domain.HostView{{
		Host:  domain.Host{ID: "beta"},
		State: domain.StateOnline,
		LastSnapshot: []domain.Metric{
			{Name: "cpu.cores", Unit: "percent", Status: domain.StatusOK, Kind: domain.KindPerCore, PerCore: []float64{10, 20}},
		},
	}}
	out := RenderOpenMetrics(views)
	for _, w := range []string{
		`heimdall_cpu_cores{core="0",host="beta"} 10`,
		`heimdall_cpu_cores{core="1",host="beta"} 20`,
	} {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q in:\n%s", w, out)
		}
	}
}
