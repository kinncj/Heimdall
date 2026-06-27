// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

func seed(t *testing.T) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	add := func(id, hub, env string, alerts []string) {
		reg.Enroll(domain.Host{ID: domain.HostID(id), DisplayName: id,
			Context: domain.HostContext{OS: "linux", Labels: map[string]string{"hub": hub, "env": env}}}, now)
		reg.Observe(domain.HostID(id), []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 10}}, nil, now)
		reg.SetAlerts(domain.HostID(id), alerts)
	}
	add("alpha", "home", "prod", []string{"hot-cpu"})
	add("beta", "remote", "dev", nil)
	add("gamma", "home", "dev", nil)
	return reg
}

func TestOrderedHostsGroupsByHub(t *testing.T) {
	m := Model{reg: seed(t), groupBy: 1} // dim 1 == hub
	hosts, groups := m.orderedHosts()
	if len(hosts) != 3 {
		t.Fatalf("got %d hosts, want 3", len(hosts))
	}
	// grouped by hub, then by id: home(alpha,gamma), remote(beta)
	wantG := []string{"home", "home", "remote"}
	wantH := []string{"alpha", "gamma", "beta"}
	for i := range hosts {
		if groups[i] != wantG[i] || string(hosts[i].Host.ID) != wantH[i] {
			t.Fatalf("pos %d = %s/%s, want %s/%s", i, groups[i], hosts[i].Host.ID, wantG[i], wantH[i])
		}
	}
}

func TestFilterNarrowsByTagAndName(t *testing.T) {
	m := Model{reg: seed(t), filter: "prod"}
	if got := m.orderedList(); len(got) != 1 || got[0].Host.ID != "alpha" {
		t.Fatalf("filter prod = %v, want [alpha]", got)
	}
	m = Model{reg: seed(t), filter: "beta"}
	if got := m.orderedList(); len(got) != 1 || got[0].Host.ID != "beta" {
		t.Fatalf("filter beta = %v, want [beta]", got)
	}
	m = Model{reg: seed(t), filter: "nope"}
	if got := m.orderedList(); len(got) != 0 {
		t.Fatalf("no-match filter should be empty, got %v", got)
	}
}

func TestOSGroupingReadsHostOSMetric(t *testing.T) {
	// The dashboard gets OS only as the host.os metric (Context.OS is empty over
	// the wire), so grouping must read the metric, not Context.
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	reg.Enroll(domain.Host{ID: "m", DisplayName: "m"}, now)
	reg.Observe("m", []domain.Metric{
		{Name: "host.os", Unit: "info", Status: domain.StatusOK, Detail: "darwin 27.0"},
		{Name: "cpu.util", Status: domain.StatusOK, Gauge: 5},
	}, nil, now)
	m := Model{reg: reg, groupBy: 2} // dim 2 == os
	if _, groups := m.orderedHosts(); len(groups) != 1 || groups[0] != "darwin" {
		t.Fatalf("os group = %v, want [darwin]", groups)
	}
}

func TestAlertCount(t *testing.T) {
	if n := alertCount(seed(t).Hosts()); n != 1 {
		t.Fatalf("alertCount = %d, want 1", n)
	}
}

func TestDimensionsIncludeHubOsAndTags(t *testing.T) {
	m := Model{reg: seed(t)}
	names := map[string]bool{}
	for _, d := range m.dimensions(m.reg.Hosts()) {
		names[d.name] = true
	}
	for _, want := range []string{"hub", "os", "env"} {
		if !names[want] {
			t.Errorf("dimensions missing %q (got %v)", want, names)
		}
	}
}
