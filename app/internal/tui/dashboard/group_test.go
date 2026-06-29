// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"sort"
	"strings"
	"testing"
	"time"

	"heimdall/app/internal/domain"
)

// gseed builds a fleet matching the filter-grammar scenarios: hosts bar, baz, qux;
// env values foo and bar; bar/baz online on hub home, qux offline on hub remote.
func gseed(t *testing.T) *domain.HostRegistry {
	t.Helper()
	reg := domain.NewHostRegistry(10*time.Second, 30*time.Second)
	now := time.Unix(1_700_000_000, 0)
	add := func(id, hub, env string, seen time.Time) {
		reg.Enroll(domain.Host{ID: domain.HostID(id), DisplayName: id,
			Context: domain.HostContext{OS: "linux", Labels: map[string]string{"hub": hub, "env": env}}}, seen)
		reg.Observe(domain.HostID(id), []domain.Metric{{Name: "cpu.util", Status: domain.StatusOK, Gauge: 10}}, nil, seen)
	}
	add("bar", "home", "foo", now)
	add("baz", "home", "bar", now)
	add("qux", "remote", "bar", now.Add(-40*time.Second)) // last seen > offlineAfter -> offline
	reg.Evaluate(now)
	return reg
}

func ids(hosts []domain.HostView) []string {
	out := make([]string, len(hosts))
	for i, h := range hosts {
		out[i] = string(h.Host.ID)
	}
	sort.Strings(out)
	return out
}

func TestFilterTermGrammar(t *testing.T) {
	// dim 3 == env tag (hub=1, os=2, env=3) — used by the group-alias case.
	cases := []struct {
		name    string
		filter  string
		groupBy int
		want    []string
	}{
		{"bare term searches every field", "ba", 0, []string{"bar", "baz", "qux"}},
		{"host scope searches only the name", "host=ba", 0, []string{"bar", "baz"}},
		{"tag scope searches only that tag", "env=fo", 0, []string{"bar"}},
		{"hub scope", "hub=home", 0, []string{"bar", "baz"}},
		{"os scope", "os=linux", 0, []string{"bar", "baz", "qux"}},
		{"state scope", "state=offline", 0, []string{"qux"}},
		{"group alias to active dimension", "group=foo", 3, []string{"bar"}},
		{"multiple terms narrow conjunctively", "env=bar state=offline", 0, []string{"qux"}},
		{"empty filter shows everything", "", 0, []string{"bar", "baz", "qux"}},
		{"unknown field is a literal value", "zzz=qqq", 0, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := Model{reg: gseed(t), filter: tc.filter, groupBy: tc.groupBy}
			if got := ids(m.orderedList()); !equalStrings(got, tc.want) {
				t.Fatalf("filter %q = %v, want %v", tc.filter, got, tc.want)
			}
		})
	}
}

func TestFilterIdenticalGroupedAndUngrouped(t *testing.T) {
	ungrouped := Model{reg: gseed(t), filter: "ba", groupBy: 0}
	grouped := Model{reg: gseed(t), filter: "ba", groupBy: 3}
	if a, b := ids(ungrouped.orderedList()), ids(grouped.orderedList()); !equalStrings(a, b) {
		t.Fatalf("grouped %v != ungrouped %v for same filter", b, a)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestWindowClampsAndKeepsCursorVisible(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "L" + itoa(i)
	}
	contains := func(ls []string, s string) bool {
		for _, l := range ls {
			if strings.Contains(l, s) {
				return true
			}
		}
		return false
	}

	// fits: returned unchanged.
	if got := window([]string{"a", "b", "c"}, 1, 10); len(got) != 3 {
		t.Fatalf("fits: got %d lines, want 3", len(got))
	}

	// cursor at top: no top indicator, bottom indicator, height == max.
	top := window(lines, 0, 10)
	if len(top) != 10 {
		t.Fatalf("top: got %d lines, want 10", len(top))
	}
	if !contains(top, "L0") {
		t.Fatalf("top: cursor row L0 not visible: %v", top)
	}
	if !strings.Contains(top[len(top)-1], "more") {
		t.Fatalf("top: expected a bottom 'more' indicator, got %q", top[len(top)-1])
	}

	// cursor at bottom: top indicator, no bottom indicator.
	bot := window(lines, 19, 10)
	if len(bot) != 10 || !contains(bot, "L19") {
		t.Fatalf("bottom: cursor row L19 not visible in %v", bot)
	}
	if !strings.Contains(bot[0], "more") {
		t.Fatalf("bottom: expected a top 'more' indicator, got %q", bot[0])
	}

	// cursor in the middle: both indicators, cursor still visible.
	mid := window(lines, 10, 10)
	if len(mid) != 10 || !contains(mid, "L10") {
		t.Fatalf("mid: cursor row L10 not visible in %v", mid)
	}
	if !strings.Contains(mid[0], "more") || !strings.Contains(mid[len(mid)-1], "more") {
		t.Fatalf("mid: expected both 'more' indicators, got %q / %q", mid[0], mid[len(mid)-1])
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

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
