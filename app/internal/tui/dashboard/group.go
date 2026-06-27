// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"sort"
	"strings"

	"heimdall/app/internal/domain"
)

// groupDim is one grouping axis (Yggdrasil): a name and a key extractor over a
// host view. Adding an axis means registering another dim, not editing a switch.
type groupDim struct {
	name string
	of   func(domain.HostView) string
}

// dimensions returns the grouping axes for the current fleet: origin hub, OS,
// then each distinct tag key. `groupBy` indexes this list with 1..N; 0 = none.
func (m Model) dimensions(hosts []domain.HostView) []groupDim {
	dims := []groupDim{
		{"hub", func(h domain.HostView) string { return labelOr(h, "hub", "(no hub)") }},
		{"os", osOf},
	}
	seen := map[string]bool{"hub": true}
	var keys []string
	for _, h := range hosts {
		for k := range h.Host.Context.Labels {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		k := k
		dims = append(dims, groupDim{k, func(h domain.HostView) string { return labelOr(h, k, "(no "+k+")") }})
	}
	return dims
}

// osOf derives a host's OS family for grouping. The dashboard receives OS only
// as the host.os inventory metric (e.g. "darwin 27.0") — Context.OS is empty
// over the wire — so read the metric first (family = first token) and fall back
// to Context.OS (set in --demo) before "(unknown)".
func osOf(h domain.HostView) string {
	for _, m := range h.LastSnapshot {
		if m.Name == "host.os" && m.Detail != "" {
			if f := strings.Fields(m.Detail); len(f) > 0 {
				return f[0]
			}
		}
	}
	if h.Host.Context.OS != "" {
		return h.Host.Context.OS
	}
	return "(unknown)"
}

func labelOr(h domain.HostView, key, fallback string) string {
	if v := h.Host.Context.Labels[key]; v != "" {
		return v
	}
	return fallback
}

// activeDim returns the current grouping dimension, or false when grouping is off.
func (m Model) activeDim(hosts []domain.HostView) (groupDim, bool) {
	if m.groupBy <= 0 {
		return groupDim{}, false
	}
	dims := m.dimensions(hosts)
	if m.groupBy-1 >= len(dims) {
		return groupDim{}, false
	}
	return dims[m.groupBy-1], true
}

// matchesFilter reports whether a host matches the filter query — its name/id or
// any tag (key=value), case-insensitive. An empty query matches everything.
func (m Model) matchesFilter(h domain.HostView) bool {
	q := strings.ToLower(strings.TrimSpace(m.filter))
	if q == "" {
		return true
	}
	if strings.Contains(strings.ToLower(h.Host.DisplayName+" "+string(h.Host.ID)), q) {
		return true
	}
	for k, v := range h.Host.Context.Labels {
		if strings.Contains(strings.ToLower(k+"="+v), q) {
			return true
		}
	}
	return false
}

// orderedHosts returns the hosts in display order: filtered, and when grouped,
// sorted so each group is contiguous. The parallel []string holds each host's
// group label (nil when not grouped) for section headers.
func (m Model) orderedHosts() ([]domain.HostView, []string) {
	var hosts []domain.HostView
	for _, h := range m.reg.Hosts() {
		if m.matchesFilter(h) {
			hosts = append(hosts, h)
		}
	}
	dim, grouped := m.activeDim(m.reg.Hosts())
	if !grouped {
		return hosts, nil
	}
	sort.SliceStable(hosts, func(i, j int) bool {
		if ki, kj := dim.of(hosts[i]), dim.of(hosts[j]); ki != kj {
			return ki < kj
		}
		return hosts[i].Host.ID < hosts[j].Host.ID
	})
	labels := make([]string, len(hosts))
	for i, h := range hosts {
		labels[i] = dim.of(h)
	}
	return hosts, labels
}

// orderedList is orderedHosts without the group labels.
func (m Model) orderedList() []domain.HostView {
	hosts, _ := m.orderedHosts()
	return hosts
}

// alertCount returns how many of the given hosts have at least one firing alert.
func alertCount(hosts []domain.HostView) int {
	n := 0
	for _, h := range hosts {
		if len(h.Alerts) > 0 {
			n++
		}
	}
	return n
}
